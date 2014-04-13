package main

// An example of shifting 8 bit data to a 74HC595 shift register.
// Implements a continuous 8-bit binary counter.

import (
	"github.com/mrmorphic/hwio"
)

func main() {
	// Get the pins we're going to use. These are on a beaglebone.
	dataPin, _ := hwio.GetPin("P8.3")  // connected to pin 14
	clockPin, _ := hwio.GetPin("P8.4") // connected to pin 11
	storePin, _ := hwio.GetPin("P8.5") // connected to pin 12

	// Make them all outputs
	hwio.PinMode(dataPin, hwio.OUTPUT)
	hwio.PinMode(clockPin, hwio.OUTPUT)
	hwio.PinMode(storePin, hwio.OUTPUT)

	data := 0

	// Set the initial state of the clock and store pins to low. These both
	// trigger on the rising edge.
	hwio.DigitalWrite(clockPin, hwio.LOW)
	hwio.DigitalWrite(storePin, hwio.LOW)

	for {
		// Shift out 8 bits
		hwio.ShiftOut(dataPin, clockPin, uint(data), hwio.MSBFIRST)

		// You can use this line instead to clock out 16 bits if you have
		// two 74HC595's chained together
		// hwio.ShiftOutSize(dataPin, clockPin, uint(data), hwio.MSBFIRST, 16)

		// Pulse the store pin once
		hwio.DigitalWrite(storePin, hwio.HIGH)
		hwio.DigitalWrite(storePin, hwio.LOW)

		// Poor humans cannot read binary as fast as the machine can display it
		hwio.Delay(100)

		data++
	}
}
