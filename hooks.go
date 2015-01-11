package core

import (
	"github.com/jzaikovs/core/loggy"
)

func (this *App) RegisterHook(name string, callback func()) {
	loggy.Info("registered hook:", name)

	arr, ok := this.hooks[name]
	if !ok {
		arr = make([]func(), 0)
	}
	arr = append(arr, callback)
	this.hooks[name] = arr
}

func (this *App) executeHook(name string) {
	loggy.Info("-----------", name, "-------------")
	loggy.Info("executing hook:", name)

	if arr, ok := this.hooks[name]; ok {
		for _, hook := range arr {
			loggy.Info("executing hook func:", name)
			hook()
		}
	}

	loggy.Info("hook", name, "executed.")
	loggy.Info("-------------------------------------")
}
