package main

// An example of shifting 8 bit data to a TLC5940 shift register.
// Implements a continuous 8-bit binary counter.

import (
	"fmt"

	"github.com/mrmorphic/hwio"
)

func main() {
	// Get the pins we're going to use. These are on a beaglebone.
	sinPin, _ := hwio.GetPin("P9.11")
	sclkPin, _ := hwio.GetPin("P9.12")
	xlatPin, _ := hwio.GetPin("P9.13")
	gsclkPin, _ := hwio.GetPin("P9.14")
	blankPin, _ := hwio.GetPin("P9.15")

	fmt.Printf("Pins are: sin=%d, sclk=%d, xlat=%d, gsclk=%d, blank=%d", sinPin, sclkPin, xlatPin, gsclkPin, blankPin)
	// Make them all outputs
	e := hwio.PinMode(sinPin, hwio.OUTPUT)
	if e == nil {
		hwio.PinMode(sclkPin, hwio.OUTPUT)
	}
	if e == nil {
		hwio.PinMode(xlatPin, hwio.OUTPUT)
	}
	if e == nil {
		hwio.PinMode(gsclkPin, hwio.OUTPUT)
	}
	if e == nil {
		hwio.PinMode(blankPin, hwio.OUTPUT)
	}
	if e != nil {
		fmt.Printf("Could not initialise pins: %s", e)
		return
	}

	// set clocks low
	hwio.DigitalWrite(sclkPin, hwio.LOW)
	hwio.DigitalWrite(xlatPin, hwio.LOW)
	hwio.DigitalWrite(gsclkPin, hwio.LOW)

	// run GS clock in it's own space
	hwio.DigitalWrite(blankPin, hwio.HIGH)
	hwio.DigitalWrite(blankPin, hwio.LOW)
	clockData(gsclkPin)

	for b := 0; b < 4096; b++ {
		writeData(uint(b), sinPin, sclkPin, xlatPin)

		for j := 0; j < 10; j++ {
			hwio.DigitalWrite(blankPin, hwio.HIGH)
			hwio.DigitalWrite(blankPin, hwio.LOW)
			clockData(gsclkPin)
		}

		//		hwio.Delay(100)
	}

	//		hwio.ShiftOut(dataPin, clockPin, uint(data), hwio.MSBFIRST)
}

// val is a 12-bit int
func writeData(val uint, sinPin hwio.Pin, sclkPin hwio.Pin, xlatPin hwio.Pin) {
	fmt.Printf("writing data %d\n", val)
	bits := 0
	mask := uint(1) << 11
	for i := 0; i < 16; i++ {
		v := val
		for j := 0; j < 12; j++ {
			if (v & mask) != 0 {
				hwio.DigitalWrite(sinPin, hwio.HIGH)
			} else {
				hwio.DigitalWrite(sinPin, hwio.HIGH)
			}
			hwio.DigitalWrite(sclkPin, hwio.HIGH)
			hwio.DigitalWrite(sclkPin, hwio.LOW)

			v = v << 1
			bits++
		}
	}

	hwio.DigitalWrite(xlatPin, hwio.HIGH)
	hwio.DigitalWrite(xlatPin, hwio.LOW)
	fmt.Printf("Wrote %d bits\n", bits)
}

func clockData(gsclkPin hwio.Pin) {
	for g := 0; g < 4096; g++ {
		hwio.DigitalWrite(gsclkPin, hwio.HIGH)
		hwio.DigitalWrite(gsclkPin, hwio.LOW)
	}
}
