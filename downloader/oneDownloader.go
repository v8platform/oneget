package downloader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type OneConfig struct {
	Login    string
	Password string
	BasePath string
	Project  string
	Version  string
	Filter   string
}

type OneDownloader struct {
	OneConfig
	httpClient *http.Client
	urlCh      chan *FileToDownload
	wg         sync.WaitGroup
}

func NewDownloader(config OneConfig) *OneDownloader {

	cj, _ := cookiejar.New(nil)

	return &OneDownloader{
		OneConfig: config,
		httpClient: &http.Client{
			Jar: cj,
		},
		wg: sync.WaitGroup{},
	}

}

func (dr *OneDownloader) Get() ([]os.FileInfo, error) {

	files := make([]os.FileInfo, 0)

	ticketUrl, err := dr.getURL()
	if err != nil {
		dr.handleError(err)
		return files, err
	}

	dr.urlCh = make(chan *FileToDownload, 10000)
	dr.wg.Add(1)
	dr.findLinks(ticketUrl, dr.findProject)
	go func() {
		dr.wg.Wait()
		close(dr.urlCh)
	}()

	for fileToDownload := range dr.urlCh {
		fileInfo, err := dr.downloadFile(fileToDownload)
		if err != nil {
			return nil, err
		}
		if fileInfo != nil {
			files = append(files, fileInfo)
		}
	}

	dr.urlCh = nil

	return files, nil

}

func (dr *OneDownloader) getURL() (string, error) {

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

func (dr *OneDownloader) findLinks(rawUrl string, f func(string, string, *html.Node)) {

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

func (dr *OneDownloader) eachNode(node *html.Node, u string, f func(string, string, *html.Node)) {

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

func (dr *OneDownloader) findProject(_, href string, _ *html.Node) {
	projectName := strings.ToLower(strings.TrimLeft(href, projectHrefPrefix))
	log.Debugf("Finding project in href %s", projectName)
	if strings.EqualFold(projectName, dr.Project) {
		dr.findLinks(releasesURL+href, dr.findVersion)
	}

}

func (dr *OneDownloader) findVersion(_, href string, node *html.Node) {

	if !strings.HasPrefix(href, versionFilesHrefPrefix) {
		return
	}

	log.Debugf("Filtering href %s by %s", href, dr.Version)
	matched, err := regexp.MatchString(dr.Version, href)
	if err != nil {
		dr.handleError(err)
		return
	}
	if !matched {
		return
	}

	dr.findLinks(releasesURL+href, dr.findToDownloadLink)

}

func (dr *OneDownloader) findToDownloadLink(_, href string, _ *html.Node) {

	lowerHref := strings.ToLower(href)
	isRO := strings.Contains(lowerHref, "path=ro\\")

	if dr.Filter == "" {
		if strings.HasSuffix(lowerHref, "rar") ||
			(strings.HasSuffix(lowerHref, "zip") && !isRO) ||
			strings.HasSuffix(lowerHref, "gz") ||
			strings.HasSuffix(lowerHref, "exe") ||
			strings.HasSuffix(lowerHref, "msi") ||
			strings.HasSuffix(lowerHref, "deb") ||
			strings.HasSuffix(lowerHref, "rpm") ||
			strings.HasSuffix(lowerHref, "epf") ||
			strings.HasSuffix(lowerHref, "erf") {

			dr.findLinks(releasesURL+href, dr.findFileServerLink)

		} else if strings.HasSuffix(lowerHref, "txt") ||
			strings.HasSuffix(lowerHref, "pdf") ||
			strings.HasSuffix(lowerHref, "html") ||
			strings.HasSuffix(lowerHref, "htm") ||
			(strings.HasSuffix(lowerHref, "zip") && isRO) {
			dr.addFileToChannel(href, releasesURL+href)
		}
	} else {
		matched, err := regexp.MatchString(dr.Filter, href)
		if err != nil {
			dr.handleError(err)
			return
		}
		if matched {
			dr.findLinks(releasesURL+href, dr.findFileServerLink)
		}
	}

}

func (dr *OneDownloader) findFileServerLink(u, href string, _ *html.Node) {

	if strings.Contains(href, fileServerHrefPrefix) {
		dr.addFileToChannel(u, href)
	}

}

func (dr *OneDownloader) addFileToChannel(u, href string) {
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

func (dr *OneDownloader) fileNameFromUrl(rawUrl string) (string, string, error) {

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

func (dr *OneDownloader) downloadFile(fileToDownload *FileToDownload) (os.FileInfo, error) {

	workDir := filepath.Join(dr.BasePath, strings.ToLower(fileToDownload.path))
	fileName := filepath.Join(workDir, fileToDownload.name)
	fileInfo, err := os.Stat(fileName)
	if os.IsExist(err) {

		return fileInfo, nil

	} else if os.IsNotExist(err) {

		if _, err := os.Stat(workDir); os.IsNotExist(err) {
			err = os.MkdirAll(filepath.Join(workDir), 0777)
			if err != nil {
				return nil, err
			}
			// https://wenzr.wordpress.com/2018/03/27/go-file-permissions-on-unix/
			os.Chmod(workDir, 0777)
		}
		log.Debugf("Workspace directory: %s", workDir)
		log.Debugf("Getting a file from url: %s", fileToDownload.url)

		acquireSemaConnections()
		resp, err := dr.httpClient.Get(fileToDownload.url)
		if err != nil {
			return nil, err
		}

		f, err := os.Create(fileName + tempFileSuffix)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		_, err = io.Copy(f, resp.Body)
		releaseSemaConnections()
		if err != nil {
			return nil, err
		}
		f.Close()

		log.Debugf("End of receiving file by url: %s", fileToDownload.url)
		log.Debugf("File saved to: %s", fileName)

		err = os.Rename(fileName+tempFileSuffix, fileName)
		if err != nil {
			return nil, err
		}

		fileInfo, err := os.Stat(fileName)
		if err != nil {
			return nil, err
		}

		return fileInfo, nil

	} else if err != nil {

	}

	return nil, nil

}

func (dr *OneDownloader) handleError(err error) {
	if err == nil {
		return
	}
	log.Error(err.Error())
}
