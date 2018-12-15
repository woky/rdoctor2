package main

import (
	"os/user"
	"path"
	//"github.com/go-ini/ini"
	"gopkg.in/ini.v1"
)

type Config struct {
	location    string
	iniConf     *ini.File
	mainSection *ini.Section
}

func (conf *Config) getString(key string) (value string, present bool) {
	present = conf.mainSection.HasKey(key)
	if present {
		value = conf.mainSection.Key(key).String()
	}
	return
}

func (conf *Config) setString(key string, value string) {
	conf.mainSection.Key(key).SetValue(value)
}

func (conf *Config) Location() string {
	return conf.location
}
func (conf *Config) GetApiKey() (string, bool) {
	return conf.getString("apikey")
}

func (conf *Config) SetApiKey(value string) {
	conf.setString("apikey", value)
}

func (conf *Config) Save() {
	err := conf.iniConf.SaveTo(conf.location)
	if err != nil {
		Die("Could not save config file '%s': %s", conf.location, err)
	}
}

func configFileLocation() string {
	u, err := user.Current()
	if err != nil {
		Die("Could not get current user: %s", err)
	}
	if len(u.HomeDir) == 0 {
		Die("User does not have a home directory")
	}
	return path.Join(u.HomeDir, ".rdoctor.ini")
}

func LoadConfig() *Config {
	location := configFileLocation()
	iniConf, err := ini.LooseLoad(location)
	if err != nil {
		Die("Could not read config file '%s': %s", location, err)
	}
	return &Config{location, iniConf, iniConf.Section("")}
}
