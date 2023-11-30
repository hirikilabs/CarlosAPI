package config

import (
	"github.com/BurntSushi/toml"
	"sync/atomic"
)

const (
	version string = "0.1"
)

type Config struct {
	Addr		string  `toml:"addr"`
	Port		int     `toml:"port"`
	RecordPath  string  `toml:"record_path"`
	RecordCmd   string  `toml:"record_cmd"`
	RecordArgs  string  `toml:"record_args"`
	Database	string  `toml:"database"`
	Version     string
}

// atomic so is thread safe
var recording atomic.Bool

// return configuration data from TOML file
func GetConfig() Config {
	var config Config
	_, err := toml.DecodeFile("config.toml", &config)
	if err != nil {
		panic(err)
	}
	config.Version = version
	return config
}

func IsRecording() bool {
	return recording.Load()
}

func Recording() {
	recording.Store(true)
}

func NoRecording() {
	recording.Store(false)
}
