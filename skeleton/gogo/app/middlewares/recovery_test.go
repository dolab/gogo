package middlewares

import (
	"testing"

	"github.com/dolab/gogo"
)

func Test_Recovery(t *testing.T) {
	testApp.Use(Recovery())
	defer testApp.CleanModdilewares()

	// register temp resource for testing
	testApp.GET("/middlewares/recovery", func(ctx *gogo.Context) {
		panic("Recover testing")
	})

	testClient.Get(t, "/middlewares/recovery", nil)
	testClient.AssertOK()
}
