package downloader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.uber.org/multierr"
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

type Filter interface {
	MatchString(source string) bool
}

type GetConfig struct {
	BasePath string
	Project  string
	Version  string
	Filters  []Filter
}

type OnegetDownloader struct {
	Login, Password string

	cookie *cookiejar.Jar

	urlCh chan *FileToDownload
	wg    *sync.WaitGroup

	parser *HtmlParser
}

func NewDownloader(login, password string) *OnegetDownloader {

	cj, _ := cookiejar.New(nil)
	parser, _ := NewHtmlParser()

	return &OnegetDownloader{
		cookie:   cj,
		Login:    login,
		Password: password,
		wg:       &sync.WaitGroup{},
		parser:   parser,
	}

}

func (dr *OnegetDownloader) Get(config ...GetConfig) ([]os.FileInfo, error) {

	if err := dr.authRequest(); err != nil {
		return nil, err
	}

	files := make([]os.FileInfo, 0)

	downloadCh := make(chan *FileToDownload, 100)

	for _, getConfig := range config {
		err := dr.getFiles(getConfig, downloadCh)
		if err != nil {
			return nil, err
		}
	}

	go func() {
		dr.wg.Wait()
		close(downloadCh)
	}()

	limit := make(chan struct{}, 10)
	mu := &sync.Mutex{}
	for fileToDownload := range downloadCh {
		go func(file *FileToDownload) {

			limit <- struct{}{}

			fileInfo, err := dr.downloadFile(fileToDownload)
			if err != nil {
				dr.wg.Done()
				log.Errorf(err.Error())
			}
			if fileInfo != nil {
				mu.Lock()
				files = append(files, fileInfo)
				mu.Unlock()
			}
			dr.wg.Done()

			<-limit

		}(fileToDownload)

	}

	downloadCh = nil

	return files, nil

}

func (dr *OnegetDownloader) authRequest() error {

	ticketUrl, err := dr.getURL(releasesURL)
	if err != nil {
		log.Errorf("Error get ticket: ", err.Error())
		return err
	}

	client := dr.getClient()
	_, err = client.Get(ticketUrl)
	if err != nil {
		log.Errorf("Error auth request: ", err.Error())
		return err
	}

	return nil
}

func (dr *OnegetDownloader) getFiles(config GetConfig, downloadCh chan *FileToDownload) error {

	releases, err := dr.getProjectReleases(config)
	if err != nil {
		return err
	}

	for _, release := range releases {
		dr.wg.Add(1)
		go func(info *ProjectVersionInfo, cfg GetConfig) {
			_ = dr.getReleaseFiles(info, config, downloadCh)
			dr.wg.Done()
		}(release, config)
	}

	return nil
}

func (dr *OnegetDownloader) getReleaseFiles(release *ProjectVersionInfo, config GetConfig, downloadCh chan *FileToDownload) error {

	client := dr.getClient()
	resp, err := client.Get(releasesURL + release.Url)
	defer resp.Body.Close()

	if err != nil {
		return err
	}

	releaseFiles, err := dr.parser.ParseProjectRelease(resp.Body)
	if err != nil {
		return err
	}
	files := filterReleaseFiles(releaseFiles, config.Filters)

	var merr error

	for _, file := range files {

		err := dr.addFileToChannel(file.url, config, downloadCh)
		if err != nil {
			log.Errorf("Error get file from <%s>: %s", err.Error())
			multierr.Append(merr, err)
		}

	}

	return merr

}

func (dr *OnegetDownloader) getProjectReleases(config GetConfig) ([]*ProjectVersionInfo, error) {

	client := dr.getClient()

	url := releasesURL + projectHrefPrefix + config.Project
	resp, err := client.Get(url)
	defer resp.Body.Close()

	if err != nil {
		return nil, err
	}
	responseBodyData, _ := ioutil.ReadAll(resp.Body)

	releases, err := dr.parser.ParseProjectReleases(bytes.NewBuffer(responseBodyData))
	if err != nil {
		return nil, fmt.Errorf("error parse project <%s> releases: %s, html: <%s>", config.Project, err.Error(), string(responseBodyData))
	}

	return filterProjectVersionInfo(releases, config.Version), nil

}
func filterReleaseFiles(list []ReleaseFileInfo, filters []Filter) (filteredList []ReleaseFileInfo) {

	if len(filters) == 0 || len(list) == 0 {
		return list
	}

	matchInfo := func(i ReleaseFileInfo) bool {

		for _, filter := range filters {

			matchName := filter.MatchString(i.name)
			matchUrl := filter.MatchString(i.url)

			if matchName || matchUrl {
				return true
			}

		}

		return false
	}

	for _, info := range list {

		if matchInfo(info) {
			filteredList = append(filteredList, info)
		}

	}

	return

}

