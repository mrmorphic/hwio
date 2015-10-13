package hwio

// A mock driver used for unit testing.
import (
	// 	"errors"
	"fmt"
)

type testDriverPin struct {
	names   []string
	modules []string
	extra   string
}

type testDriverPinMap map[Pin]*testDriverPin

type TestDriver struct {
	pinDefs []*testDriverPin
	modules map[string]Module

	verbose bool
}

func (d *TestDriver) Init() error {
	d.createPinData()
	d.initialiseModules()

	return nil
}

func (d *TestDriver) Close() {

}

func (d *TestDriver) MatchesHardwareConfig() bool {
	return true
}

// func (d *TestDriver) SetVerbosity(verbose bool) {
// 	d.verbose = verbose
// }

// // Mock records the pin mode being assigned.
// func (d *TestDriver) PinMode(pin Pin, mode PinIOMode) (e error) {
// 	if d.verbose {
// 		fmt.Printf("PinMode(%d, %s)\n", pin, mode.String())
// 	}
// 	m := d.getPinModes()
// 	m[pin] = mode
// 	return nil
// }

// // Mock emulates DigitalWrite by writing the value to pinValues.
// func (d *TestDriver) DigitalWrite(pin Pin, value int) (e error) {
// 	if d.verbose {
// 		fmt.Printf("DigitalWrite(%d, %d)\n", pin, value)
// 	}
// 	m := d.getPinValues()
// 	m[pin] = value
// 	return nil
// }

// func (d *TestDriver) DigitalRead(pin Pin) (value int, e error) {
// 	m := d.getPinValues()
// 	return m[pin], nil
// }

// func (d *TestDriver) AnalogWrite(pin Pin, value int) (e error) {
// 	// just the same as DigitalWrite, store value in the pinValues map
// 	return DigitalWrite(pin, value)
// }

// func (d *TestDriver) AnalogRead(pin Pin) (value int, e error) {
// 	if pin == 6 {
// 		return 1, nil
// 	}
// 	if pin == 7 {
// 		return 1000, nil
// 	}
// 	return 0, errors.New("analog read got error")
// }

// // Mock has a fixed set of hardcoded pins with different capabilities
func (d *TestDriver) PinMap() HardwarePinMap {
	result := make(HardwarePinMap)

	for i, hw := range d.pinDefs {
		result.add(Pin(i), hw.names, hw.modules)
	}

	return result
}

// // Getter that gets the pinModes map on demand, and creating it on first
// // instance.
// func (d *TestDriver) getPinModes() map[Pin]PinIOMode {
// 	if d.pinModes == nil {
// 		d.pinModes = make(map[Pin]PinIOMode)
// 	}
// 	return d.pinModes
// }

// func (d *TestDriver) getPinValues() map[Pin]int {
// 	if d.pinValues == nil {
// 		d.pinValues = make(map[Pin]int)
// 	}
// 	return d.pinValues
// }

// func (d *TestDriver) MockGetPinMode(pin Pin) PinIOMode {
// 	m := d.getPinModes()
// 	return m[pin]
// }

// func (d *TestDriver) MockGetPinValue(pin Pin) int {
// 	m := d.getPinValues()
// 	return m[pin]
// }

// func (d *TestDriver) MockSetPinValue(pin Pin, value int) {
// 	m := d.getPinValues()
// 	m[pin] = value
// }

func (d *TestDriver) GetModules() map[string]Module {
	return d.modules
}

func (d *TestDriver) createPinData() {
	d.pinDefs = []*testDriverPin{
		d.makeTestPin([]string{"P1", "gpio1"}, []string{"gpio"}, "1"),
		d.makeTestPin([]string{"P2", "gpio2"}, []string{"gpio"}, "2"),
		d.makeTestPin([]string{"P3", "gpio3"}, []string{"gpio"}, "3"),
		d.makeTestPin([]string{"P4", "gpio4"}, []string{"gpio"}, "4"),
		d.makeTestPin([]string{"P5", "gpio5"}, []string{"gpio"}, "5"),
		d.makeTestPin([]string{"P6", "gpio6"}, []string{"gpio"}, "6"),
		d.makeTestPin([]string{"P7", "gpio7"}, []string{"gpio"}, "7"),
		d.makeTestPin([]string{"P8", "gpio8"}, []string{"gpio"}, "8"),
		d.makeTestPin([]string{"P9", "gpio9"}, []string{"gpio"}, "9"),
		d.makeTestPin([]string{"P10", "gpio10"}, []string{"gpio"}, "10"),
		d.makeTestPin([]string{"P11", "ain4"}, []string{"analog"}, "11"),
		d.makeTestPin([]string{"P12", "ain6"}, []string{"analog"}, "12"),
	}
}

