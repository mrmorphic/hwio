package hwio

// A driver for Raspberry Pi where device tree is supported (linux kernel 3.7+)
//
// Things known to work (tested on raspian 3.10+ kernel, rev 1 board):
// - digital write on all support ed GPIO pins
// - digital read on all GPIO pins, for modes INPUT.
//
// Known issues:
// - INPUT_PULLUP and INPUT_PULLDOWN not implemented yet.
// - no support yet for SPI, I2C, serial
//
// References:
// - http://elinux.org/RPi_Low-level_peripherals
// - https://projects.drogon.net/raspberry-pi/wiringpi/
// - BCM2835 technical reference

// Represents info we need to know about a pin on the Pi.
type RPiPin struct {
	names   []string // This intended for the P8.16 format name (currently unused)
	modules []string // Names of modules that may allocate this pin

	gpioLogical int // logical number for GPIO, for pins used by "gpio" module. This is the GPIO port number plus the GPIO pin within the port.
}

type RaspberryPiDTDriver struct { // all pins understood by the driver
	pins []*RPiPin

	// a map of module names to module objects, created at initialisation
	modules map[string]Module
}

func (d *RaspberryPiDTDriver) Init() error {
	d.createPinData()
	d.initialiseModules()

	return nil
}

func (d *RaspberryPiDTDriver) makePin(names []string, modules []string, gpioLogical int) *RPiPin {
	return &RPiPin{names, modules, gpioLogical}
}

func (d *RaspberryPiDTDriver) createPinData() {
	d.pins = []*RPiPin{
		d.makePin([]string{"null"}, []string{"unassignable"}, 0), // 0 - spacer
		d.makePin([]string{"3.3v"}, []string{"unassignable"}, 0),
		d.makePin([]string{"5v"}, []string{"unassignable"}, 0),
		d.makePin([]string{"sda"}, []string{"i2c"}, 0),
		d.makePin([]string{"do-not-connect1"}, []string{"unassignable"}, 0),
		d.makePin([]string{"scl"}, []string{"i2c"}, 0),
		d.makePin([]string{"ground"}, []string{"unassignable"}, 0),
		d.makePin([]string{"gpio4"}, []string{"gpio"}, 4),
		d.makePin([]string{"txd"}, []string{"serial"}, 0),
		d.makePin([]string{"do-not-connect-2"}, []string{"unassignable"}, 0),
		d.makePin([]string{"rxd"}, []string{"serial"}, 0),
		d.makePin([]string{"gpio17"}, []string{"gpio"}, 17),
		d.makePin([]string{"gpio18"}, []string{"gpio"}, 18), // also supports PWM
		d.makePin([]string{"gpio21"}, []string{"gpio"}, 21),
		d.makePin([]string{"do-not-connect-3"}, []string{"unassignable"}, 0),
		d.makePin([]string{"gpio22"}, []string{"gpio"}, 22),
		d.makePin([]string{"gpio23"}, []string{"gpio"}, 23),
		d.makePin([]string{"do-not-connect-4"}, []string{"unassignable"}, 0),
		d.makePin([]string{"gpio24"}, []string{"gpio"}, 24),
		d.makePin([]string{"mosi"}, []string{"spi"}, 0),
		d.makePin([]string{"do-not-connect-5"}, []string{"unassignable"}, 0),
		d.makePin([]string{"miso"}, []string{"spi"}, 0),
		d.makePin([]string{"gpio25"}, []string{"gpio"}, 25),
		d.makePin([]string{"sclk"}, []string{"spi"}, 0),
		d.makePin([]string{"ce0n"}, []string{"spi"}, 0),
		d.makePin([]string{"do-not-connect-6"}, []string{"unassignable"}, 0),
		d.makePin([]string{"ce1n"}, []string{"spi"}, 0),
	}
}

func (d *RaspberryPiDTDriver) initialiseModules() error {
	d.modules = make(map[string]Module)

	gpio := NewDTGPIOModule("gpio")
	e := gpio.SetOptions(d.getGPIOOptions())
	if e != nil {
		return e
	}

	i2c := NewDTI2CModule("i2c")
	e = i2c.SetOptions(d.getI2COptions())
	if e != nil {
		return e
	}

	d.modules["gpio"] = gpio
	d.modules["i2c"] = i2c

	return nil
}

// Get options for GPIO module, derived from the pin structure
func (d *RaspberryPiDTDriver) getGPIOOptions() map[string]interface{} {
	result := make(map[string]interface{})

	pins := make(DTGPIOModulePinDefMap)

	// Add the GPIO pins to this map
	for i, hw := range d.pins {
		if hw.modules[0] == "gpio" {
			pins[Pin(i)] = &DTGPIOModulePinDef{pin: Pin(i), gpioLogical: hw.gpioLogical}
		}
	}
	result["pins"] = pins

	return result
}

func (d *RaspberryPiDTDriver) getI2COptions() map[string]interface{} {
	result := make(map[string]interface{})

	pins := make(DTI2CModulePins, 0)
	pins = append(pins, Pin(3))
	pins = append(pins, Pin(5))

	result["pins"] = pins

	if d.BoardRevision() == 1 {
		result["device"] = "/dev/i2c-0"
	} else {
		result["device"] = "/dev/i2c-1"
	}

	return result
}

// Determine the version of Raspberry Pi.
// This discussion http://www.raspberrypi.org/phpBB3/viewtopic.php?f=44&t=23989
// was used to determine the algorithm, specifically the comment by gordon@drogon.net
// It will return 1 or 2.
func (d *RaspberryPiDTDriver) BoardRevision() int {
	revision := CpuInfo(0, "Revision")
	if revision == "0002" || revision == "0003" {
		return 1
	}
	return 2
}

func (d *RaspberryPiDTDriver) GetModules() map[string]Module {
	return d.modules
}

func (d *RaspberryPiDTDriver) Close() {
	// Disable all the modules
	for _, module := range d.modules {
		module.Disable()
	}
}

func (d *RaspberryPiDTDriver) PinMap() (pinMap HardwarePinMap) {
	pinMap = make(HardwarePinMap)

	for i, hw := range d.pins {
		pinMap.add(Pin(i), hw.names, hw.modules)
	}

	return
}
