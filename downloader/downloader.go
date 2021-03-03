package downloader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

var releasesURL = "https://releases.1c.ru"
var loginURL = "https://login.1c.ru"

const projectHrefPrefix = "/project/"
const versionFilesHrefPrefix = "/version_files"
const fileServerHrefPrefix = "/public/file/get"

const tempFileSuffix = ".d1c"

var semaMaxConnections = make(chan struct{}, 10)
var logOutput = io.Writer(os.Stdout)

type FileToDownload struct {
	url  string
	path string
	name string
}

type Downloader struct {
	Login         string
	Password      string
	BasePath      string
	StartDate     time.Time
	Nicks         map[string]bool
	VersionFilter string
	DistribFilter string
	httpClient    *http.Client
	urlCh         chan *FileToDownload
	wg            sync.WaitGroup
	logger        *log.Logger
}

func New(config *Downloader) *Downloader {

	cj, _ := cookiejar.New(nil)
	config.httpClient = &http.Client{
		Jar: cj,
	}

	config.logger = log.New(logOutput, "", log.LstdFlags)
	return config

}

func (dr *Downloader) Get() ([]os.FileInfo, error) {

	files := make([]os.FileInfo, 0)

	ticketUrl, err := dr.getURL()
	if err != nil {
		dr.handleError(err)
		return files, err
	}

	dr.urlCh = make(chan *FileToDownload, 10000)
	dr.wg.Add(1)
	go dr.findLinks(ticketUrl, dr.findProject)
	go func() {
		dr.wg.Wait()
		close(dr.urlCh)
	}()

	for fileToDownload := range dr.urlCh {
		if fileInfo, ok := dr.downloadFile(fileToDownload); ok {
			files = append(files, fileInfo)
		}
	}

	dr.urlCh = nil

	return files, nil

}

func (dr *Downloader) getURL() (string, error) {

	type loginParams struct {
		Login       string `json:"login"`
		Password    string `json:"password"`
		ServiceNick string `json:"serviceNick"`
	}

	type ticket struct {
		Ticket string `json:"ticket"`
	}

	postBody, err := json.Marshal(
		loginParams{dr.Login, dr.Password, releasesURL})
	if err != nil {
		return "", err
	}

	acquireSemaConnections()
	resp, err := dr.httpClient.Post(
		loginURL+"/rest/public/ticket/get",
		"application/json",
		bytes.NewReader(postBody))
	releaseSemaConnections()
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	responseBodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s", string(responseBodyData))
	}

	var tick ticket
	err = json.Unmarshal(responseBodyData, &tick)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(loginURL+"/ticket/auth?token=%s", tick.Ticket), nil

}

func (dr *Downloader) findLinks(rawUrl string, f func(string, string, *html.Node)) {

	defer dr.wg.Done()

	acquireSemaConnections()
	resp, err := dr.httpClient.Get(rawUrl)
	releaseSemaConnections()

	if err != nil {
		dr.handleError(err)
		return
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		dr.handleError(err)
		return
	}

	dr.eachNode(doc, rawUrl, f)

}

func (dr *Downloader) eachNode(node *html.Node, u string, f func(string, string, *html.Node)) {

	if node.Type == html.ElementNode && node.Data == "a" {
		for _, attr := range node.Attr {
			if attr.Key == "href" {
				f(u, attr.Val, node)
				break
			}
		}
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		dr.eachNode(c, u, f)
	}

}

func (dr *Downloader) findProject(_, href string, _ *html.Node) {
	projectName := strings.ToLower(strings.TrimLeft(href, projectHrefPrefix))
	if (dr.Nicks == nil && strings.HasPrefix(href, projectHrefPrefix)) || dr.Nicks[projectName] {
		dr.wg.Add(1)
		go dr.findLinks(releasesURL+href, dr.findVersion)
	}

}

func (dr *Downloader) findVersion(_, href string, node *html.Node) {

	if strings.HasPrefix(href, versionFilesHrefPrefix) {

		vDateRaw := strings.Trim(node.Parent.NextSibling.NextSibling.FirstChild.Data, " \n")
		vDate, err := time.Parse("02.01.06", vDateRaw)
		if err != nil {
			dr.handleError(err)
			return
		}

		if dr.VersionFilter != "" {
			matched, err := regexp.MatchString(dr.VersionFilter, href)
			if err != nil {
				dr.handleError(err)
				return
			}
			if !matched {
				return
			}
		}


		if vDate.After(dr.StartDate) {
			dr.wg.Add(1)
			go dr.findLinks(releasesURL+href, dr.findToDownloadLink)
		}

	}

}

