package downloader

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/khorevaa/logos"
)

var releasesURL = "https://releases.1c.ru"
var loginURL = "https://login.1c.ru"

var log = logos.New("github.com/v8platform/oneget/downloader").Sugar()

const (
	projectHrefPrefix = "/project/"
	projectHrefSuffix = "?allUpdates=true"
	tempFileSuffix    = ".d1c"
)

var executionRe = regexp.MustCompile(`name="execution"\s+value="([^"]+)"`)

func NewClient(loginUrl string, baseUrl string, login string, password string) (*Client, error) {

	cj, _ := cookiejar.New(nil)

	c := &Client{
		login:    login,
		password: password,
		loginUrl: loginUrl,
		baseUrl:  baseUrl,
		cookie:   cj,
	}

	err := c.casLogin()
	if err != nil {
		return nil, err
	}

	return c, nil
}

type Client struct {
	cookie   *cookiejar.Jar
	login    string
	password string
	loginUrl string
	baseUrl  string
}

// casLogin выполняет аутентификацию через CAS form-based login.
// 1. GET login page — получить execution token и session cookies
// 2. POST login form — аутентификация и установка CAS cookies
func (c *Client) casLogin() error {
	serviceURL := c.baseUrl + "/public/security_check"
	casLoginURL := c.loginUrl + "/login?service=" + url.QueryEscape(serviceURL)

	// Шаг 1: GET страницу логина для получения execution token
	log.Debugf("CAS login: getting login page from %s", casLoginURL)
	req, err := http.NewRequest("GET", casLoginURL, nil)
	if err != nil {
		return fmt.Errorf("CAS login: create request error: %s", err.Error())
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return fmt.Errorf("CAS login: get login page error: %s", err.Error())
	}
	defer resp.Body.Close()

	body, err := readBody(resp.Body)
	if err != nil {
		return fmt.Errorf("CAS login: read login page error: %s", err.Error())
	}

	matches := executionRe.FindSubmatch(body)
	if matches == nil {
		return fmt.Errorf("CAS login: execution token not found on login page")
	}
	executionToken := string(matches[1])
	log.Debugf("CAS login: got execution token (length=%d)", len(executionToken))

	// Шаг 2: POST form с credentials
	formData := url.Values{
		"username":   {c.login},
		"password":   {c.password},
		"execution":  {executionToken},
		"_eventId":   {"submit"},
		"geolocation": {""},
	}

	req, err = http.NewRequest("POST", casLoginURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("CAS login: create POST request error: %s", err.Error())
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp2, err := c.doRequest(req)
	if err != nil {
		return fmt.Errorf("CAS login: POST error: %s", err.Error())
	}
	defer resp2.Body.Close()

	// Проверяем что аутентификация прошла — финальный URL не должен быть страницей логина
	finalURL := resp2.Request.URL.String()
	log.Debugf("CAS login: final URL after auth: %s", finalURL)

	if strings.Contains(finalURL, "login.1c.ru/login") {
		// Читаем тело чтобы проверить ошибку
		body2, _ := readBody(resp2.Body)
		if strings.Contains(string(body2), "Неверный логин или пароль") ||
			strings.Contains(string(body2), "Incorrect login or password") {
			return fmt.Errorf("CAS login: неверный логин или пароль")
		}
		return fmt.Errorf("CAS login: аутентификация не удалась, финальный URL: %s", finalURL)
	}

	log.Debugf("CAS login: authenticated successfully")
	return nil
}

func (c *Client) Get(getUrl string) (*http.Response, error) {

	// для URL вида /total/
	if strings.HasPrefix(getUrl, "/") {
		getUrl = c.baseUrl + getUrl
	}

	req, err := http.NewRequest("GET", getUrl, nil)

	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)

	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		log.Debugf("Re-auth required, performing CAS login again for: %s", getUrl)
		if err := c.casLogin(); err != nil {
			return nil, fmt.Errorf("re-auth failed: %s", err.Error())
		}

		req, err := http.NewRequest("GET", getUrl, nil)
		if err != nil {
			return nil, err
		}

		return c.doRequest(req)

	case http.StatusBadRequest, http.StatusNotFound:

		return nil, fmt.Errorf("response CODE:%d  ERR:%s",
			resp.StatusCode, readBodyMustString(resp.Body))
	case http.StatusOK:
		return resp, nil
	default:
		return resp, fmt.Errorf("unknown response CODE: <%d>", resp.StatusCode)
	}

}

func (c *Client) client() *http.Client {

	return &http.Client{
		Jar: c.cookie,
	}
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {

	return c.client().Do(req)

}

func readBody(body io.ReadCloser) ([]byte, error) {
	buf := get()
	defer put(buf)
	_, err := io.Copy(buf, body)
	if err != nil {
		return nil, err
	}

	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result, nil
}

func readBodyMustString(body io.ReadCloser) string {
	buf, err := readBody(body)
	if err != nil {
		log.Errorf("must read body err: %s", err.Error())
		return ""
	}

	return string(buf)
}

// checkResponseBody проверяет HTML-ответ на наличие страницы логина или ошибки сервера.
func checkResponseBody(body string) error {
	if strings.Contains(body, "<title>Личные данные</title>") ||
		strings.Contains(body, `id="loginForm"`) {
		return fmt.Errorf("сессия истекла, получена страница логина вместо данных")
	}
	if strings.Contains(body, "Ошибка на нашем сервере") ||
		strings.Contains(body, "сервис временно недоступен") {
		return fmt.Errorf("сервер releases.1c.ru временно недоступен")
	}
	if strings.Contains(body, "Ошибка доступа") {
		return fmt.Errorf("ошибка доступа на releases.1c.ru — нет прав на данный ресурс")
	}
	return nil
}

var pool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

func get() *bytes.Buffer {
	return pool.Get().(*bytes.Buffer)
}

func put(buf *bytes.Buffer) {
	buf.Reset()
	pool.Put(buf)
}
