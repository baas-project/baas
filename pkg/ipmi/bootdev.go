package ipmi

import (
	"fmt"
	"github.com/gebn/bmc/pkg/ipmi"
	"github.com/gebn/bmc/pkg/layerexts"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// IPMI Spec
// https://www.intel.com/content/dam/www/public/us/en/documents/product-briefs/ipmi-second-gen-interface-spec-v2-rev1-1.pdf

var LayerTypeGetBootDevResp = gopacket.RegisterLayerType(
	1999,
	gopacket.LayerTypeMetadata{
		Name: "Get Boot Device Response",
		Decoder: layerexts.BuildDecoder(func() layerexts.LayerDecodingLayer {
			return &GetBootDevRsp{}
		}),
	},
)

var LayerTypeGetBootDevReq = gopacket.RegisterLayerType(
	1998,
	gopacket.LayerTypeMetadata{
		Name: "Get Boot Device Request",
	},
)


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
	ClearCMOS          bool
	LockKeyboard       bool
	BootDevice         BootDevice
	ScreenBlank        bool
	LockOutResetButton bool

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

type GetBootDevReq struct {
	layers.BaseLayer

	ParameterSelector uint8

	SetSelector uint8

	BlockSelector uint8
}

func (g GetBootDevReq) LayerType() gopacket.LayerType {
	return LayerTypeGetBootDevReq
}

func (g GetBootDevReq) SerializeTo(b gopacket.SerializeBuffer, _ gopacket.SerializeOptions) error {
	bytes, err := b.PrependBytes(3)
	if err != nil {
		return err
	}

	bytes[0] = g.ParameterSelector
	bytes[1] = g.SetSelector
	bytes[2] = g.BlockSelector

	return nil
}

type GetBootDevRsp struct {
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
	BootFlags BootFlags

	// Parameter6: Not implemented
	// Parameter7: Not implemented
}

func (g *GetBootDevRsp) LayerType() gopacket.LayerType {
	return LayerTypeGetBootDevResp
}

func (g *GetBootDevRsp) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error {
	if len(data) < 4 {
		df.SetTruncated()
		return fmt.Errorf("GetBootDevRsp must be at least 4 bytes; got %v", len(data))
	}

	g.BaseLayer.Contents = data[:4]
	g.BaseLayer.Payload = data[4:]

	// Header
	g.CompletionCode = ipmi.CompletionCode(data[0])
	g.ParameterNotSupported = data[0]&0x80 != 0
	g.ParameterVersion = data[1]
	g.ParameterValid = !(data[3]&0b10000000 != 0)
	g.BootOptionsParameterSelector = data[3] & 0b01111111

	// Parameter
	switch g.BootOptionsParameterSelector {
	case 5:
		// Byte 4+1
		g.BootFlags.Valid = data[4]&(1<<7) != 0
		g.BootFlags.Persistent = data[4]&(1<<6) != 0
		g.BootFlags.UEFI = data[4]&(1<<5) != 0

		// Byte 4+2
		g.BootFlags.ClearCMOS = data[6]&(1<<7) != 0
		g.BootFlags.BootDevice = BootDevice((data[6] & 0b00111100) >> 2)
		g.BootFlags.ScreenBlank = data[6]&(1<<1) != 0
		g.BootFlags.LockOutResetButton = data[6]&(1<<0) != 0

		// Byte 4+3
		g.BootFlags.LockOutPowerButton = data[7]&(1<<7) != 0
		g.BootFlags.FirmwareVerbosity = FirmwareVerbosity((data[7] & 0b01100000) >> 5)
		g.BootFlags.ForceProgressEventTraps = data[7]&(1<<4) != 0
		g.BootFlags.UserPasswordBypass = data[7]&(1<<3) != 0
		g.BootFlags.LockOutSleepButton = data[7]&(1<<2) != 0
		g.BootFlags.ConsoleRedirectionControl = ConsoleRedirectionControl(data[7] & 0b00000011)

		// Byte 4+4
		g.BootFlags.BIOSSharedModeOverride = data[8]&(1<<3) != 0
		g.BootFlags.BIOSMuxControlOverride = BIOSMuxControlOverride(data[8] & 0b00000011)

		// Byte 4+5
		g.BootFlags.DeviceInstanceSelector = data[9] & 0b00011111

	default:
		return fmt.Errorf("unsupported parameter type %v", g.BootOptionsParameterSelector)
	}

	return nil
}

func (g *GetBootDevRsp) CanDecode() gopacket.LayerClass {
	return g.LayerType()
}

func (g *GetBootDevRsp) NextLayerType() gopacket.LayerType {
	return gopacket.LayerTypePayload
}

type GetBootDevCmd struct {
	Req GetBootDevReq
	Rsp GetBootDevRsp
}

// Name returns the name of the command, without request/response suffix
// e.g. "Get Device ID". This is used for metrics.
func (GetBootDevCmd) Name() string {
	return "Change Boot Device"
}

var OperationGetBootDevReq = ipmi.Operation{
	Function: ipmi.NetworkFunctionChassisReq,
	Command:  0x09, // 0x09 == get (Appendix G)
}

func (GetBootDevCmd) Operation() *ipmi.Operation {
	return &OperationGetBootDevReq
}

func (b GetBootDevCmd) Request() gopacket.SerializableLayer {
	return &b.Req
}

func (b GetBootDevCmd) Response() gopacket.DecodingLayer {
	return &b.Rsp
}
