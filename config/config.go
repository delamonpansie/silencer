package config

import (
	"flag"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"

	"github.com/buildkite/interpolate"
	"github.com/creasty/defaults"
	"github.com/go-yaml/yaml"
)

type Rule struct {
	Name     string        `yaml:"name"`
	Re       []string      `yaml:"re"`
	Duration time.Duration `yaml:"duration"`
}

type LogFile struct {
	FileName string        `yaml:"file_name"`
	Rule     []Rule        `yaml:"rule"`
	Duration time.Duration `yaml:"duration"`
}

type IPTables struct {
	Chain string `yaml:"chain"`
}

type IPSet struct {
	Set string `yaml:"set"`
}
type Filter struct {
	IPTables *IPTables `yaml:"iptables,omitempty"`
	IPSet    *IPSet    `yaml:"ipset,omitempty"`
}

type Config struct {
	Duration  time.Duration     `yaml:"duration" default:"168h" `
	LogFile   []LogFile         `yaml:"log_file"`
	Env       map[string]string `yaml:"env"`
	Filter    Filter            `yaml:"filter"`
	Whitelist []net.IPNet       `yaml:"whitelist"`
}

var configName = flag.String("config", "silencer.yaml", "path to configuration file")

func expand(s string, mapEnv map[string]string) string {
	env := os.Environ()
	for k, v := range mapEnv {
		env = append(env, k+"="+v)
	}
	s, err := interpolate.Interpolate(interpolate.NewSliceEnv(env), s)
	if err != nil {
		log.Fatal(err)
	}
	return s
}

func Load() Config {
	var data []byte
	var err error

	data, err = ioutil.ReadFile(*configName)
	if err != nil {
		log.Fatal(err)
	}

	config := Config{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		panic(err)
	}
	// replace ${var} or $var in the string according to env & config.Env
	data = []byte(expand(string(data), config.Env))

	config = Config{}
	defaults.Set(&config)
	if err := yaml.UnmarshalStrict(data, &config); err != nil {
		panic(err)
	}
	if config.Duration == 0 {
		log.Fatal("default duration is 0")
	}

	for _, subnet := range config.Whitelist {
		if len(subnet.Mask) != 4 {
			log.Fatal("net mask length not equal 4")
		}
	}

	if (config.Filter.IPSet != nil && config.Filter.IPTables != nil) ||
		(config.Filter.IPSet == nil && config.Filter.IPTables == nil) {
		log.Fatal("exactly one filter must be configured")
	}

	for i := range config.LogFile {
		if config.LogFile[i].Duration == 0 {
			config.LogFile[i].Duration = config.Duration
		}
		for j := range config.LogFile[i].Rule {
			if config.LogFile[i].Rule[j].Duration == 0 {
				config.LogFile[i].Rule[j].Duration = config.LogFile[i].Duration
			}
		}
	}

	return config
}
