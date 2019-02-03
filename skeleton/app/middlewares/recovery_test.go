package middlewares

import (
	"net/http"
	"testing"

	"github.com/dolab/gogo"
)

func Test_Recovery(t *testing.T) {
	gogoapp.Use(Recovery())
	defer gogoapp.CleanFilters()

	// register temp resource for testing
	gogoapp.GET("/filters/recovery", func(ctx *gogo.Context) {
		panic("Recover testing")
	})

	request := gogotesting.New(t)
	request.Get("/filters/recovery", nil)
	request.AssertStatus(http.StatusInternalServerError)
	request.AssertNotEmpty()
}
