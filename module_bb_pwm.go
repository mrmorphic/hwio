// Implementation of PWM module interface for systems using device tree.
// It follows a similar pattern as the DT GPIO module. A module instance can handle
// multiple pins.

// period = nanoseconds, 1,000,000,000 is a second
// duty = active period
// polarity = 0: duty high; polarity = 1: duty low
// run = 0: disable; 1: enabled

package hwio

// References:
// - http://digital-drive.com/?p=146

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type BBPWMModule struct {
	name        string
	definedPins BBPWMModulePinDefMap
	openPins    map[Pin]*BBPWMModuleOpenPin
}

type BBPWMModulePinDef struct {
	pin Pin

	// used to derive the slot if not there, and the folder which contains the PWM files. This is of the form
	// "P8_13" and is case-sensitive
	name string
}

type BBPWMModulePinDefMap map[Pin]*BBPWMModulePinDef

type BBPWMModuleOpenPin struct {
	pin          Pin
	periodFile   string
	dutyFile     string
	polarityFile string
	runFile      string
}

func (pinDef BBPWMModulePinDef) overlayName() string {
	return "bone_pwm_" + pinDef.name
}

func (pinDef BBPWMModulePinDef) deviceDir() string {
	s, _ := findFirstMatchingFile("/sys/devices/ocp.*/pwm_test_" + pinDef.name + ".*")
	return s + "/"
}

func NewBBPWMModule(name string) (result *BBPWMModule) {
	result = &BBPWMModule{name: name}
	result.openPins = make(map[Pin]*BBPWMModuleOpenPin)
	return result
}

// Set options of the module. Parameters we look for include:
// - "pins" - an object of type DTGPIOModulePinDefMap
func (module *BBPWMModule) SetOptions(options map[string]interface{}) error {
	v := options["pins"]
	if v == nil {
		return fmt.Errorf("Module '%s' SetOptions() did not get 'pins' values", module.GetName())
	}

	module.definedPins = v.(BBPWMModulePinDefMap)
	return nil
}

// enable PWM module. It doesn't allocate any pins immediately. It does check of am33xx_pwm is present
// in the capemgr slots, and adds it if not. By default, this is not enabled on the BB but can be added
// easily.
func (module *BBPWMModule) Enable() error {
	// ensure that the PWM module is loaded
	return module.ensureSlot("am33xx_pwm")
}

// disables module and release any pins assigned.
func (module *BBPWMModule) Disable() error {
	for _, openPin := range module.openPins {
		openPin.closePin()
	}
	return nil
}

func (module *BBPWMModule) GetName() string {
	return module.name
}

// Enable a specific PWM pin. You need to call this explicitly after enabling the module, as the
// module will not by default allocate all pins, since there are a few.
func (module *BBPWMModule) EnablePin(pin Pin, enabled bool) error {
	if module.definedPins[pin] == nil {
		return fmt.Errorf("Pin %d is not known as a PWM pin on module %s", pin, module.GetName())
	}

	openPin := module.openPins[pin]
	if enabled {
		// ensure pin is enabled by creating an open pin
		if openPin == nil {
			p, e := module.makeOpenPin(pin)
			if e != nil {
				return e
			}
			module.openPins[pin] = p
			return p.enabled(true)
		}
	} else {
		// disable the pin if enabled
		if openPin != nil {
			return openPin.enabled(false)
		}
	}
	return nil
}

// Set the period of this pin, in nanoseconds
func (module *BBPWMModule) SetPeriod(pin Pin, ns int64) error {
	openPin := module.openPins[pin]
	if openPin == nil {
		return fmt.Errorf("PWM pin is being written but is not enabled. Have you called EnablePin?")
	}

	return openPin.setPeriod(ns)
}

// Set the duty time, the amount of time during each period that that output is HIGH.
func (module *BBPWMModule) SetDuty(pin Pin, ns int64) error {
	openPin := module.openPins[pin]
	if openPin == nil {
		return fmt.Errorf("PWM pin is being written but is not enabled. Have you called EnablePin?")
	}

	return openPin.setDuty(ns)
}

// create an openPin object and put it in the map.
func (module *BBPWMModule) makeOpenPin(pin Pin) (*BBPWMModuleOpenPin, error) {
	p := module.definedPins[pin]
	if p == nil {
		return nil, fmt.Errorf("Pin %d is not known to PWM module %s", pin, module.GetName())
	}

	e := AssignPin(pin, module)
	if e != nil {
		return nil, e
	}

	// Ensure that the cape manager knows about it
	e = module.ensureSlot(p.overlayName())
	if e != nil {
		return nil, e
	}

	dir := p.deviceDir()
	result := &BBPWMModuleOpenPin{pin: pin}
	result.periodFile = dir + "period"
	result.dutyFile = dir + "duty"
	result.runFile = dir + "run"
	result.polarityFile = dir + "polarity"

	module.openPins[pin] = result

	// ensure polarity is 0, so that the duty time represents the time the signal is high.
	e = writeStringToFile(result.polarityFile, "0")
	if e != nil {
		return nil, e
	}

	return result, nil
}

// Add the named thing to the capemanager slots file.
// @todo refactor for beaglebone black driver, and refactor analog as well which does the same thing.
func (module *BBPWMModule) ensureSlot(item string) error {
	path, e := findFirstMatchingFile("/sys/devices/bone_capemgr.*/slots")
	if e != nil {
		return e
	}

	file, e := os.Open(path)
	if e != nil {
		return e
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Index(line, item) > 0 {
			return nil
		}
	}

	// enable the item
	return writeStringToFile(path, item)
}

// Needs to be called to allocate the GPIO pin
func (op *BBPWMModuleOpenPin) closePin() error {
	// @todo how do we close this pin?

	return nil
}

// @todo capture the stdout message on writestring, which happens if the driver doesn't like the value.
// Set the period in nanoseconds. On BBB, maximum is 1 second (1,000,000,000ns)
func (op *BBPWMModuleOpenPin) setPeriod(ns int64) error {
	s := strconv.FormatInt(int64(ns), 10)
	e := writeStringToFile(op.periodFile, s)
	if e != nil {
		return e
	}

	return nil
}

func (op *BBPWMModuleOpenPin) setDuty(ns int64) error {
	s := strconv.FormatInt(int64(ns), 10)
	e := writeStringToFile(op.dutyFile, s)
	if e != nil {
		return e
	}

	return nil
}

func (op *BBPWMModuleOpenPin) enabled(e bool) error {
	if e {
		return writeStringToFile(op.runFile, "1")
	} else {
		return writeStringToFile(op.runFile, "0")
	}
}
