package floki

import (
	"encoding/hex"
	"log"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

// type Tenant struct {
// 	Group map[string][]string
// }

// type YAMLConfig struct {
// 	Tenants Tenant `yaml:"tenants"`
// }

type ConfigManager struct {
	tenantFile    string
	trackFilePath string
	config        *map[string]interface{}
}

func NewConfig() *ConfigManager {
	var conf map[string]interface{}
	return &ConfigManager{
		tenantFile:    "/opt/proxy/conf/tenant.yaml",
		trackFilePath: "/opt/proxy/track",
		config:        &conf,
	}
}

func (c ConfigManager) configFromFile(path string) (map[string]interface{}, error) {
	f, err := ReadFile(path)
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

// if groups, exists := conf["tenants"].(map[interface{}]interface{}); exists {
// 	for group := range groups {
// 		if group == "children" {
// 			fmt.Printf("Map: %v", groups[group])
// 		} else if group == "Share_Grant_Email_Group" {
// 			fmt.Println("GROUP")
// 		}
// 		switch groups[group].(type) {
// 		default:
// 			for _, val := range groups[group].([]interface{}) {
// 				fmt.Println(val)
// 			}
// 		case map[interface{}]interface{}:
// 			// continue
// 			fmt.Println(groups[group])
// 		}
// 	}
// }

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

	(*c).config = nil
	(*c).config = &conf
}

func (c *ConfigManager) configUpdated() bool {
	f, err := ReadFile(c.tenantFile)
	if err != nil {
		log.Println(err)
		return false
	}

	sig := genSignature(f)

	t, err := ReadFile(c.trackFilePath)
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

func ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func genSignature(data []byte) string {
	return hex.EncodeToString(data)
}

func storeSignature(path string, signature []byte) error {
	return os.WriteFile(path, signature, 0660)
}
