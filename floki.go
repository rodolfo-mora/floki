package floki

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

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
	Config     *ConfigManager
}

func NewFloki(url string, port string, done chan bool) Floki {
	log.Printf("Proxying requests for Loki %s", url)

	return Floki{
		LokiServer: url,
		Port:       port,
		Store:      NewMemoryStore(),
		Config:     NewTenantConfig(done),
		Exporter:   exporter.NewPrometheusExporter(":3100"),
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
	return f.Config.getTenants(groups...)
}

func Unauthorized(w http.ResponseWriter) {
	err := http.StatusText(http.StatusUnauthorized)
	code := http.StatusUnauthorized
	http.Error(w, err, code)
}
