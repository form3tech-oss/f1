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

func Result(failed bool) ResultType {
	if failed {
		return FailedResult
	}
	return SucessResult
}
