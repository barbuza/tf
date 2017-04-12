package libtf

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"gopkg.in/yaml.v2"
)

func envDictToString(input map[string]interface{}) (string, error) {
	data, err := yaml.Marshal(input)
	if err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(data), nil
}

func envBoolToString(input bool) string {
	if input {
		return "yes"
	}
	return "no"
}

func envListToString(input []string) string {
	return strings.Join(input, ",")
}

func envStringToList(input string) []string {
	return strings.Split(input, ",")
}

func envStringToBool(input string) (bool, error) {
	switch input {
	case "yes":
		return true, nil
	case "no":
		return false, nil
	default:
		return false, fmt.Errorf("value '%s' is not bool", input)
	}
}

func envStringToDict(input string) (map[string]interface{}, error) {
	data, err := base64.RawStdEncoding.DecodeString(input)
	if err != nil {
		return nil, err
	}
	var dict map[string]interface{}
	err = yaml.Unmarshal(data, &dict)
	if err != nil {
		return nil, err
	}
	fixInterfacesInMap(dict)
	return dict, nil
}

func structToEnv(input map[string]interface{}) (map[string]string, error) {
	res := make(map[string]string)
	for key, value := range input {
		key = EnvKey(key)
		switch value.(type) {
		case string:
			res[key] = value.(string)
		case bool:
			res[key] = envBoolToString(value.(bool))
		case []string:
			res[key] = envListToString(value.([]string))
		case map[string]interface{}:
			dict, err := envDictToString(value.(map[string]interface{}))
			if err != nil {
				return nil, err
			}
			res[key] = dict
		default:
			return nil, fmt.Errorf("struct value is unsupported: %s", spew.Sdump(value))
		}
	}
	return res, nil
}
