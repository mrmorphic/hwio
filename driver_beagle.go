package hwio

// A driver for BeagleBone, based on PyBBIO at:
// 		https://github.com/alexanderhiam/PyBBIO/blob/master/bbio/config.py
// The hardware control logic has been ported to Go, using a memory mapped file
// to get at the control registers direct.
//
// This driver is very specific to BeagleBone (I've built to revision A5) and the
// TI chip that powers it. It may work on other Beagle boards but this is
// completely untested.
//
// Things known to work (tested on hardware):
// - digital output on all GPIO pins that are exposed on P8 and P9 of the board, and USR0 to USR3
// - digital input on all GPIO pins that are exposed on P8 and P9 of the board, including pull-up,
//	 pull-down and pull-disabled modes.
//
// WARNINGS:
// - THIS IS STILL UNDER DEVELOPMENT
// - UNTESTED FEATURES MAY FRY YOUR BOARD
// - ANY CHANGES YOU MAKE TO THIS MAY FRY YOUR BOARD
// Don't say you weren't warned.

// @todo: Use set and clear register locations instead for digital write, instead of bit manipulation
// @todo: Analog pin support
// @todo: Timers
// @todo: Interupts
// @todo: PWM support

import (
	"os"
	"strconv"
	"syscall"
	"errors"
	"fmt"
	"time"
)

// Represents info we need to know about a pin on the BeagleBone.
// @todo Determine if 'hwPin' is required
type BeaglePin struct {
	hwPin     string // This intended for the P8.16 format name (currently unused)
	gpioName  string // This is used for a human readable name
	port      uint   // The GPIO port
	bit       uint   // A single bit in the position of the I/O value on the port
	mode0Name string // mode 0 signal name, used by the muxer
	adcEnable uint   // bit mask for analog pin control
}

func (p BeaglePin) GetName() string {
	return p.gpioName
}

// internal function to identify an analog pin from config
func (p BeaglePin) isAnalogPin() bool {
	return p.adcEnable != 0
}

