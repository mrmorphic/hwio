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
	case 1:
		d.pinConfigs = []*DTPinConfig{
			&DTPinConfig{[]string{"null"}, []string{"unassignable"}, 0, 0}, // 0 - spacer
			&DTPinConfig{[]string{"3.3v"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"5v"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"sda"}, []string{"i2c"}, 0, 0},
			&DTPinConfig{[]string{"do-not-connect-1"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"scl"}, []string{"i2c"}, 0, 0},
			&DTPinConfig{[]string{"ground"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"gpio4"}, []string{"gpio"}, 4, 0},
			&DTPinConfig{[]string{"txd"}, []string{"serial"}, 0, 0},
			&DTPinConfig{[]string{"do-not-connect-2"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"rxd"}, []string{"serial"}, 0, 0},
			&DTPinConfig{[]string{"gpio17"}, []string{"gpio"}, 17, 0},
			&DTPinConfig{[]string{"gpio18"}, []string{"gpio"}, 18, 0}, // also supports PWM
			&DTPinConfig{[]string{"gpio21"}, []string{"gpio"}, 21, 0},
			&DTPinConfig{[]string{"do-not-connect-3"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"gpio22"}, []string{"gpio"}, 22, 0},
			&DTPinConfig{[]string{"gpio23"}, []string{"gpio"}, 23, 0},
			&DTPinConfig{[]string{"do-not-connect-4"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"gpio24"}, []string{"gpio"}, 24, 0},
			&DTPinConfig{[]string{"mosi"}, []string{"spi"}, 0, 0},
			&DTPinConfig{[]string{"do-not-connect-5"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"miso"}, []string{"spi"}, 0, 0},
			&DTPinConfig{[]string{"gpio25"}, []string{"gpio"}, 25, 0},
			&DTPinConfig{[]string{"sclk"}, []string{"spi"}, 0, 0},
			&DTPinConfig{[]string{"ce0n"}, []string{"spi"}, 0, 0},
			&DTPinConfig{[]string{"do-not-connect-6"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"ce1n"}, []string{"spi"}, 0, 0},
		}
	case 2:
		d.pinConfigs = []*DTPinConfig{
			&DTPinConfig{[]string{"null"}, []string{"unassignable"}, 0, 0}, // 0 - spacer
			&DTPinConfig{[]string{"3.3v-1"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"5v-1"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"sda"}, []string{"i2c"}, 0, 0},
			&DTPinConfig{[]string{"5v-2"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"scl"}, []string{"i2c"}, 0, 0},
			&DTPinConfig{[]string{"ground-1"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"gpio4"}, []string{"gpio"}, 4, 0},
			&DTPinConfig{[]string{"txd"}, []string{"serial"}, 0, 0},
			&DTPinConfig{[]string{"ground-2"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"rxd"}, []string{"serial"}, 0, 0},
			&DTPinConfig{[]string{"gpio17"}, []string{"gpio"}, 17, 0},
			&DTPinConfig{[]string{"gpio18"}, []string{"gpio"}, 18, 0}, // also supports PWM
			&DTPinConfig{[]string{"gpio27"}, []string{"gpio"}, 27, 0},
			&DTPinConfig{[]string{"ground-3"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"gpio22"}, []string{"gpio"}, 22, 0},
			&DTPinConfig{[]string{"gpio23"}, []string{"gpio"}, 23, 0},
			&DTPinConfig{[]string{"3.3v-2"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"gpio24"}, []string{"gpio"}, 24, 0},
			&DTPinConfig{[]string{"mosi"}, []string{"spi"}, 0, 0},
			&DTPinConfig{[]string{"ground-4"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"miso"}, []string{"spi"}, 0, 0},
			&DTPinConfig{[]string{"gpio25"}, []string{"gpio"}, 25, 0},
			&DTPinConfig{[]string{"sclk"}, []string{"spi"}, 0, 0},
			&DTPinConfig{[]string{"gpio8"}, []string{"gpio"}, 8, 0},
			&DTPinConfig{[]string{"ground-5"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"gpio7"}, []string{"gpio"}, 7, 0},
		}
	default: // B+
		d.pinConfigs = []*DTPinConfig{
			&DTPinConfig{[]string{"null"}, []string{"unassignable"}, 0, 0}, // 0 - spacer
			&DTPinConfig{[]string{"3.3v-1"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"5v-1"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"sda"}, []string{"i2c"}, 0, 0},
			&DTPinConfig{[]string{"5v-2"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"scl"}, []string{"i2c"}, 0, 0},
			&DTPinConfig{[]string{"ground-1"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"gpio4"}, []string{"gpio"}, 4, 0},
			&DTPinConfig{[]string{"txd"}, []string{"serial"}, 0, 0},
			&DTPinConfig{[]string{"ground-2"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"rxd"}, []string{"serial"}, 0, 0},
			&DTPinConfig{[]string{"gpio17"}, []string{"gpio"}, 17, 0},
			&DTPinConfig{[]string{"gpio18"}, []string{"gpio"}, 18, 0}, // also supports PWM
			&DTPinConfig{[]string{"gpio27"}, []string{"gpio"}, 21, 0},
			&DTPinConfig{[]string{"ground-3"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"gpio22"}, []string{"gpio"}, 22, 0},
			&DTPinConfig{[]string{"gpio23"}, []string{"gpio"}, 23, 0},
			&DTPinConfig{[]string{"3.3v-2"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"gpio24"}, []string{"gpio"}, 24, 0},
			&DTPinConfig{[]string{"mosi"}, []string{"spi"}, 0, 0},
			&DTPinConfig{[]string{"ground-4"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"miso"}, []string{"spi"}, 0, 0},
			&DTPinConfig{[]string{"gpio25"}, []string{"gpio"}, 25, 0},
			&DTPinConfig{[]string{"sclk"}, []string{"spi"}, 0, 0},
			&DTPinConfig{[]string{"gpio8"}, []string{"gpio"}, 8, 0},
			&DTPinConfig{[]string{"ground-5"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"gpio7"}, []string{"gpio"}, 7, 0},
			&DTPinConfig{[]string{"do-not-connect-1"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"do-not-connect-2"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"gpio5"}, []string{"gpio"}, 5, 0},
			&DTPinConfig{[]string{"ground-6"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"gpio6"}, []string{"gpio"}, 6, 0},
			&DTPinConfig{[]string{"gpio12"}, []string{"gpio"}, 12, 0},
			&DTPinConfig{[]string{"gpio13"}, []string{"gpio"}, 13, 0},
			&DTPinConfig{[]string{"ground-7"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"gpio19"}, []string{"gpio"}, 19, 0},
			&DTPinConfig{[]string{"gpio16"}, []string{"gpio"}, 16, 0},
			&DTPinConfig{[]string{"gpio26"}, []string{"gpio"}, 26, 0},
			&DTPinConfig{[]string{"gpio20"}, []string{"gpio"}, 20, 0},
			&DTPinConfig{[]string{"ground-8"}, []string{"unassignable"}, 0, 0},
			&DTPinConfig{[]string{"gpio21"}, []string{"gpio"}, 21, 0},
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

	if d.BoardRevision() == 1 {
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
func (d *RaspberryPiDTDriver) BoardRevision() int {
	revision := CpuInfo(0, "Revision")
	switch revision {
	case "0002", "0003":
		return 1
	case "0010":
		return 3
	}

	// Pi 2 boards have different strings, but pinout is the same as B+
	revision = CpuInfo(0, "CPU revision")
	switch revision {
	case "5":
		return 3
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

	for i, hw := range d.pinConfigs {
		pinMap.add(Pin(i), hw.names, hw.modules)
	}

	return
}