func (d *TestDriver) makeTestPin(names []string, modules []string, extra string) *testDriverPin {
	result := &testDriverPin{names, modules, extra}
	return result
}

func (d *TestDriver) initialiseModules() {
	d.modules = make(map[string]Module)

	gpio := newTestGPIOModule("gpio")
	gpio.SetOptions(d.getModuleOptions("gpio"))

	analog := newTestAnalogModule("analog")
	analog.SetOptions(d.getModuleOptions("analog"))

	// i2c1 := NewDTI2CModule("i2c1")

	d.modules["gpio"] = gpio
	d.modules["analog"] = analog
	// d.modules["i2c1"] = i2c1
}

func (d *TestDriver) getModuleOptions(module string) map[string]interface{} {
	result := make(map[string]interface{})

	pins := make(testDriverPinMap)

	for i, hw := range d.pinDefs {
		if hw.modules[0] == module {
			pins[Pin(i)] = hw
		}
	}
	result["pins"] = pins

	return result
}

// Mock module to replicate GPIO behaviour
type testGPIOModule struct {
	name string

	pinDefs testDriverPinMap

	pinModes map[Pin]PinIOMode

	// this simulates actual pin values. DigitalWrite ends up settin
	pinValues map[Pin]int
}

func newTestGPIOModule(name string) *testGPIOModule {
	result := &testGPIOModule{name: name}
	result.pinModes = make(map[Pin]PinIOMode)
	result.pinValues = make(map[Pin]int)
	return result
}

func (module *testGPIOModule) SetOptions(map[string]interface{}) error {
	return nil
}

func (module *testGPIOModule) Enable() error {
	return nil
}

func (module *testGPIOModule) Disable() error {
	return nil
}

func (module *testGPIOModule) GetName() string {
	return module.name
}

func (module *testGPIOModule) PinMode(pin Pin, mode PinIOMode) error {
	module.pinModes[pin] = mode
	return nil
}

func (module *testGPIOModule) DigitalWrite(pin Pin, value int) error {
	if module.pinModes[pin] == 0 {
		return fmt.Errorf("Pin %d has not had mode set", pin)
	}
	module.pinValues[pin] = value
	return nil
}

func (module *testGPIOModule) DigitalRead(pin Pin) (int, error) {
	return module.pinValues[pin], nil

}

func (module *testGPIOModule) ClosePin(pin Pin) error {
	return nil
}

func (module *testGPIOModule) MockGetPinMode(pin Pin) PinIOMode {
	return module.pinModes[pin]
}

func (module *testGPIOModule) MockGetPinValue(pin Pin) int {
	return module.pinValues[pin]
}

func (module *testGPIOModule) MockSetPinValue(pin Pin, value int) {
	module.pinValues[pin] = value
}

// Mock module to replicate analog module behaviour.
type testAnalogModule struct {
	name string

	pinDefs testDriverPinMap
}

func newTestAnalogModule(name string) *testAnalogModule {
	return &testAnalogModule{name: name}
}

func (module *testAnalogModule) SetOptions(map[string]interface{}) error {
	return nil
}

func (module *testAnalogModule) Enable() error {
	return nil
}

func (module *testAnalogModule) Disable() error {
	return nil
}

func (module *testAnalogModule) GetName() string {
	return module.name
}

func (module *testAnalogModule) AnalogRead(pin Pin) (result int, e error) {
	if pin == 10 {
		return 1, nil
	}
	if pin == 11 {
		return 1000, nil
	}
	return 0, nil
}
