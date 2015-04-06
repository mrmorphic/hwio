/*
	Package hwio implements a simple Arduino-like interface for controlling
	hardware I/O, with configurable backends depending on the device.
*/
package hwio

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
	pin    Pin    // pin being assigned
	module Module // module that has assigned this pin
	// pinIOMode PinIOMode // mode that was assigned to this pin
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
	assignedPins = make(map[Pin]*assignedPin)
	determineDriver()
}

func fileExists(name string) bool {
	_, err := os.Stat(name)
	if err != nil {
		return false
	}

	return true
}

// Work out the driver from environment if we can. If we have any problems,
// don't generate an error, just return with the driver not set.
func determineDriver() {
	drivers := [...]HardwareDriver{NewBeagleboneBlackDTDriver(), NewRaspPiDTDriver(), NewOdroidC1Driver()}
	for _, d := range drivers {
		if d.MatchesHardwareConfig() {
			SetDriver(d)
			return
		}
	}

	fmt.Printf("Unable to select a suitable driver for this board.\n")
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

// Ensure that any resources external to the program that have been allocated are tidied up.
func CloseAll() {
	if driver == nil {
		return
	}
	driver.Close()
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
func GetPin(pinName string) (Pin, error) {
	pl := strings.ToLower(pinName)
	for pin, pinDef := range definedPins {
		for _, name := range pinDef.names {
			if strings.ToLower(name) == pl {
				return pin, nil
			}
		}
	}

	return Pin(0), fmt.Errorf("Could not find a pin called %s", pinName)
}

// Shortcut for calling GetPin and then PinMode.
func GetPinWithMode(cname string, mode PinIOMode) (pin Pin, e error) {
	p, e := GetPin(cname)
	if e != nil {
		return
	}

	e = PinMode(p, mode)
	return p, e
}

// Set error checking. This should be called before pin assignments.
func SetErrorChecking(check bool) {
	errorChecking = check
}

// Helper function to get GPIO module
func GetGPIOModule() (GPIOModule, error) {
	m, e := GetModule("gpio")
	if e != nil {
		return nil, e
	}

	if m == nil {
		return nil, errors.New("Driver does not support GPIO")
	}

	return m.(GPIOModule), nil
}

// Given an internal pin number, return the canonical name for the pin, as defined by the driver. If the pin
// is not to the driver, return "".
func PinName(pin Pin) string {
	p := definedPins[pin]
	if p == nil {
		return ""
	}
	return p.names[0]
}

// Set the mode of a pin. Analogous to Arduino pin mode.
func PinMode(pin Pin, mode PinIOMode) error {
	gpio, e := GetGPIOModule()
	if e != nil {
		return e
	}

	return gpio.PinMode(pin, mode)
}

// Close a specific pin that has been assigned as GPIO by PinMode
func ClosePin(pin Pin) error {
	gpio, e := GetGPIOModule()
	if e != nil {
		return e
	}

	return gpio.ClosePin(pin)
}

// Assign a pin to a module. This is typically called by modules when they allocate pins. If the pin is already assigned,
// an error is generated. ethod is public in case it is needed to hack around default driver settings.
func AssignPin(pin Pin, module Module) error {
	if a := assignedPins[pin]; a != nil {
		return fmt.Errorf("Pin %d is already assigned to module %s", pin, a.module.GetName())
	}
	assignedPins[pin] = &assignedPin{pin, module}
	return nil
}

// Assign a set of pins. Method is public in case it is needed to hack around default driver settings.
func AssignPins(pins PinList, module Module) error {
	for _, pin := range pins {
		e := AssignPin(pin, module)
		if e != nil {
			return e
		}
	}
	return nil
}

// Unassign a pin. Method is public in case it is needed to hack around default driver settings.
func UnassignPin(pin Pin) error {
	delete(assignedPins, pin)
	return nil
}

// Unassign a set of pins. Method is public in case it is needed to hack around default driver settings.
func UnassignPins(pins PinList) (er error) {
	er = nil

	for _, pin := range pins {
		e := UnassignPin(pin)
		if e != nil {
			er = e
		}
	}

	return
}

// Write a value to a digital pin
func DigitalWrite(pin Pin, value int) (e error) {
	gpio, e := GetGPIOModule()
	if e != nil {
		return e
	}

	return gpio.DigitalWrite(pin, value)
}

// Read a value from a digital pin
func DigitalRead(pin Pin) (result int, e error) {
	// @todo consider memoizing
	gpio, e := GetGPIOModule()
	if e != nil {
		return 0, e
	}

	return gpio.DigitalRead(pin)
}

// given a logic level of HIGH or LOW, return the opposite. Invalid values returned as LOW.
func Negate(logicLevel int) int {
	if logicLevel == LOW {
		return HIGH
	}
	return LOW
}

// Helper function to pulse a pin, which must have been set as GPIO.
// 'active' is LOW or HIGH. Pulse sets pin to inactive, then active for
// 'durationMicroseconds' and the back to inactive.
func Pulse(pin Pin, active int, durationMicroseconds int) error {
	// set to inactive state, in case it wasn't already
	e := DigitalWrite(pin, Negate(active))
	if e != nil {
		return e
	}

	// set to active state
	e = DigitalWrite(pin, active)
	if e != nil {
		return e
	}

	DelayMicroseconds(durationMicroseconds)

	// finally reset to inactive state
	return DigitalWrite(pin, Negate(active))
}

// Helper function to get GPIO module
func GetAnalogModule() (AnalogModule, error) {
	m, e := GetModule("analog")
	if e != nil {
		return nil, e
	}

	if m == nil {
		return nil, errors.New("Driver does not support analog")
	}

	return m.(AnalogModule), nil
}

// Read an analog value from a pin. The range of values is hardware driver dependent.
func AnalogRead(pin Pin) (int, error) {
	analog, e := GetAnalogModule()
	if e != nil {
		return 0, e
	}

	return analog.AnalogRead(pin)
}

// Helper to turn an on-board LED on or off. Uses LED module
func Led(name string, on bool) error {
	m, e := GetModule("leds")
	if e != nil {
		return e
	}

	leds := m.(LEDModule)
	led, e := leds.GetLED(name)
	if e != nil {
		return e
	}

	e = led.SetTrigger("none")
	if e != nil {
		return e
	}

	return led.SetOn(on)
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

// Given an integer and a list of GPIO pins (that must have been set up as outputs), write the integer across
// the pins. The number of bits is determined by the length of the pins. The most-significant output pin is first.
// Bits are written MSB first.
// Maximum number of bits that can be shifted is 32.
// Note that the bits are not written out instantaneously, although very quickly. If you need instantaneous changing of
// all pins, you need to consider an output buffer.
func WriteUIntToPins(value uint32, pins []Pin) error {
	if len(pins) > 31 {
		return errors.New("WriteUIntToPins only supports up to 32 bits")
	}

	bit := uint32(0)
	v := value
	mask := uint32(1) << uint32((len(pins) - 1))
	for i := uint32(0); i < uint32(len(pins)); i++ {
		bit = v & mask
		if bit != 0 {
			bit = 1
		}
		v = v << 1
		// write to data pin
		//		fmt.Printf("Writing %s to pin %s\n", bit, pins[i])
		e := DigitalWrite(pins[i], int(bit))
		if e != nil {
			return e
		}
	}
	return nil
}

// Write a string to a file and close it again.
func WriteStringToFile(filename string, value string) error {
	//	fmt.Printf("writing %s to file %s\n", value, filename)
	f, e := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC, 0666)
	if e != nil {
		return e
	}
	defer f.Close()

	_, e = f.WriteString(value)
	return e
}

// Given a glob pattern, return the full path of the first matching file
func findFirstMatchingFile(glob string) (string, error) {
	matches, e := filepath.Glob(glob)
	if e != nil {
		return "", e
	}

	if len(matches) >= 1 {
		return matches[0], nil
	}
	return "", nil
}

func Map(value int, fromLow int, fromHigh int, toLow int, toHigh int) int {
	return (value-fromLow)*(toHigh-toLow)/(fromHigh-fromLow) + toLow
}

// Given a high and low byte, combine to form a single 16-bit value
func UInt16FromUInt8(highByte byte, lowByte byte) uint16 {
	return uint16(uint16(highByte)<<8) | uint16(lowByte)
}

func ReverseBytes16(value uint16) uint16 {
	// @todo implement ReverseBytes16()
	return 0
}

func ReverseBytes32(value uint32) uint32 {
	// @todo implement ReverseBytes32()
	return 0
}

// Get a module by name. If driver is not set, it will return an error. If the driver does not support that module,
//
func GetModule(name string) (Module, error) {
	driver := GetDriver()
	if driver == nil {
		return nil, errors.New("GetModule: Driver is not set")
	}

	modules := driver.GetModules()
	return modules[name], nil
}

// This is the interface that hardware drivers implement. Generally all drivers are created
// but not initialised. If MatchesHardwareConfig() is true and the driver is selected, Init()
// will be called.
type HardwareDriver interface {
	// Each driver is responsible for evaluating whether it applies to the current hardware
	// configuration or not. If this function returns false, the driver will not be used and Init
	// will not be called. If this function returns true, the driver may be called, in which case
	// Init will then be called
	MatchesHardwareConfig() bool

	// Initialise the driver.
	Init() (e error)

	// Return a module by name, or nil if undefined. The module names can be different between types of boards.
	GetModules() map[string]Module

	// Return the pin map for the driver, listing all supported pins and their capabilities
	PinMap() (pinMap HardwarePinMap)

	// Close the driver before destruction
	Close()
}
