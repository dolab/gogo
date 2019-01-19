package render

// Render represents HTTP response render
type Render interface {
	ContentType() string
	Render(v interface{}) error
}
