package config

import (
	log "github.com/Sirupsen/logrus"
	"flag"
	"github.com/BurntSushi/toml"
)

var cfg Cfg

type Cfg struct {
	Logger LoggerConfig   `toml:"logger"`
	App    AppConfig      `toml:"app"`
}

type LoggerConfig struct {
	LogLevel string `toml:"log_level"`
	LogFile  string `toml:"log_file"`
}

type AppConfig struct {
	AppPort string `toml:"app_port"`
}

func init() {
	fileName := flag.String("c", "config.toml", "config file name")

	flag.Parse()
	_, err := toml.DecodeFile(*fileName, &cfg)
	if err != nil {
		log.Fatal("decode: ", err)
		return
	}
}

func GetConfig() Cfg {
	return cfg
}
