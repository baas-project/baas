package machines

import (
	"fmt"
	"github.com/krolaw/dhcp4"
	"log"
)

type dhcpHandler struct {
	machineStore MachineStore
}

func (handler dhcpHandler) ServeDHCP(req dhcp4.Packet, _ dhcp4.MessageType, options dhcp4.Options) dhcp4.Packet {
	arch := options[dhcp4.OptionClientArchitecture]
	if len(arch) != 2 {
		// commented because this happens quite often
		// fmt.Printf("Received dhcp packet with invalid arch option (%v)\n", arch)
		return dhcp4.Packet{}
	}

	archId := int(arch[0]) << 8 | int(arch[1])

	mac := req.CHAddr()

	fmt.Printf("Identified mac address %v as architecture id %v --", mac, archId)

	var systemArchitecture SystemArchitecture

	// From https://www.ietf.org/assignments/dhcpv6-parameters/dhcpv6-parameters.xml#processor-architecture
	switch archId {
	case 0: systemArchitecture = X86_64 		// x86 with bios (32/64?)
	case 6 | 7: systemArchitecture = X86_64 	// x86 with uefi (32/64?)
	case 10: systemArchitecture = Unknown 		// Arm 32 bits with uefi (unknown because we dont support arm32 (yet))
	case 11: systemArchitecture = Arm64 		// Arm 64 bits with uefi

	default: systemArchitecture = Unknown
	}


	machine := Machine{
		MacAddress:   mac.String(),
		Architecture: systemArchitecture,
	}

	err := handler.machineStore.UpdateMachine(machine)
	if err != nil {
		log.Printf("An error occured: %v", err)
	}

	return dhcp4.Packet{}
}

func WatchArchitecturesDhcp(store MachineStore) {
	err := dhcp4.ListenAndServe(dhcpHandler {
		store,
	})

	if err != nil {
		println(err.Error())
	}
}
