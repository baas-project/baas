// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package api defines structures which are transferred over the network
package api

// Port is the port on which the control server listens
const Port int = 4848

// BootInformRequest is the data which the machine (client) sends to the control server on initial boot
type BootInformRequest struct {
}
