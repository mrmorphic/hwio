/*
	Package hwio implements a simple Arduino-like interface for controlling
	hardware I/O, with configurable backends depending on the device.
*/
package hwio

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type BitShiftOrder byte

const (
	LSBFIRST BitShiftOrder = iota
	MSBFIRST
)

// Reference to driver we're using
var driver HardwareDriver

// Retrieved from the driver, this is the map of the hardware pins supported by
// the driver and their capabilities
var definedPins HardwarePinMap

// A private type for associating a pin's definition with the current IO mode
// and any other dynamic properties of the pin.
type assignedPin struct {
	pinDef    *PinDef   // definition of pin
	pinIOMode PinIOMode // mode that was assigned to this pin
}

// A map of pin numbers to the assigned dynamic properties of the pin. This is
// set by PinMode when errorChecking is on, and can be used by other functions
// to determine if the request is valid given the assigned properties of the pin.
var assignedPins map[Pin]*assignedPin

// If set to true, functions should test that their constraints are met.
// e.g. test that the pin is capable of doing what is asked. This can be set
// with SetErrorChecking(). Setting to false bypasses checks for performance.
// By default turned on, which is a better default for beginners.
var errorChecking bool = true

// init() attempts to determine from the environment what the driver is. The
// intent is that the consumer of the library would not generally have to worry
// about it, it would just work. If it cannot determine the driver, it doesn't
// set the driver to anything.
func init() {
	determineDriver()
}

// Work out the driver from environment if we can. If we have any problems,
// don't generate an error, just return with the driver not set.
// @todo use reflection to determine all implementors of the driver interface, and
// @todo   call a method on the interface to self-detect. init and 
// @todo   constructor of drivers should do no setup in this case, esp of hardware
func determineDriver() {
	uname, e := exec.Command("uname", "-a").Output()
	if e != nil {
		return
	}

	s := string(uname)
	if strings.Contains(s, "beaglebone") {
		SetDriver(new(BeagleBoneDriver))
	} else if strings.Contains(s, "raspberrypi") || strings.Contains(s, "adafruit") {
		SetDriver(new(RaspberryPiDriver))
	}
}

// Check if the driver is assigned. If not, return an error to indicate that,
// otherwise return no error.
func assertDriver() error {
	if driver == nil {
		return errors.New("hwio has no configured driver")
	}
	return nil
}

// Set the driver. Also calls Init on the driver, and loads the capabilities
// of the device.
func SetDriver(d HardwareDriver) {
	driver = d
	e := driver.Init()
	if e != nil {
		fmt.Printf("Could not initialise driver: %s", e)
	}
	definedPins = driver.PinMap()
	assignedPins = make(map[Pin]*assignedPin)
}

// Retrieve the current hardware driver.
func GetDriver() HardwareDriver {
	return driver
}

// Returns a map of the hardware pins. This will only work once the driver is
// set.
func GetDefinedPins() HardwarePinMap {
	return definedPins
}

// Returns a Pin given a canonical name for the pin.
// e.g. to get the pin number of P8.13 on a beaglebone,
//     pin := hwio.GetPin("P8.13")
// Order of search is:
// - search hwRefs in the pin map in order.
// This function should not generally be relied on for performance. For max speed, call this
// for each pin you use once on init, and use the returned Pin values thereafter.
// Search is case sensitive at the moment
// @todo GetPin: consider making it case-insensitive on name
// @todo GetPin: consider allowing an int or int as string to identify logical pin directly
func GetPin(cname string) (Pin, error) {
	for pin, pinDef := range definedPins {
		for _, name := range pinDef.hwPinRefs {
			if name == cname {
				return pin, nil
			}
		}
	}
	return Pin(0), errors.New(fmt.Sprintf("Could not find a pin called %s", cname))
}

// Set error checking. This should be called before pin assignments.
func SetErrorChecking(check bool) {
	errorChecking = check
}

// Set the mode of a pin. Analogous to Arduino pin mode.
func PinMode(pin Pin, mode PinIOMode) (e error) {
	if errorChecking {
		if e = assertDriver(); e != nil {
			return
		}

		pd := definedPins[pin]
		if pd == nil {
			return errors.New(fmt.Sprintf("Pin %d is not defined by the current driver", pin))
		}

		if e = checkPinMode(mode, pd); e != nil {
			return
		}

		// assign this pin
		assignedPins[pin] = &assignedPin{pinDef: pd, pinIOMode: mode}
	}

	return driver.PinMode(pin, mode)
}

func checkPinMode(mode PinIOMode, pd *PinDef) (e error) {
	ok := false
	switch mode {
	case INPUT:
		ok = pd.HasCapability(CAP_INPUT) || pd.HasCapability(CAP_ANALOG_IN)
	case OUTPUT:
		ok = pd.HasCapability(CAP_OUTPUT)
	case INPUT_PULLUP:
		ok = pd.HasCapability(CAP_INPUT_PULLUP)
	case INPUT_PULLDOWN:
		ok = pd.HasCapability(CAP_INPUT_PULLDOWN)
	}
	if ok {
		return nil
	}
	return errors.New(fmt.Sprintf("Pin %d can't be set to mode %s because it does not support that capability", pd.pin, mode.String()))
}

// Write a value to a digital pin
func DigitalWrite(pin Pin, value int) (e error) {
	if errorChecking {
		if e = assertDriver(); e != nil {
			return
		}

		a := assignedPins[pin]
		if a == nil {
			return errors.New(fmt.Sprintf("DigitalWrite: pin %d mode has not been set", pin))
		}
		if a.pinIOMode != OUTPUT {
			return errors.New(fmt.Sprintf("DigitalWrite: pin %d mode is not set for output", pin))
		}
	}

	return driver.DigitalWrite(pin, value)
}

