package retriever

import "time"

// TimeRange defines a time window for message retrieval
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// WorkersCount is a constant to represent number of workers to be used in ExecuteJobs
var WorkersCount = 100
