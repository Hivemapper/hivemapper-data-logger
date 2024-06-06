package logger

// Define the custom type
type CamType int

// Define constants for the custom type
const (
	Unknown CamType = iota
	HDC
	HDCS
)

// Map of string values to custom type constants
var stringToEnum = map[string]CamType{
	"hdc":  HDC,
	"hdcs": HDCS,
}

// Function to get the enum value from a string
func GetEnumFromString(s string) (CamType, error) {
	if val, ok := stringToEnum[s]; ok {
		return val, nil
	}
	panic("invalid string value")
}

// Function to get the string representation from the enum value
func (e CamType) String() string {
	switch e {
	case HDC:
		return "hdc"
	case HDCS:
		return "hdcs"
	default:
		return "unknown"
	}
}