// Read a value from a digital pin
func DigitalRead(pin Pin) (result int, e error) {
	if errorChecking {
		if e = assertDriver(); e != nil {
			return 0, e
		}

		a := assignedPins[pin]
		if a == nil {
			e = errors.New(fmt.Sprintf("DigitalRead: pin %d mode has not been set", pin))
			return
		}
		if a.pinIOMode != INPUT && a.pinIOMode != INPUT_PULLUP && a.pinIOMode != INPUT_PULLDOWN {
			e = errors.New(fmt.Sprintf("DigitalRead: pin %d mode is not set for input", pin))
			return
		}
	}

	return driver.DigitalRead(pin)
}

// Read an analog value from a pin. The range of values is hardware driver dependent.
func AnalogRead(pin Pin) (result int, e error) {
	if errorChecking {
		if e = assertDriver(); e != nil {
			return
		}

		a := assignedPins[pin]
		if a == nil {
			e = errors.New(fmt.Sprintf("AnalogRead: pin %d mode has not been set", pin))
			return
		}
		if a.pinIOMode != INPUT && a.pinIOMode != INPUT_PULLUP {
			e = errors.New(fmt.Sprintf("AnalogRead: pin %d mode is not set for input", pin))
			return
		}
	}

	return driver.AnalogRead(pin)
}

// Write an analog value. The interpretation is hardware dependent, but is
// generally implemented using PWM.
func AnalogWrite(pin Pin, value int) (e error) {
	if errorChecking {
		if e = assertDriver(); e != nil {
			return
		}

		a := assignedPins[pin]
		if a == nil {
			return errors.New(fmt.Sprintf("AnalogWrite: pin %d mode has not been set", pin))
		}
		if a.pinIOMode != OUTPUT {
			return errors.New(fmt.Sprintf("AnalogWrite: pin %d mode is not set for output", pin))
		}
	}

	return driver.AnalogWrite(pin, value)
}

// Delay execution by the specified number of milliseconds. This is a helper
// function for similarity with Arduino. It is implemented using standard go
// time package.
func Delay(duration int) {
	time.Sleep(time.Duration(duration) * time.Millisecond)
}

// Delay execution by the specified number of microseconds. This is a helper
// function for similarity with Arduino. It is implemented using standard go
// time package
func DelayMicroseconds(duration int) {
	time.Sleep(time.Duration(duration) * time.Microsecond)
}

// @todo DebugPinMap: sort
func DebugPinMap() {
	fmt.Println("HardwarePinMap:")
	for key, val := range definedPins {
		fmt.Printf("Pin %d: %s\n", key, val.String())
	}
	fmt.Printf("\n")
}

// The approximate mapping of Arduino shiftOut, this shifts a byte out on the
// data pin, pulsing the clock pin high and then low.
func ShiftOut(dataPin Pin, clockPin Pin, value uint, order BitShiftOrder) error {
	return ShiftOutSize(dataPin, clockPin, value, order, 8)
}

// More generic version of ShiftOut which shifts out n of data from value. The
// value shifted out is always the lowest n bits of the value, but 'order'
// determines whether the msb or lsb from that value are shifted first
func ShiftOutSize(dataPin Pin, clockPin Pin, value uint, order BitShiftOrder, n uint) error {
	bit := uint(0)
	v := value
	mask := uint(1) << (n - 1)
	for i := uint(0); i < n; i++ {
		// get the next bit
		if order == LSBFIRST {
			bit = v & 1
			v = v >> 1
		} else {
			bit = v & mask
			if bit != 0 {
				bit = 1
			}
			v = v << 1
		}
		// write to data pin
		e := DigitalWrite(dataPin, int(bit))
		if e != nil {
			return e
		}
		// pulse clock high and then low
		e = DigitalWrite(clockPin, HIGH)
		if e != nil {
			return e
		}
		DigitalWrite(clockPin, LOW)
	}
	return nil
}

// def toggle(gpio_pin):
//   """ Toggles the state of the given digital pin. """
//   assert (gpio_pin in GPIO), "*Invalid GPIO pin: '%s'" % gpio_pin
//   _xorReg(GPIO[gpio_pin][0]+GPIO_DATAOUT, GPIO[gpio_pin][1])

// def pinState(gpio_pin):
//   """ Returns the state of a digital pin if it is configured as
//       an output. Returns None if it is configuredas an input. """
//   assert (gpio_pin in GPIO), "*Invalid GPIO pin: '%s'" % gpio_pin
//   if (_getReg(GPIO[gpio_pin][0]+GPIO_OE) & GPIO[gpio_pin][1]):
//     return None
//   if (_getReg(GPIO[gpio_pin][0]+GPIO_DATAOUT) & GPIO[gpio_pin][1]):
//     return HIGH
//   return LOW

// @todo Implement other core Arduino function equivalents:
//	AnalogReference
//	Tone
//	NoTone
//	ShiftOut
//	ShiftIn
//	PulseIn
//	Millis
//	Micros
//	RandomSeed
//	Random
//	AttachInterupt
//	DetachInterupt

// This is the interface that hardware drivers implement.
type HardwareDriver interface {
	// Initialise the driver after creation
	Init() (e error)

	// Set mode of a pin
	PinMode(pin Pin, mode PinIOMode) (e error)

	// Write digital output
	DigitalWrite(pin Pin, value int) error

	// Read digital input
	DigitalRead(Pin) (int, error)

	// PWM write
	AnalogWrite(pin Pin, value int) error

	// Analog input. Resolution is device dependent.
	AnalogRead(pin Pin) (int, error)

	// Return the pin map for the driver, listing all supported pins and their capabilities
	PinMap() (pinMap HardwarePinMap)

	// Close the driver before destruction
	Close()
}
