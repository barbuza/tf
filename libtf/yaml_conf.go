package libtf

import (
	"errors"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

type yamlConfVariable struct {
	Type     string `yaml:"type"`
	Optional bool   `yaml:"optional"`
}

type yamlConfService struct {
	Name    string            `yaml:"name"`
	Image   string            `yaml:"image"`
	Ecs     string            `yaml:"ecs"`
	Command string            `yaml:"command"`
	Compose bool              `yaml:"compose"`
	Memory  int               `yaml:"memory"`
	NoEnv   bool              `yaml:"noenv"`
	NoLog   bool              `yaml:"nolog"`
	Env     map[string]string `yaml:"env"`
	Links   []string          `yaml:"links"`
	Ports   []int             `yaml:"ports"`
}

type yamlConfGlobal struct {
	BaseImage   string `yaml:"base_image"`
	ProjectName string `yaml:"project_name"`
}

type YamlConf struct {
	Keys          map[string]string
	Global        yamlConfGlobal              `yaml:"global"`
	Services      []yamlConfService           `yaml:"services"`
	Env           map[string]yamlConfVariable `yaml:"env"`
	Targets       []string
	SortedEnvKeys []string
}

var yamlConfDefaultEnv = []string{
	"env_name",
	"aws_key",
	"aws_secret",
	"aws_region",
	"tf_state_bucket",
}

type ByString []string

func (a ByString) Len() int {
	return len(a)
}

func (a ByString) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByString) Less(i, j int) bool {
	return strings.Compare(a[i], a[j]) == -1
}

func LoadYamlConf(filename string, conf *YamlConf) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(data, conf); err != nil {
		return err
	}
	for _, name := range yamlConfDefaultEnv {
		conf.Env[name] = yamlConfVariable{
			Type: "string",
		}
	}
	sortedEnvKeys := make([]string, len(conf.Env))
	idx := 0
	for key := range conf.Env {
		sortedEnvKeys[idx] = key
		idx++
	}
	sort.Sort(ByString(sortedEnvKeys))
	conf.SortedEnvKeys = sortedEnvKeys
	conf.Targets = findTerraformTargets()
	return nil
}

func (conf *YamlConf) Validate() error {

	config := TfConfig{}
	if err := LoadTfConfig(&config); err != nil {
		return err
	}
	conf.Keys = config.Keys

	if len(conf.Global.BaseImage) == 0 {
		return errors.New("global.base_image is not defined")
	}
	if len(conf.Global.ProjectName) == 0 {
		return errors.New("global.project_name is not defined")
	}
	if _, found := conf.Keys[conf.Global.ProjectName]; !found {
		return fmt.Errorf("encryption key is not defined for project %s", conf.Global.ProjectName)
	}

	for index, service := range conf.Services {
		if len(service.Name) == 0 {
			return fmt.Errorf("services[%d].name is not defined", index)
		}
		if len(service.Ecs) == 0 && !service.Compose {
			return fmt.Errorf("both compose and ecs disabled for service.%s", service.Name)
		}
		if len(service.Ecs) != 0 && service.Memory == 0 {
			return fmt.Errorf("services.%s.memory is not defined", service.Name)
		}
	}

	for name, variable := range conf.Env {
		switch variable.Type {
		case "string", "bool", "dict", "list", "int":
		default:
			return fmt.Errorf("env.%s.type is invalid", name)
		}
	}

	return nil
}
