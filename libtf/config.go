package libtf

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os/user"
	"path"
)

type TfConfig struct {
	Keys map[string]string `yaml:"keys"`
}

func LoadTfConfig(config *TfConfig) error {
	currentUser, err := user.Current()
	if err != nil {
		return nil
	}
	data, err := ioutil.ReadFile(path.Join(currentUser.HomeDir, ".tfrc"))
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(data, config); err != nil {
		return err
	}
	for key, value := range config.Keys {
		if len(value) != 32 {
			return fmt.Errorf("key for %s must be 32 chars", key)
		}
	}
	return nil
}
