package statustype

// error, failure, pending, success
const (
	// Error is the status of a failed check
	Error Status = "error"
	// Failure is the status of a failed check
	Failure Status = "failure"
	// Pending is the status of a pending check
	Pending Status = "pending"
	// Success is the status of a successful check
	Success Status = "success"
)

type Status string

// String returns the string representation of the status
func (s Status) String() string {
	return string(s)
}
