package libtf

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInt(t *testing.T) {
	res1 := envIntToString(10)
	assert.Equal(t, "10", res1)

	res2, err2 := envStringToInt("10")
	assert.Nil(t, err2)
	assert.Equal(t, 10, res2)

	_, err3 := envStringToInt("sdf")
	assert.Error(t, err3)
}

func TestList(t *testing.T) {
	res1, err1 := envListToString([]interface{}{"foo", "bar", "spam"})
	assert.Nil(t, err1)
	assert.Equal(t, "foo,bar,spam", res1)

	_, err2 := envListToString([]interface{}{"foo", 1})
	assert.Error(t, err2)

	res3 := envStringToList("foo,bar,spam")
	assert.Equal(t, []interface{}{"foo", "bar", "spam"}, res3)
}

func TestBool(t *testing.T) {
	assert.Equal(t, "true", envBoolToString(true))
	assert.Equal(t, "false", envBoolToString(false))

	res1, err1 := envStringToBool("yes")
	assert.Nil(t, err1)
	assert.True(t, res1)

	res2, err2 := envStringToBool("true")
	assert.Nil(t, err2)
	assert.True(t, res2)

	res3, err3 := envStringToBool("no")
	assert.Nil(t, err3)
	assert.False(t, res3)

	res4, err4 := envStringToBool("false")
	assert.Nil(t, err4)
	assert.False(t, res4)

	_, err5 := envStringToBool("foo")
	assert.Error(t, err5)
}

func TestDict(t *testing.T) {
	data1 := map[string]interface{}{
		"foo":  "test",
		"bar":  1,
		"spam": []interface{}{"foo", "bar"},
		"eggs": map[string]interface{}{
			"foo":  "bar",
			"spam": "eggs",
		},
	}
	str, err1 := envDictToString(data1)
	assert.Nil(t, err1)
	data2, err2 := envStringToDict(str)
	assert.Nil(t, err2)
	assert.True(t, reflect.DeepEqual(data1, data2))
}
