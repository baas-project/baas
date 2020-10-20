package machines

// SystemArchitecture defines constants describing the architecture of machines.
type SystemArchitecture string

const (
	// Arm64 is the 64 bit Arm architecture
	Arm64 SystemArchitecture = "Arm64"
	// X8664 is the 64 bit x86 architecture
	X8664 SystemArchitecture = "x86_64"
	// Unknown is any architecture which baas could not identify.
	Unknown SystemArchitecture = "unknown"
)

// Name gets the name of an architecture as a string. Convenience function,
// but actually does very little as the name is also the value of the constant.
func (id *SystemArchitecture) Name() string {
	return string(*id)
}
