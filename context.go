package core

type Context interface {
	Input
	Output
	//In  Input
	//Out Output
}

type t_context struct {
	Input
	Output
}
