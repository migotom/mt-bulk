package entities

// Result response from worker.
type Result struct {
	Host      Host
	Job       Job
	Responses []string
	Error     error
}
