package hwio

// A driver for Raspberry Pi where device tree is supported (linux kernel 3.7+)
//
// Things known to work (tested on raspian 3.10+ kernel, rev 1 board):
// - digital write on all support ed GPIO pins
// - digital read on all GPIO pins, for modes INPUT.
//
// Known issues:
// - INPUT_PULLUP and INPUT_PULLDOWN not implemented yet.
// - no support yet for SPI, serial
//
// References:
// - http://elinux.org/RPi_Low-level_peripherals
// - https://projects.drogon.net/raspberry-pi/wiringpi/
// - BCM2835 technical reference

import (
	"os/exec"
	"strings"
)

type pinoutRevision int

const (
	//type0ne is used for a Raspberry 1
	type0ne = iota
	//typeTwo is used for a Raspberry 2
	typeTwo
	//typeAplusBPlusZeroPi2 is used fo  Raspberry Pi 1 Models A+ and B+, Pi 2 Model B, Pi 3 Model B and Pi Zero (and Zero W)
	typeAplusBPlusZeroPi2
)

type RaspberryPiDTDriver struct { // all pins understood by the driver
	pinConfigs []*DTPinConfig

	// a map of module names to module objects, created at initialisation
	modules map[string]Module
}

func NewRaspPiDTDriver() *RaspberryPiDTDriver {
	return &RaspberryPiDTDriver{}
}

func (d *RaspberryPiDTDriver) MatchesHardwareConfig() bool {
	cpuinfo, e := exec.Command("cat", "/proc/cpuinfo").Output()
	if e != nil {
		return false
	}
	s := string(cpuinfo)
	if strings.Contains(s, "BCM2708") || strings.Contains(s, "BCM2709") || strings.Contains(s, "BCM2835") {
		return true
	}

	return false
}

func (d *RaspberryPiDTDriver) Init() error {
	d.createPinData()
	d.initialiseModules()

	return nil
}

