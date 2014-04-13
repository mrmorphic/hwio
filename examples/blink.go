// Blink
//
// Blinks an LED for one second on, one second off.
// This is heavily annotated, with error handling.

package main

import (
	"fmt"
	"os"

	"github.com/mrmorphic/hwio"
)

func main() {
	// get a pin by name. You could also just use the logical pin number, but this is
	// more readable. On BeagleBone, USR0 is an on-board LED.
	ledPin, err := hwio.GetPin("USR1")

	// Generally we wouldn't expect an error, but perhaps someone is running this a
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Set the mode of the pin to output. This will return an error if, for example,
	// we were trying to set an analog input to an output.
	err = hwio.PinMode(ledPin, hwio.OUTPUT)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Run the blink forever
	for {
		hwio.DigitalWrite(ledPin, hwio.HIGH)
		hwio.Delay(1000)
		hwio.DigitalWrite(ledPin, hwio.LOW)
		hwio.Delay(1000)
	}
}
