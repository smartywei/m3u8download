package config

import (
	"gopkg.in/ini.v1"
)

var configKV map[string]interface{}

var CFG *ini.File

func InitConfig() {
	now_cfg, err := ini.Load("./config/m3u8Tool.ini")
	if err != nil {
		panic(err)
	}
	CFG = now_cfg
}