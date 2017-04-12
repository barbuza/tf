package libtf

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"fmt"
)

type TfConfig struct {
	Keys map[string]string `yaml:"keys"`
}

func LoadTfConfig(config *TfConfig) error {
	data, err := ioutil.ReadFile(os.ExpandEnv("/Users/barbuza/.tfrc"))
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
