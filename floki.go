package floki

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
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
	Config     ConfigManager
}

func NewFloki(url string, port string, apiurl string) Floki {
	log.Printf("Proxying requests for Loki %s", url)

	return Floki{
		LokiServer: url,
		Port:       port,
		APIUrl:     apiurl,
		Store:      NewMemoryStore(),
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
	if err := http.ListenAndServe(fmt.Sprintf(":%s", f.Port), nil); err != nil {
		log.Fatal(err)
	}
}

func (f Floki) Handler(w http.ResponseWriter, r *http.Request) {
	lokiUrl, _ := url.Parse(f.LokiServer)

	reverseProxy := httputil.NewSingleHostReverseProxy(lokiUrl)
	if r.Header.Get("X-Grafana-User") == "" {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	err := f.UpdateHeaders(r, lokiUrl)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
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

func (f Floki) queryTenantAPI(group string) (string, error) {
	group = strings.Replace(group, " ", "%20", -1)

	res, err := http.Get(f.APIUrl + "?groups=" + group)
	if err != nil {
		return "", err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	if len(body) == 2 {
		return "", nil
	}

	return string(body), nil

}

func (f Floki) GetTenants(user string) (string, error) {
	var tenants []string
	groups := f.Store.GetSSOGroups(user)
	for _, group := range groups {
		tenant, err := f.queryTenantAPI(group)
		if err != nil {
			return "", err
		}

		tenant = strings.Replace(tenant, "\"", "", -1)
		if tenant == "" {
			continue
		}

		tenants = append(tenants, tenant)
	}
	if len(tenants) > 0 {
		return strings.Join(tenants, "|"), nil
	}
	return "", nil
}
