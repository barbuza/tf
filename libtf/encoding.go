package libtf

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/barbuza/tf/json_compat"
	"github.com/davecgh/go-spew/spew"
	"gopkg.in/yaml.v2"
	"strconv"
)

func envIntToString(input int) string {
	return strconv.Itoa(input)
}

func envStringToInt(input string) (int, error) {
	return strconv.Atoi(input)
}

func envDictToString(input map[string]interface{}) (string, error) {
	data, err := yaml.Marshal(input)
	if err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(data), nil
}

func envBoolToString(input bool) string {
	if input {
		return "true"
	}
	return "false"
}

func envListToString(input []interface{}) (string, error) {
	res := make([]string, len(input))
	for idx, value := range input {
		switch value.(type) {
		case string:
			res[idx] = value.(string)
		default:
			return "", fmt.Errorf("unsupported list value %s", spew.Sdump(value))
		}
	}
	return strings.Join(res, ","), nil
}

func envStringToList(input string) []interface{} {
	s := strings.Split(input, ",")
	res := make([]interface{}, len(s))
	for idx, item := range s {
		res[idx] = item
	}
	return res
}

func envStringToBool(input string) (bool, error) {
	switch input {
	case "yes":
		return true, nil
	case "true":
		return true, nil
	case "false":
		return false, nil
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
	var dict map[interface{}]interface{}
	err = yaml.Unmarshal(data, &dict)
	if err != nil {
		return nil, err
	}
	return json_compat.ConvertMap(dict)
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
		case int:
			res[key] = envIntToString(value.(int))
		case []interface{}:
			list, err := envListToString(value.([]interface{}))
			if err != nil {
				return nil, err
			}
			res[key] = list
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
