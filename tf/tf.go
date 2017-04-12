package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strings"
	"syscall"

	"github.com/barbuza/tf/libtf"
	"github.com/fatih/color"
	"gopkg.in/yaml.v2"
)

func commandRunEnv(conf libtf.YamlConf, vault libtf.Vault) {
	bin, err := exec.LookPath(flag.Arg(1))
	if err != nil {
		panic(err)
	}
	env := make([]string, len(vault.Raw))
	idx := 0
	for key, value := range vault.Raw {
		env[idx] = fmt.Sprintf("%s=%s", key, value)
		idx++
	}
	syscall.Exec(bin, flag.Args()[1:], append(os.Environ(), env...))
}

func commandRun(conf libtf.YamlConf, vault libtf.Vault) {
	data, err := json.MarshalIndent(vault.Env, "", " ")
	if err != nil {
		panic(err)
	}
	bin, err := exec.LookPath(flag.Arg(1))
	if err != nil {
		panic(err)
	}
	env := fmt.Sprintf("TR_JSON=%s", string(data))
	syscall.Exec(bin, flag.Args()[1:], append(os.Environ(), env))
}

func commandDump(conf libtf.YamlConf, vault libtf.Vault) {
	data, err := json.MarshalIndent(vault.Env, "", "  ")
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, bytes.NewBuffer(data))
}

func commandCompose(conf libtf.YamlConf, vault libtf.Vault) {
	data, err := yaml.Marshal(conf.AsCompose(vault))
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(".compose.yml", data, 0600); err != nil {
		panic(err)
	}
	bin, err := exec.LookPath("docker-compose")
	if err != nil {
		panic(err)
	}
	syscall.Exec(bin, append([]string{"docker-compose", "-f", ".compose.yml"}, flag.Args()[1:]...), os.Environ())
}

func commandTerraform(conf libtf.YamlConf, vault libtf.Vault, target string) {
	if err := os.Chdir(target); err != nil {
		panic(err)
	}

	services := map[string][]libtf.EcsServiceConfig{}
	conf.AsEcs(vault, services)

	if err := libtf.RimRaf(".ecs-def"); err != nil {
		panic(err)
	}

	if err := os.Mkdir(".ecs-def", 0700); err != nil {
		panic(err)
	}

	for key, value := range services {
		data, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			panic(err)
		}

		err = ioutil.WriteFile(fmt.Sprintf(".ecs-def/%s.json", key), data, 0600)
		if err != nil {
			panic(err)
		}
	}

	vault.InitRemoteState(target)

	terraformBin, err := exec.LookPath("terraform")
	if err != nil {
		panic(err)
	}

	env := make([]string, len(vault.Raw)+len(services)+len(conf.Targets)+1)
	idx := 0
	for key, value := range vault.Raw {
		env[idx] = fmt.Sprintf("%s=%s", key, value)
		idx++
	}
	for service := range services {
		env[idx] = fmt.Sprintf("%s=.ecs-def/%s.json", libtf.EnvKey(libtf.EcsTemplateVar(service)), service)
		idx++
	}
	for _, target := range conf.Targets {
		env[idx] = fmt.Sprintf("%s=%s", libtf.EnvKey(libtf.StateKeyVar(target)), libtf.StateKey(vault.EnvName(), target))
		idx++
	}
	env[idx] = "TF_INPUT=0"

	syscall.Exec(terraformBin, append([]string{"terraform"}, flag.Args()[1:]...), append(os.Environ(), env...))
}

func commandVariables(conf libtf.YamlConf, vault libtf.Vault) {
	keys := []string{}

	services := map[string][]libtf.EcsServiceConfig{}
	conf.AsEcs(vault, services)

	for service := range services {
		keys = append(keys, libtf.EcsTemplateVar(service))
	}

	for _, target := range conf.Targets {
		keys = append(keys, libtf.StateKeyVar(target))
	}

	for key, variable := range conf.Env {
		if !variable.Optional {
			keys = append(keys, key)
		}
	}

	sort.Sort(libtf.ByString(keys))

	for _, key := range keys {
		fmt.Printf("variable \"%s\" {}\n", key)
	}
}

func main() {
	configFile := flag.String("config", ".tf.yml", "")
	vaultFile := flag.String("vault", "env", "")
	flag.Parse()

	conf := libtf.YamlConf{}
	if err := libtf.LoadYamlConf(*configFile, &conf); err != nil {
		panic(err)
	}
	if err := conf.Validate(); err != nil {
		panic(err)
	}

	libtf.GetGitVersion()

	vault := libtf.Vault{}
	var err error
	if *vaultFile == "env" {
		err = conf.LoadEnv(&vault)
	} else {
		err = conf.LoadVault(*vaultFile, &vault)
	}

	if err != nil {
		color.Red("%s", err)
		os.Exit(1)
	}

	switch flag.Arg(0) {
	case "run":
		commandRun(conf, vault)
	case "run-env":
		commandRunEnv(conf, vault)
	case "dump":
		commandDump(conf, vault)
	case "compose":
		commandCompose(conf, vault)
	case "variables":
		commandVariables(conf, vault)
	default:

		found := false
		for _, target := range conf.Targets {
			if target == flag.Arg(0) {
				found = true
				commandTerraform(conf, vault, target)
				break
			}
		}
		if !found {
			commands := strings.Join(append([]string{"run", "run-env", "dump", "compose"}, conf.Targets...), "|")
			fmt.Printf("usage: tf -config=.tf.yml -vault=env|vault.yml %s\n", commands)
			os.Exit(1)
		}
	}
}
