package machines

type SystemArchitecture string
const(
	Arm64 SystemArchitecture = "Arm64"
	X86_64 SystemArchitecture = "x86_64"
	Unknown SystemArchitecture = "unknown"
)
