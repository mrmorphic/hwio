package hwio

// A user-land driver for Raspberry Pi
//
// Things known to work (tested on hardware):
// - nothing yet
//
// WARNINGS:
// - THIS IS STILL UNDER DEVELOPMENT
// - UNTESTED FEATURES MAY FRY YOUR BOARD
// - ANY CHANGES YOU MAKE TO THIS MAY FRY YOUR BOARD
// Don't say you weren't warned.

// @todo Implement GPIO output
// @todo Implement GPIO input

import (
	"os"
	"strconv"
	"syscall"
	"errors"
	"fmt"
	"time"
)

// Represents info we need to know about a pin on the Pi.
// @todo Determine if 'hwPin' is required
type RaspberryPiPin struct {
	hwPin     string // This intended for the P8.16 format name (currently unused)
	profile   []Capability
	gpioName  string // This is used for a human readable name
	bit       uint   // A single bit in the position of the I/O value on the port
	mode0Name string // mode 0 signal name, used by the muxer
}

func (p RaspberryPiPin) GetName() string {
	return p.gpioName
}

var piPins []*RaspberryPiPin
var gpioProfile []Capability
var unusedProfile []Capability

func init() {
	gpioProfile = []Capability{
		CAP_OUTPUT,
		CAP_INPUT,
		CAP_INPUT_PULLUP,
		CAP_INPUT_PULLDOWN,
	}
	unusedProfile = []Capability {
	}

	// The pins are numbered as they are on the connector. This means introducing
	// artificial pins for things like power, to keep the numbering.
	p := []*RaspberryPiPin{
		&RaspberryPiPin{"NULL", unusedProfile, "", 0, "gpmc_ad6"}, // 0 - spacer
		&RaspberryPiPin{"3.3V", unusedProfile, "", 0, "gpmc_ad6"},
		&RaspberryPiPin{"5V", unusedProfile, "", 0, "gpmc_ad7"},
		&RaspberryPiPin{"SDA", unusedProfile, "GPIO0", 1 << 0, "gpmc_ad2"}, //also gpio
		&RaspberryPiPin{"DONOTCONNECT1", unusedProfile, "", 0, "gpmc_ad3"},
		&RaspberryPiPin{"SCL", unusedProfile, "GPIO1", 1 << 1, "gpmc_ad3"}, // also gpio
		&RaspberryPiPin{"GROUND", unusedProfile, "", 0, "gpmc_advn_ale"},
		&RaspberryPiPin{"GPIO4", gpioProfile, "GPIO4", 1 << 4, "gpmc_oen_ren"},
		&RaspberryPiPin{"TXD", unusedProfile, "GPIO14", 1 << 14, "gpmc_ben0_cle"},
		&RaspberryPiPin{"DONOTCONNECT2", unusedProfile, "", 0, "gpmc_wen"},
		&RaspberryPiPin{"RXD", unusedProfile, "GPIO15", 1 << 15, "gpmc_ad13"},
		&RaspberryPiPin{"GPIO17", gpioProfile, "GPIO17", 1 << 17, "gpmc_ad12"},
		&RaspberryPiPin{"GPIO18", gpioProfile, "GPIO18", 1 << 18, "gpmc_ad9"}, // also supports PWM
		&RaspberryPiPin{"GPIO21", gpioProfile, "GPIO21", 1 << 21, "gpmc_ad10"},
		&RaspberryPiPin{"DONOTCONNECT3", unusedProfile, "", 0, "gpmc_ad15"},
		&RaspberryPiPin{"GPIO22", gpioProfile, "GPIO22", 1 << 22, "gpmc_ad14"},
		&RaspberryPiPin{"GPIO23", gpioProfile, "GPIO23", 1 << 23, "gpmc_ad11"},
		&RaspberryPiPin{"DONOTCONNECT4", unusedProfile, "", 0, "gpmc_clk"},
		&RaspberryPiPin{"GPIO24", gpioProfile, "GPIO24", 1 << 24, "gpmc_ad8"},
		&RaspberryPiPin{"MOSI", unusedProfile, "GPIO10", 1 << 10, "gpmc_csn2"},
		&RaspberryPiPin{"DONOTCONNECT5", unusedProfile, "", 0, "gpmc_csn1"},
		&RaspberryPiPin{"MISO", unusedProfile, "GPIO9", 1 << 9, "gpmc_ad5"},
		&RaspberryPiPin{"GPIO25", gpioProfile, "GPIO25", 1 << 25, "gpmc_ad4"},
		&RaspberryPiPin{"SCLK", unusedProfile, "GPIO11", 1 << 11, "gpmc_ad1"},
		&RaspberryPiPin{"CE0N", unusedProfile, "GPIO8", 1 << 8, "gpmc_ad0"},
		&RaspberryPiPin{"DONOTCONNECT6", unusedProfile, "", 0, "gpmc_csn0"},
		&RaspberryPiPin{"CE1N", unusedProfile, "GPIO7", 1 << 7, "lcd_vsync"},
	}
	piPins = p
}

type RaspberryPiDriver struct {
	// Mapped memory for directly accessing hardware registers
	mmap []byte
}

