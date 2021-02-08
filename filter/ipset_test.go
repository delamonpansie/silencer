package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseIsetList(t *testing.T) {
	input := []byte(`
create silencer hash:ip family inet hashsize 1024 maxelem 65536
add silencer 1.1.1.1
add silencer 2.2.2.2
add silencer 3.3.3.3
`)
	expected := []string{
		"1.1.1.1",
		"2.2.2.2",
		"3.3.3.3",
	}

	list := parseIpsetList(input)
	output := make([]string, len(list))
	for i, l := range list {
		output[i] = l.String()
	}
	assert.Equal(t, expected, output)
}