const (
	MMAP_OFFSET = 0x44c00000
	MMAP_SIZE   = 0x48ffffff - MMAP_OFFSET

	GPIO0 = 0x44e07000 - MMAP_OFFSET
	GPIO1 = 0x4804c000 - MMAP_OFFSET
	GPIO2 = 0x481ac000 - MMAP_OFFSET
	GPIO3 = 0x481ae000 - MMAP_OFFSET

	//	CM_PER = 0x44e00000-MMAP_OFFSET
	CM_WKUP = 0x44e00400-MMAP_OFFSET

	//	CM_PER_EPWMSS0_CLKCTRL = 0xd4+CM_PER
	//	CM_PER_EPWMSS1_CLKCTRL = 0xcc+CM_PER
	//	CM_PER_EPWMSS2_CLKCTRL = 0xd8+CM_PER

	CM_WKUP_ADC_TSC_CLKCTRL = 0xbc+CM_WKUP

	MODULEMODE_ENABLE = 0x02
	IDLEST_MASK = 0x03<<16

	// //# To enable module clock:
	// //#  _setReg(CM_WKUP_module_CLKCTRL, MODULEMODE_ENABLE)
	// //#  while (_getReg(CM_WKUP_module_CLKCTRL) & IDLEST_MASK): pass
	// //# To disable module clock:
	// //#  _andReg(CM_WKUP_module_CLKCTRL, ~MODULEMODE_ENABLE)
	// //#-----------------------------
	PINMUX_PATH = "/sys/kernel/debug/omap_mux/"

	CONF_SLEW_SLOW    = 1 << 6
	CONF_RX_ACTIVE    = 1 << 5
	CONF_PULLUP       = 1 << 4
	CONF_PULLDOWN     = 0x00
	CONF_PULL_DISABLE = 1 << 3

	CONF_GPIO_MODE   = 0x07
	CONF_GPIO_OUTPUT = CONF_GPIO_MODE
	CONF_GPIO_INPUT  = CONF_GPIO_MODE + CONF_RX_ACTIVE
	// CONF_ADC_PIN     = CONF_RX_ACTIVE+CONF_PULL_DISABLE

	// CONF_UART_TX     = CONF_PULLUP
	// CONF_UART_RX     = CONF_RX_ACTIVE

	GPIO_OE      = 0x134
	GPIO_DATAIN  = 0x138
	GPIO_DATAOUT = 0x13c

// GPIO_CLEARDATAOUT = 0x190
// GPIO_SETDATAOUT   = 0x194


	// Start of ADC config
	ADC_TSC = 0x44e0d000-MMAP_OFFSET

	// SYSCONFIG register
	ADC_SYSCONFIG = ADC_TSC+0x10

	// @todo find out the meaning of ADC_SOFTRESET. This bit in SYSCONFIG register is marked unused
	ADC_SOFTRESET = 0x01

	// CTLR register
	ADC_CTRL = ADC_TSC+0x40

	// Bit in CTRL to disable write protect on step config registers
	ADC_STEPCONFIG_WRITE_PROTECT_OFF = 0x01<<2
// # Write protect default on, must first turn off to change stepconfig:
// #  _setReg(ADC_CTRL, ADC_STEPCONFIG_WRITE_PROTECT_OFF)
// # To set write protect on:
// #  _clearReg(ADC_CTRL, ADC_STEPCONFIG_WRITE_PROTECT_OFF)

	// Bit in CTRL to enable ADC, which should be done after setting up other ADC registers
	TSC_ADC_SS_ENABLE = 0x01
// # To enable:
// # _setReg(ADC_CTRL, TSC_ADC_SS_ENABLE)
// #  This will turn STEPCONFIG write protect back on 
// # To keep write protect off:
// # _orReg(ADC_CTRL, TSC_ADC_SS_ENABLE)
// #----------------

// ADC_CLKDIV = ADC_TSC+0x4c  # Write desired value-1

	// STEPENABLE register
	ADC_STEPENABLE = ADC_TSC+0x54

// ADC_ENABLE = lambda AINx: 0x01<<(ADC[AINx]+1)
// #----------------------

// ADC_IDLECONFIG = ADC_TSC+0x58

	// ADC STEPCONFIG registers
	ADCSTEPCONFIG1 = ADC_TSC+0x64
	ADCSTEPDELAY1  = ADC_TSC+0x68
	ADCSTEPCONFIG2 = ADC_TSC+0x6c
	ADCSTEPDELAY2  = ADC_TSC+0x70
	ADCSTEPCONFIG3 = ADC_TSC+0x74
	ADCSTEPDELAY3  = ADC_TSC+0x78
	ADCSTEPCONFIG4 = ADC_TSC+0x7c
	ADCSTEPDELAY4  = ADC_TSC+0x80
	ADCSTEPCONFIG5 = ADC_TSC+0x84
	ADCSTEPDELAY5  = ADC_TSC+0x88
	ADCSTEPCONFIG6 = ADC_TSC+0x8c
	ADCSTEPDELAY6  = ADC_TSC+0x90
	ADCSTEPCONFIG7 = ADC_TSC+0x94
	ADCSTEPDELAY7  = ADC_TSC+0x98
	ADCSTEPCONFIG8 = ADC_TSC+0x9c
	ADCSTEPDELAY8  = ADC_TSC+0xa0
// # Only need the first 8 steps - 1 for each AIN pin

// ADC_RESET = 0x00 # Default value of STEPCONFIG

// ADC_AVG2  = 0x01<<2
// ADC_AVG4  = 0x02<<2
// ADC_AVG8  = 0x03<<2
// ADC_AVG16 = 0x04<<2

// #SEL_INP = lambda AINx: (ADC[AINx]+1)<<19
// # Set input with _orReg(ADCSTEPCONFIGx, SEL_INP(AINx))
// # ADC[AINx]+1 because positive AMUX input 0 is VREFN 
// #  (see user manual section 12.3.7)
// SEL_INP = lambda AINx: (ADC[AINx])<<19

// SAMPLE_DELAY = lambda cycles: (cycles&0xff)<<24
// # SAMPLE_DELAY is the number of cycles to sample for
// # Set delay with _orReg(ADCSTEPDELAYx, SAMPLE_DELAY(cycles))

// #----------------------

	// ADC FIFO
	ADC_FIFO0DATA = ADC_TSC+0x100
	ADC_FIFO_MASK = 0xfff

// ## ADC pins:

// # And some constants so the user doesn't need to use strings:
// AIN0 = A0 = 'AIN0'
// AIN1 = A1 = 'AIN1'
// AIN2 = A2 = 'AIN2'
// AIN3 = A3 = 'AIN3'
// AIN4 = A4 = 'AIN4'
// AIN5 = A5 = 'AIN5'
// AIN6 = A6 = 'AIN6'
// AIN7 = A7 = VSYS = 'AIN7'

// ##--- End ADC config -------##
// ##############################

// ##############################
// ##--- Start UART config: ---##

// # UART ports must be in form: 
// #    [port, tx_pinmux_filename, tx_pinmux_mode, 
// #           rx_pinmux_filename, rx_pinmux_mode]

// UART = {
//   'UART1' : ['/dev/ttyO1', 'uart1_txd', 0,  'uart1_rxd', 0],
//   'UART2' : ['/dev/ttyO2',   'spi0_d0', 1,  'spi0_sclk', 1],
//   'UART4' : ['/dev/ttyO4',  'gpmc_wpn', 6, 'gpmc_wait0', 6],
//   'UART5' : ['/dev/ttyO5', 'lcd_data8', 4,  'lcd_data9', 4]
// }

// # Formatting constants to mimic Arduino's serial.print() formatting:
// DEC = 'DEC'
// BIN = 'BIN'
// OCT = 'OCT'
// HEX = 'HEX'

)

