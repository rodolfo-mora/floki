package floki

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type User struct {
	Email     string   `json:"email"`
	SSOGroups []string `json:"sso_groups"`
}

type Floki struct {
	LokiServer string
	Port       string
	APIUrl     string
	Store      *MemoryStore
	Config     *ConfigManager
}

func NewFloki(url string, port string, apiurl string) Floki {
	log.Printf("Proxying requests for Loki %s", url)

	return Floki{
		LokiServer: url,
		Port:       port,
		APIUrl:     apiurl,
		Store:      NewMemoryStore(),
		Config:     NewTenantConfig(),
	}
}

func (f Floki) RegisterUser(user string, groups []string) {
	f.Store.Save(user, groups)
}

func (f Floki) Start() {
	f.registerRoutes()
}

func (f Floki) UpdarteConfig() {

}

func (f Floki) registerRoutes() {
	http.HandleFunc("/", f.Handler)
	addr := fmt.Sprintf(":%s", f.Port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func (f Floki) Handler(w http.ResponseWriter, r *http.Request) {
	lokiUrl, _ := url.Parse(f.LokiServer)

	reverseProxy := httputil.NewSingleHostReverseProxy(lokiUrl)
	if r.Header.Get("X-Grafana-User") == "" {
		Unauthorized(w)
		return
	}
	err := f.UpdateHeaders(r, lokiUrl)
	if err != nil {
		Unauthorized(w)
		return
	}
	reverseProxy.ServeHTTP(w, r)
}

func (f Floki) UpdateHeaders(r *http.Request, u *url.URL) error {
	(*r).URL.Scheme = u.Scheme
	(*r).URL.Host = u.Host
	(*r).Host = u.Host
	(*r).Header.Set("X-Forwarded-Host", u.Host)

	tenants, err := f.GetTenants(r.Header.Get("X-Grafana-User"))
	if err != nil {
		return err
	}
	(*r).Header.Set("X-Scope-OrgID", tenants)
	return nil
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
