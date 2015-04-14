package core

import (
	"encoding/json"
	"io/ioutil"

	"github.com/jzaikovs/core/loggy"
	"github.com/jzaikovs/t"
)

type configStruct struct {
	Host          string                 `json:"host"`
	BaseURL       string                 `json:"base_url"`
	FCGI          bool                   `json:"fcgi"`
	HandleContent bool                   `json:"handle_content"`
	Port          int                    `json:"port"`
	Subdir        string                 `json:"subdir"`
	Views         map[string]string      `json:"views"`
	Data          map[string]interface{} `json:"data"`

	err_object_func func(code int, err error) interface{}
}

// constructor for t_config object
func newConfigStruct() *configStruct {
	this := new(configStruct)

	// defaul rest error handler
	this.SetRESTErrObjectFunc(func(code int, err error) interface{} {
		obj := t.Map{"code": code}
		if err != nil {
			obj["error"] = err.Error()
		}
		return obj
	})

	return this
}

// Load is function for loading configuration from json file specified by path parameter.
func (config *configStruct) Load(path string) error {
	// default configurations
	if config.Port == 0 {
		config.Port = 8080
	}

	if len(config.Host) == 0 {
		// by default we listen to all ip
		config.Host = "0.0.0.0"
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(bytes, config); err != nil {
		return err
	}
	loggy.Info("Configuration loaded from file:", path)

	b, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		loggy.Info(err)
	}

	loggy.Info("Configs:\n", string(b))

	return nil
}

func (config *configStruct) SetRESTErrObjectFunc(fn func(code int, err error) interface{}) {
	config.err_object_func = fn
}
