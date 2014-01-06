# hwio

## Introduction

hwio is a Go library for interfacing with hardware I/O, particularly on ARM SoC-based boards.
It is loosely modelled on the Arduino programming style, but deviating where that doesn't make sense in
Go. It makes use of a thin hardware abstraction via an interface so a program written against the library
for say a beaglebone could be easily compiled to run on a raspberry pi, maybe only changing pin references.

To use hwio, you just need to import it into your Go project, initialise modules and pins as
required, and then use functions that manipulate the pins.

## Basic Usage

Initialising a pin looks like this:

	myPin, err := hwio.GetPin("gpio4")
	err = hwio.PinMode(myPin, hwio.OUTPUT)

Or the shorter, more convenient form:

	myPin, err := GetPinWithMode("gpio4", hwio.OUTPUT)

Unlike Arduino, where the pins are directly numbered and you just use the number, in hwio
you get the pin first, by name. This is necessary as different hardware drivers may provide
different pins.

The mode constants include:

 *  INPUT - set pin to digital input
 *  OUTPUT - set pin to digital output

(Pull-ups and pull-downs are not currently supported by the drivers, as this is not apparently exposed to file system.)

Writing a value to a pin looks like this:

	hwio.DigitalWrite(myPin, hwio.HIGH)

Reading a value from a digital pin looks like this, returning a HIGH or LOW:

	value, err := hwio.DigitalRead(myPin)

Reading an analog value looks like this:

	value, err := hwio.Analogread(somePin)

Analog values (on BeagleBone Black at least) are integers typically between 0-1800, which is the number of millivolts. 
(Note that you cannot drive analog inputs more than 1.8 volts on the BeagleBone, and you should use the analog voltage
references it provides)

## Utility Functions

To delay a number of milliseconds:

	hwio.Delay(500)  // delay 500ms

Or to delay by microseconds:

	hwio.DelayMicroseconds(1500)  // delay 1500 usec, or 1.5 milliseconds

The Arduino ShiftOut function is supported in a simplified form for 8 bits:

	e := hwio.ShiftOut(dataPin, clockPin, 127, hwio.MSBFIRST)   // write 8 bits, MSB first

or in a bigger variant that supports different sizes:

	e := hwio.ShiftOutSize(dataPin, clockPin, someValue, hwio.LSBFIRST, 12)  // write 12 bits LSB first

Sometimes you might want to write an unsigned int to a set of digital pins (e.g. a parallel port). This can be done as
follows:

	somePins := []hwio.Pin{myPin3, myPin2, myPin1, myPin0}
	e := hwio.WriteUIntToPins(myValue, somePins)

This will write out the n lowest bits of myValue, with the most significant bit of that value written to myPin3 etc. It uses DigitalWrite
so the outputs are not written instantaneously.

There is an implementation of the Arduino map() function:

	// map a value in range 0-1800 to new range 0-1023
	i := hwio.Map(value, 0, 1800, 0, 1023)

## I2C

I2C is supported on BeagleBone Black and Raspberry Pi. It is accessible through the "i2c" module (BBB i2c2 pins), as follows:

	m, e := hwio.GetModule("i2c")
	if e != nil {
		fmt.Printf("could not get i2c module: %s\n", e)
		return
	}
	i2c := m.(hwio.I2CModule)

	// Uncomment on Raspberry pi, which doesn't automatically enable i2c bus. BeagleBone does, as the default device tree enables it.
	// i2c.Enable()
	// defer i2c.Disable()

	device := i2c.GetDevice(0x68)

Once you have a device, you can use Write, WriteBytes, Read or ReadBytes to set or get data from the i2c device.

e.g.

	device.WriteByte(controlRegister, someValue)

While you can use the i2c types to directly talk to i2c devices, the specific device may already have higher-level support in the
hwio/devices package, so check there first, as the hard work may be done already.

## PWM

