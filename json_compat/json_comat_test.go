package json_compat

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValid(t *testing.T) {
	res, err := Convert(map[interface{}]interface{}{
		"foo": []interface{}{
			map[interface{}]interface{}{
				"bar": 1,
			},
		},
		"spam": map[interface{}]interface{}{
			"eggs": 1,
		},
	})

	assert.Nil(t, err)
	assert.Equal(t, map[string]interface{}{
		"foo": []interface{}{
			map[string]interface{}{
				"bar": 1,
			},
		},
		"spam": map[string]interface{}{
			"eggs": 1,
		},
	}, res)
}

func TestIntKey(t *testing.T) {
	res, err := Convert(map[interface{}]interface{}{
		1: 2,
	})
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestInvalidValue(t *testing.T) {
	res, err := Convert(map[interface{}]interface{}{
		"foo": int64(10),
	})
	assert.Nil(t, res)
	assert.Error(t, err)
}