func filterProjectVersionInfo(list []*ProjectVersionInfo, filter string) (filteredList []*ProjectVersionInfo) {

	if len(filter) == 0 || len(list) == 0 {
		return list
	}

	if strings.EqualFold(strings.ToLower(filter), "latest") {
		return append(filteredList, list[0])
	}

	re, _ := regexp.Compile(filter)

	for _, info := range list {
		if re.MatchString(info.Name) {
			filteredList = append(filteredList, info)
		}
	}

	return

}

func (dr *OnegetDownloader) getClient() *http.Client {
	return &http.Client{
		Jar:       dr.cookie,
		Transport: nil,
	}
}

func (dr *OnegetDownloader) getURL(url string) (string, error) {

	type loginParams struct {
		Login       string `json:"login"`
		Password    string `json:"password"`
		ServiceNick string `json:"serviceNick"`
	}

	type ticket struct {
		Ticket string `json:"ticket"`
	}

	postBody, err := json.Marshal(
		loginParams{dr.Login, dr.Password, url})
	if err != nil {
		return "", err
	}

	client := dr.getClient()

	resp, err := client.Post(
		loginURL+"/rest/public/ticket/get",
		"application/json",
		bytes.NewReader(postBody))
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

func (dr *OnegetDownloader) addFileToChannel(href string, config GetConfig, downloadCh chan *FileToDownload) (err error) {

	downloadHref := []string{releasesURL + href}

	if !directLink(href) {
		downloadHref, err = dr.getDownloadFileLinks(href, config)
		if err != nil {
			return err
		}
	}

	fileName, filePath, err := dr.fileNameFromUrl(href)
	if err != nil {
		return err
	}

	if len(downloadHref) == 0 {
		return nil
	}

	dr.wg.Add(1)

	log.Debugf("Add to download: %s", href)
	downloadCh <- &FileToDownload{
		url:      downloadHref,
		path:     filePath,
		name:     fileName,
		basePath: config.BasePath,
	}

	return nil
}

func directLink(href string) bool {
	return strings.HasSuffix(href, "txt") ||
		strings.HasSuffix(href, "pdf") ||
		strings.HasSuffix(href, "html") ||
		strings.HasSuffix(href, "htm") ||
		(strings.HasSuffix(href, "zip") &&
			strings.Contains(href, "path=ro\\"))
}

func (dr *OnegetDownloader) fileNameFromUrl(rawUrl string) (string, string, error) {

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

func (dr *OnegetDownloader) downloadFile(fileToDownload *FileToDownload) (os.FileInfo, error) {

	workDir := filepath.Join(fileToDownload.basePath, strings.ToLower(fileToDownload.path))
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

		downloadUrl := fileToDownload.url[0]

		log.Debugf("Workspace directory: %s", workDir)
		log.Debugf("Getting a file from url: %s", downloadUrl)
		client := dr.getClient()
		resp, err := client.Get(downloadUrl)
		defer resp.Body.Close()

		if err != nil {
			return nil, err
		}

		f, err := os.Create(fileName + tempFileSuffix)
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(f, resp.Body)
		if err != nil {
			return nil, err
		}
		f.Close()

		log.Debugf("End of receiving file by url: %s", downloadUrl)
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

func (dr *OnegetDownloader) handleError(err error) {
	if err == nil {
		return
	}
	log.Error(err.Error())
}

func (dr *OnegetDownloader) getDownloadFileLinks(href string, config GetConfig) ([]string, error) {

	client := dr.getClient()
	resp, err := client.Get(releasesURL + href)
	if err != nil {
		return []string{}, err
	}
	defer resp.Body.Close()

	fileLinks, err := dr.parser.ParseReleaseFiles(resp.Body)
	if err != nil {
		return nil, err
	}

	return fileLinks, nil

}
