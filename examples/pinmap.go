// pinmap
//
// This just prints a map of pins for the device you're using. The maps shows the
// logical pin number, the name or names that the driver knows the pin by,
// and the set of capabilities that the driver supports for that pin.

package main

import (
	"github.com/mrmorphic/hwio"
)

func main() {
	hwio.DebugPinMap()
}
