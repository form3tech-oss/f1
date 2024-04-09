package metrics

type ResultType string

const (
	SucessResult  ResultType = "success"
	FailedResult  ResultType = "fail"
	DroppedResult ResultType = "dropped"
	UnknownResult ResultType = "unknown"
)

func (r ResultType) String() string {
	return string(r)
}

func ResultTypeFromString(result string) ResultType {
	switch result {
	case SucessResult.String():
		return SucessResult
	case FailedResult.String():
		return FailedResult
	case DroppedResult.String():
		return DroppedResult
	case UnknownResult.String():
		return UnknownResult
	default:
		return UnknownResult
	}
}

func Result(failed bool) ResultType {
	if failed {
		return FailedResult
	}
	return SucessResult
}
