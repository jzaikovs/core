package core

import (
	"encoding/json"
	"io/ioutil"
)

type configs struct {
	BaseUrl       string                 `json:"base_url"`
	FCGI          bool                   `json:"fcgi"`
	HandleContent bool                   `json:"handle_content"`
	Port          int                    `json:"port"`
	Subdir        string                 `json:"subdir"`
	Views         map[string]string      `json:"views"`
	Data          map[string]interface{} `json:"data"`
}

// function for loading configuration from json file specified by path parameter.
func (this *configs) Load(path string) error {
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
	b, _ := json.MarshalIndent(this, "", "  ")
	Log.Info("Configs:\n", string(b))
	return nil
}