func (dr *Downloader) findToDownloadLink(_, href string, _ *html.Node) {

	lowerHref := strings.ToLower(href)
	isRO := strings.Contains(lowerHref, "path=ro\\")

	if dr.DistribFilter == "" {
		if strings.HasSuffix(lowerHref, "rar") ||
			(strings.HasSuffix(lowerHref, "zip") && !isRO) ||
			strings.HasSuffix(lowerHref, "gz") ||
			strings.HasSuffix(lowerHref, "exe") ||
			strings.HasSuffix(lowerHref, "msi") ||
			strings.HasSuffix(lowerHref, "deb") ||
			strings.HasSuffix(lowerHref, "rpm") ||
			strings.HasSuffix(lowerHref, "epf") ||
			strings.HasSuffix(lowerHref, "erf") {

			dr.wg.Add(1)
			dr.findLinks(releasesURL+href, dr.findFileServerLink)

		} else if strings.HasSuffix(lowerHref, "txt") ||
			strings.HasSuffix(lowerHref, "pdf") ||
			strings.HasSuffix(lowerHref, "html") ||
			strings.HasSuffix(lowerHref, "htm") ||
			(strings.HasSuffix(lowerHref, "zip") && isRO) {
				dr.addFileToChannel(href, releasesURL+href)
			}
	}else {
		matched, err := regexp.MatchString(dr.DistribFilter, href)
		if err != nil {
			dr.handleError(err)
			return
		}
		if matched {
			dr.wg.Add(1)
			dr.findLinks(releasesURL+href, dr.findFileServerLink)
		}
	}



}

func (dr *Downloader) findFileServerLink(u, href string, _ *html.Node) {

	if strings.Contains(href, fileServerHrefPrefix) {
		dr.addFileToChannel(u, href)
	}

}

func (dr *Downloader) addFileToChannel(u, href string) {
	fileName, filePath, err := dr.fileNameFromUrl(u)
	if err == nil {
		fileToDownload := FileToDownload{
			url:  href,
			path: filePath,
			name: fileName,
		}
		dr.urlCh <- &fileToDownload
	} else {
		dr.handleError(err)
	}
}

func (dr *Downloader) fileNameFromUrl(rawUrl string) (string, string, error) {

	fileName := strings.Builder{}
	filePath := strings.Builder{}

	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return "", "", err
	}

	query, err := url.ParseQuery(parsedUrl.RawQuery)
	if err != nil {
		return "", "", err
	}

	path := strings.Split(query.Get("path"), "\\")
	fileName.WriteString(path[len(path)-1])

	nick := query.Get("nick")
	ver := query.Get("ver")

	filePath.WriteString(nick)
	filePath.WriteRune(os.PathSeparator)
	filePath.WriteString(ver)
	filePath.WriteRune(os.PathSeparator)

	return fileName.String(), filePath.String(), nil
}

func (dr *Downloader) downloadFile(fileToDownload *FileToDownload) (os.FileInfo, bool) {

	fullpath := dr.BasePath + fileToDownload.path + fileToDownload.name
	fileInfo, err := os.Stat(fullpath)
	if os.IsExist(err) {

		return fileInfo, true

	} else if os.IsNotExist(err) {

		dr.handleOutput(fmt.Sprintf("Getting a file from url: %s\n", fileToDownload.url))
		acquireSemaConnections()
		resp, err := dr.httpClient.Get(fileToDownload.url)
		if err != nil {
			return nil, false
		}

		err = os.MkdirAll(dr.BasePath+fileToDownload.path, os.ModeDir)
		if err != nil {
			return nil, false
		}

		f, err := os.Create(fullpath + tempFileSuffix)
		if err != nil {
			return nil, false
		}

		defer resp.Body.Close()

		_, err = io.Copy(f, resp.Body)
		releaseSemaConnections()
		if err != nil {
			return nil, false
		}
		f.Close()

		dr.handleOutput(fmt.Sprintf("End of receiving file by url: %s\n", fileToDownload.url))
		dr.handleOutput(fmt.Sprintf("File saved to: %s\n", fullpath))

		err = os.Rename(fullpath+tempFileSuffix, fullpath)
		if err != nil {
			return nil, false
		}

		fileInfo, err := os.Stat(fullpath)
		if err != nil {
			return nil, false
		}

		return fileInfo, true

	} else if err != nil {

	}

	return nil, false

}

func (dr *Downloader) handleError(err error) {
	_ = fmt.Errorf("%s", err)
	dr.logger.Println(err)
}

func (dr *Downloader) handleOutput(text string) {
	fmt.Print(text)
	dr.logger.Print(text)
}

func acquireSemaConnections() {
	semaMaxConnections <- struct{}{}
}

func releaseSemaConnections() {
	_ = <-semaMaxConnections
}

func (dr *Downloader) LogOutput() io.Writer {
	return logOutput
}

func (dr *Downloader) SetLogOutput(out io.Writer) {
	logOutput = out
}