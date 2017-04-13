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
	"github.com/kyokomi/emoji"
	"gopkg.in/yaml.v2"
)

func commandRunEnv(conf libtf.HclConf, vault libtf.Vault) {
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
	syscall.Exec(bin, flag.Args()[1:], append(env, os.Environ()...))
}

func commandRun(conf libtf.HclConf, vault libtf.Vault) {
	data, err := json.MarshalIndent(vault.Env, "", " ")
	if err != nil {
		panic(err)
	}
	bin, err := exec.LookPath(flag.Arg(1))
	if err != nil {
		panic(err)
	}
	env := fmt.Sprintf("TR_JSON=%s", string(data))
	syscall.Exec(bin, flag.Args()[1:], append([]string{env}, os.Environ()...))
}

func commandDump(conf libtf.HclConf, vault libtf.Vault) {
	data, err := json.MarshalIndent(vault.Env, "", "  ")
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, bytes.NewBuffer(data))
}

func commandCompose(conf libtf.HclConf, vault libtf.Vault) {
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

func commandTerraform(conf libtf.HclConf, vault libtf.Vault, target string) {
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

	env := []string{}
	for key, value := range vault.Raw {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	for service := range services {
		env = append(env, fmt.Sprintf("%s=.ecs-def/%s.json", libtf.EnvKey(libtf.EcsTemplateVar(service)), service))
	}
	for _, target := range conf.Targets {
		env = append(env, fmt.Sprintf("%s=%s", libtf.EnvKey(libtf.StateKeyVar(target)), libtf.StateKey(vault.EnvName(), target)))
	}
	//env = append(env, "TF_INPUT=0")
	//env[idx] = "TF_INPUT=0"

	env = append(env, []string{
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", vault.AwsKey()),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", vault.AwsSecret()),
		fmt.Sprintf("AWS_DEFAULT_REGION=%s", vault.AwsRegion()),
	}...)

	syscall.Exec(terraformBin, append([]string{"terraform"}, flag.Args()[1:]...), append(env, os.Environ()...))
}

func commandVariables(conf libtf.HclConf, vault libtf.Vault) {
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

func commandEncrypt(conf libtf.HclConf, vault libtf.Vault) {
	output := flag.Arg(1)
	keyString := conf.Keys[conf.Global.ProjectName]
	if len(keyString) == 0 {
		panic("no key found in ~/.tfrc")
	}
	data, err := vault.Encode(keyString)
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(output, data, 0600); err != nil {
		panic(err)
	}
	emoji.Printf(":ok_hand: %s\n", output)
}

func commandDecrypt(conf libtf.HclConf, vault libtf.Vault) {
	output := flag.Arg(1)
	data, err := yaml.Marshal(vault.Env)
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(output, data, 0600); err != nil {
		panic(err)
	}
	emoji.Printf(":ok_hand: %s\n", output)
}

func commandRunEcsTask(conf libtf.HclConf, vault libtf.Vault, allInstances bool) {
	if err := libtf.RunEcsTask(vault, flag.Arg(1), allInstances); err != nil {
		panic(err)
	}
}

func main() {

	configFile := flag.String("config", ".tf.hcl", "")
	vaultFile := flag.String("vault", "env", "")
	allInstances := flag.Bool("all_instances", false, "")

	flag.Parse()

	conf := libtf.HclConf{}

	if err := libtf.LoadHclConf(*configFile, &conf); err != nil {
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
	} else if strings.HasSuffix(*vaultFile, ".yml") {
		err = conf.LoadYamlFile(*vaultFile, &vault)
	} else if strings.HasSuffix(*vaultFile, ".vault") {
		err = conf.LoadVault(*vaultFile, &vault)
	} else {
		panic("invalid vault filename")
	}

	vault.AddDefaults()

	if err != nil {
		if os.IsNotExist(err) {
			panic(err)
		}
		color.Red("%s", err)
		os.Exit(1)
	}

	switch flag.Arg(0) {
	case "run":
		commandRun(conf, vault)
	case "run-env":
		commandRunEnv(conf, vault)
	case "ecs-task":
		commandRunEcsTask(conf, vault, *allInstances)
	case "dump":
		commandDump(conf, vault)
	case "compose":
		commandCompose(conf, vault)
	case "variables":
		commandVariables(conf, vault)
	case "encrypt":
		commandEncrypt(conf, vault)
	case "decrypt":
		commandDecrypt(conf, vault)
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
			commands := strings.Join(append([]string{"run", "run-env", "dump", "ecs-task", "compose", "variables", "encrypt", "decrypt"}, conf.Targets...), "|")
			fmt.Printf("usage: tf -config=.tf.hcl -vault=env|name.yml|name.vault %s\n", commands)
			os.Exit(1)
		}
	}
}
