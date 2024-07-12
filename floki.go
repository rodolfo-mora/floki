package floki

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/rodolfo-mora/floki/pkg/config"
	"github.com/rodolfo-mora/floki/pkg/exporter"
)

type User struct {
	Email     string   `json:"email"`
	SSOGroups []string `json:"sso_groups"`
}

type Floki struct {
	LokiServer string
	Port       string
	APIUrl     string
	Exporter   exporter.Exporter
	Store      *MemoryStore
	Tenants    *TenantManager
}

func NewFloki(url string, port string, done chan bool) Floki {
	c := config.NewConfig()
	log.Printf("Proxying requests for Loki %s", c.LokiURL)

	return Floki{
		LokiServer: c.LokiURL,
		Port:       c.ProxyPort,
		Store:      NewMemoryStore(),
		Tenants:    NewTenantManager(done, c.Tenant),
		Exporter:   exporter.NewPrometheusExporter(c.ExporterPort),
	}
}

func (f Floki) RegisterUser(user string, groups []string) {
	f.Store.Save(user, groups)
}

func (f Floki) Start() {
	f.registerRoutes()
}

func (f Floki) registerRoutes() {
	log.Println("Starting server")

	http.Handle("/", f.Exporter.Wrapper("/", f.Handler))
	http.Handle("/metrics", f.Exporter.Export())
	addr := fmt.Sprintf(":%s", f.Port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func (f Floki) Handler(w http.ResponseWriter, r *http.Request) {
	lokiUrl, _ := url.Parse(f.LokiServer)
	if r.Header.Get("X-Grafana-User") == "" {
		Unauthorized(w)
		return
	}

	f.UpdateHeaders(r, lokiUrl)

	reverseProxy := httputil.NewSingleHostReverseProxy(lokiUrl)
	reverseProxy.ServeHTTP(w, r)
}

func (f Floki) UpdateHeaders(r *http.Request, u *url.URL) {
	(*r).URL.Scheme = u.Scheme
	(*r).URL.Host = u.Host
	(*r).Host = u.Host
	(*r).Header.Set("X-Forwarded-Host", u.Host)

	tenants, err := f.GetTenants(r.Header.Get("X-Grafana-User"))
	if err != nil {
		log.Println(err)
	}

	(*r).Header.Set("X-Scope-OrgID", tenants)
}

func (f Floki) GetTenants(user string) (string, error) {
	groups := f.Store.GetSSOGroups(user)
	return f.Tenants.getTenants(groups...)
}

func Unauthorized(w http.ResponseWriter) {
	setErrorCode(w, http.StatusUnauthorized)
}

func NotFound(w http.ResponseWriter) {
	setErrorCode(w, http.StatusNotFound)
}

func setErrorCode(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}
