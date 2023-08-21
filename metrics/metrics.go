package metrics

import "github.com/prometheus/client_golang/prometheus"

const (
	namespace         = "elevator"
	elevatorNameLabel = "elevator"
)

var (
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    namespace + "_request_duration_seconds",
			Help:    "Duration of elevator request processing",
			Buckets: []float64{0.1, 0.5, 1, 2, 5},
		},
		[]string{elevatorNameLabel},
	)
)

func init() {
	prometheus.MustRegister(requestDuration)
}

func RequestDurationHistogram(elevatorName string, seconds float64) {
	requestDuration.With(prometheus.Labels{elevatorNameLabel: elevatorName}).Observe(seconds)
}
