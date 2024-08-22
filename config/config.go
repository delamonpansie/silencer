package config

import (
	"errors"
	"flag"
	"io/fs"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/buildkite/interpolate"
	"github.com/creasty/defaults"
	"github.com/delamonpansie/silencer/logger"
	"github.com/go-yaml/yaml"
	"go.uber.org/zap"
)

var log = &logger.Log

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

type NFT struct {
	Table string `yaml:"table"`
	Set   string `yaml:"set"`
}

type IPSet struct {
	Set string `yaml:"set"`
}

type Filter struct {
	NFT   *NFT   `yaml:"nft,omitempty"`
	IPSet *IPSet `yaml:"ipset,omitempty"`
}

type Config struct {
	Duration  time.Duration     `yaml:"duration" default:"168h" `
	LogFile   []LogFile         `yaml:"log_file"`
	Env       map[string]string `yaml:"env"`
	Filter    Filter            `yaml:"filter"`
	Whitelist []net.IPNet       `yaml:"whitelist"`
}

var configName = flag.String("config", "/etc/silencer.yaml", "path to configuration file")

func expand(s string, mapEnv map[string]string) string {
	env := os.Environ()
	for k, v := range mapEnv {
		env = append(env, k+"="+v)
	}
	s, err := interpolate.Interpolate(interpolate.NewSliceEnv(env), s)
	if err != nil {
		log.Fatal("expand", zap.Error(err))
	}
	return s
}

func Load() Config {
	var data []byte
	var err error

	data, err = ioutil.ReadFile(*configName)
	if errors.Is(err, fs.ErrNotExist) {
		data, err = ioutil.ReadFile(filepath.Base(*configName))
	}
	if err != nil {
		log.Fatal("Load", zap.Error(err))
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

	switch {
	case config.Filter.IPSet == nil && config.Filter.NFT == nil:
		log.Fatal("at least one filter must be configured")
	case config.Filter.IPSet != nil && config.Filter.NFT != nil:
		log.Fatal("at most one filter must be configured")
	case config.Filter.IPSet != nil && config.Filter.IPSet.Set == "":
		log.Fatal("ipset set name cannot be empty")
	case config.Filter.NFT != nil && config.Filter.NFT.Table == "":
		log.Fatal("nft table name cannot be empty")
	case config.Filter.NFT != nil && config.Filter.NFT.Set == "":
		log.Fatal("nft set name cannot be empty")
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
