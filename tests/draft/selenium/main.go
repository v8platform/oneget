package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

const (
	// Set constants to set the address of ChromEDriver.exe and local call ports, respectively.
	seleniumPath = `chromedriver.exe`
	port         = 9515
)

var (
	chromeCaps = chrome.Capabilities{
		// Prefs: imgCaps,
		Path: "",
		Args: []string{
			// "--headless",
			"--start-maximized",
			"--window-size=1920x1080",
			"--no-sandbox",
			"--user-agent=Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36",
			"--disable-gpu",
			"--disable-impl-side-painting",
			"--disable-gpu-sandbox",
			"--disable-accelerated-2d-canvas",
			"--disable-accelerated-jpeg-decoding",
			"--test-type=ui",
			"--ignore-certificate-errors",
		},
	}
	//Set the option to set the Selenium service, set to empty. Set it as needed.
	ops     = []selenium.ServiceOption{}
	service *selenium.Service
	// Set the browser compatibility, set the browser name to Chrome
	caps     = selenium.Capabilities{"browserName": "chrome"}
	articles = make([]string, 0, 120)
)

// init in initialization a service background service
func Init() (*selenium.Service, error) {
	// 1. Open Selenium service
	_, filename, _, _ := runtime.Caller(1)
	return selenium.NewChromeDriverService(path.Join(path.Dir(filename), seleniumPath), port, ops...)
}

// [+] Traversed index subscript, one page grabbed article URL ------------------------------- -------------------------------------------------- ------------
func Run(urlgen string) (err error) {
	// 1. Load custom browser configuration
	caps.AddChrome(chromeCaps)
	// 2. Mount the browser to Selenium Driver, call the browser URLPREFIX: Test reference: defaultURLPREFIX = "http://127.0.0.1:4444/wd/hub"
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://127.0.0.1:%v/wd/hub", port))
	if err != nil {
		err = fmt.Errorf("unable create browser, err: %v", err)
		return
	}
	// 3. Open the root page
	// login https://login.1c.ru/login
	if err = wd.Get("https://login.1c.ru/login"); err != nil {
		log.Println(fmt.Sprintf("Failed to load page: %s\n", err))
	}

	// 5. Click username
	u, err := wd.FindElement(selenium.ByName, "username")
	if err != nil {
		return err
	}
	u.Clear()
	u.SendKeys("Имя")
	
	// Click password
	p, err := wd.FindElement(selenium.ByName, "password")
	if err != nil {
		return err
	}
	p.Clear()
	p.SendKeys("Секретный пароль")

	// login
	b, err := wd.FindElement(selenium.ByID, "loginButton")
	if err != nil {
		return err
	}
	b.Click()

	// go to reliases https://releases.1c.ru/total
	if err = wd.Get("https://releases.1c.ru/total"); err != nil {
		log.Println(fmt.Sprintf("Failed to load page: %s\n", err))
	}

	// Переходим на платформы: <a href="/project/Platform83">Технологическая платформа 8.3</a>
	link, err := wd.FindElement(selenium.ByLinkText, "Технологическая платформа 8.3")
	if err != nil {
		return err
	}
	link.Click()

	// Переходим на платформу 8.3.20.1838
	link, err = wd.FindElement(selenium.ByLinkText, "8.3.20.1838")
	if err != nil {
		return err
	}
	link.Click()

	// Переходим на страничку загрузки Технологическая платформа 1С:Предприятия (64-bit) для Linux
	link, err = wd.FindElement(selenium.ByLinkText, "Технологическая платформа 1С:Предприятия (64-bit) для Linux")
	if err != nil {
		return err
	}
	link.Click()

	// Нажимаем на кнопку скачть дистрибутив
	link, err = wd.FindElement(selenium.ByLinkText, "Скачать дистрибутив")
	if err != nil {
		return err
	}
	a, err := link.GetAttribute("href")
	if err != nil {
		return err
	}
	log.Println(a)

	// Пытаемся скачать файл
	if err = wd.Get(a); err != nil {
		log.Println(fmt.Sprintf("Failed to load page: %s\n", err))
	}

	// wait downloads "C:\Users\Zhdan\Downloads\server64_8_3_20_1838.tar.gz.crdownload"

	for true {
		if _, err := os.Stat("C:\\Users\\Zhdan\\Downloads\\server64_8_3_20_1838.tar.gz.crdownload"); err == nil {
			time.Sleep(5 * time.Second)
		} else if _, err := os.Stat("C:\\Users\\Zhdan\\Downloads\\server64_8_3_20_1838.tar.gz"); err == nil {
			break
		}
	}
	return wd.Close()
}


func main() {
	service, err := Init()
	if err != nil {
		log.Fatal(err)
	}
	defer service.Stop()
	// Open a web page
	if err = Run("https://login.1c.ru/login"); err != nil {
		log.Fatal(err)
	}
}
