package libtf

import (
	"errors"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/hashicorp/hcl"
)

type hclConfVariable struct {
	Type     string `hcl:"type"`
	Optional bool   `hcl:"optional"`
}

type hclConfService struct {
	Name    string            `hcl:"name"`
	Image   string            `hcl:"image"`
	Ecs     string            `hcl:"ecs"`
	Command string            `hcl:"command"`
	Compose bool              `hcl:"compose"`
	Memory  int               `hcl:"memory"`
	NoEnv   bool              `hcl:"noenv"`
	NoLog   bool              `hcl:"nolog"`
	Env     map[string]string `hcl:"env"`
	Links   []string          `hcl:"links"`
	Ports   []int             `hcl:"ports"`
}

type hclConfGlobal struct {
	BaseImage   string `hcl:"base_image"`
	ProjectName string `hcl:"project_name"`
}

type HclConf struct {
	Keys          map[string]string
	Global        hclConfGlobal              `hcl:"global"`
	Services      map[string]hclConfService  `hcl:"service"`
	Env           map[string]hclConfVariable `hcl:"env"`
	Targets       []string
	SortedEnvKeys []string
	EcsServices   map[string]bool
}

var hclConfDefaultEnv = []string{
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

func LoadHclConf(filename string, conf *HclConf) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	if err := hcl.Unmarshal(data, conf); err != nil {
		return err
	}
	conf.EcsServices = map[string]bool{}
	for name, service := range conf.Services {
		if len(service.Ecs) != 0 {
			conf.EcsServices[name] = true
		}
		conf.Services[name] = hclConfService{
			Name:    name,
			Image:   service.Image,
			Ecs:     service.Ecs,
			Command: service.Command,
			Compose: service.Compose,
			Memory:  service.Memory,
			NoEnv:   service.NoEnv,
			NoLog:   service.NoLog,
			Env:     service.Env,
			Links:   service.Links,
			Ports:   service.Ports,
		}
	}
	for name, variable := range conf.Env {
		if len(variable.Type) == 0 {
			conf.Env[name] = hclConfVariable{
				Type:     "string",
				Optional: variable.Optional,
			}
		}
	}
	for _, name := range hclConfDefaultEnv {
		conf.Env[name] = hclConfVariable{
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

func (conf *HclConf) Validate() error {

	config := TfConfig{}
	if err := LoadTfConfig(&config); err == nil {
		conf.Keys = config.Keys
	} else {
		conf.Keys = map[string]string{}
	}

	if len(conf.Global.BaseImage) == 0 {
		return errors.New("global.base_image is not defined")
	}
	if len(conf.Global.ProjectName) == 0 {
		return errors.New("global.project_name is not defined")
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
