package libtr

import (
	"fmt"
)

type ComposeConfig struct {
	Version  string                          `yaml:"version"`
	Services map[string]composeServiceConfig `yaml:"services"`
}

type composeServiceConfig struct {
	Image       string            `yaml:"image"`
	Links       []string          `yaml:"links,omitempty"`
	Command     string            `yaml:"command,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
	Ports       []int             `yaml:"ports,omitempty"`
	Logging     map[string]string `yaml:"logging,omitempty"`
}

func mergeEnv(a map[string]string, b map[string]string) map[string]string {
	res := make(map[string]string)
	for name, value := range a {
		res[name] = value
	}
	for name, value := range b {
		res[name] = value
	}
	return res
}

func (conf *YamlConf) AsCompose(vault Vault) ComposeConfig {
	config := ComposeConfig{
		Version:  "2",
		Services: make(map[string]composeServiceConfig),
	}
	for _, service := range conf.Services {
		if !service.Compose {
			continue
		}
		config.Services[service.Name] = service.asCompose(conf.Global.BaseImage, vault)
	}
	return config
}

func (service *yamlConfService) asCompose(baseImage string, vault Vault) composeServiceConfig {
	serviceEnv := service.Env
	if !service.NoEnv {
		serviceEnv = mergeEnv(vault.Raw, serviceEnv)
	}
	var logging map[string]string
	if service.NoLog {
		logging = map[string]string{
			"driver": "none",
		}
	}
	image := service.Image
	if len(image) == 0 {
		image = fmt.Sprintf("%s:%s", baseImage, getGitVersion())
	}
	return composeServiceConfig{
		Image:       image,
		Links:       service.Links,
		Ports:       service.Ports,
		Environment: serviceEnv,
		Command:     service.Command,
		Logging:     logging,
	}
}
