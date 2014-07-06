package core

/*
	TODO: REVIEW
*/

func (this *App) RegisterHook(name string, callback func()) {
	Log.Info("registered hook:", name)

	arr, ok := this.hooks[name]
	if !ok {
		arr = make([]func(), 0)
	}
	arr = append(arr, callback)
	this.hooks[name] = arr
}

func (this *App) executeHook(name string) {
	Log.Info("-----------" + name + "-------------")
	Log.Info("executing hook:", name)
	if arr, ok := this.hooks[name]; ok {
		for _, hook := range arr {
			Log.Info("executing hook func:", name)
			hook()
		}
	}
	Log.Info("hook", name, "executed.")
	Log.Info("-------------------------------------")
}
