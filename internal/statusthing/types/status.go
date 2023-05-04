package types

const (
	// redStatusString is the string representation of StatusRed
	redStatusString = "STATUS_RED"
	// greenStatusString is the string representation of StatusGreen
	greenStatusString = "STATUS_GREEN"
	// yellowStatusString is the string representation of StatusYellow
	yellowStatusString = "STATUS_YELLOW"
	// unknownStatusString is the string representation of StatusUnknown
	unknownStatusString = "STATUS_UNKNOWN"
)

// Status represents a status
type Status int64

const (
	// StatusUnknown is the default status code
	StatusUnknown Status = iota
	// StatusRed is generally the bad status
	StatusRed
	// StatusGreen is generally the good status
	StatusGreen
	// StatusYellow is generally the warning/remediation status
	StatusYellow
)

// StatusFromString returns a [types.Status] from its string representation
func StatusFromString(statusString string) Status {
	switch statusString {
	case StatusRed.String():
		return StatusRed
	case StatusGreen.String():
		return StatusGreen
	case StatusYellow.String():
		return StatusYellow
	default:
		return StatusUnknown
	}
}

// String returns the string representation of a status code
func (s Status) String() string {
	switch s {
	case StatusRed:
		return redStatusString
	case StatusGreen:
		return greenStatusString
	case StatusYellow:
		return yellowStatusString
	default:
		return unknownStatusString
	}
}
