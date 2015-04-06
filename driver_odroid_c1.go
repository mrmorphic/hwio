package hwio

// A driver for Odroid C1's running Ubuntu 14.04 with Linux kernel 3.8 or higher.
//
// Known issues:
// - INPUT_PULLUP and INPUT_PULLDOWN not implemented yet.
// - no support yet for SPI, serial, I2C
//
// GPIO are 3.3V, analog is 1.8V
//
// Articles used in building this driver:
// - http://www.hardkernel.com/main/products/prdt_info.php?g_code=G141578608433&tab_idx=2

type OdroidC1Driver struct {
	// all pins understood by the driver
	pinConfigs []*DTPinConfig

	// a map of module names to module objects, created at initialisation
	modules map[string]Module
}

func NewOdroidC1Driver() *OdroidC1Driver {
	return &OdroidC1Driver{}
}

// Examine the hardware environment and determine if this driver will handle it.
// For Odroid C1, it's easy: /proc/cpuinfo identifies it.
func (d *OdroidC1Driver) MatchesHardwareConfig() bool {
	// we need to get CPU 3, because /proc/cpuinfo on odroid has a set of properties
	// that are system wide, that are listed after CPU specific properties.
	// CpuInfo associated these with CPU 3, the last one it saw. Not ideal, but works.
	hw := CpuInfo(3, "Hardware")
	if hw == "ODROIDC" {
		return true
	}
	return false
}

func (d *OdroidC1Driver) Init() error {
	d.createPinData()
	return d.initialiseModules()
}

func (d *OdroidC1Driver) createPinData() {
	d.pinConfigs = []*DTPinConfig{
		// dummy placeholder for "pin 0"
		&DTPinConfig{[]string{"dummy"}, []string{"unassignable"}, 0, 0}, // 0 - spacer

		// Odroid has a mostly Raspberry Pi compatible header (40-pin), except GPIO numbers are different,
		// and an analog input is available.
		&DTPinConfig{[]string{"3.3v-1"}, []string{"unassignable"}, 0, 0},   // 1
		&DTPinConfig{[]string{"5v-1"}, []string{"unassignable"}, 0, 0},     // 2
		&DTPinConfig{[]string{"sda1"}, []string{"i2ca"}, 0, 0},             // 3
		&DTPinConfig{[]string{"5v-2"}, []string{"unassignable"}, 0, 0},     // 4
		&DTPinConfig{[]string{"scl1"}, []string{"i2ca"}, 0, 0},             // 5
		&DTPinConfig{[]string{"ground-1"}, []string{"unassignable"}, 0, 0}, // 6
		&DTPinConfig{[]string{"gpio83"}, []string{"gpio"}, 83, 0},          // 7
		&DTPinConfig{[]string{"txd"}, []string{"serial"}, 0, 0},            // 8
		&DTPinConfig{[]string{"ground-2"}, []string{"unassignable"}, 0, 0}, // 9
		&DTPinConfig{[]string{"rxd"}, []string{"serial"}, 0, 0},            // 10
		&DTPinConfig{[]string{"gpio88"}, []string{"gpio"}, 88, 0},          // 11
		&DTPinConfig{[]string{"gpio87"}, []string{"gpio"}, 87, 0},          // 12
		&DTPinConfig{[]string{"gpio116"}, []string{"gpio"}, 116, 0},        // 13
		&DTPinConfig{[]string{"ground-3"}, []string{"unassignable"}, 0, 0}, // 14
		&DTPinConfig{[]string{"gpio115"}, []string{"gpio"}, 115, 0},        // 15
		&DTPinConfig{[]string{"gpio104"}, []string{"gpio"}, 104, 0},        // 16
		&DTPinConfig{[]string{"3.3v-2"}, []string{"unassignable"}, 0, 0},   // 17
		&DTPinConfig{[]string{"gpio102"}, []string{"gpio"}, 102, 0},        // 18
		&DTPinConfig{[]string{"mosi"}, []string{"spi"}, 0, 0},              // 19 - may be GPIO by default - CHECK
		&DTPinConfig{[]string{"ground-4"}, []string{"unassignable"}, 0, 0}, // 20
		&DTPinConfig{[]string{"miso"}, []string{"spi"}, 0, 0},              // 21 - may be GPIO by default - CHECK
		&DTPinConfig{[]string{"gpio103"}, []string{"gpio"}, 103, 0},        // 22
		&DTPinConfig{[]string{"sclk"}, []string{"spi"}, 0, 0},              // 23 - may be GPIO by default - CHECK
		&DTPinConfig{[]string{"ce0"}, []string{"spi"}, 0, 0},               // 24 - also marked as CE0
		&DTPinConfig{[]string{"ground-5"}, []string{"unassignable"}, 0, 0}, // 25
		&DTPinConfig{[]string{"gpio118"}, []string{"gpio"}, 118, 0},        // 26
		&DTPinConfig{[]string{"sda2"}, []string{"i2cb"}, 0, 0},             // 27
		&DTPinConfig{[]string{"scl2"}, []string{"i2cb"}, 0, 0},             // 28
		&DTPinConfig{[]string{"gpio101"}, []string{"gpio"}, 101, 0},        // 29
		&DTPinConfig{[]string{"ground-6"}, []string{"unassignable"}, 0, 0}, // 30
		&DTPinConfig{[]string{"gpio100"}, []string{"gpio"}, 100, 0},        // 31
		&DTPinConfig{[]string{"gpio99"}, []string{"gpio"}, 99, 0},          // 32
		&DTPinConfig{[]string{"gpio108"}, []string{"gpio"}, 108, 0},        // 33
		&DTPinConfig{[]string{"ground-7"}, []string{"unassignable"}, 0, 0}, // 34
		&DTPinConfig{[]string{"gpio97"}, []string{"gpio"}, 97, 0},          // 35
		&DTPinConfig{[]string{"gpio98"}, []string{"gpio"}, 98, 0},          // 36
		&DTPinConfig{[]string{"ain1"}, []string{"analog"}, 26, 1},          // 37 - different from Rpi
		&DTPinConfig{[]string{"1.8v"}, []string{"unassignable"}, 0, 0},     // 38 - different from Rpi
		&DTPinConfig{[]string{"ground-8"}, []string{"unassignable"}, 0, 0}, // 39 - different from Rpi
		&DTPinConfig{[]string{"ain0"}, []string{"analog"}, 21, 0},          // 40 - different from Rpi
	}
}

