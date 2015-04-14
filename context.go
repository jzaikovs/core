package core

// Context interface combines input and output interfaces so that
// RouteFunc accepts single parameter - Context
type Context interface {
	Input
	Output
}

type context struct {
	Input
	Output
}
