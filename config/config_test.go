package config

import (
	"encoding/json"
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xorcare/golden"
)

func TestMain(m *testing.M) {
	flag.Set("config", "../silencer.yaml")
	os.Exit(m.Run())
}

func Test_Load(t *testing.T) {
	c := Load()
	buf, err := json.MarshalIndent(c, "", "  ")
	require.NoError(t, err)
	golden.Assert(t, buf)
}
