package gogo

// ControllerDispatch is an interface that allows custom router for Resource.
type ControllerDispatch interface {
	DISPATCH(c *Context)
}

// ControllerID is an interface that allows custom resource id name with Resource.
type ControllerID interface {
	ID() string
}

// ControllerIndex is an interface that wraps the Index method. It is used by Resource
// for registering GET /resource route handler.
type ControllerIndex interface {
	Index(c *Context)
}

// ControllerShow is an interface that wraps the Show method. It is used by Resource
// for registering GET /resource/:id route handler.
type ControllerShow interface {
	Show(c *Context)
}

// ControllerCreate is an interface that wraps the Create method. It is used by Resource
// for registering POST /resource route handler.
type ControllerCreate interface {
	Create(c *Context)
}

// ControllerUpdate is an interface that wraps the Update method. It is used by Resource
// for registering PUT /resource/:id route handler.
type ControllerUpdate interface {
	Update(c *Context)
}

// ControllerExplore is an interface that wraps the Explore method. It is used by Resource
// for registering HEAD /resource/:id route handler.
type ControllerExplore interface {
	Explore(c *Context)
}

// ControllerDestroy is an interface that wraps the Destroy method. It is used by Resource
// for registering DELETE /resource/:id route handler.
type ControllerDestroy interface {
	Destroy(c *Context)
}
