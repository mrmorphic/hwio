// pinmap
//
// This just prints a map of pins for the device you're using. The maps shows the
// logical pin number, the name or names that the driver knows the pin by,
// and the set of capabilities that the driver supports for that pin.

package main

import (
	"hwio"
)

func main() {
	// select which driver you're using, which depends on what kind of board you're using.
	// Here, we are using a beaglebone. These will not be required at all once hwio can
	// determine the driver directly from the running environment.
	driver := new(hwio.BeagleBoneDriver)
	hwio.SetDriver(driver)

	hwio.DebugPinMap()
}