func (d *RaspberryPiDriver) Init() error {
	// Set up the memory mapped file giving us access to hardware registers
	file, e := os.OpenFile("/dev/mem", os.O_RDWR|os.O_APPEND, 0)
	if e != nil {
		return e
	}
	mmap, e := syscall.Mmap(int(file.Fd()), MMAP_OFFSET, MMAP_SIZE, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if e != nil {
		return e
	}
	d.mmap = mmap

	d.analogInit()

	return nil
}

func (d *RaspberryPiDriver) Close() {
	syscall.Munmap(d.mmap)
}

func (d *RaspberryPiDriver) PinMode(pin Pin, mode PinIOMode) error {
	p := piPins[pin]

	if mode == OUTPUT {
		e := d.pinMux(p.mode0Name, CONF_GPIO_OUTPUT)
		if e != nil {
			return e
		}

		d.clearRegL(p.port+uint(GPIO_OE), p.bit)
	} else {
		pull := CONF_PULL_DISABLE
		// note: pull up/down modes assume that CONF_PULLDOWN resets the pull disable bit
		if mode == INPUT_PULLUP {
			pull = CONF_PULLUP
		} else if mode == INPUT_PULLDOWN {
			pull = CONF_PULLDOWN
		}

		e := d.pinMux(p.mode0Name, CONF_GPIO_INPUT|uint(pull))
		if e != nil {
			return e
		}

//		fmt.Printf("R/W dir reg BEFORE value is %x\n", d.getRegL(p.port+uint(GPIO_OE)))

		d.orRegL(p.port+uint(GPIO_OE), p.bit)
//		fmt.Printf("R/W dir reg AFTER value is %x\n", d.getRegL(p.port+uint(GPIO_OE)))
	}
	return nil
}

func (d *RaspberryPiDriver) pinMux(mux string, mode uint) error {
	// Uses kernel omap_mux files to set pin modes.
	// There's no simple way to write the control module registers from a 
	// user-level process because it lacks the proper privileges, but it's 
	// easy enough to just use the built-in file-based system and let the 
	// kernel do the work. 
	f, e := os.OpenFile(PINMUX_PATH+mux, os.O_WRONLY|os.O_TRUNC, 0666)
	if e != nil {
		return e
	}

	s := strconv.FormatInt(int64(mode), 16)
//	fmt.Printf("Writing mode %s to mux file %s\n", s, PINMUX_PATH+mux)
	f.WriteString(s)
	return nil
}

func (d *RaspberryPiDriver) DigitalWrite(pin Pin, value int) (e error) {
	p := piPins[pin]
	if value == 0 {
		d.clearRegL(p.port+GPIO_DATAOUT, p.bit)
	} else {
		d.orRegL(p.port+GPIO_DATAOUT, p.bit)
	}
	return nil
}

func (d *RaspberryPiDriver) DigitalRead(pin Pin) (value int, e error) {
	p := piPins[pin]
	reg := d.getRegL(p.port+GPIO_DATAIN)
	//	fmt.Printf("\nraw in: %x (checking bit %d)\n", reg, p.bit)
	if (reg & p.bit) != 0 {
		return HIGH, nil
	}
	return LOW, nil
}

func (d *RaspberryPiDriver) AnalogWrite(pin Pin, value int) (e error) {
	return nil
}

func (d *RaspberryPiDriver) AnalogRead(pin Pin) (value int, e error) {
	return 0, errors.New("Analog input is not supported")
}

// Sets 32 bit Register at address to its current value AND mask.
func (d *RaspberryPiDriver) andRegL(address uint, mask uint) {
	d.setRegL(address, d.getRegL(address)&mask)
}

// Sets 32 bit Register at address to its current value OR mask.
func (d *RaspberryPiDriver) orRegL(address uint, mask uint) {
	d.setRegL(address, d.getRegL(address)|mask)
}

// Clears mask bits in 32 bit register at given address.
func (d *RaspberryPiDriver) clearRegL(address uint, mask uint) {
	d.andRegL(address, ^mask)
}

// Returns unpacked 32 bit register value starting from address. Integers
// are little endian on BeagleBone
func (d *RaspberryPiDriver) getRegL(address uint) (result uint) {
	result = uint(d.mmap[address])
	result |= uint(d.mmap[address+1])<<8
	result |= uint(d.mmap[address+2])<<16
	result |= uint(d.mmap[address+3])<<24
	return result
}

func (d *RaspberryPiDriver) setRegL(address uint, value uint) {
	d.mmap[address] = byte(value & 0xff)
	d.mmap[address+1] = byte((value >> 8) & 0xff)
	d.mmap[address+2] = byte((value >> 16) & 0xff)
	d.mmap[address+3] = byte((value >> 24) & 0xff)
}

func (d *RaspberryPiDriver) PinMap() (pinMap HardwarePinMap) {
	pinMap = make(HardwarePinMap)

	for i, hw := range piPins {
		names := []string{hw.hwPin}
		if hw.hwPin != hw.gpioName {
			names = append(names, hw.gpioName)
		}
		pinMap.add(Pin(i), names, hw.profile)
	}

	return
}

func (d *RaspberryPiDriver) analogInit() {
}

