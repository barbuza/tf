package libtf

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/davecgh/go-spew/spew"
	"gopkg.in/yaml.v2"
)

type Vault struct {
	Env map[string]interface{}
	Raw map[string]string
}

func (vault *Vault) awsRegion() string {
	return vault.Env["aws_region"].(string)
}

func (vault *Vault) EnvName() string {
	return vault.Env["env_name"].(string)
}

func (vault *Vault) awsKey() string {
	return vault.Env["aws_key"].(string)
}

func (vault *Vault) awsSecret() string {
	return vault.Env["aws_secret"].(string)
}

func (vault *Vault) stateBucket() string {
	return vault.Env["tf_state_bucket"].(string)
}

func fixInterfacesInMap(input map[string]interface{}) {
	for key, value := range input {
		switch value.(type) {
		case map[interface{}]interface{}:
			input[key] = fixInterfaceMap(value.(map[interface{}]interface{}))
		case []interface{}:
			input[key] = fixInterfaceSlice(value.([]interface{}))
		default:
		}
	}
}

func fixInterfaceSlice(input []interface{}) []string {
	res := make([]string, len(input))
	for idx, item := range input {
		switch item.(type) {
		case string:
			res[idx] = item.(string)
		default:
			panic("list values must be strings")
		}
	}
	return res
}

func fixInterfaceMap(input map[interface{}]interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	for key, value := range input {
		switch key.(type) {
		case string:
			switch value.(type) {
			case map[interface{}]interface{}:
				out[key.(string)] = fixInterfaceMap(value.(map[interface{}]interface{}))
			case []interface{}:
				out[key.(string)] = fixInterfaceSlice(value.([]interface{}))
			default:
				out[key.(string)] = value
			}
		default:
			panic("can't parse non-string keys")
		}
	}
	return out
}

func EnvKey(key string) string {
	return fmt.Sprintf("TF_VAR_%s", key)
}

func StateKeyVar(target string) string {
	return fmt.Sprintf("%s_state_key", target)
}

func StateKey(envName string, target string) string {
	return fmt.Sprintf("%s-%s.tfstate", envName, target)
}

func EcsTemplateVar(service string) string {
	return fmt.Sprintf("ecs_%s_template", service)
}

func (conf *YamlConf) LoadEnv(vault *Vault) error {
	res := map[string]interface{}{}
	for _, key := range conf.SortedEnvKeys {
		variable := conf.Env[key]
		value, found := os.LookupEnv(EnvKey(key))
		if !found && variable.Optional {
			continue
		}
		if !found {
			return fmt.Errorf("%s is not defined in env", EnvKey(key))
		}
		switch variable.Type {
		case "string":
			res[key] = value
		case "bool":
			boolValue, err := envStringToBool(value)
			if err != nil {
				return err
			}
			res[key] = boolValue
		case "list":
			res[key] = envStringToList(value)
		case "dict":
			dictValue, err := envStringToDict(value)
			if err != nil {
				return err
			}
			res[key] = dictValue
		default:
			return fmt.Errorf("unknown type %s", variable.Type)
		}
	}
	vault.Env = res
	var err error
	vault.Raw, err = structToEnv(vault.Env)
	return err
}

func (conf *YamlConf) LoadVault(filename string, vault *Vault) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	decoded := map[string]interface{}{}
	if err := yaml.Unmarshal(data, &decoded); err != nil {
		return err
	}

	fixInterfacesInMap(decoded)
	res := map[string]interface{}{}

	for _, key := range conf.SortedEnvKeys {
		variable := conf.Env[key]
		value, found := decoded[key]
		if !found && variable.Optional {
			continue
		}
		if !found {
			return fmt.Errorf("%s is not defined in vault", key)
		}
		switch variable.Type {
		case "string":
			switch value.(type) {
			case string:
				res[key] = value
			default:
				return fmt.Errorf("%s is not of type string", spew.Sdump(value))
			}
		case "bool":
			switch value.(type) {
			case bool:
				res[key] = value
			default:
				return fmt.Errorf("%s is not of type bool", spew.Sdump(value))
			}
		case "list":
			switch value.(type) {
			case []string:
				res[key] = value
			default:
				return fmt.Errorf("%s is not of type list", spew.Sdump(value))
			}
		case "dict":
			switch value.(type) {
			case map[string]interface{}:
				res[key] = value
			default:
				return fmt.Errorf("%s is not of type dict", spew.Sdump(value))
			}
		default:
			return fmt.Errorf("unknown type %s", variable.Type)
		}
	}

	vault.Env = res
	vault.Raw, err = structToEnv(vault.Env)
	return err
}
