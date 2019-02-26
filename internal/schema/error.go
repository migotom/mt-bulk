package schema

// Error response from worker.
type Error struct {
	Host
	Message string
}
