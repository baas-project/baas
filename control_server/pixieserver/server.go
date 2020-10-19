package pixieserver

import (
	"go.universe.tf/netboot/out/ipxe"
	"go.universe.tf/netboot/pixiecore"
	"log"
	"time"
)

func StartPixiecore(url string) {
	Ipxe := map[pixiecore.Firmware][]byte{}

	Ipxe[pixiecore.FirmwareX86PC] = ipxe.MustAsset("third_party/ipxe/src/bin/undionly.kpxe")
	Ipxe[pixiecore.FirmwareEFI32] = ipxe.MustAsset("third_party/ipxe/src/bin-i386-efi/ipxe.efi")
	Ipxe[pixiecore.FirmwareEFI64] = ipxe.MustAsset("third_party/ipxe/src/bin-x86_64-efi/ipxe.efi")
	Ipxe[pixiecore.FirmwareEFIBC] = ipxe.MustAsset("third_party/ipxe/src/bin-x86_64-efi/ipxe.efi")
	Ipxe[pixiecore.FirmwareX86Ipxe] = ipxe.MustAsset("third_party/ipxe/src/bin/ipxe.pxe")

	b, err := pixiecore.APIBooter(url, time.Second*5)
	if err != nil {
		log.Fatalf("Couldn't create booter %v", err)
	}

	s := pixiecore.Server{
		Booter:     b,
		Ipxe:       Ipxe,
		Log:        logWithStdLog,
		DHCPNoBind: true,
		Address:    "0.0.0.0",
	}

	err = s.Serve()
	if err != nil {
		log.Fatalf("Error while serving: %v", err)
	}
}
