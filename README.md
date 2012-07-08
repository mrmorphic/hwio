# hwio

## Introduction

hwio is a library for interfacing with hardware I/O. It is loosely modelled on
the Arduino programming style, but deviating where that doesn't make sense in
go. It makes use of a thin hardware abstraction via an interface so a program
written against the library for say a beaglebone could be easily compiled to run
on a raspberry pi (big caveat: someone needs to write that driver ;-), maybe
only changing logical pin numbers.

Hardware drivers implement the interface that allows different devices to
implement the I/O handling as appropriate. This allows for drivers that use
different techniques to access the I/O. e.g. some drivers may use GPIO (a file
system approach), whereas for maximum performance another driver might use
direct memory access to I/O registers, but is less portable.

Some general principles the library attempts to adhere to include:

 *	Pin references are logical, and are mapped to hardware pins by the driver.
 *	Not implement Arduino functions for their own sake if go's framework
 	naturally supports them better, unless we can provide a simpler interface
 	to those functions and keep close to the Arduino semantics.
 *	Make no assumption about the state of a pin whose mode has not been set.
 	Specifically, pins that don't have mode set may not on a particular hardware
 	configuration even be configured as general purpose I/O. For example, many
 	beaglebone pins have overloaded functions set using a multiplexer. Any pin
 	whose mode is set by PinMode can be assumed to be general purpose I/O, and
 	likewise if it is not set, it could have any multiplexed behaviour assigned
 	to it. A consequence is that unlike Arduino, PinMode *must* be called before
 	a pin is used.
 *	The library should be as fast as possible so that applications that require
 	very high speed I/O should achieve maximal throughput, given an appropriate
 	driver.
 *	The library should be tolerant towards beginners, and give meaningful errors
 	when the hardware is requested to do things it can't. But this checking
 	should be able to be disabled for maximum performance.
 *	Make simple stuff simple, and harder stuff possible. In particular, while
 	Arduino-like methods have uniform interface and semantics across drivers,
 	we don't hide the driver itself, so special features of a driver can still
 	be used, albeit non-portably.
 *	Sub packages can be added as required that approximately parallel Arduino
 	libraries (e.g. perhaps an SD card package). Where possible, these
 	implementations should be generic across I/O pins. e.g. not assuming
 	specific pin behaviour as happens with Arduino.

## BIG SHINY DISCLAIMER

REALLY IMPORTANT THINGS TO KNOW ABOUT THIS ABOUT THIS LIBRARY:

 *	It is under development. If you're lucky, it might work.
 *	Currently only BeagleBone is supported
 *	It is not properly tested.
 *	IT MAY FRY YOUR BOARD
 *	IF YOU CHANGE IT, OR LOOK AT IT THE WRONG WAY, IT MAY FRY YOUR BOARD
 *	I DON'T WANT PEOPLE GETTING ANGRY WITH ME IF THIS CODE FRIES THEIR BOARD.
 	I GET TO DEAL WITH ENOUGH SHIT EVERY DAY, I DON'T NEED MORE.
 *	IT MAY FRY YOUR BOARD
 *	If you don't want to risk frying your board, you can still run the
 	unit tests ;-)

## What you need to know to use hwio

### Pins

Pins are logical representation of physical pins on the hardware. To provide
some abstraction, pins are numbered, much like on an Arduino. Unlike Arduino,
there is no single mapping to hardware pins - this is done by the hardware
driver.

Each driver must implement a method that defines the mapping from logical pins
to physical pins as understood by that piece of hardware. Additionally, the
driver also publishes the capabilities of each pin, so that hwio can ensure
that constraints of the hardware are met. For example, if a pin implements PWM
in hardware, that is a capability. A method (e.g. motor driver) that requires
hardware PWM can ensure that a pin has the appropriate capability. Because each
pin can have a different set of capabilities, there is no distinction between
analog and digital pins as there is in Arduino; there is one set of pins, which
may support digital and/or analog capabilities.

The caller generally works with logic pin numbers. The hardware driver exposes
the hardware-specific names, however.

### Prerequisites

### Installation

__I'll do this when it is in a more stable state.__

## Driver Specific

### BeagleBone

