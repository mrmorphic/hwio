# Getting Started with BeagleBone and hwio

## The hardware

The BeagleBone has lots of hardware on it, but the thing we're most interested in when starting out are
the basics:

 * User LEDs on the board
 * Expansion headers P8 and P9

The expansion headers surface many of the GPIO (general purpose I/O) pins from the processor chip. Most of these pins can be programmed to perform other functions as well, although hwio is mostly concerned with their GPIO behaviour. hwio
takes care of the mode setting, so you can expect GPIO pins to just work the way digital I/O pins work on an Arduino.

There are also analog inputs on the expansion headers, and hwio makes it really easy to read these too.

NOTE: THE ANALOG INPUT PINS HAVE A MAXIMUM RATING OF 1.8 VOLTS. OTHER GPIO PINS HAVE A MAXIMUM RATING OF 3.3 VOLTS.
EXCEEDING THESE LIMITS WILL LIKELY DAMAGE YOUR BOARD.

A 1.8 volt reference voltage is available on the expansion header, but is low current, and should only be used as a reference.

## Writing to digital outputs

	// configure the pin as output
	pin, e := hwio.GetPin("gpio3_16")
	e = hwio.PinMode(pin, hwio.OUTPUT)

	// Operate the pin
	hwio.DigitalWrite(pin, hwio.HIGH)
	hwio.DigitalWrite(pin, hwio.LOW)

## Reading digital inputs

	pin, e := hwio.GetPin("gpio3_16")
	e = hwio.PinMode(pin, hwio.INPUT_PULLUP)

	val := hwio.DigitalRead(pin)
	if val == hwio.HIGH {
	}

BeagleBone GPIO inputs can be configured to have pull-up or pull-down resistors enabled, or none.
The modes are:

	hwio.INPUT (no resistor)
	hwio.INPUT_PULLUP
	hwio.INPUT_PULLDOWN

## Reading analog inputs

_to be written_

## Map of pins

_to be written_
