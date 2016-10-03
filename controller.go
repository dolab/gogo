package gogo

type ControllerID interface {
	ID() string
}
type ControllerIndex interface {
	Index(c *Context)
}

type ControllerShow interface {
	Show(c *Context)
}

type ControllerCreate interface {
	Create(c *Context)
}

type ControllerDestroy interface {
	Destroy(c *Context)
}

type ControllerUpdate interface {
	Update(c *Context)
}

type ControllerExplore interface {
	Explore(c *Context)
}
