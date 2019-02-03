package middlewares

import (
	"net/http"
	"runtime"
	"strings"

	"github.com/dolab/gogo"
)

func Recovery() gogo.FilterFunc {
	return func(ctx *gogo.Context) {
		defer func() {
			if panicErr := recover(); panicErr != nil {
				// where does panic occur? try max 20 depths
				pcs := make([]uintptr, 20)
				max := runtime.Callers(2, pcs)

				if max == 0 {
					ctx.Logger.Warn("No pcs available.")
				} else {
					frames := runtime.CallersFrames(pcs[:max])
					for {
						frame, more := frames.Next()

						// To keep this example's output stable
						// even if there are changes in the testing package,
						// stop unwinding when we leave package runtime.
						if strings.Contains(frame.Function, "runtime.") {
							if more {
								continue
							} else {
								break
							}
						}

						tmp := strings.SplitN(frame.File, "/src/", 2)
						if len(tmp) == 2 {
							ctx.Logger.Errorf("(src/%s:%d: %v)", tmp[1], frame.Line, panicErr)
						} else {
							ctx.Logger.Errorf("(%s:%d: %v)", frame.File, frame.Line, panicErr)
						}

						break
					}
				}

				ctx.SetStatus(http.StatusInternalServerError)
				ctx.Return(panicErr)
			}
		}()

		ctx.Next()
	}
}