var beaglePins []*BeaglePin

func init() {
	// Note: Logical pin numbers are implicitly assigned from 0 in the order in
	// which they occur in this slice. i.e. Pin 0 will be gpmc_a5, pin 1 will be
	// gpmc_a5 and so on.
	// @todo Review the actual pins on the BeagleBone and see if there are others.
	// @todo Review for correctness against specs. Notable mistake is USR0 and USR1 having the same pin mask
	p := []*BeaglePin{
		// P8
		&BeaglePin{"P8.3", "GPIO1_6", GPIO1, 1 << 6, "gpmc_ad6", 0},
		&BeaglePin{"P8.4", "GPIO1_7", GPIO1, 1 << 7, "gpmc_ad7", 0},
		&BeaglePin{"P8.5", "GPIO1_2", GPIO1, 1 << 2, "gpmc_ad2", 0},
		&BeaglePin{"P8.6", "GPIO1_3", GPIO1, 1 << 3, "gpmc_ad3", 0},
		&BeaglePin{"P8.7", "GPIO2_2", GPIO2, 1 << 2, "gpmc_advn_ale", 0},
		&BeaglePin{"P8.8", "GPIO2_3", GPIO2, 1 << 3, "gpmc_oen_ren", 0},
		&BeaglePin{"P8.9", "GPIO2_5", GPIO2, 1 << 5, "gpmc_ben0_cle", 0},
		&BeaglePin{"P8.10", "GPIO2_4", GPIO2, 1 << 4, "gpmc_wen", 0},
		&BeaglePin{"P8.11", "GPIO1_13", GPIO1, 1 << 13, "gpmc_ad13", 0},
		&BeaglePin{"P8.12", "GPIO1_12", GPIO1, 1 << 12, "gpmc_ad12", 0},
		&BeaglePin{"P8.13", "GPIO0_23", GPIO0, 1 << 23, "gpmc_ad9", 0},
		&BeaglePin{"P8.14", "GPIO0_26", GPIO0, 1 << 26, "gpmc_ad10", 0},
		&BeaglePin{"P8.15", "GPIO1_15", GPIO1, 1 << 15, "gpmc_ad15", 0},
		&BeaglePin{"P8.16", "GPIO1_14", GPIO1, 1 << 14, "gpmc_ad14", 0},
		&BeaglePin{"P8.17", "GPIO0_27", GPIO0, 1 << 27, "gpmc_ad11", 0},
		&BeaglePin{"P8.18", "GPIO2_1", GPIO2, 1 << 1, "gpmc_clk", 0},
		&BeaglePin{"P8.19", "GPIO0_22", GPIO0, 1 << 22, "gpmc_ad8", 0},
		&BeaglePin{"P8.20", "GPIO1_31", GPIO1, 1 << 31, "gpmc_csn2", 0},
		&BeaglePin{"P8.21", "GPIO1_30", GPIO1, 1 << 30, "gpmc_csn1", 0},
		&BeaglePin{"P8.22", "GPIO1_5", GPIO1, 1 << 5, "gpmc_ad5", 0},
		&BeaglePin{"P8.23", "GPIO1_4", GPIO1, 1 << 4, "gpmc_ad4", 0},
		&BeaglePin{"P8.24", "GPIO1_1", GPIO1, 1 << 1, "gpmc_ad1", 0},
		&BeaglePin{"P8.25", "GPIO1_0", GPIO1, 1, "gpmc_ad0", 0},
		&BeaglePin{"P8.26", "GPIO1_29", GPIO1, 1 << 29, "gpmc_csn0", 0},
		&BeaglePin{"P8.27", "GPIO2_22", GPIO2, 1 << 22, "lcd_vsync", 0},
		&BeaglePin{"P8.28", "GPIO2_24", GPIO2, 1 << 24, "lcd_pclk", 0},
		&BeaglePin{"P8.29", "GPIO2_23", GPIO2, 1 << 23, "lcd_hsync", 0},
		&BeaglePin{"P8.30", "GPIO2_25", GPIO2, 1 << 25, "lcd_ac_bias_en", 0},
		&BeaglePin{"P8.31", "GPIO0_10", GPIO0, 1 << 10, "lcd_data14", 0},
		&BeaglePin{"P8.32", "GPIO0_11", GPIO0, 1 << 11, "lcd_data15", 0},
		&BeaglePin{"P8.33", "GPIO0_9", GPIO0, 1 << 9, "lcd_data13", 0},
		&BeaglePin{"P8.34", "GPIO2_17", GPIO2, 1 << 17, "lcd_data11", 0},
		&BeaglePin{"P8.35", "GPIO0_8", GPIO0, 1 << 8, "lcd_data12", 0},
		&BeaglePin{"P8.36", "GPIO2_16", GPIO2, 1 << 16, "lcd_data10", 0},
		&BeaglePin{"P8.37", "GPIO2_14", GPIO2, 1 << 14, "lcd_data8", 0},
		&BeaglePin{"P8.38", "GPIO2_15", GPIO2, 1 << 15, "lcd_data9", 0},
		&BeaglePin{"P8.39", "GPIO2_12", GPIO2, 1 << 12, "lcd_data6", 0},
		&BeaglePin{"P8.40", "GPIO2_13", GPIO2, 1 << 13, "lcd_data7", 0},
		&BeaglePin{"P8.41", "GPIO2_10", GPIO2, 1 << 10, "lcd_data4", 0},
		&BeaglePin{"P8.42", "GPIO2_11", GPIO2, 1 << 11, "lcd_data5", 0},
		&BeaglePin{"P8.43", "GPIO2_8", GPIO2, 1 << 8, "lcd_data2", 0},
		&BeaglePin{"P8.44", "GPIO2_9", GPIO2, 1 << 9, "lcd_data3", 0},
		&BeaglePin{"P8.45", "GPIO2_6", GPIO2, 1 << 6, "lcd_data0", 0},
		&BeaglePin{"P8.46", "GPIO2_7", GPIO2, 1 << 7, "lcd_data1", 0},

		// P9
		&BeaglePin{"P9.11", "GPIO0_30", GPIO0, 1 << 30, "gpmc_wait0", 0},
		&BeaglePin{"P9.12", "GPIO1_28", GPIO1, 1 << 28, "gpmc_ben1", 0},
		&BeaglePin{"P9.13", "GPIO0_31", GPIO0, 1 << 31, "gpmc_wpn", 0},
		&BeaglePin{"P9.14", "GPIO1_18", GPIO1, 1 << 18, "gpmc_a2", 0},
		&BeaglePin{"P9.15", "GPIO1_16", GPIO1, 1 << 16, "gpmc_a0", 0},
		&BeaglePin{"P9.16", "GPIO1_19", GPIO1, 1 << 19, "gpmc_a3", 0},
		&BeaglePin{"P9.17", "GPIO0_5", GPIO0, 1 << 5, "spi0_cs0", 0},
		&BeaglePin{"P9.18", "GPIO0_4", GPIO0, 1 << 4, "spi0_d1", 0},
		&BeaglePin{"P9.19", "GPIO0_13", GPIO0, 1 << 13, "uart1_rtsn", 0},
		&BeaglePin{"P9.20", "GPIO0_12", GPIO0, 1 << 12, "uart1_ctsn", 0},
		&BeaglePin{"P9.21", "GPIO0_3", GPIO0, 1 << 3, "spi0_d0", 0},
		&BeaglePin{"P9.22", "GPIO0_2", GPIO0, 1 << 2, "spi0_sclk", 0},
		&BeaglePin{"P9.23", "GPIO1_17", GPIO1, 1 << 17, "gpmc_a1", 0},
		&BeaglePin{"P9.24", "GPIO0_15", GPIO0, 1 << 15, "uart1_txd", 0},
		&BeaglePin{"P9.25", "GPIO3_21", GPIO3, 1 << 21, "mcasp0_ahclkx", 0},
		&BeaglePin{"P9.26", "GPIO0_14", GPIO0, 1 << 14, "uart1_rxd", 0},
		&BeaglePin{"P9.27", "GPIO3_19", GPIO3, 1 << 19, "mcasp0_fsr", 0},
		&BeaglePin{"P9.28", "GPIO3_17", GPIO3, 1 << 17, "mcasp0_ahclkr", 0},
		&BeaglePin{"P9.29", "GPIO3_15", GPIO3, 1 << 15, "mcasp0_fsx", 0},
		&BeaglePin{"P9.30", "GPIO3_16", GPIO3, 1 << 16, "mcasp0_axr0", 0},
		&BeaglePin{"P9.31", "GPIO3_14", GPIO3, 1 << 14, "mcasp0_aclkx", 0},
		&BeaglePin{"P9.33", "AIN4", 0, 0, "ain4", 1<<5},
		&BeaglePin{"P9.35", "AIN6", 0, 0, "ain6", 1<<7},
		&BeaglePin{"P9.36", "AIN5", 0, 0, "ain5", 1<<6},
		&BeaglePin{"P9.37", "AIN2", 0, 0, "ain2", 1<<3},
		&BeaglePin{"P9.38", "AIN3", 0, 0, "ain3", 1<<4},
		&BeaglePin{"P9.39", "AIN0", 0, 0, "ain0", 1<<1},
		&BeaglePin{"P9.40", "AIN1", 0, 0, "ain1", 1<<2},
		&BeaglePin{"P9.41", "GPIO0_20", GPIO0, 1 << 20, "xdma_event_intr1", 0},
		&BeaglePin{"P9.42", "GPIO0_7", GPIO0, 1 << 7, "ecap0_in_pwm0_out", 0},

		// USR LEDs
		&BeaglePin{"USR0", "USR0", GPIO1, 1 << 21, "gpmc_a5", 0},
		&BeaglePin{"USR1", "USR1", GPIO1, 1 << 22, "gpmc_a6", 0},
		&BeaglePin{"USR2", "USR2", GPIO1, 1 << 23, "gpmc_a7", 0},
		&BeaglePin{"USR3", "USR3", GPIO1, 1 << 24, "gpmc_a8", 0},
	}
	beaglePins = p
}

