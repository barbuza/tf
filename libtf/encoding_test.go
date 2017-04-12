package libtf

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	res1 := envListToString([]string{"foo", "bar", "spam"})
	assert.Equal(t, "foo,bar,spam", res1)

	res3 := envStringToList("foo,bar,spam")
	assert.Equal(t, []string{"foo", "bar", "spam"}, res3)
}

func TestBool(t *testing.T) {
	assert.Equal(t, "yes", envBoolToString(true))
	assert.Equal(t, "no", envBoolToString(false))

	res1, err1 := envStringToBool("yes")
	assert.Nil(t, err1)
	assert.True(t, res1)

	res2, err2 := envStringToBool("no")
	assert.Nil(t, err2)
	assert.False(t, res2)

	_, err3 := envStringToBool("foo")
	assert.Error(t, err3)
}

func TestDict(t *testing.T) {
	data1 := map[string]interface{}{
		"foo":  "test",
		"bar":  1,
		"spam": []string{"foo", "bar"},
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
