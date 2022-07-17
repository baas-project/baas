// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package pixieserver starts pixiecore. Usually pixiecore runs as a separate application,
// but we just import and run it internally.
package pixieserver

import (
	"time"

	log "github.com/sirupsen/logrus"

	"go.universe.tf/netboot/out/ipxe"
	"go.universe.tf/netboot/pixiecore"
)

// StartPixiecore starts the pixiecore server(s) (dhcp, tftp and http).
func StartPixiecore(url string) {
	// This is essentially the same as what pixiecore does when ran as a command line application.
	Ipxe := map[pixiecore.Firmware][]byte{}

	Ipxe[pixiecore.FirmwareX86PC] = ipxe.MustAsset("third_party/ipxe/src/bin/undionly.kpxe")
	Ipxe[pixiecore.FirmwareEFI32] = ipxe.MustAsset("third_party/ipxe/src/bin-i386-efi/ipxe.efi")
	Ipxe[pixiecore.FirmwareEFI64] = ipxe.MustAsset("third_party/ipxe/src/bin-x86_64-efi/ipxe.efi")
	Ipxe[pixiecore.FirmwareEFIBC] = ipxe.MustAsset("third_party/ipxe/src/bin-x86_64-efi/ipxe.efi")
	Ipxe[pixiecore.FirmwareX86Ipxe] = ipxe.MustAsset("third_party/ipxe/src/bin/ipxe.pxe")

	log.Info("Starting pixiecore")

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
