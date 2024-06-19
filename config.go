package floki

import (
	"encoding/hex"
	"errors"
	"log"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

type ConfigManager struct {
	tenantFile    string
	trackFilePath string
	tenants       *map[string]interface{}
}

func NewConfig() *ConfigManager {
	c := ConfigManager{
		tenantFile:    "/opt/proxy/conf/tenant.yaml",
		trackFilePath: "/opt/proxy/track",
	}

	tenants, err := c.configFromFile(c.tenantFile)
	if err != nil {
		log.Fatal(err)
	}

	c.tenants = &tenants
	return &c
}

func (c *ConfigManager) Start() {
	var done chan bool
	go c.ConfigWatcher(done)
	<-done
}

func (c ConfigManager) configFromFile(path string) (map[string]interface{}, error) {
	f, err := readFile(path)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	conf := make(map[string]interface{})
	err = yaml.Unmarshal(f, &conf)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return conf, nil
}

func (c *ConfigManager) ConfigWatcher(done chan bool) {
	ticker := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if c.configUpdated() {
				log.Println("Tenant configuration changes detected. Updating local config.")
				c.updateConfig()
			}
		}
	}
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

	(*c).tenants = nil
	(*c).tenants = &conf
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
		storeSignature(c.trackFilePath, f)
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

func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func genSignature(data []byte) string {
	return hex.EncodeToString(data)
}

func storeSignature(path string, signature []byte) error {
	return os.WriteFile(path, signature, 0660)
}
