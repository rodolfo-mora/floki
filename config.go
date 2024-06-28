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

type ConfigManager struct {
	exporterEnabled bool
	tenantFile      string
	trackFilePath   string
	tenantConfig    JSONConfig
}

func NewTenantConfig(done chan bool) *ConfigManager {
	cm := ConfigManager{
		exporterEnabled: true,
		tenantFile:      "/Users/rodgon/code/personal/go/floki/config/tenants.yaml",
		trackFilePath:   "/Users/rodgon/code/personal/go/floki/app/track",
	}

	tenants, err := cm.configFromFile(cm.tenantFile)
	if err != nil {
		log.Fatal(err)
	}

	cm.tenantConfig = tenants
	go cm.ConfigWatcher(done)
	return &cm
}

/*
Description
*/
func (c ConfigManager) configFromFile(path string) (JSONConfig, error) {
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

func (c *ConfigManager) ConfigWatcher(done chan bool) {
	ticker := time.NewTicker(1000 * time.Millisecond)
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if c.configUpdated() {
				log.Println("Tenant configuration changes detected. Updating local config.")
				c.updateConfig()
				c.printConfig()
			}
		}
	}
}

func (c *ConfigManager) printConfig() {
	log.Println((*c).tenantConfig)
}

func (c *ConfigManager) updateConfig() {
	var mux sync.Mutex

	mux.Lock()
	defer mux.Unlock()

	conf, err := c.configFromFile(c.tenantFile)
	if err != nil {
		log.Println(err)
		return
	}

	(*c).tenantConfig = JSONConfig{}
	(*c).tenantConfig = conf
}

func (c *ConfigManager) configUpdated() bool {
	f, err := readFile(c.tenantFile)
	if err != nil {
		log.Println(err)
		return false
	}

	sig := genSignature(f)
	/*
	  If trackfile doesn't exist this might be our first time
	  executing. We store the signature and return false.
	*/
	if !c.trackFileExists() {
		storeSignature(c.trackFilePath, []byte(sig))
		return false
	}

	t, err := readFile(c.trackFilePath)
	if err != nil {
		log.Println(err)
		return false
	}

	if sig != string(t) {
		storeSignature(c.trackFilePath, []byte(sig))
		return true
	}

	return false
}

func (c *ConfigManager) trackFileExists() bool {
	if _, err := os.Stat(c.trackFilePath); errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}

func (c *ConfigManager) getTenants(groups ...string) (string, error) {
	var tenants []string

	for _, group := range groups {
		tenants = append(tenants, strings.Join(c.tenantConfig.Tenants[group], "|"))
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
