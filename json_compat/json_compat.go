package json_compat

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
)

func ConvertMap(value map[interface{}]interface{}) (map[string]interface{}, error) {
	res := map[string]interface{}{}
	for name, item := range value {
		val, err := Convert(item)
		if err != nil {
			return nil, err
		}
		switch name.(type) {
		case string:
			res[name.(string)] = val
		default:
			return nil, fmt.Errorf("invalid key %s", spew.Sdump(name))
		}
	}
	return res, nil
}

func convertSlice(value []interface{}) ([]interface{}, error) {
	res := make([]interface{}, len(value))
	for idx, item := range value {
		val, err := Convert(item)
		if err != nil {
			return nil, err
		}
		res[idx] = val
	}
	return res, nil
}

func Convert(value interface{}) (interface{}, error) {
	switch value.(type) {
	case string:
		return value, nil
	case int:
		return value, nil
	case bool:
		return value, nil
	case map[interface{}]interface{}:
		return ConvertMap(value.(map[interface{}]interface{}))
	case []interface{}:
		return convertSlice(value.([]interface{}))
	default:
		return nil, fmt.Errorf("invalid value %s", spew.Sdump(value))
	}
	return value, nil
}