PWM support for BeagleBone Black has been added. To use a PWM pin, you need to fetch the module that the PWM belongs to, enable the PWM module
and pin, and then you can manipulate the period and duty cycle. e.g.

	// Get the module
	m, e := hwio.GetModule("pwm2")
	if e != nil {
		fmt.Printf("could not get pwm2 module: %s\n", e)
		return
	}

	pwm := m.(hwio.PWMModule)

	// Enable it.
	pwm.Enable()

	// Get the PWM pin
	pwm8_13, _ := hwio.GetPin("P8.13")
	e = pwm.EnablePin(pwm8_13, true)
	if e != nil {
		fmt.Printf("Error enabling pin: %s\n", e)
		return
	}

	// Set the period and duty cycle, in nanoseconds. This is a 1/10th second cycle
	pwm.SetPeriod(pwm8_13, 100000000)
	pwm.SetDuty(pwm8_13, 90000000)

On BeagleBone Black, there are 3 PWM modules, "pwm0", "pwm1" and "pwm2". I am not sure if "pwm1" pins can be assigned, as they are pre-allocaed in the default device tree config, but in theory it should be possible to use them. By default, these pins can be used:

  * pwm0: P9.21 (ehrpwm0B) and P9.22 (ehrpwm0A)
  *	pwm2: P8.13 (ehrpwm2A) and P8.19 (ehrpwm2A)

This is a preliminary implementation; only P8.13 (pwm2) has been tested. PWM pins are not present in default device tree. The module will add them as necessary to bonemgr/slots; this will override defaults.

## Servo

There is a servo implementation in the hwio/servo package. See README.md in that package.

## Devices

There are sub-packages under 'devices' that have been made to work with hwio. The currently supported devices include:

  *	GY-520 gyroscope/accelerometer using I2C.
  * HD-44780 multi-line LCD display. Currently implemented over I2C converter only.

See README.md files in respective directories.


## Driver Selection

The intention of the hwio library is to use uname to attempt to detect the platform and select an appropriate driver (see drivers section below), 
so for some platforms this may auto-detect. However, with the variety of boards around and the variety of operating systems, you may find that autodetection
doesn't work. If you need to set the driver automatically, you can do:

	hwio.SetDriver(new(BeagleBoneBlackDriver))

This needs to be done before any other hwio calls.


## More Information

For more information about the library, including pin diagrams for supported boards, see http://stuffwemade.net/hwio.


## BIG SHINY DISCLAIMER

REALLY IMPORTANT THINGS TO KNOW ABOUT THIS ABOUT THIS LIBRARY:

 *	It is under development. If you're lucky, it might work. It should be considered
	Alpha.
 *	If you don't want to risk frying your board, you can still run the
 	unit tests ;-)


## Board Support

Currently there are 3 drivers:

  *	BeagleBoneBlackDriver - for BeagleBone boards running linux kernel 3.7 or higher, including
  	BeagleBone Black. This is untested on older BeagleBone boards with updated kernels.
  * RaspberryPiDTDriver - for Raspberry Pi modules running linux kernel 3.7 or higher, which includes newer Raspian
    kernels and some late Occidental kernels.
  * TestDriver - for unit tests.

Old pre-kernel-3.7 drivers for BeagleBone and Raspberry Pi have been deprecated as I have no test beds for these. If you want
to use these, you can check out the 'legacy' branch that contains the older drivers, but no new features will be added.


### BeagleBoneBlackDriver

This driver accesses hardware via the device interfaces exposed in the file system on linux kernels 3.8 or higher, where
device tree is mandated. This should be a robust driver as the hardware access is maintained by device driver authors,
but is likely to be not as fast as direct memory I/O to the hardware as there is file system overhead.

Status:

  * In active development.
  * Tested for gpio reads and writes, analog reads and i2c. Test device was BeagleBone Black running rev A5C, running angstrom.
  * Driver automatically blocks out the GPIO pins that are allocated to LCD and MMC on the default BeagleBone Black boards.
  * GPIOs not assigned at boot to other modules are known to read and write.
  * PWM is known to work on erhpwm2A and B ports.
  * USR0-USR3 don't work and cannot be accessed as GPIO, as the LED driver reserves them.
  * GPIO pull-ups is not yet supported.
  * i2c is enabled by default.

### RaspberryPiDTDriver

This driver is very similar to the BeagleBone Black driver in that it uses the modules compiled into the kernel and
configured using device tree. It uses the same GPIO and i2c implementatons, just with different pins.

