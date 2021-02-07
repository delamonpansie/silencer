package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseIptablesList(t *testing.T) {
	input := []byte(`
Chain silence (0 references)
target     prot opt source               destination
DROP       all  --  141.98.10.143        0.0.0.0/0
DROP       all  --  77.40.61.86          0.0.0.0/0
DROP       all  --  165.232.39.56        0.0.0.0/0
DROP       all  --  156.96.118.58        0.0.0.0/0
DROP       all  --  37.49.225.115        0.0.0.0/0
DROP       all  --  193.169.254.106      0.0.0.0/0
DROP       all  --  141.98.10.192        0.0.0.0/0
DROP       all  --  103.253.42.54        0.0.0.0/0
DROP       all  --  45.125.65.105        0.0.0.0/0
DROP       all  --  45.125.65.39         0.0.0.0/0
DROP       all  --  185.36.81.33         0.0.0.0/0
DROP       all  --  141.98.10.136        0.0.0.0/0
DROP       all  --  141.98.10.143        0.0.0.0/0
RETURN     all  --  0.0.0.0/0            0.0.0.0/0
`)
	expected := []string{
		"141.98.10.143",
		"77.40.61.86",
		"165.232.39.56",
		"156.96.118.58",
		"37.49.225.115",
		"193.169.254.106",
		"141.98.10.192",
		"103.253.42.54",
		"45.125.65.105",
		"45.125.65.39",
		"185.36.81.33",
		"141.98.10.136",
		"141.98.10.143",
	}

	list := parseIptablesList(input)
	output := make([]string, len(list))
	for i, l := range list {
		output[i] = l.String()
	}
	assert.Equal(t, expected, output)
}
