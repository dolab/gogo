package middlewares

import (
	"runtime"
	"strings"

	"github.com/dolab/gogo"
)

func Recovery() gogo.Middleware {
	return func(ctx *gogo.Context) {
		defer func() {
			if panicErr := recover(); panicErr != nil {
				// where does panic occur? try max 20 depths
				pcs := make([]uintptr, 20)
				max := runtime.Callers(2, pcs)
				for i := 0; i < max; i++ {
					pcfunc := runtime.FuncForPC(pcs[i])
					if strings.HasPrefix(pcfunc.Name(), "runtime.") {
						continue
					}

					pcfile, pcline := pcfunc.FileLine(pcs[i])

					tmp := strings.SplitN(pcfile, "/src/", 2)
					if len(tmp) == 2 {
						pcfile = "src/" + tmp[1]
					}
					ctx.Logger.Errorf("(%s:%d: %v)", pcfile, pcline, panicErr)

					break
				}

				ctx.Abort()
			}
		}()

		ctx.Next()
	}
}