Current status:

 *	DigitalRead and DigitalWrite (GPIO) have been tested and work correctly on supported GPIO pins. Test platform was
 	Raspberry Pi (revision 1), Raspian 2013-12-20-wheezy-raspbian, kernel 3.10.24+.
 *	GPIO pins are gpio4, gpio17, gpio18, gpio21, gpio22, gpio23, gpio24 and gpio25.
 *	I2C is working on raspian. You need to enable it on the board first.
 	Follow [these instructions](http://www.abelectronics.co.uk/i2c-raspbian-wheezy/info.aspx "i2c and spi support on raspian")

GetPin references on this driver return the pin numbers that are on the headers. Pin 0 is unimplemented.

Note: before using this, check your kernel is 3.7 or higher. There are a number of pre-3.7 distributions still in use, and this driver
does not support pre-3.7.

## Implementation Notes

Some general principles the library attempts to adhere to include:

 *	Pin references are logical, and are mapped to hardware pins by the driver. The pin
    numbers you get back from GetPin are, unless otherwise specified, related to the pin numbers
    on extension headers.
 *	Drivers provide pin names, so you can look them up by meaningful names
	instead of relying on device specific numbers. On boards such as BeagleBone, where pins
	are multiplexed to internal functions, pins can have multiple names.
 *	The library does not implement Arduino functions for their own sake if go's
	framework naturally supports them better, unless we can provide a simpler interface
 	to those functions and keep close to the Arduino semantics.
 *	Drivers are very thin layers; most of the I/O functionality is provided by **modules**.
    These aim to be as generic as possible so that different drivers on similar kernels can
    assemble the modules that are enabled in device tree, with appropriate pin configuration.
    This also makes it easier to add in new modules to support various SoC functions.
 *	Make no assumption about the state of a pin whose mode has not been set.
 	Specifically, pins that don't have mode set may not on a particular hardware
 	configuration even be configured as general purpose I/O. For example, many
 	beaglebone pins have overloaded functions set using a multiplexer, and some are be pre-assigned
 	by the default device tree configuration.
 *	Any pin whose mode is set by PinMode can be assumed to be general purpose I/O, and
 	likewise if it is not set, it could have any multiplexed behaviour assigned
 	to it. A consequence is that unlike Arduino, PinMode *must* be called before
 	a pin is used.
 *	The library should be as fast as possible so that applications that require
 	very high speed I/O should achieve maximal throughput, given an appropriate
 	driver.
 *	Make simple stuff simple, and harder stuff possible. In particular, while
 	Arduino-like methods have uniform interface and semantics across drivers,
 	we don't hide the driver itself or the modules it uses, so special features of a driver or module
 	can still be used, albeit non-portably.
 *	Sub-packages can be added as required that approximately parallel Arduino
 	libaries (e.g. perhaps an SD card package).


### Pins

Pins are logical representation of physical pins on the hardware. To provide
some abstraction, pins are numbered, much like on an Arduino. Unlike Arduino,
there is no single mapping to hardware pins - this is done by the hardware
driver. To make it easier to work with, drivers can give one or more names to
a pin, and you can use GetPin to get a reference to the pin by one of those
names.

Each driver must implement a method that defines the mapping from logical pins
to physical pins as understood by that piece of hardware. Additionally, the
driver also publishes the modules that the hardware configuration supports, so
that hwio can ensure that constraints of the hardware are met. For example, if a
pin implements PWM and GPIO in hardware, it is associated with two modules. When
the PWM module is enabled, it will assign the pin to itself. Because each
pin can have a different set of capabilities, there is no distinction between
analog and digital pins as there is in Arduino; there is one set of pins, which
may support any number of capabilities including digital and analog.

The caller generally works with logical pin numbers retrieved by GetPin.


## Things to be done

 *	Interupts (lib, BeagleBone and R-Pi)
 *	Serial support for UART pins (lib, BeagleBone and R-Pi)
 *	SPI support; consider augmenting ShiftIn and ShiftOut to use hardware pins
 	if appropriate (Beaglebone and R-Pi)
 *	Stepper (lib)
 *	TLC5940 (lib)
