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
	Room RoomConfig `toml:"room"`
}

type LoggerConfig struct {
	LogLevel string `toml:"log_level"`
	LogFile  string `toml:"log_file"`
}

type AppConfig struct {
	AppPort string `toml:"app_port"`
}

type RoomConfig struct {
	CommandStalePeriodMs int `toml:"command_stale_period_ms"`
	RoomTickerPeriodMs int `toml:"room_ticker_period_ms"`
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
