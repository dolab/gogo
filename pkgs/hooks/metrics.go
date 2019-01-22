package hooks

// import (
// 	"net/http"

// 	metrics "github.com/armon/go-metrics"
// 	"github.com/armon/go-metrics/prometheus"
// )

// // A MetricsHook defines gogo metrics
// type MetricsHook struct {
// 	sink metrics.MetricSink
// }

// func NewMetricsHook() *MetricsHook {
// 	sink, _ := prometheus.NewPrometheusSink()
// 	metrics.NewGlobal(metrics.DefaultConfig("gogo"), sink)

// 	return &MetricsHook{
// 		sink: sink,
// 	}
// }

// func (hook *MetricsHook) NewRequestReceivedHook() NamedHook {
// 	requestReceivedParts := []string{"request", "received"}

// 	return NamedHook{
// 		Name: "__metrics@request_received",
// 		Apply: func(w http.ResponseWriter, r *http.Request) bool {
// 			hook.sink.IncrCounter(requestReceivedParts, 1)
// 			return true
// 		},
// 	}
// }

// func (hook *MetricsHook) NewRequestRoutedHook() NamedHook {
// 	requestRoutedParts := []string{"request", "routed"}

// 	return NamedHook{
// 		Name: "__metrics@request_routed",
// 		Apply: func(w http.ResponseWriter, r *http.Request) bool {
// 			hook.sink.IncrCounter(requestRoutedParts, 1)
// 			return true
// 		},
// 	}
// }

// func (hook *MetricsHook) NewResponseAlwaysHook() NamedHook {
// 	requestDurationParts := []string{"request", "duration"}

// 	return NamedHook{
// 		Name: "__metrics@request_duration",
// 		Apply: func(w http.ResponseWriter, r *http.Request) bool {
// 			hook.sink.IncrCounter(requestDurationParts, 1)
// 			return true
// 		},
// 	}
// }
