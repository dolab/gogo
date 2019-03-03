package interceptors

// All phases of interceptors
const (
	_phase Phase = iota
	RequestReceived
	RequestRouted
	ResponseReady
	ResponseAlways
	phase_
)

// A Phase defines processing of a server
type Phase int

// IsValid returns true if phase has defined, otherwise returns false.
func (p Phase) IsValid() bool {
	return p > _phase && phase_ > p
}

func (p Phase) String() string {
	switch p {
	case RequestReceived:
		return "RequestReceived"
	case RequestRouted:
		return "RequestRouted"
	case ResponseReady:
		return "ResponseReady"
	case ResponseAlways:
		return "ResponseAlways"
	}

	return "Unknown"
}
