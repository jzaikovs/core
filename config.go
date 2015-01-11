package core

import (
	"encoding/json"
	"github.com/jzaikovs/core/loggy"
	"io/ioutil"
)

type t_configs struct {
	Host          string                 `json:"host"`
	BaseUrl       string                 `json:"base_url"`
	FCGI          bool                   `json:"fcgi"`
	HandleContent bool                   `json:"handle_content"`
	Port          int                    `json:"port"`
	Subdir        string                 `json:"subdir"`
	Views         map[string]string      `json:"views"`
	Data          map[string]interface{} `json:"data"`
}

// function for loading configuration from json file specified by path parameter.
func (this *t_configs) Load(path string) error {
	// default configurations
	if this.Port == 0 {
		this.Port = 8080
	}

	if len(this.Host) == 0 {
		// by default we listen to all ip
		this.Host = "0.0.0.0"
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(bytes, this); err != nil {
		return err
	}
	loggy.Info("Configuration loaded from file:", path)

	b, err := json.MarshalIndent(this, "", "  ")
	if err != nil {
		loggy.Info(err)
	}

	loggy.Info("Configs:\n", string(b))
	return nil
}
