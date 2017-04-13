package libtf

import (
	"fmt"
	"sort"
	"strings"
)

type ecsPortMapping struct {
	ContainerPort int `json:"containerPort"`
}

type ecsEnvVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type byEcsEnvName []ecsEnvVariable

func (a byEcsEnvName) Len() int {
	return len(a)
}

func (a byEcsEnvName) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a byEcsEnvName) Less(i, j int) bool {
	return strings.Compare(a[i].Name, a[j].Name) == -1
}

type byEcsPort []ecsPortMapping

func (a byEcsPort) Len() int {
	return len(a)
}

func (a byEcsPort) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a byEcsPort) Less(i, j int) bool {
	return a[i].ContainerPort < a[j].ContainerPort
}

type ecsLogOptions struct {
	Group  string `json:"awslogs-group"`
	Region string `json:"awslogs-region"`
}

type ecsLogConfiguration struct {
	LogDriver string        `json:"logDriver"`
	Options   ecsLogOptions `json:"options"`
}

type EcsServiceConfig struct {
	Name              string               `json:"name"`
	Image             string               `json:"image"`
	Command           []string             `json:"command,omitempty"`
	Links             []string             `json:"links,omitempty"`
	Environment       []ecsEnvVariable     `json:"environment,omitempty"`
	Essential         bool                 `json:"essential"`
	MemoryReservation int                  `json:"memoryReservation"`
	PortMappings      []ecsPortMapping     `json:"portMappings,omitempty"`
	LogConfiguration  *ecsLogConfiguration `json:"logConfiguration,omitempty"`
	DNSSearchDomains  []string             `json:"dnsSearchDomains"`
}

func (service *hclConfService) asEcs(conf *HclConf, vault Vault) EcsServiceConfig {
	image := service.Image
	if len(image) == 0 {
		image = fmt.Sprintf("%s:%s", conf.Global.BaseImage, GetGitVersion())
	}
	portMappings := []ecsPortMapping{}
	for _, port := range service.Ports {
		portMappings = append(portMappings, ecsPortMapping{
			ContainerPort: port,
		})
		sort.Sort(byEcsPort(portMappings))
	}
	var logConfigration *ecsLogConfiguration
	if !service.NoLog {
		logConfigration = new(ecsLogConfiguration)
		*logConfigration = ecsLogConfiguration{
			LogDriver: "awslogs",
			Options: ecsLogOptions{
				Group:  fmt.Sprintf("%s-%s", vault.EnvName(), service.Name),
				Region: vault.AwsRegion(),
			},
		}
	}
	env := []ecsEnvVariable{}
	if !service.NoEnv {
		for key, value := range vault.Raw {
			env = append(env, ecsEnvVariable{
				Name:  key,
				Value: value,
			})
		}
		for key, value := range service.Env {
			env = append(env, ecsEnvVariable{
				Name:  key,
				Value: value,
			})
		}
		sort.Sort(byEcsEnvName(env))
	}
	links := []string{}
	for _, name := range service.Links {
		_, found := conf.EcsServices[name]
		if found {
			links = append(links, name)
		}
	}
	return EcsServiceConfig{
		Essential:         true,
		Name:              service.Name,
		Image:             image,
		Command:           strings.Split(service.Command, " "),
		MemoryReservation: service.Memory,
		PortMappings:      portMappings,
		LogConfiguration:  logConfigration,
		Environment:       env,
		Links:             links,
		DNSSearchDomains:  []string{"internal"},
	}
}

func (conf *HclConf) AsEcs(vault Vault, services map[string][]EcsServiceConfig) {
	for _, service := range conf.Services {
		if len(service.Ecs) == 0 {
			continue
		}
		_, ok := services[service.Ecs]
		if !ok {
			services[service.Ecs] = []EcsServiceConfig{}
		}
		services[service.Ecs] = append(services[service.Ecs], service.asEcs(conf, vault))
	}
}