type BeagleBoneDriver struct {
	// Mapped memory for directly accessing hardware registers
	mmap []byte
}

func (d *BeagleBoneDriver) Init() error {
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
	//	with open(MEM_FILE, "r+b") as f:
	//  __mmap = mmap(f.fileno(), MMAP_SIZE, offset=MMAP_OFFSET)
}

func (d *BeagleBoneDriver) Close() {
	syscall.Munmap(d.mmap)
}

func (d *BeagleBoneDriver) PinMode(pin Pin, mode PinIOMode) error {
	p := beaglePins[pin]

	// handle analog first, they are simplest from PinMode perspective
	if p.isAnalogPin() {
		if mode != INPUT {
			return errors.New(fmt.Sprintf("Pin %d is an analog pin, and the mode must be INPUT", p))
		}
		return nil	// nothing to set up
	}

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

func (d *BeagleBoneDriver) pinMux(mux string, mode uint) error {
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

func (d *BeagleBoneDriver) DigitalWrite(pin Pin, value int) (e error) {
	p := beaglePins[pin]
	if value == 0 {
		d.clearRegL(p.port+GPIO_DATAOUT, p.bit)
	} else {
		d.orRegL(p.port+GPIO_DATAOUT, p.bit)
	}
	return nil
}

func (d *BeagleBoneDriver) DigitalRead(pin Pin) (value int, e error) {
	p := beaglePins[pin]
	reg := d.getRegL(p.port+GPIO_DATAIN)
	//	fmt.Printf("\nraw in: %x (checking bit %d)\n", reg, p.bit)
	if (reg & p.bit) != 0 {
		return HIGH, nil
	}
	return LOW, nil
}

func (d *BeagleBoneDriver) AnalogWrite(pin Pin, value int) (e error) {
	return nil
}

func (d *BeagleBoneDriver) AnalogRead(pin Pin) (value int, e error) {
//	if d.getRegL(CM_WKUP_ADC_TSC_CLKCTRL) & IDLEST_MASK == 0 {
//		// if for any reason the ADC module clock has been shut off, turn it back on
//		d.analogInit()
//	}

	p := beaglePins[pin]

	// if (_getReg(CM_WKUP_ADC_TSC_CLKCTRL) & IDLEST_MASK):
	//   # The ADC module clock has been shut off, e.g. by a different 
	//   # PyBBIO script stopping while this one was running, turn back on:
	//   _analog_init() 

	// ADC_ENABLE = lambda AINx: 0x01<<(ADC[AINx]+1)
	d.setRegL(ADC_STEPENABLE, p.adcEnable)
	// # Enable sequncer step that's set for given input:
	// _setReg(ADC_STEPENABLE, ADC_ENABLE(analog_pin))

	for d.getRegL(ADC_STEPENABLE) & p.adcEnable != 0 { }

	// # Sequencer starts automatically after enabling step, wait for complete:
	// while(_getReg(ADC_STEPENABLE) & ADC_ENABLE(analog_pin)): pass
	// # Return 12-bit value from the ADC FIFO register:
	// return _getReg(ADC_FIFO0DATA) & ADC_FIFO_MASK
	res := d.getRegL(ADC_FIFO0DATA) & ADC_FIFO_MASK
	fmt.Printf("register output %x\n", res)
	return int(res), nil
}

// def inVolts(adc_value, bits=12, vRef=1.8):
//   """ Converts and returns the given ADC value to a voltage according
//       to the given number of bits and reference voltage. """
//   return adc_value*(vRef/2**bits)

// Sets 32 bit Register at address to its current value AND mask.
func (d *BeagleBoneDriver) andRegL(address uint, mask uint) {
	d.setRegL(address, d.getRegL(address)&mask)
}

// Sets 32 bit Register at address to its current value OR mask.
func (d *BeagleBoneDriver) orRegL(address uint, mask uint) {
	d.setRegL(address, d.getRegL(address)|mask)
}

// Clears mask bits in 32 bit register at given address.
func (d *BeagleBoneDriver) clearRegL(address uint, mask uint) {
	d.andRegL(address, ^mask)
}

// Returns unpacked 32 bit register value starting from address. Integers
// are little endian on BeagleBone
func (d *BeagleBoneDriver) getRegL(address uint) (result uint) {
	result = uint(d.mmap[address])
	result |= uint(d.mmap[address+1])<<8
	result |= uint(d.mmap[address+2])<<16
	result |= uint(d.mmap[address+3])<<24
	return result
}

func (d *BeagleBoneDriver) setRegL(address uint, value uint) {
	d.mmap[address] = byte(value & 0xff)
	d.mmap[address+1] = byte((value >> 8) & 0xff)
	d.mmap[address+2] = byte((value >> 16) & 0xff)
	d.mmap[address+3] = byte((value >> 24) & 0xff)
}

func (d *BeagleBoneDriver) PinMap() (pinMap HardwarePinMap) {
	gpioCap := []Capability{
		CAP_OUTPUT,
		CAP_INPUT,
		CAP_INPUT_PULLUP,
		CAP_INPUT_PULLDOWN,
	}
	//	analog := []Capability {CAP_INPUT,CAP_OUTPUT,CAP_ANALOG_IN}
	//	pwm := []Capability {CAP_INPUT,CAP_OUTPUT,CAP_PWM}
	//	readonly := []Capability {CAP_INPUT}

	pinMap = make(HardwarePinMap)

	for i, hw := range beaglePins {
		// @todo select profile based on extra info added to beaglePins. Notable
		// exception is analog pins.
		profile := gpioCap
		pinMap.add(Pin(i), hw.gpioName, profile)
	}

	return
}

// # _UART_PORT is a wrapper class for pySerial to enable Arduino-like access
// # to the UART1, UART2, UART4, and UART5 serial ports on the expansion headers:
// class _UART_PORT(object):
//   def __init__(self, uart):
//     assert uart in UART, "*Invalid UART: %s" % uart
//     self.config = uart
//     self.baud = 0
//     self.open = False
//     self.ser_port = None
//     self.peek_char = ''

//   def begin(self, baud, timeout=1):
//     """ Starts the serial port at the given baud rate. """
//     # Set proper pinmux to match expansion headers:
//     tx_pinmux_filename = UART[self.config][1]
//     tx_pinmux_mode     = UART[self.config][2]+CONF_UART_TX
//     _pinMux(tx_pinmux_filename, tx_pinmux_mode)

//     rx_pinmux_filename = UART[self.config][3]
//     rx_pinmux_mode     = UART[self.config][4]+CONF_UART_RX
//     _pinMux(rx_pinmux_filename, rx_pinmux_mode)    

//     port = UART[self.config][0]
//     self.baud = baud
//     self.ser_port = serial.Serial(port, baud, timeout=timeout)
//     self.open = True 

//   def end(self):
//     """ Closes the serial port if open. """
//     if not(self.open): return
//     self.flush()
//     self.ser_port.close()
//     self.ser_port = None
//     self.baud = 0
//     self.open = False

//   def available(self):
//     """ Returns the number of bytes currently in the receive buffer. """
//     return self.ser_port.inWaiting() + len(self.peek_char)

//   def read(self):
//     """ Returns first byte of data in the receive buffer or -1 if timeout reached. """
//     if (self.peek_char):
//       c = self.peek_char
//       self.peek_char = ''
//       return c
//     byte = self.ser_port.read(1)
//     return -1 if (byte == None) else byte

//   def peek(self):
//     """ Returns the next char from the receive buffer without removing it, 
//         or -1 if no data available. """
//     if (self.peek_char):
//       return self.peek_char
//           if self.available():
//       self.peek_char = self.ser_port.read(1)
//       return self.peek_char
//     return -1    

//   def flush(self):
//     """ Waits for current write to finish then flushes rx/tx buffers. """
//     self.ser_port.flush()
//     self.peek_char = ''

//   def prints(self, data, base=None):
//     """ Prints string of given data to the serial port. Returns the number
//         of bytes written. The optional 'base' argument is used to format the
//         data per the Arduino serial.print() formatting scheme, see:
//         http://arduino.cc/en/Serial/Print """
//     return self.write(self._process(data, base))

//   def println(self, data, base=None):
//     """ Prints string of given data to the serial port followed by a 
//         carriage return and line feed. Returns the number of bytes written.
//         The optional 'base' argument is used to format the data per the Arduino
//         serial.print() formatting scheme, see: http://arduino.cc/en/Serial/Print """
//     return self.write(self._process(data, base)+"\r\n")

//   def write(self, data):
//     """ Writes given data to serial port. If data is list or string each
//         element/character is sent sequentially. If data is float it is 
//         converted to an int, if data is int it is sent as a single byte 
//         (least significant if data > 1 byte). Returns the number of bytes
//         written. """
//     assert self.open, "*%s not open, call begin() method before writing" %\
//                       UART[self.config][0]

//     if (type(data) == float): data = int(data)
//     if (type(data) == int): data = chr(data & 0xff)

//     elif ((type(data) == list) or (type(data) == tuple)):
//       bytes_written = 0
//       for i in data:
//         bytes_written += self.write(i)  
//       return bytes_written

//     else:
//       # Type not supported by write, e.g. dict; use prints().
//       return 0

//     written = self.ser_port.write(data)
//     # Serial.serial.write() returns None if no bits written, we want 0:
//     return written if written else 0

//      def _process(self, data, base):
//     """ Processes and returns given data per Arduino format specified on 
//         serial.print() page: http://arduino.cc/en/Serial/Print """
//     if (type(data) == str):
//       # Can't format if already a string:
//       return data

//     if (type(data) is int):
//       if not (base): base = DEC # Default for ints
//       if (base == DEC):
//         return str(data) # e.g. 20 -> "20"
//       if (base == BIN):
//         return bin(data)[2:] # e.g. 20 -> "10100"
//       if (base == OCT):
//         return oct(data)[1:] # e.g. 20 -> "24"
//       if (base == HEX):
//         return hex(data)[2:] # e.g. 20 -> "14"

//     elif (type(data) is float):
//       if not (base): base = 2 # Default for floats
//       if ((base == 0)):
//         return str(int(data))
//       if ((type(base) == int) and (base > 0)):
//         return ("%0." + ("%i" % base) + "f") % data

//     # If we get here data isn't supported by this formatting scheme,
//     # just convert to a string and return:
//     return str(data)

// # Initialize the global serial port instances:
// Serial1 = _UART_PORT('UART1')
// Serial2 = _UART_PORT('UART2')
// Serial4 = _UART_PORT('UART4')
// Serial5 = _UART_PORT('UART5')

func (d *BeagleBoneDriver) analogInit() {
	// """ Initializes the on-board 8ch 12bit ADC. """
	// # Enable ADC module clock, though should already be enabled on
	// # newer Angstrom images:
	d.setRegL(CM_WKUP_ADC_TSC_CLKCTRL, MODULEMODE_ENABLE)
	// _setReg(CM_WKUP_ADC_TSC_CLKCTRL, MODULEMODE_ENABLE)
	// # Wait for enable complete:
	for d.getRegL(CM_WKUP_ADC_TSC_CLKCTRL) & IDLEST_MASK != 0 {
		time.Sleep(100 * time.Microsecond)
	}
	// while (_getReg(CM_WKUP_ADC_TSC_CLKCTRL) & IDLEST_MASK): time.sleep(0.1)

	// # Software reset:
	d.setRegL(ADC_SYSCONFIG, ADC_SOFTRESET)
	// _setReg(ADC_SYSCONFIG, ADC_SOFTRESET)
	for d.getRegL(ADC_SYSCONFIG) & ADC_SOFTRESET != 0 { }
	// while(_getReg(ADC_SYSCONFIG) & ADC_SOFTRESET): pass

	// # Make sure STEPCONFIG write protect is off:
	d.setRegL(ADC_CTRL, ADC_STEPCONFIG_WRITE_PROTECT_OFF)
	// _setReg(ADC_CTRL, ADC_STEPCONFIG_WRITE_PROTECT_OFF)

	// # Set STEPCONFIG1-STEPCONFIG8 to correspond to ADC inputs 0-7:
	d.setRegL(ADCSTEPCONFIG1, 0<<19)
	d.setRegL(ADCSTEPCONFIG2, 1<<19)
	d.setRegL(ADCSTEPCONFIG3, 2<<19)
	d.setRegL(ADCSTEPCONFIG4, 3<<19)
	d.setRegL(ADCSTEPCONFIG5, 4<<19)
	d.setRegL(ADCSTEPCONFIG6, 5<<19)
	d.setRegL(ADCSTEPCONFIG7, 6<<19)
	d.setRegL(ADCSTEPCONFIG8, 7<<19)

	// for i in xrange(8):
	//   config = SEL_INP('AIN%i' % i)
	//   _setReg(eval('ADCSTEPCONFIG%i' % (i+1)), config)
	// # Now we can enable ADC subsystem, leaving write protect off:

	d.orRegL(ADC_CTRL, TSC_ADC_SS_ENABLE)
	// _orReg(ADC_CTRL, TSC_ADC_SS_ENABLE)
}


// def _analog_cleanup():
//   # Software reset:
//   _setReg(ADC_SYSCONFIG, ADC_SOFTRESET)
//   while(_getReg(ADC_SYSCONFIG) & ADC_SOFTRESET): pass