// http://www.hobbytronics.co.uk/raspberry-pi-gpio-pinout
func (d *RaspberryPiDTDriver) createPinData() {
	switch d.BoardRevision() {
	case type0ne:
		d.pinConfigs = []*DTPinConfig{
			{[]string{"null"}, []string{"unassignable"}, 0, 0}, // 0 - spacer
			{[]string{"3.3v"}, []string{"unassignable"}, 0, 0},
			{[]string{"5v"}, []string{"unassignable"}, 0, 0},
			{[]string{"sda"}, []string{"i2c"}, 0, 0},
			{[]string{"do-not-connect-1"}, []string{"unassignable"}, 0, 0},
			{[]string{"scl"}, []string{"i2c"}, 0, 0},
			{[]string{"ground"}, []string{"unassignable"}, 0, 0},
			{[]string{"gpio4"}, []string{"gpio"}, 4, 0},
			{[]string{"txd"}, []string{"serial"}, 0, 0},
			{[]string{"do-not-connect-2"}, []string{"unassignable"}, 0, 0},
			{[]string{"rxd"}, []string{"serial"}, 0, 0},
			{[]string{"gpio17"}, []string{"gpio"}, 17, 0},
			{[]string{"gpio18"}, []string{"gpio"}, 18, 0}, // also supports PWM
			{[]string{"gpio21"}, []string{"gpio"}, 21, 0},
			{[]string{"do-not-connect-3"}, []string{"unassignable"}, 0, 0},
			{[]string{"gpio22"}, []string{"gpio"}, 22, 0},
			{[]string{"gpio23"}, []string{"gpio"}, 23, 0},
			{[]string{"do-not-connect-4"}, []string{"unassignable"}, 0, 0},
			{[]string{"gpio24"}, []string{"gpio"}, 24, 0},
			{[]string{"mosi"}, []string{"spi"}, 0, 0},
			{[]string{"do-not-connect-5"}, []string{"unassignable"}, 0, 0},
			{[]string{"miso"}, []string{"spi"}, 0, 0},
			{[]string{"gpio25"}, []string{"gpio"}, 25, 0},
			{[]string{"sclk"}, []string{"spi"}, 0, 0},
			{[]string{"ce0n"}, []string{"spi"}, 0, 0},
			{[]string{"do-not-connect-6"}, []string{"unassignable"}, 0, 0},
			{[]string{"ce1n"}, []string{"spi"}, 0, 0},
		}
	case typeTwo:
		d.pinConfigs = []*DTPinConfig{
			{[]string{"null"}, []string{"unassignable"}, 0, 0}, // 0 - spacer
			{[]string{"3.3v-1"}, []string{"unassignable"}, 0, 0},
			{[]string{"5v-1"}, []string{"unassignable"}, 0, 0},
			{[]string{"sda"}, []string{"i2c"}, 0, 0},
			{[]string{"5v-2"}, []string{"unassignable"}, 0, 0},
			{[]string{"scl"}, []string{"i2c"}, 0, 0},
			{[]string{"ground-1"}, []string{"unassignable"}, 0, 0},
			{[]string{"gpio4"}, []string{"gpio"}, 4, 0},
			{[]string{"txd"}, []string{"serial"}, 0, 0},
			{[]string{"ground-2"}, []string{"unassignable"}, 0, 0},
			{[]string{"rxd"}, []string{"serial"}, 0, 0},
			{[]string{"gpio17"}, []string{"gpio"}, 17, 0},
			{[]string{"gpio18"}, []string{"gpio"}, 18, 0}, // also supports PWM
			{[]string{"gpio27"}, []string{"gpio"}, 27, 0},
			{[]string{"ground-3"}, []string{"unassignable"}, 0, 0},
			{[]string{"gpio22"}, []string{"gpio"}, 22, 0},
			{[]string{"gpio23"}, []string{"gpio"}, 23, 0},
			{[]string{"3.3v-2"}, []string{"unassignable"}, 0, 0},
			{[]string{"gpio24"}, []string{"gpio"}, 24, 0},
			{[]string{"mosi"}, []string{"spi"}, 0, 0},
			{[]string{"ground-4"}, []string{"unassignable"}, 0, 0},
			{[]string{"miso"}, []string{"spi"}, 0, 0},
			{[]string{"gpio25"}, []string{"gpio"}, 25, 0},
			{[]string{"sclk"}, []string{"spi"}, 0, 0},
			{[]string{"gpio8"}, []string{"gpio"}, 8, 0},
			{[]string{"ground-5"}, []string{"unassignable"}, 0, 0},
			{[]string{"gpio7"}, []string{"gpio"}, 7, 0},
		}
	case typeAplusBPlusZeroPi2: // B+
		d.pinConfigs = []*DTPinConfig{
			{[]string{"null"}, []string{"unassignable"}, 0, 0}, // 0 - spacer
			{[]string{"3.3v-1"}, []string{"unassignable"}, 0, 0},
			{[]string{"5v-1"}, []string{"unassignable"}, 0, 0},
			{[]string{"sda"}, []string{"i2c"}, 0, 0},
			{[]string{"5v-2"}, []string{"unassignable"}, 0, 0},
			{[]string{"scl"}, []string{"i2c"}, 0, 0},
			{[]string{"ground-1"}, []string{"unassignable"}, 0, 0},
			{[]string{"gpio4"}, []string{"gpio"}, 4, 0},
			{[]string{"txd"}, []string{"serial"}, 0, 0},
			{[]string{"ground-2"}, []string{"unassignable"}, 0, 0},
			{[]string{"rxd"}, []string{"serial"}, 0, 0},
			{[]string{"gpio17"}, []string{"gpio"}, 17, 0},
			{[]string{"gpio18"}, []string{"gpio"}, 18, 0}, // also supports PWM
			{[]string{"gpio27"}, []string{"gpio"}, 21, 0},
			{[]string{"ground-3"}, []string{"unassignable"}, 0, 0},
			{[]string{"gpio22"}, []string{"gpio"}, 22, 0},
			{[]string{"gpio23"}, []string{"gpio"}, 23, 0},
			{[]string{"3.3v-2"}, []string{"unassignable"}, 0, 0},
			{[]string{"gpio24"}, []string{"gpio"}, 24, 0},
			{[]string{"mosi"}, []string{"spi"}, 0, 0},
			{[]string{"ground-4"}, []string{"unassignable"}, 0, 0},
			{[]string{"miso"}, []string{"spi"}, 0, 0},
			{[]string{"gpio25"}, []string{"gpio"}, 25, 0},
			{[]string{"sclk"}, []string{"spi"}, 0, 0},
			{[]string{"gpio8"}, []string{"gpio"}, 8, 0},
			{[]string{"ground-5"}, []string{"unassignable"}, 0, 0},
			{[]string{"gpio7"}, []string{"gpio"}, 7, 0},
			{[]string{"do-not-connect-1"}, []string{"unassignable"}, 0, 0},
			{[]string{"do-not-connect-2"}, []string{"unassignable"}, 0, 0},
			{[]string{"gpio5"}, []string{"gpio"}, 5, 0},
			{[]string{"ground-6"}, []string{"unassignable"}, 0, 0},
			{[]string{"gpio6"}, []string{"gpio"}, 6, 0},
			{[]string{"gpio12"}, []string{"gpio"}, 12, 0},
			{[]string{"gpio13"}, []string{"gpio"}, 13, 0},
			{[]string{"ground-7"}, []string{"unassignable"}, 0, 0},
			{[]string{"gpio19"}, []string{"gpio"}, 19, 0},
			{[]string{"gpio16"}, []string{"gpio"}, 16, 0},
			{[]string{"gpio26"}, []string{"gpio"}, 26, 0},
			{[]string{"gpio20"}, []string{"gpio"}, 20, 0},
			{[]string{"ground-8"}, []string{"unassignable"}, 0, 0},
			{[]string{"gpio21"}, []string{"gpio"}, 21, 0},
		}
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

	// Create the leds module which is BBB-specific. There are no options.
	leds := NewDTLEDModule("leds")
	e = leds.SetOptions(d.getLEDOptions("leds"))
	if e != nil {
		return e
	}

	d.modules["gpio"] = gpio
	d.modules["i2c"] = i2c
	d.modules["leds"] = leds

	return nil
}

// Get options for GPIO module, derived from the pin structure
func (d *RaspberryPiDTDriver) getGPIOOptions() map[string]interface{} {
	result := make(map[string]interface{})

	pins := make(DTGPIOModulePinDefMap)

	// Add the GPIO pins to this map
	for i, hw := range d.pinConfigs {
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

	if d.BoardRevision() == type0ne {
		result["device"] = "/dev/i2c-0"
	} else {
		result["device"] = "/dev/i2c-1"
	}

	return result
}

func (d *RaspberryPiDTDriver) getLEDOptions(name string) map[string]interface{} {
	result := make(map[string]interface{})

	pins := make(DTLEDModulePins)
	pins["ok"] = "/sys/class/leds/led0/"

	result["pins"] = pins

	return result
}

// Determine the version of Raspberry Pi.
// This discussion http://www.raspberrypi.org/phpBB3/viewtopic.php?f=44&t=23989
// was used to determine the algorithm, specifically the comment by gordon@drogon.net
// It will return 1 or 2.
// https://elinux.org/RPi_HardwareHistory#Board_Revision_History
func (d *RaspberryPiDTDriver) BoardRevision() pinoutRevision {
	revision := CpuInfo(0, "Revision")
	switch revision {
	case "0002", "0003":
		return type0ne
	case "0010":
		return typeAplusBPlusZeroPi2
	}

	revision = CpuInfo(3, "Revision")
	switch revision {
	case "a02082", "a22082": // PI 3 Model B
		return typeAplusBPlusZeroPi2
	case "a020d3": // PI 3 Model B+
		return typeAplusBPlusZeroPi2
	}

	// Pi 2 boards have different strings, but pinout is the same as B+
	revision = CpuInfo(0, "CPU revision")
	switch revision {
	case "5":
		return typeAplusBPlusZeroPi2
	case "7": //PI zero +
		return typeAplusBPlusZeroPi2
	}

	return typeTwo
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

	for i, hw := range d.pinConfigs {
		pinMap.Add(Pin(i), hw.names, hw.modules)
	}

	return
}
