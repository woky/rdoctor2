package main

import (
	"fmt"
	"net/url"
	"os"
	"os/user"
	"path"
	"runtime"
	"strings"

	//"github.com/go-ini/ini"
	"gopkg.in/ini.v1"
)

type Config struct {
	location    string
	iniConf     *ini.File
	mainSection *ini.Section
}

func (conf *Config) Location() string {
	return conf.location
}

func getFromEnv(key string) (val string, present bool) {
	varName := "RDOCTOR_" + strings.ToUpper(key)
	val, present = os.LookupEnv(varName)
	return
}

func (conf *Config) Has(key string) bool {
	if conf.mainSection.HasKey(key) {
		return true
	}
	_, present := getFromEnv(key)
	return present
}

func (conf *Config) Get(key string) string {
	if conf.mainSection.HasKey(key) {
		return conf.mainSection.Key(key).String()
	}
	val, present := getFromEnv(key)
	if !present {
		panic("Configuration key '" + key + "' not set!")
	}
	return val
}

func (conf *Config) GetOrElse(key string, def string) string {
	if conf.Has(key) {
		return conf.Get(key)
	}
	return def
}

func (conf *Config) GetBool(key string, def bool) bool {
	if !conf.Has(key) {
		return def
	}
	val := conf.Get(key)
	return !(val == "0" || strings.EqualFold(val, "no") ||
		strings.EqualFold(val, "false"))
}

func (conf *Config) HasApiKey() bool {
	return conf.Has("apikey")
}

func (conf *Config) GetApiKey() string {
	return conf.Get("apikey")
}

func (conf *Config) SetApiKey(value string) {
	conf.mainSection.Key("apikey").SetValue(value)
}

func (conf *Config) GetServiceAddress() string {
	return conf.GetOrElse("address", "rdoctor.rchain-dev.tk")
}

func (conf *Config) IsServiceSecure() bool {
	return conf.GetBool("secure", true)
}

func (conf *Config) GetServiceHttpUrl(path string) string {
	proto := "https://"
	if !conf.IsServiceSecure() {
		proto = "http://"
	}
	return proto + conf.GetServiceAddress() + path
}

func (conf *Config) GetServiceWsUrl(path string) string {
	proto := "wss://"
	if !conf.IsServiceSecure() {
		proto = "ws://"
	}
	return proto + conf.GetServiceAddress() + path
}

func (conf *Config) GetLatestVersionUrl() string {
	return conf.GetServiceHttpUrl("/download/latest/version.txt")
}

func (conf *Config) GetLatestClientUrl() string {
	path := fmt.Sprintf("/download/latest/%s.%s/rdoctor", runtime.GOOS,
		runtime.GOARCH)
	return conf.GetServiceHttpUrl(path)
}

func (conf *Config) GetNewKeyUrl(identity string) string {
	query := url.Values{"identity": {identity}}
	return conf.GetServiceHttpUrl("/api/newkey?" + query.Encode())
}

func (conf *Config) GetConfirmKeyUrl(apiKey string) string {
	return conf.GetServiceHttpUrl("/confirm?key=" + apiKey)
}

func (conf *Config) GetPingUrl() string {
	return conf.GetServiceHttpUrl("/api/ping?key=" + conf.GetApiKey())
}

func (conf *Config) GetSubmitLogUrl() string {
	return conf.GetServiceWsUrl("/api/ws/submitlog?key=" + conf.GetApiKey())
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
