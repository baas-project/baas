package machines

import (
	"log"

	"github.com/krolaw/dhcp4"
)

type dhcpHandler struct {
	machineStore MachineStore
}

// ServeDHCP does the actual work explained at the WatchArchitecturesDhcp documentation.
func (handler dhcpHandler) ServeDHCP(req dhcp4.Packet, _ dhcp4.MessageType, options dhcp4.Options) dhcp4.Packet {
	arch := options[dhcp4.OptionClientArchitecture]
	if len(arch) != 2 {
		// commented because this happens quite often
		// fmt.Printf("Received dhcp packet with invalid arch option (%v)\n", arch)
		return dhcp4.Packet{}
	}

	archID := int(arch[0])<<8 | int(arch[1])

	mac := req.CHAddr()

	var systemArchitecture SystemArchitecture

	// From https://www.ietf.org/assignments/dhcpv6-parameters/dhcpv6-parameters.xml#processor-architecture
	switch archID {
	case 0:
		systemArchitecture = X86_64 // x86 with bios (32/64?)
	case 6 | 7:
		systemArchitecture = X86_64 // x86 with uefi (32/64?)
	case 10:
		systemArchitecture = Unknown // Arm 32 bits with uefi (unknown because we dont support arm32 (yet))
	case 11:
		systemArchitecture = Arm64 // Arm 64 bits with uefi

	default:
		systemArchitecture = Unknown
	}

	log.Printf("Identified mac address %v as architecture id %v (%v)\n", mac, archID, systemArchitecture.Name())

	machine := Machine{
		MacAddress:   mac.String(),
		Architecture: systemArchitecture,
	}

	err := handler.machineStore.UpdateMachine(machine)
	if err != nil {
		log.Printf("An error occurred: %v\n", err)
	}

	return dhcp4.Packet{}
}

// WatchArchitecturesDhcp starts listening for dhcp requests.
// In these requests, machines announce their architecture, which is stored
// in the machine store, to be used when pixiecore requests what OS the machine should boot.
func WatchArchitecturesDhcp(store MachineStore) {
	log.Printf("Starting Architecture Watcher")
	err := dhcp4.ListenAndServe(dhcpHandler{
		store,
	})

	if err != nil {
		println(err.Error())
	}
}
