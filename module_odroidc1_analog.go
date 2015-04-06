package hwio

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

// ODroidC1AnalogModule is a module for handling the Odroid C1 analog hardware, which is not generic.
type ODroidC1AnalogModule struct {
	name string

	analogInitialised bool

	definedPins ODroidC1AnalogModulePinDefMap

	openPins map[Pin]*ODroidC1AnalogModuleOpenPin
}

// Represents the definition of an analog pin, which should contain all the info required to open, close, read and write the pin
// using FS drivers.
type ODroidC1AnalogModulePinDef struct {
	pin           Pin
	analogLogical int
}

// A map of GPIO pin definitions.
type ODroidC1AnalogModulePinDefMap map[Pin]*ODroidC1AnalogModulePinDef

type ODroidC1AnalogModuleOpenPin struct {
	pin           Pin
	analogLogical int

	// path to file representing analog pin
	analogFile string

	valueFile *os.File
}

func NewODroidC1AnalogModule(name string) (result *ODroidC1AnalogModule) {
	result = &ODroidC1AnalogModule{name: name}
	result.openPins = make(map[Pin]*ODroidC1AnalogModuleOpenPin)
	return result
}

// Set options of the module. Parameters we look for include:
// - "pins" - an object of type ODroidC1AnalogModulePinDefMap
func (module *ODroidC1AnalogModule) SetOptions(options map[string]interface{}) error {
	v := options["pins"]
	if v == nil {
		return fmt.Errorf("Module '%s' SetOptions() did not get 'pins' values", module.GetName())
	}

	module.definedPins = v.(ODroidC1AnalogModulePinDefMap)
	return nil
}

// enable GPIO module. It doesn't allocate any pins immediately.
func (module *ODroidC1AnalogModule) Enable() error {
	// once-off initialisation of analog
	if !module.analogInitialised {
		module.analogInitialised = true

		// attempt to assign all pins to this module
		for pin, _ := range module.definedPins {
			// attempt to assign this pin for this module.
			e := AssignPin(pin, module)
			if e != nil {
				return e
			}
			e = module.makeOpenAnalogPin(pin)
			if e != nil {
				return e
			}
		}
	}
	return nil
}

// disables module and release any pins assigned.
func (module *ODroidC1AnalogModule) Disable() error {
	// Unassign any pins we may have assigned
	for pin, _ := range module.definedPins {
		// attempt to assign this pin for this module.
		UnassignPin(pin)
	}

	// if there are any open analog pins, close them
	for _, openPin := range module.openPins {
		openPin.analogClose()
	}
	return nil
}

func (module *ODroidC1AnalogModule) GetName() string {
	return module.name
}

func (module *ODroidC1AnalogModule) AnalogRead(pin Pin) (value int, e error) {
	openPin := module.openPins[pin]
	if openPin == nil {
		return 0, errors.New("Pin is being read for analog value but has not been opened. Have you called PinMode?")
	}
	return openPin.analogGetValue()
}

func (module *ODroidC1AnalogModule) makeOpenAnalogPin(pin Pin) error {
	p := module.definedPins[pin]
	if p == nil {
		return fmt.Errorf("Pin %d is not known to analog module", pin)
	}

	path := fmt.Sprintf("/sys/class/saradc/saradc_ch%d", p.analogLogical)
	result := &ODroidC1AnalogModuleOpenPin{pin: pin, analogLogical: p.analogLogical, analogFile: path}

	module.openPins[pin] = result

	e := result.analogOpen()
	if e != nil {
		return e
	}

	return nil
}

func (op *ODroidC1AnalogModuleOpenPin) analogOpen() error {
	// Open analog input file computed from the calculated path of actual analog files and the analog pin name
	f, e := os.OpenFile(op.analogFile, os.O_RDONLY, 0666)
	op.valueFile = f

	return e
}

func (op *ODroidC1AnalogModuleOpenPin) analogGetValue() (int, error) {
	var b []byte
	b = make([]byte, 5)
	n, e := op.valueFile.ReadAt(b, 0)

	// if there's an error and no byte were read, quit now. If we didn't get all the bytes we asked for, which
	// is generally the case, we will get an error as well but would have got some bytes.
	if e != nil && n == 0 {
		return 0, e
	}

	value, e := strconv.Atoi(string(b[:n-1]))

	return value, e
}

func (op *ODroidC1AnalogModuleOpenPin) analogClose() error {
	return op.valueFile.Close()
}
