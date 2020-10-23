package ipmi

import (
	"github.com/gebn/bmc/pkg/ipmi"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// IPMI Spec
// https://www.intel.com/content/dam/www/public/us/en/documents/product-briefs/ipmi-second-gen-interface-spec-v2-rev1-1.pdf

type BootDevice uint8

// In same order as the spec
const (
	BootDeviceNoOverride         BootDevice = 0b0000
	BootDeviceForcePXE           BootDevice = 0b0001
	BootDeviceForceHDD           BootDevice = 0b0010
	BootDeviceForceHDDSafe       BootDevice = 0b0011
	BootDeviceForceDiagnostics   BootDevice = 0b0100
	BootDeviceForceDVD           BootDevice = 0b0101
	BootDeviceForceBIOS          BootDevice = 0b0110
	BootDeviceForceRemoteFloppy  BootDevice = 0b0111
	BootDeviceForceRemotePrimary BootDevice = 0b1001
	BootDeviceForceRemoteDVD     BootDevice = 0b1000
	BootDeviceForceRemoteHDD     BootDevice = 0b1011
	BootDeviceForceFloppy        BootDevice = 0b1111
)

type FirmwareVerbosity uint8

const (
	FirmwareVerbosityDefault FirmwareVerbosity = 0b00
	FirmwareVerbosityQuiet   FirmwareVerbosity = 0b01
	FirmwareVerbosityVerbose FirmwareVerbosity = 0b10
)

type ConsoleRedirectionControl uint8

const (
	ConsoleRedirectionControlDefault  ConsoleRedirectionControl = 0b00
	ConsoleRedirectionControlSuppress ConsoleRedirectionControl = 0b01
	ConsoleRedirectionControlEnabled  ConsoleRedirectionControl = 0b10
)

type BIOSMuxControlOverride uint8

const (
	BIOSMuxControlOverrideRecommended BIOSMuxControlOverride = 0b000
	BIOSMuxControlOverrideBMC         BIOSMuxControlOverride = 0b001
	BIOSMuxControlOverrideSystem      BIOSMuxControlOverride = 0b010
)

type BootFlags struct {
	// Byte 1
	Valid      bool // The bit should be set to indicate that valid flag data is present
	Persistent bool // If the options should be persistent or apply to next boot only
	UEFI       bool // Use UEFI or BIOS

	// Byte 2
	ClearCMOS           bool
	LockKeyboard        bool
	BootDevice          BootDevice
	ScreenBlank         bool
	LockOutResetButtons bool

	// Byte 3
	LockOutPowerButton        bool
	FirmwareVerbosity         FirmwareVerbosity
	ForceProgressEventTraps   bool
	UserPasswordBypass        bool
	LockOutSleepButton        bool
	ConsoleRedirectionControl ConsoleRedirectionControl

	// Byte 4
	BIOSSharedModeOverride bool
	BIOSMuxControlOverride BIOSMuxControlOverride

	// Byte 5
	DeviceInstanceSelector uint8
}

type GetBootDevResp struct {
	layers.BaseLayer

	CompletionCode        ipmi.CompletionCode
	ParameterNotSupported bool

	ParameterVersion uint8

	ParameterValid bool

	BootOptionsParameterSelector uint8

	// Parameter0: Not implemented
	// Parameter1: Not implemented
	// Parameter2: Not implemented
	// Parameter3: Not implemented
	// Parameter4: Not implemented
	// Parameter5: Boot Flags
	BootFlags
}

type GetBootDev struct {
	Rsp GetBootDevResp
}

func (b *GetBootDev) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error {
	panic("implement me")
}

func (b *GetBootDev) CanDecode() gopacket.LayerClass {
	panic("implement me")
}

func (b *GetBootDev) NextLayerType() gopacket.LayerType {
	panic("implement me")
}

func (b *GetBootDev) LayerPayload() []byte {
	panic("implement me")
}

//OperationGetChassisStatusReq
var OperationGetBootDevReq = ipmi.Operation{
	Function: ipmi.NetworkFunctionChassisReq,
	Command:  0x09, // 0x09 == get (Appendix G)
}

// Name returns the name of the command, without request/response suffix
// e.g. "Get Device ID". This is used for metrics.
func (b *GetBootDev) Name() string {
	return "Change Boot Device"
}

// Operation returns the operation parameters for the request. This should
// avoid allocation, referencing a value in static memory. Technically, this
// should be a member of a Request interface that embeds
// gopacket.SerializableLayer, however it is here to allow Request() to
// return nil for commands not requiring a request payload, which would
// otherwise need to have a no-op layer created.
func (b *GetBootDev) Operation() *ipmi.Operation {
	return &OperationGetBootDevReq
}

// Request returns the possibly nil request layer that we send to the
// managed system. This should not allocate any additional memory.
func (b *GetBootDev) Request() gopacket.SerializableLayer {
	return nil
}

// Response returns the possibly nil response layer that we expect back from
// the managed system following our request. This should not allocate any
// additional memory.
func (b *GetBootDev) Response() gopacket.DecodingLayer {

}
