package middlewares

import (
	"testing"

	"github.com/dolab/gogo"
)

func Test_Recovery(t *testing.T) {
	gogoapp.Use(Recovery())
	defer gogoapp.CleanModdilewares()

	// register temp resource for testing
	gogoapp.GET("/middlewares/recovery", func(ctx *gogo.Context) {
		panic("Recover testing")
	})

	request := gogotest.New(t)
	request.Get("/middlewares/recovery", nil)
	request.AssertOK()
}
