package hwio

// A driver for BeagleBone's running Linux kernel 3.8 or higher, which use device trees instead
// of the old driver.
//
// Notable differences between this driver and the other BeagleBone driver:
// - this uses the file system for everything.
// - will only work on linux kernel 3.8 and higher, irrespective of the board version.
// - memory mapping is no longer used, as it was unsupported anyway.
// - this will probably not have the raw performance of the memory map technique (this is yet to be measured)
// - this driver will likely support alot more functions, as it's leveraging drivers that already exist.
//
// This driver shares some information from the other driver, since the pin configuration information is essentially the same.
//
// Articles used in building this driver:
// GPIO:
// - http://www.avrfreaks.net/wiki/index.php/Documentation:Linux/GPIO#Example_of_GPIO_access_from_within_a_C_program
// Analog:
// - http://hipstercircuits.com/reading-analog-adc-values-on-beaglebone-black/
// Background on changes in linux kernal 3.8:
// - https://docs.google.com/document/d/17P54kZkZO_-JtTjrFuVz-Cp_RMMg7GB_8W9JK9sLKfA/edit?hl=en&forcehl=1#heading=h.mfjmczsbv38r

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type BeagleBoneFSOpenPin struct {
	pin          Pin
	gpioLogical  int
	gpioBaseName string
	valueFile    *os.File
}

// Write a string to a file and close it again.
func writeStringToFile(filename string, value string) error {
	//	fmt.Printf("writing %s to file %s\n", value, filename)
	f, e := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC, 0666)
	if e != nil {
		return e
	}
	defer f.Close()

	f.WriteString(value)
	return nil
}

// Needs to be called to allocate the GPIO pin
func (op *BeagleBoneFSOpenPin) gpioExport() error {
	s := strconv.FormatInt(int64(op.gpioLogical), 10)
	e := writeStringToFile("/sys/class/gpio/export", s)
	if e != nil {
		return e
	}

	// calculate the base name for the gpio pin
	op.gpioBaseName = "/sys/class/gpio/gpio" + strconv.Itoa(op.gpioLogical)
	return nil
}

// Once exported, the direction of a GPIO can be set
func (op *BeagleBoneFSOpenPin) gpioDirection(dir string) error {
	if dir != "in" && dir != "out" {
		return errors.New("direction must be in or out")
	}
	f := op.gpioBaseName + "/direction"
	e := writeStringToFile(f, dir)

	mode := os.O_WRONLY | os.O_TRUNC
	if dir == "in" {
		mode = os.O_RDONLY
	}

	// open the value file with the correct mode. Put that file in 'op'. Note that we keep this file open
	// continuously for performance.
	// Preliminary tests on 200,000 DigitalWrites indicate an order of magnitude improvement when we don't have
	// to re-open the file each time. Re-seeking and writing a new value suffices.
	op.valueFile, e = os.OpenFile(op.gpioBaseName+"/value", mode, 0666)

	return e
}

// Get the value. Will return HIGH or LOW
func (op *BeagleBoneFSOpenPin) gpioGetValue() (int, error) {
	var b []byte
	b = make([]byte, 1)
	n, e := op.valueFile.ReadAt(b, 0)
	value := 0
	if n > 0 {
		if b[0] == '1' {
			value = HIGH
		} else {
			value = LOW
		}
	}
	return value, e
}

// Set the value, Expects HIGH or LOW
func (op *BeagleBoneFSOpenPin) gpioSetValue(value int) error {
	if op.valueFile == nil {
		fmt.Printf("value file no set\n")
		return errors.New("value file is not defined")
	}

	// Seek the start of the value file before writing. This is sufficient for the driver to accept a new value.
	_, e := op.valueFile.Seek(0, 0)
	if e != nil {
		return e
	}

	// Write a 1 or 0.
	// @todo investigate if we'd get better performance if we have precalculated []byte values with 0 and 1, and
	// use write directly instead of WriteString. Probably only marginal.
	if value == 0 {
		op.valueFile.WriteString("0")
	} else {
		op.valueFile.WriteString("1")
	}

	return nil
}

type BeagleBoneFSDriver struct {
	openPins map[Pin]*BeagleBoneFSOpenPin
}

func (d *BeagleBoneFSDriver) Init() error {
	d.openPins = make(map[Pin]*BeagleBoneFSOpenPin)
	return nil
}

func (d *BeagleBoneFSDriver) Close() {
	// @todo call unexport on all open pins
}

// create an openPin object and put it in the map.
func (d *BeagleBoneFSDriver) makeOpenPin(pin Pin, gpioLogicalPin int) *BeagleBoneFSOpenPin {
	result := &BeagleBoneFSOpenPin{pin: pin, gpioLogical: gpioLogicalPin}
	d.openPins[pin] = result
	return result
}

// For GPIO:
// - write GPIO pin to /sys/class/gpio/export. This is the port number plus pin on that port. Ports 0, 32, 64, 96.
// - write direction to /sys/class/gpio/gpio{nn}/direction. Values are 'in' and 'out'

func (d *BeagleBoneFSDriver) PinMode(pin Pin, mode PinIOMode) error {
	p := beaglePins[pin]

	// handle analog first, they are simplest from PinMode perspective
	if p.isAnalogPin() {
		if mode != INPUT {
			return errors.New(fmt.Sprintf("Pin %d is an analog pin, and the mode must be INPUT", p))
		}
		// @todo set up the analog pin
		return nil // nothing to set up
	}

	// Create an open pin object
	openPin := d.makeOpenPin(pin, p.gpioLogical)
	e := openPin.gpioExport()
	if e != nil {
		return e
	}

	if mode == OUTPUT {
		fmt.Printf("about to set pin %d to output\n", pin)
		e = openPin.gpioDirection("out")
		if e != nil {
			return e
		}
	} else {
		e = openPin.gpioDirection("in")
		// pull := BB_CONF_PULL_DISABLE
		// // note: pull up/down modes assume that CONF_PULLDOWN resets the pull disable bit
		// if mode == INPUT_PULLUP {
		// 	pull = BB_CONF_PULLUP
		// } else if mode == INPUT_PULLDOWN {
		// 	pull = BB_CONF_PULLDOWN
		// }

		if e != nil {
			return e
		}
	}
	return nil
}

func (d *BeagleBoneFSDriver) DigitalWrite(pin Pin, value int) (e error) {
	openPin := d.openPins[pin]
	if openPin == nil {
		return errors.New("Pin is being written but has not been opened")
	}
	openPin.gpioSetValue(value)
	return nil
}

func (d *BeagleBoneFSDriver) DigitalRead(pin Pin) (value int, e error) {
	openPin := d.openPins[pin]
	return openPin.gpioGetValue()
}

func (d *BeagleBoneFSDriver) AnalogWrite(pin Pin, value int) (e error) {
	return nil
}

func (d *BeagleBoneFSDriver) AnalogRead(pin Pin) (value int, e error) {
	return 0, nil
}

func (d *BeagleBoneFSDriver) PinMap() (pinMap HardwarePinMap) {
	pinMap = make(HardwarePinMap)

	for i, hw := range beaglePins {
		names := []string{hw.hwPin}
		if hw.hwPin != hw.gpioName {
			names = append(names, hw.gpioName)
		}
		pinMap.add(Pin(i), names, hw.profile)
	}

	return
}
