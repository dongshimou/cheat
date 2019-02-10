package config

import (
	"cheat/model"
	"encoding/json"
	"io/ioutil"
	"os"
)

var (
	projectConfig *model.Config
)

func Get() *model.Config {
	return projectConfig
}
func init() {

	file, err := os.Open("config.cfg")
	if err != nil {
		panic(err)
		return
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
		return
	}
	cfg := model.Config{}

	if err := json.Unmarshal(data, &cfg); err != nil {
		panic(err)
		return
	}
	projectConfig = &cfg
}
