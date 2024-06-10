package Floki

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Floki struct {
	LokiServer string
	Port       string
}

func NewFloki(url string, port string) Floki {
	log.Printf("Proxying requests for Loki %s", url)
	return Floki{
		LokiServer: url,
		Port:       port,
	}
}

func (f Floki) registerRoutes() {
	http.HandleFunc("/", f.Handler)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", f.Port), nil); err != nil {
		log.Fatal(err)
	}
}

func (f Floki) Handler(w http.ResponseWriter, r *http.Request) {
	lokiUrl, _ := url.Parse(f.LokiServer)

	reverseProxy := httputil.NewSingleHostReverseProxy(lokiUrl)
	UpdateHeaders(r, lokiUrl)
	reverseProxy.ServeHTTP(w, r)
}

func UpdateHeaders(r *http.Request, u *url.URL) {
	(*r).URL.Scheme = u.Scheme
	(*r).URL.Host = u.Host
	(*r).Host = u.Host
	(*r).Header.Set("X-Forwarded-Host", u.Host)
	(*r).Header.Set("X-Scope-OrgID", "fake")
}
