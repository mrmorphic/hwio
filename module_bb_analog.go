package hwio

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// BBAnalogModule handles BeagleBone-specific analog.
type BBAnalogModule struct {
	name string

	analogInitialised    bool
	analogValueFilesPath string

	definedPins BBAnalogModulePinDefMap

	openPins map[Pin]*BBAnalogModuleOpenPin
}

// Represents the definition of an analog pin, which should contain all the info required to open, close, read and write the pin
// using FS drivers.
type BBAnalogModulePinDef struct {
	pin           Pin
	analogLogical int
}

// A map of GPIO pin definitions.
type BBAnalogModulePinDefMap map[Pin]*BBAnalogModulePinDef

type BBAnalogModuleOpenPin struct {
	pin           Pin
	analogLogical int

	// path to file representing analog pin
	analogFile string

	valueFile *os.File
}

func NewBBAnalogModule(name string) (result *BBAnalogModule) {
	result = &BBAnalogModule{name: name}
	result.openPins = make(map[Pin]*BBAnalogModuleOpenPin)
	return result
}

// Set options of the module. Parameters we look for include:
// - "pins" - an object of type BBAnalogModulePinDefMap
func (module *BBAnalogModule) SetOptions(options map[string]interface{}) error {
	v := options["pins"]
	if v == nil {
		return fmt.Errorf("Module '%s' SetOptions() did not get 'pins' values", module.GetName())
	}

	module.definedPins = v.(BBAnalogModulePinDefMap)
	return nil
}

// enable GPIO module. It doesn't allocate any pins immediately.
func (module *BBAnalogModule) Enable() error {
	// once-off initialisation of analog
	if !module.analogInitialised {
		path, e := findFirstMatchingFile("/sys/devices/bone_capemgr.*/slots")
		if e != nil {
			return e
		}

		// enable analog
		e = WriteStringToFile(path, "cape-bone-iio")
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
func (module *BBAnalogModule) Disable() error {
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

func (module *BBAnalogModule) GetName() string {
	return module.name
}

// func (module *BBAnalogModule) AnalogWrite(pin Pin, value int) (e error) {
// 	return nil
// }

func (module *BBAnalogModule) AnalogRead(pin Pin) (value int, e error) {
	openPin := module.openPins[pin]
	if openPin == nil {
		return 0, errors.New("Pin is being read for analog value but has not been opened. Have you called PinMode?")
	}
	return openPin.analogGetValue()
}

func (module *BBAnalogModule) makeOpenAnalogPin(pin Pin) (*BBAnalogModuleOpenPin, error) {
	p := module.definedPins[pin]
	if p == nil {
		return nil, fmt.Errorf("Pin %d is not known to analog module", pin)
	}

	path := module.analogValueFilesPath + fmt.Sprintf("ain%d", p.analogLogical)
	result := &BBAnalogModuleOpenPin{pin: pin, analogLogical: p.analogLogical, analogFile: path}
	module.openPins[pin] = result

	return result, nil
}

func (op *BBAnalogModuleOpenPin) analogOpen() error {
	// Open analog input file computed from the calculated path of actual analog files and the analog pin name
	f, e := os.OpenFile(op.analogFile, os.O_RDONLY, 0666)
	op.valueFile = f

	return e
}

func (op *BBAnalogModuleOpenPin) analogGetValue() (int, error) {
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

func (op *BBAnalogModuleOpenPin) analogClose() error {
	return op.valueFile.Close()
}
