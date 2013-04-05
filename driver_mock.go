package hwio

// A mock driver used for unit testing.
import (
	"fmt"
)

type TestDriver struct {
	pinModes  map[Pin]PinIOMode
	pinValues map[Pin]int
	verbose   bool
}

func (d *TestDriver) Init() error {
	return nil
}

func (d *TestDriver) Close() {

}

func (d *TestDriver) SetVerbosity(verbose bool) {
	d.verbose = verbose
}

// Mock records the pin mode being assigned.
func (d *TestDriver) PinMode(pin Pin, mode PinIOMode) (e error) {
	if d.verbose {
		fmt.Printf("PinMode(%d, %s)\n", pin, mode.String())
	}
	m := d.getPinModes()
	m[pin] = mode
	return nil
}

// Mock emulates DigitalWrite by writing the value to pinValues.
func (d *TestDriver) DigitalWrite(pin Pin, value int) (e error) {
	if d.verbose {
		fmt.Printf("DigitalWrite(%d, %d)\n", pin, value)
	}
	m := d.getPinValues()
	m[pin] = value
	return nil
}

func (d *TestDriver) DigitalRead(pin Pin) (value int, e error) {
	m := d.getPinValues()
	return m[pin], nil
}

func (d *TestDriver) AnalogWrite(pin Pin, value int) (e error) {
	// just the same as DigitalWrite, store value in the pinValues map
	return DigitalWrite(pin, value)
}

func (d *TestDriver) AnalogRead(pin Pin) (value int, e error) {
	return 0, nil
}

// Mock has a fixed set of hardcoded pins with different capabilities
func (d *TestDriver) PinMap() (pinMap HardwarePinMap) {
	general := []Capability{CAP_INPUT, CAP_OUTPUT}
	analog := []Capability{CAP_INPUT, CAP_OUTPUT, CAP_ANALOG_IN}
	pwm := []Capability{CAP_INPUT, CAP_OUTPUT, CAP_PWM}
	readonly := []Capability{CAP_INPUT}
	writeonly := []Capability{CAP_OUTPUT}

	pinMap = make(HardwarePinMap)

	pinMap.add(0, []string{"HWPin0"}, general)
	pinMap.add(1, []string{"HWPin1"}, readonly)
	pinMap.add(2, []string{"HWPin2"}, writeonly)
	pinMap.add(3, []string{"HWPin3"}, general)
	pinMap.add(4, []string{"HWPin4"}, general)
	pinMap.add(5, []string{"HWPin5"}, general)
	pinMap.add(6, []string{"HWPin6"}, analog)
	pinMap.add(7, []string{"HWPin7"}, pwm)
	return
}

// Getter that gets the pinModes map on demand, and creating it on first
// instance. 
func (d *TestDriver) getPinModes() map[Pin]PinIOMode {
	if d.pinModes == nil {
		d.pinModes = make(map[Pin]PinIOMode)
	}
	return d.pinModes
}

func (d *TestDriver) getPinValues() map[Pin]int {
	if d.pinValues == nil {
		d.pinValues = make(map[Pin]int)
	}
	return d.pinValues
}

func (d *TestDriver) MockGetPinMode(pin Pin) PinIOMode {
	m := d.getPinModes()
	return m[pin]
}

func (d *TestDriver) MockGetPinValue(pin Pin) int {
	m := d.getPinValues()
	return m[pin]
}

func (d *TestDriver) MockSetPinValue(pin Pin, value int) {
	m := d.getPinValues()
	m[pin] = value
}