func (d *OdroidC1Driver) initialiseModules() error {
	d.modules = make(map[string]Module)

	gpio := NewDTGPIOModule("gpio")
	e := gpio.SetOptions(d.getGPIOOptions())
	if e != nil {
		return e
	}

	analog := NewODroidC1AnalogModule("analog")
	e = analog.SetOptions(d.getAnalogOptions())
	if e != nil {
		return e
	}

	i2ca := NewDTI2CModule("i2ca")
	e = i2ca.SetOptions(d.getI2COptions("i2ca"))
	if e != nil {
		return e
	}
	i2cb := NewDTI2CModule("i2cb")
	e = i2cb.SetOptions(d.getI2COptions("i2cb"))
	if e != nil {
		return e
	}

	d.modules["gpio"] = gpio
	d.modules["analog"] = analog
	d.modules["i2ca"] = i2ca
	d.modules["i2cb"] = i2cb

	// alias i2c to i2c2. This is for portability; getting the i2c module on any device should return the default i2c interface,
	// but should not preclude addition of other i2c busses.
	d.modules["i2c"] = i2ca

	// initialise by default, which will assign P9.19 and P9.20. This is configured by default in device tree and these pins cannot be assigned.
	i2ca.Enable()
	i2cb.Enable()
	analog.Enable()

	return nil
}

// Get options for GPIO module, derived from the pin structure
func (d *OdroidC1Driver) getGPIOOptions() map[string]interface{} {
	result := make(map[string]interface{})

	pins := make(DTGPIOModulePinDefMap)

	// Add the GPIO pins to this map
	for i, pinConf := range d.pinConfigs {
		if pinConf.usedBy("gpio") {
			pins[Pin(i)] = &DTGPIOModulePinDef{pin: Pin(i), gpioLogical: pinConf.gpioLogical}
		}
	}
	result["pins"] = pins

	return result
}

// Get options for analog module, derived from the pin structure
func (d *OdroidC1Driver) getAnalogOptions() map[string]interface{} {
	result := make(map[string]interface{})

	pins := make(ODroidC1AnalogModulePinDefMap)

	// Add the GPIO pins to this map
	for i, pinConf := range d.pinConfigs {
		if pinConf.usedBy("analog") {
			pins[Pin(i)] = &ODroidC1AnalogModulePinDef{pin: Pin(i), analogLogical: pinConf.analogLogical}
		}
	}
	result["pins"] = pins

	return result
}

// Return the i2c options required to initialise that module.
func (d *OdroidC1Driver) getI2COptions(module string) map[string]interface{} {
	result := make(map[string]interface{})

	pins := make(DTI2CModulePins, 0)
	for i, pinConf := range d.pinConfigs {
		if pinConf.usedBy(module) {
			pins = append(pins, Pin(i))
		}
	}

	result["pins"] = pins

	// TODO CALCULATE THIS FROM MODULE
	// this should really look at the device structure to ensure that I2C2 on hardware maps to /dev/i2c1. This confusion seems
	// to happen because of the way the kernel initialises the devices at boot time.
	if module == "i2ca" {
		result["device"] = "/dev/i2c-1"
	} else {
		result["device"] = "/dev/i2c-2"
	}

	return result
}

// internal function to get a Pin. It does not use GetPin because that relies on the driver having already been initialised. This
// method can be called while still initialising. Only matches names[0], which is the Pn.nn expansion header name.
func (d *OdroidC1Driver) getPin(name string) Pin {
	for i, hw := range d.pinConfigs {
		if hw.names[0] == name {
			return Pin(i)
		}
	}
	return Pin(0)
}

func (d *OdroidC1Driver) GetModules() map[string]Module {
	return d.modules
}

func (d *OdroidC1Driver) Close() {
	// Disable all the modules
	for _, module := range d.modules {
		module.Disable()
	}
}

func (d *OdroidC1Driver) PinMap() (pinMap HardwarePinMap) {
	pinMap = make(HardwarePinMap)

	for i, hw := range d.pinConfigs {
		pinMap.add(Pin(i), hw.names, hw.modules)
	}

	return
}
