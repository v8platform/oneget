package server

import (
	"fmt"
	"github.com/khorevaa/logos"
	"net/http"
)

var log = logos.New("github.com/v8platform/oneget/server").Sugar()

func Run(workspace string, port string) {

	host := fmt.Sprintf("0.0.0.0:%s", port)
	log.Infof("http server is enabled for port: %s", port)
	fs := http.FileServer(http.Dir(workspace))
	log.Errorf("Error file server", http.ListenAndServe(host, fs))
}
