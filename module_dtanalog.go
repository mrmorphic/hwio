package hwio

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type DTAnalogModule struct {
	name string

	analogInitialised    bool
	analogValueFilesPath string

	definedPins DTAnalogModulePinDefMap

	openPins map[Pin]*DTAnalogModuleOpenPin
}

// Represents the definition of an analog pin, which should contain all the info required to open, close, read and write the pin
// using FS drivers.
type DTAnalogModulePinDef struct {
	pin           Pin
	analogLogical int
}

// A map of GPIO pin definitions.
type DTAnalogModulePinDefMap map[Pin]*DTAnalogModulePinDef

type DTAnalogModuleOpenPin struct {
	pin           Pin
	analogLogical int

	// path to file representing analog pin
	analogFile string

	valueFile *os.File
}

func NewDTAnalogModule(name string) (result *DTAnalogModule) {
	result = &DTAnalogModule{name: name}
	result.openPins = make(map[Pin]*DTAnalogModuleOpenPin)
	return result
}

// Set options of the module. Parameters we look for include:
// - "pins" - an object of type DTAnalogModulePinDefMap
func (module *DTAnalogModule) SetOptions(options map[string]interface{}) error {
	v := options["pins"]
	if v == nil {
		return fmt.Errorf("Module '%s' SetOptions() did not get 'pins' values", module.GetName())
	}

	module.definedPins = v.(DTAnalogModulePinDefMap)
	return nil
}

// enable GPIO module. It doesn't allocate any pins immediately.
func (module *DTAnalogModule) Enable() error {
	// once-off initialisation of analog
	if !module.analogInitialised {

		path, e := findFirstMatchingFile("/sys/devices/bone_capemgr.*/slots")
		if e != nil {
			return e
		}

		// enable analog
		WriteStringToFile(path, "cape-bone-iio")
		if e != nil {
			return e
		}

		// determine path where analog files are
		path, e = findFirstMatchingFile("/sys/devices/ocp.*/helper.*/AIN0")
		if e != nil {
			return e
		}
		if path == "" {
			return errors.New("Could not locate /sys/devices/ocp.*/helper.*/AIN0")
		}

		// remove AIN0 to get the path where these files are
		module.analogValueFilesPath = strings.TrimSuffix(path, "AIN0")

		module.analogInitialised = true

		// attempt to assign all pins to this module
		for pin, _ := range module.definedPins {
			// attempt to assign this pin for this module.
			e = AssignPin(pin, module)
			if e != nil {
				return e
			}
		}
	}
	return nil
}

// disables module and release any pins assigned.
func (module *DTAnalogModule) Disable() error {
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

func (module *DTAnalogModule) GetName() string {
	return module.name
}

// func (module *DTAnalogModule) AnalogWrite(pin Pin, value int) (e error) {
// 	return nil
// }

func (module *DTAnalogModule) AnalogRead(pin Pin) (value int, e error) {
	openPin := module.openPins[pin]
	if openPin == nil {
		openPin, e = module.makeOpenAnalogPin(pin)
	}
	return openPin.analogGetValue()
}

func (module *DTAnalogModule) makeOpenAnalogPin(pin Pin) (*DTAnalogModuleOpenPin, error) {
	p := module.definedPins[pin]
	if p == nil {
		return nil, fmt.Errorf("Pin %d is not known to analog module", pin)
	}

	path := module.analogValueFilesPath + fmt.Sprintf("AIN%d", p.analogLogical)
	result := &DTAnalogModuleOpenPin{pin: pin, analogLogical: p.analogLogical, analogFile: path}
	module.openPins[pin] = result

	return result, nil
}

func (op *DTAnalogModuleOpenPin) analogOpen() error {
	// Open analog input file computed from the calculated path of actual analog files and the analog pin name
	f, e := os.OpenFile(op.analogFile, os.O_RDONLY, 0666)
	op.valueFile = f

	return e
}

func (op *DTAnalogModuleOpenPin) analogGetValue() (int, error) {
	op.analogOpen()

	var value int
	_, e := fmt.Fscanf(op.valueFile, "%d\n", &value)
	op.analogClose()

	return value, e
}

func (op *DTAnalogModuleOpenPin) analogClose() error {
	return op.valueFile.Close()
}
