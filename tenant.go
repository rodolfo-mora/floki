package floki

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	jsonyaml "github.com/ghodss/yaml"
)

type JSONConfig struct {
	Tenants map[string][]string `json:"tenants"`
}

type TenantManager struct {
	exporterEnabled bool
	tenantFile      string
	trackFilePath   string
	tenantConfig    JSONConfig
}

func NewTenantManager(done chan bool, tenantfile string, trackfile string) *TenantManager {
	tm := TenantManager{
		exporterEnabled: true,
		tenantFile:      tenantfile,
		trackFilePath:   trackfile,
	}

	tenants, err := tm.configFromFile(tm.tenantFile)
	if err != nil {
		log.Fatal(err)
	}

	tm.tenantConfig = tenants
	go tm.ConfigWatcher(done)
	return &tm
}

/*
Description
*/
func (t TenantManager) configFromFile(path string) (JSONConfig, error) {
	var conf JSONConfig

	f, err := readFile(path)
	if err != nil {
		return conf, err
	}

	yconf, err := jsonyaml.YAMLToJSON(f)
	if err != nil {
		return conf, err
	}

	err = json.Unmarshal(yconf, &conf)
	if err != nil {
		return conf, err
	}

	return conf, nil
}

func (t *TenantManager) ConfigWatcher(done chan bool) {
	ticker := time.NewTicker(1000 * time.Millisecond)
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if t.configUpdated() {
				log.Println("Tenant configuration changes detected. Updating local config.")
				t.updateConfig()
			}
		}
	}
}

func (t *TenantManager) updateConfig() {
	var mux sync.Mutex

	mux.Lock()
	defer mux.Unlock()

	conf, err := t.configFromFile(t.tenantFile)
	if err != nil {
		log.Println(err)
		return
	}

	(*t).tenantConfig = JSONConfig{}
	(*t).tenantConfig = conf
}

func (t *TenantManager) configUpdated() bool {
	f, err := readFile(t.tenantFile)
	if err != nil {
		log.Println(err)
		return false
	}

	sig := genSignature(f)
	/*
	  If trackfile doesn't exist this might be our first time
	  executing. We store the signature and return false.
	*/
	if !t.trackFileExists() {
		storeSignature(t.trackFilePath, []byte(sig))
		return false
	}

	tf, err := readFile(t.trackFilePath)
	if err != nil {
		log.Println(err)
		return false
	}

	if sig != string(tf) {
		storeSignature(t.trackFilePath, []byte(sig))
		return true
	}

	return false
}

func (t *TenantManager) trackFileExists() bool {
	if _, err := os.Stat(t.trackFilePath); errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}

func (t *TenantManager) getTenants(groups ...string) (string, error) {
	var tenants []string

	for _, group := range groups {
		tenants = append(tenants, strings.Join(t.tenantConfig.Tenants[group], "|"))
	}

	return strings.Join(tenants, "|"), nil
}

func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func genSignature(data []byte) string {
	sig := md5.Sum(data)
	return string(sig[:])
}

func storeSignature(path string, signature []byte) error {
	return os.WriteFile(path, signature, 0660)
}
