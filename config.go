package core

import (
	"encoding/json"
	"io/ioutil"
)

type confg struct {
	BaseUrl       string                 `json:"base_url"`
	FCGI          bool                   `json:"fcgi"`
	HandleContent bool                   `json:"handle_content"`
	Port          int                    `json:"port"`
	Subdir        string                 `json:"subdir"`
	Views         map[string]string      `json:"views"`
	Data          map[string]interface{} `json:"data"`
}

var Config = &confg{}

func (this *confg) Load(path string) error {
	// default configurations
	if this.Port == 0 {
		this.Port = 8080
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(bytes, this); err != nil {
		return err
	}
	Log.Info("Configuration loaded from file:", path)
	return nil
}
