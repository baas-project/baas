// Package api defines structures which are transferred over the network
package api

import (
	"encoding/json"
	"fmt"

	"github.com/baas-project/baas/pkg/model"
)

// Port is the port on which the control server listens
const Port int = 4848

// BootInformRequest is the data which the machine (client) sends to the control server on initial boot
type BootInformRequest struct {
}

// ReprovisioningInfo is to inform the management OS of the previous and next machine state and session
type ReprovisioningInfo struct {
	Prev model.MachineSetup
	Next model.MachineSetup
}

// PrettyPrintStruct prints a nice looking version of a struct
// TODO: Give this a better place in the code than randomly inside API
func PrettyPrintStruct(a interface{}) {
	// If I had a nickel for every time that the best way in a language to pretty print a datastructure is to cast it into a JSON
	// structure and printing that, I would have two nickels. That is not a lot, but it is funny that it happened twice.
	s, _ := json.MarshalIndent(a, "", "\t")
	fmt.Println(string(s))
}
