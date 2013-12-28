package hwio

// // A driver for BeagleBone, based on PyBBIO at:
// // 		https://github.com/alexanderhiam/PyBBIO/blob/master/bbio/config.py
// // The hardware control logic has been ported to Go, using a memory mapped file
// // to get at the control registers direct.
// //
// // NOTE: THIS DRIVER WILL NOT WORK ON THE BEAGLEBONE BLACK, OR OLDER BEAGLEBONES RUNNING
// // LINUX KERNEL 3.8 OR HIGHER. USE DRIVER_BEAGLE_FS INSTEAD. This is because the new kernel
// // uses device trees, and does not support the old muxing technique.
// //
// // This driver is very specific to BeagleBone (I've built to revision A5) and the
// // TI chip that powers it. It may work on other Beagle boards but this is
// // completely untested.
// //
// // Things known to work (tested on hardware):
// // - digital output on all GPIO pins that are exposed on P8 and P9 of the board, and USR0 to USR3
// // - digital input on all GPIO pins that are exposed on P8 and P9 of the board, including pull-up,
// //	 pull-down and pull-disabled modes.

// import (
// 	"errors"
// 	"fmt"
// 	"os"
// 	"strconv"
// 	"syscall"
// 	"unsafe"
// )

// // Represents info we need to know about a pin on the BeagleBone.
// // @todo Determine if 'hwPin' is required
// type BeaglePin struct {
// 	hwPin       string // This intended for the P8.16 format name (currently unused)
// 	profile     []Capability
// 	gpioName    string // This is used for a human readable name
// 	port        uint   // The GPIO port
// 	bit         uint   // A single bit in the position of the I/O value on the port
// 	mode0Name   string // mode 0 signal name, used by the muxer
// 	adcEnable   uint   // bit mask for analog pin control
// 	setAddr     uint   // derived, port+set reg offset
// 	clrAddr     uint   // derived, port+clr reg offset
// 	gpioLogical int    // logical number for GPIO, used by FS driver. This is the GPIO port number plus the GPIO pin within the port
// }

// func (p BeaglePin) GetName() string {
// 	return p.gpioName
// }

// // internal function to identify an analog pin from config
// func (p BeaglePin) isAnalogPin() bool {
// 	return p.adcEnable != 0
// }

// // map logical ports 0 to 3 to their uint offsets in the mmap
// var BB_logicalPorts = [4]uint{BB_GPIO0, BB_GPIO1, BB_GPIO2, BB_GPIO3}

// func makeBeaglePin(hwPin string, profile []Capability, gpioName string, gpioLogicalPort int, gpioPinOnPort int, mode0Name string, adcEnable uint) *BeaglePin {
// 	//			makeBeaglePin("P8.3", bbGpioProfile, "GPIO1_6", BB_GPIO1, 1<<6, "gpmc_ad6", 0, 38),
// 	port := BB_logicalPorts[gpioLogicalPort]
// 	return &BeaglePin{hwPin, profile, gpioName, port, 1 << uint(gpioPinOnPort), mode0Name, adcEnable, port + BB_GPIO_SETDATAOUT, port + BB_GPIO_CLEARDATAOUT, gpioLogicalPort*32 + gpioPinOnPort}
// }

// const (
// 	// Memory map parameters. Note that while the offset is measured in
// 	// bytes so we can create the memory map, all offsets are measured
// 	// in uint32 offsets, since we access the memory map as 32-bit values only.
// 	BB_MMAP_OFFSET = 0x44c00000
// 	BB_MMAP_SIZE   = 0x48ffffff - BB_MMAP_OFFSET

// 	// Size of mmap in uint32
// 	BB_MMAP_N_UINT32 = BB_MMAP_SIZE >> 2

// 	BB_GPIO0 = (0x44e07000 - BB_MMAP_OFFSET) >> 2
// 	BB_GPIO1 = (0x4804c000 - BB_MMAP_OFFSET) >> 2
// 	BB_GPIO2 = (0x481ac000 - BB_MMAP_OFFSET) >> 2
// 	BB_GPIO3 = (0x481ae000 - BB_MMAP_OFFSET) >> 2

// 	//	BB_CM_PER = 0x44e00000-BB_MMAP_OFFSET
// 	BB_CM_WKUP = (0x44e00400 - BB_MMAP_OFFSET) >> 2

// 	//	BB_CM_PER_EPWMSS0_CLKCTRL = 0xd4+CM_PER
// 	//	BB_CM_PER_EPWMSS1_CLKCTRL = 0xcc+CM_PER
// 	//	BB_CM_PER_EPWMSS2_CLKCTRL = 0xd8+CM_PER

// 	BB_CM_WKUP_ADC_TSC_CLKCTRL = 0xbc + BB_CM_WKUP

// 	BB_MODULEMODE_ENABLE = 0x02
// 	BB_IDLEST_MASK       = 0x03 << 16

// 	// //# To enable module clock:
// 	// //#  _setReg(CM_WKUP_module_CLKCTRL, MODULEMODE_ENABLE)
// 	// //#  while (_getReg(CM_WKUP_module_CLKCTRL) & IDLEST_MASK): pass
// 	// //# To disable module clock:
// 	// //#  _andReg(CM_WKUP_module_CLKCTRL, ~MODULEMODE_ENABLE)
// 	// //#-----------------------------
// 	BB_PINMUX_PATH = "/sys/kernel/debug/omap_mux/"

// 	BB_CONF_SLEW_SLOW    = 1 << 6
// 	BB_CONF_RX_ACTIVE    = 1 << 5
// 	BB_CONF_PULLUP       = 1 << 4
// 	BB_CONF_PULLDOWN     = 0x00
// 	BB_CONF_PULL_DISABLE = 1 << 3

// 	BB_CONF_GPIO_MODE   = 0x07
// 	BB_CONF_GPIO_OUTPUT = BB_CONF_GPIO_MODE
// 	BB_CONF_GPIO_INPUT  = BB_CONF_GPIO_MODE + BB_CONF_RX_ACTIVE
// 	// BB_CONF_ADC_PIN     = BB_CONF_RX_ACTIVE+BB_CONF_PULL_DISABLE

// 	// BB_CONF_UART_TX     = BB_CONF_PULLUP
// 	// BB_CONF_UART_RX     = BB_CONF_RX_ACTIVE

// 	BB_GPIO_OE      = 0x134 >> 2
// 	BB_GPIO_DATAIN  = 0x138 >> 2
// 	BB_GPIO_DATAOUT = 0x13c >> 2

// 	// When setting or clearing output bits, these registers provide a faster way to do it, without
// 	// requiring bit manipulation of the output register
// 	BB_GPIO_CLEARDATAOUT = 0x190 >> 2
// 	BB_GPIO_SETDATAOUT   = 0x194 >> 2

// 	// Start of ADC config
// 	BB_ADC_TSC = (0x44e0d000 - BB_MMAP_OFFSET) >> 2

// 	// SYSCONFIG register
// 	BB_ADC_SYSCONFIG = BB_ADC_TSC + (0x10 >> 2)

// 	// @todo find out the meaning of ADC_SOFTRESET. This bit in SYSCONFIG register is marked unused
// 	BB_ADC_SOFTRESET = 0x01

// 	// CTLR register
// 	BB_ADC_CTRL = BB_ADC_TSC + (0x40 >> 2)

// 	// Bit in CTRL to disable write protect on step config registers
// 	BB_ADC_STEPCONFIG_WRITE_PROTECT_OFF = 0x01 << 2
// 	// # Write protect default on, must first turn off to change stepconfig:
// 	// #  _setReg(ADC_CTRL, ADC_STEPCONFIG_WRITE_PROTECT_OFF)
// 	// # To set write protect on:
// 	// #  _clearReg(ADC_CTRL, ADC_STEPCONFIG_WRITE_PROTECT_OFF)

// 	// Bit in CTRL to enable ADC, which should be done after setting up other ADC registers
// 	BB_TSC_ADC_SS_ENABLE = 0x01
// 	// # To enable:
// 	// # _setReg(ADC_CTRL, TSC_ADC_SS_ENABLE)
// 	// #  This will turn STEPCONFIG write protect back on
// 	// # To keep write protect off:
// 	// # _orReg(ADC_CTRL, TSC_ADC_SS_ENABLE)
// 	// #----------------

// 	// ADC_CLKDIV = ADC_TSC+0x4c  # Write desired value-1

// 	// STEPENABLE register
// 	BB_ADC_STEPENABLE = BB_ADC_TSC + (0x54 >> 2)

// 	// ADC_ENABLE = lambda AINx: 0x01<<(ADC[AINx]+1)
// 	// #----------------------

// 	// ADC_IDLECONFIG = ADC_TSC+0x58

// 	// ADC STEPCONFIG registers
// 	BB_ADCSTEPCONFIG1 = BB_ADC_TSC + (0x64 >> 2)
// 	BB_ADCSTEPDELAY1  = BB_ADC_TSC + (0x68 >> 2)
// 	BB_ADCSTEPCONFIG2 = BB_ADC_TSC + (0x6c >> 2)
// 	BB_ADCSTEPDELAY2  = BB_ADC_TSC + (0x70 >> 2)
// 	BB_ADCSTEPCONFIG3 = BB_ADC_TSC + (0x74 >> 2)
// 	BB_ADCSTEPDELAY3  = BB_ADC_TSC + (0x78 >> 2)
// 	BB_ADCSTEPCONFIG4 = BB_ADC_TSC + (0x7c >> 2)
// 	BB_ADCSTEPDELAY4  = BB_ADC_TSC + (0x80 >> 2)
// 	BB_ADCSTEPCONFIG5 = BB_ADC_TSC + (0x84 >> 2)
// 	BB_ADCSTEPDELAY5  = BB_ADC_TSC + (0x88 >> 2)
// 	BB_ADCSTEPCONFIG6 = BB_ADC_TSC + (0x8c >> 2)
// 	BB_ADCSTEPDELAY6  = BB_ADC_TSC + (0x90 >> 2)
// 	BB_ADCSTEPCONFIG7 = BB_ADC_TSC + (0x94 >> 2)
// 	BB_ADCSTEPDELAY7  = BB_ADC_TSC + (0x98 >> 2)
// 	BB_ADCSTEPCONFIG8 = BB_ADC_TSC + (0x9c >> 2)
// 	BB_ADCSTEPDELAY8  = BB_ADC_TSC + (0xa0 >> 2)
// 	// # Only need the first 8 steps - 1 for each AIN pin

// 	// BB_ADC_RESET = 0x00 # Default value of STEPCONFIG

// 	// BB_ADC_AVG2  = 0x01<<2
// 	// BB_ADC_AVG4  = 0x02<<2
// 	// BB_ADC_AVG8  = 0x03<<2
// 	// BB_ADC_AVG16 = 0x04<<2

// 	// #SEL_INP = lambda AINx: (ADC[AINx]+1)<<19
// 	// # Set input with _orReg(ADCSTEPCONFIGx, SEL_INP(AINx))
// 	// # ADC[AINx]+1 because positive AMUX input 0 is VREFN
// 	// #  (see user manual section 12.3.7)
// 	// SEL_INP = lambda AINx: (ADC[AINx])<<19

// 	// SAMPLE_DELAY = lambda cycles: (cycles&0xff)<<24
// 	// # SAMPLE_DELAY is the number of cycles to sample for
// 	// # Set delay with _orReg(ADCSTEPDELAYx, SAMPLE_DELAY(cycles))

// 	// #----------------------

// 	// ADC FIFO
// 	BB_ADC_FIFO0DATA = BB_ADC_TSC + (0x100 >> 2)
// 	BB_ADC_FIFO_MASK = 0xfff

// // ## ADC pins:

// // # And some constants so the user doesn't need to use strings:
// // AIN0 = A0 = 'AIN0'
// // AIN1 = A1 = 'AIN1'
// // AIN2 = A2 = 'AIN2'
// // AIN3 = A3 = 'AIN3'
// // AIN4 = A4 = 'AIN4'
// // AIN5 = A5 = 'AIN5'
// // AIN6 = A6 = 'AIN6'
// // AIN7 = A7 = VSYS = 'AIN7'

// // ##--- End ADC config -------##
// // ##############################

// // ##############################
// // ##--- Start UART config: ---##

// // # UART ports must be in form:
// // #    [port, tx_pinmux_filename, tx_pinmux_mode,
// // #           rx_pinmux_filename, rx_pinmux_mode]

// // UART = {
// //   'UART1' : ['/dev/ttyO1', 'uart1_txd', 0,  'uart1_rxd', 0],
// //   'UART2' : ['/dev/ttyO2',   'spi0_d0', 1,  'spi0_sclk', 1],
// //   'UART4' : ['/dev/ttyO4',  'gpmc_wpn', 6, 'gpmc_wait0', 6],
// //   'UART5' : ['/dev/ttyO5', 'lcd_data8', 4,  'lcd_data9', 4]
// // }

// // # Formatting constants to mimic Arduino's serial.print() formatting:
// // DEC = 'DEC'
// // BIN = 'BIN'
// // OCT = 'OCT'
// // HEX = 'HEX'

// )

// var beaglePins []*BeaglePin
// var bbGpioProfile []Capability
// var bbAnalogInProfile []Capability
// var bbUsrLedProfile []Capability

// //	analog := []Capability {CAP_INPUT,CAP_OUTPUT,CAP_ANALOG_IN}
// //	pwm := []Capability {CAP_INPUT,CAP_OUTPUT,CAP_PWM}
// //	readonly := []Capability {CAP_INPUT}

// func init() {
// 	// Note: Logical pin numbers are implicitly assigned from 0 in the order in
// 	// which they occur in this slice. i.e. Pin 0 will be gpmc_a5, pin 1 will be
// 	// gpmc_a5 and so on.

// 	bbGpioProfile = []Capability{
// 		CAP_OUTPUT,
// 		CAP_INPUT,
// 		CAP_INPUT_PULLUP,
// 		CAP_INPUT_PULLDOWN,
// 	}
// 	bbAnalogInProfile = []Capability{
// 		CAP_ANALOG_IN,
// 	}
// 	bbUsrLedProfile = []Capability{
// 		CAP_OUTPUT,
// 	}

// 	p := []*BeaglePin{
// 		// P8
// 		makeBeaglePin("P8.3", bbGpioProfile, "GPIO1_6", 1, 6, "gpmc_ad6", 0),
// 		makeBeaglePin("P8.4", bbGpioProfile, "GPIO1_7", 1, 7, "gpmc_ad7", 0),
// 		makeBeaglePin("P8.5", bbGpioProfile, "GPIO1_2", 1, 2, "gpmc_ad2", 0),
// 		makeBeaglePin("P8.6", bbGpioProfile, "GPIO1_3", 1, 3, "gpmc_ad3", 0),
// 		makeBeaglePin("P8.7", bbGpioProfile, "GPIO2_2", 2, 2, "gpmc_advn_ale", 0),
// 		makeBeaglePin("P8.8", bbGpioProfile, "GPIO2_3", 2, 3, "gpmc_oen_ren", 0),
// 		makeBeaglePin("P8.9", bbGpioProfile, "GPIO2_5", 2, 5, "gpmc_ben0_cle", 0),
// 		makeBeaglePin("P8.10", bbGpioProfile, "GPIO2_4", 2, 4, "gpmc_wen", 0),
// 		makeBeaglePin("P8.11", bbGpioProfile, "GPIO1_13", 1, 13, "gpmc_ad13", 0),
// 		makeBeaglePin("P8.12", bbGpioProfile, "GPIO1_12", 1, 12, "gpmc_ad12", 0),
// 		makeBeaglePin("P8.13", bbGpioProfile, "GPIO0_23", 0, 23, "gpmc_ad9", 0),
// 		makeBeaglePin("P8.14", bbGpioProfile, "GPIO0_26", 0, 26, "gpmc_ad10", 0),
// 		makeBeaglePin("P8.15", bbGpioProfile, "GPIO1_15", 1, 15, "gpmc_ad15", 0),
// 		makeBeaglePin("P8.16", bbGpioProfile, "GPIO1_14", 1, 14, "gpmc_ad14", 0),
// 		makeBeaglePin("P8.17", bbGpioProfile, "GPIO0_27", 0, 27, "gpmc_ad11", 0),
// 		makeBeaglePin("P8.18", bbGpioProfile, "GPIO2_1", 2, 1, "gpmc_clk", 0),
// 		makeBeaglePin("P8.19", bbGpioProfile, "GPIO0_22", 0, 22, "gpmc_ad8", 0),
// 		makeBeaglePin("P8.20", bbGpioProfile, "GPIO1_31", 1, 31, "gpmc_csn2", 0),
// 		makeBeaglePin("P8.21", bbGpioProfile, "GPIO1_30", 1, 30, "gpmc_csn1", 0),
// 		makeBeaglePin("P8.22", bbGpioProfile, "GPIO1_5", 1, 5, "gpmc_ad5", 0),
// 		makeBeaglePin("P8.23", bbGpioProfile, "GPIO1_4", 1, 4, "gpmc_ad4", 0),
// 		makeBeaglePin("P8.24", bbGpioProfile, "GPIO1_1", 1, 1, "gpmc_ad1", 0),
// 		makeBeaglePin("P8.25", bbGpioProfile, "GPIO1_0", 1, 1, "gpmc_ad0", 0),
// 		makeBeaglePin("P8.26", bbGpioProfile, "GPIO1_29", 1, 29, "gpmc_csn0", 0),
// 		makeBeaglePin("P8.27", bbGpioProfile, "GPIO2_22", 2, 22, "lcd_vsync", 0),
// 		makeBeaglePin("P8.28", bbGpioProfile, "GPIO2_24", 2, 24, "lcd_pclk", 0),
// 		makeBeaglePin("P8.29", bbGpioProfile, "GPIO2_23", 2, 23, "lcd_hsync", 0),
// 		makeBeaglePin("P8.30", bbGpioProfile, "GPIO2_25", 2, 25, "lcd_ac_bias_en", 0),
// 		makeBeaglePin("P8.31", bbGpioProfile, "GPIO0_10", 0, 10, "lcd_data14", 0),
// 		makeBeaglePin("P8.32", bbGpioProfile, "GPIO0_11", 0, 11, "lcd_data15", 0),
// 		makeBeaglePin("P8.33", bbGpioProfile, "GPIO0_9", 0, 9, "lcd_data13", 0),
// 		makeBeaglePin("P8.34", bbGpioProfile, "GPIO2_17", 2, 17, "lcd_data11", 0),
// 		makeBeaglePin("P8.35", bbGpioProfile, "GPIO0_8", 0, 8, "lcd_data12", 0),
// 		makeBeaglePin("P8.36", bbGpioProfile, "GPIO2_16", 2, 16, "lcd_data10", 0),
// 		makeBeaglePin("P8.37", bbGpioProfile, "GPIO2_14", 2, 14, "lcd_data8", 0),
// 		makeBeaglePin("P8.38", bbGpioProfile, "GPIO2_15", 2, 15, "lcd_data9", 0),
// 		makeBeaglePin("P8.40", bbGpioProfile, "GPIO2_13", 2, 13, "lcd_data7", 0),
// 		makeBeaglePin("P8.41", bbGpioProfile, "GPIO2_10", 2, 10, "lcd_data4", 0),
// 		makeBeaglePin("P8.42", bbGpioProfile, "GPIO2_11", 2, 11, "lcd_data5", 0),
// 		makeBeaglePin("P8.43", bbGpioProfile, "GPIO2_8", 2, 8, "lcd_data2", 0),
// 		makeBeaglePin("P8.44", bbGpioProfile, "GPIO2_9", 2, 9, "lcd_data3", 0),
// 		makeBeaglePin("P8.45", bbGpioProfile, "GPIO2_6", 2, 6, "lcd_data0", 0),
// 		// makeBeaglePin("P8.46", bbGpioProfile, "GPIO2_7", 2, 7, "lcd_data1", 0),

// 		// P9
// 		makeBeaglePin("P9.11", bbGpioProfile, "GPIO0_30", 0, 30, "gpmc_wait0", 0),
// 		makeBeaglePin("P9.12", bbGpioProfile, "GPIO1_28", 1, 28, "gpmc_ben1", 0),
// 		makeBeaglePin("P9.13", bbGpioProfile, "GPIO0_31", 0, 31, "gpmc_wpn", 0),
// 		makeBeaglePin("P9.14", bbGpioProfile, "GPIO1_18", 1, 18, "gpmc_a2", 0),
// 		makeBeaglePin("P9.15", bbGpioProfile, "GPIO1_16", 1, 16, "gpmc_a0", 0),
// 		makeBeaglePin("P9.16", bbGpioProfile, "GPIO1_19", 1, 19, "gpmc_a3", 0),
// 		makeBeaglePin("P9.17", bbGpioProfile, "GPIO0_5", 0, 5, "spi0_cs0", 0),
// 		makeBeaglePin("P9.18", bbGpioProfile, "GPIO0_4", 0, 4, "spi0_d1", 0),
// 		makeBeaglePin("P9.19", bbGpioProfile, "GPIO0_13", 0, 13, "uart1_rtsn", 0),
// 		makeBeaglePin("P9.20", bbGpioProfile, "GPIO0_12", 0, 12, "uart1_ctsn", 0),
// 		makeBeaglePin("P9.21", bbGpioProfile, "GPIO0_3", 0, 3, "spi0_d0", 0),
// 		makeBeaglePin("P9.22", bbGpioProfile, "GPIO0_2", 0, 2, "spi0_sclk", 0),
// 		makeBeaglePin("P9.23", bbGpioProfile, "GPIO1_17", 1, 17, "gpmc_a1", 0),
// 		makeBeaglePin("P9.24", bbGpioProfile, "GPIO0_15", 0, 15, "uart1_txd", 0),
// 		makeBeaglePin("P9.25", bbGpioProfile, "GPIO3_21", 3, 21, "mcasp0_ahclkx", 0),
// 		makeBeaglePin("P9.26", bbGpioProfile, "GPIO0_14", 0, 14, "uart1_rxd", 0),
// 		makeBeaglePin("P9.27", bbGpioProfile, "GPIO3_19", 3, 19, "mcasp0_fsr", 0),
// 		makeBeaglePin("P9.28", bbGpioProfile, "GPIO3_17", 3, 17, "mcasp0_ahclkr", 0),
// 		makeBeaglePin("P9.29", bbGpioProfile, "GPIO3_15", 3, 15, "mcasp0_fsx", 0),
// 		makeBeaglePin("P9.30", bbGpioProfile, "GPIO3_16", 3, 16, "mcasp0_axr0", 0),
// 		makeBeaglePin("P9.31", bbGpioProfile, "GPIO3_14", 3, 14, "mcasp0_aclkx", 0),
// 		makeBeaglePin("P9.33", bbAnalogInProfile, "AIN4", 0, 0, "ain4", 1<<5),
// 		makeBeaglePin("P9.35", bbAnalogInProfile, "AIN6", 0, 0, "ain6", 1<<7),
// 		makeBeaglePin("P9.36", bbAnalogInProfile, "AIN5", 0, 0, "ain5", 1<<6),
// 		makeBeaglePin("P9.37", bbAnalogInProfile, "AIN2", 0, 0, "ain2", 1<<3),
// 		makeBeaglePin("P9.38", bbAnalogInProfile, "AIN3", 0, 0, "ain3", 1<<4),
// 		makeBeaglePin("P9.39", bbAnalogInProfile, "AIN0", 0, 0, "ain0", 1<<1),
// 		makeBeaglePin("P9.40", bbAnalogInProfile, "AIN1", 0, 0, "ain1", 1<<2),
// 		makeBeaglePin("P9.41", bbGpioProfile, "GPIO0_20", 0, 20, "xdma_event_intr1", 0),
// 		makeBeaglePin("P9.42", bbGpioProfile, "GPIO0_7", 0, 7, "ecap0_in_pwm0_out", 0),

// 		// // USR LEDs
// 		makeBeaglePin("USR0", bbUsrLedProfile, "USR0", 1, 21, "gpmc_a5", 0),
// 		makeBeaglePin("USR1", bbUsrLedProfile, "USR1", 1, 22, "gpmc_a6", 0),
// 		makeBeaglePin("USR2", bbUsrLedProfile, "USR2", 1, 23, "gpmc_a7", 0),
// 		makeBeaglePin("USR3", bbUsrLedProfile, "USR3", 1, 24, "gpmc_a8", 0),
// 	}
// 	beaglePins = p
// }

// type BeagleBoneDriver struct {
// 	// Mapped memory for directly accessing hardware registers
// 	mmap []byte

// 	// Memory accessable as an array of long.
// 	memArray *[BB_MMAP_N_UINT32]uint
// }

// func (d *BeagleBoneDriver) Init() error {
// 	// The following snippet is from:
// 	// https://groups.google.com/forum/?fromgroups=#!topic/golang-nuts/1omyttb2Hlo
// 	// This creates a memory map of longs. Given that all our access is 32-bit, this
// 	// could speed up access significantly. Need to adjust getRegL to use this and divide
// 	// the address offset by 4.
// 	//	if F==nil {panic(err)}
// 	//	fd := F.Fd();
// 	//	addr, _, errno := syscall.Syscall6(syscall.SYS_MMAP,
// 	//							0, uintptr(HandRankSz)*4,
// 	//							1 /* syscall.PROT_READ */,
// 	//							0, uintptr(fd), 0);
// 	//	if errno != 0 {
// 	//		log.Exitf("mmap display: %s", os.Errno(errno))
// 	//	}
// 	//	HandRanks = (*[HandRankSz]int32)(unsafe.Pointer(addr));
// 	//	// mmap without touching the pages would be cheating...
// 	//	var sum int32;
// 	//	for hr:=range(HandRanks) {
// 	//		sum ^= HandRanks[hr];
// 	//	}

// 	// Set up the memory mapped file giving us access to hardware registers
// 	file, e := os.OpenFile("/dev/mem", os.O_RDWR|os.O_APPEND, 0)
// 	if e != nil {
// 		return e
// 	}
// 	mmap, e := syscall.Mmap(int(file.Fd()), BB_MMAP_OFFSET, BB_MMAP_SIZE, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
// 	if e != nil {
// 		return e
// 	}
// 	d.mmap = mmap
// 	d.memArray = (*[BB_MMAP_N_UINT32]uint)(unsafe.Pointer(&mmap[0]))

// 	d.analogInit()

// 	return nil
// }

// func (d *BeagleBoneDriver) Close() {
// 	syscall.Munmap(d.mmap)
// }

// func (d *BeagleBoneDriver) PinMode(pin Pin, mode PinIOMode) error {
// 	p := beaglePins[pin]

// 	// handle analog first, they are simplest from PinMode perspective
// 	if p.isAnalogPin() {
// 		if mode != INPUT {
// 			return errors.New(fmt.Sprintf("Pin %d is an analog pin, and the mode must be INPUT", p))
// 		}
// 		return nil // nothing to set up
// 	}

// 	if mode == OUTPUT {
// 		e := d.pinMux(p.mode0Name, BB_CONF_GPIO_OUTPUT)
// 		if e != nil {
// 			return e
// 		}

// 		d.clearRegL(p.port+uint(BB_GPIO_OE), p.bit)
// 	} else {
// 		pull := BB_CONF_PULL_DISABLE
// 		// note: pull up/down modes assume that CONF_PULLDOWN resets the pull disable bit
// 		if mode == INPUT_PULLUP {
// 			pull = BB_CONF_PULLUP
// 		} else if mode == INPUT_PULLDOWN {
// 			pull = BB_CONF_PULLDOWN
// 		}

// 		e := d.pinMux(p.mode0Name, BB_CONF_GPIO_INPUT|uint(pull))
// 		if e != nil {
// 			return e
// 		}

// 		//		fmt.Printf("R/W dir reg BEFORE value is %x\n", d.getRegL(p.port+uint(GPIO_OE)))

// 		d.orRegL(p.port+uint(BB_GPIO_OE), p.bit)
// 		//		fmt.Printf("R/W dir reg AFTER value is %x\n", d.getRegL(p.port+uint(GPIO_OE)))
// 	}
// 	return nil
// }

// func (d *BeagleBoneDriver) pinMux(mux string, mode uint) error {
// 	// Uses kernel omap_mux files to set pin modes.
// 	// There's no simple way to write the control module registers from a
// 	// user-level process because it lacks the proper privileges, but it's
// 	// easy enough to just use the built-in file-based system and let the
// 	// kernel do the work.
// 	f, e := os.OpenFile(BB_PINMUX_PATH+mux, os.O_WRONLY|os.O_TRUNC, 0666)
// 	if e != nil {
// 		return e
// 	}

// 	s := strconv.FormatInt(int64(mode), 16)
// 	//	fmt.Printf("Writing mode %s to mux file %s\n", s, PINMUX_PATH+mux)
// 	f.WriteString(s)
// 	return nil
// }

// func (d *BeagleBoneDriver) DigitalWrite(pin Pin, value int) (e error) {
// 	p := beaglePins[pin]
// 	if value == 0 {
// 		d.memArray[p.clrAddr] = p.bit
// 	} else {
// 		d.memArray[p.setAddr] = p.bit
// 	}
// 	return nil
// }

// func (d *BeagleBoneDriver) DigitalRead(pin Pin) (value int, e error) {
// 	p := beaglePins[pin]
// 	reg := d.getRegL(p.port + BB_GPIO_DATAIN)
// 	//	fmt.Printf("\nraw in: %x (checking bit %d)\n", reg, p.bit)
// 	if (reg & p.bit) != 0 {
// 		return HIGH, nil
// 	}
// 	return LOW, nil
// }

// func (d *BeagleBoneDriver) AnalogWrite(pin Pin, value int) (e error) {
// 	return nil
// }

// func (d *BeagleBoneDriver) AnalogRead(pin Pin) (value int, e error) {
// 	//	if d.getRegL(CM_WKUP_ADC_TSC_CLKCTRL) & IDLEST_MASK == 0 {
// 	//		// if for any reason the ADC module clock has been shut off, turn it back on
// 	//		d.analogInit()
// 	//	}

// 	p := beaglePins[pin]

// 	// if (_getReg(CM_WKUP_ADC_TSC_CLKCTRL) & IDLEST_MASK):
// 	//   # The ADC module clock has been shut off, e.g. by a different
// 	//   # PyBBIO script stopping while this one was running, turn back on:
// 	//   _analog_init()

// 	// ADC_ENABLE = lambda AINx: 0x01<<(ADC[AINx]+1)
// 	d.setRegL(BB_ADC_STEPENABLE, p.adcEnable)
// 	// # Enable sequncer step that's set for given input:
// 	// _setReg(ADC_STEPENABLE, ADC_ENABLE(analog_pin))

// 	for d.getRegL(BB_ADC_STEPENABLE)&p.adcEnable != 0 {
// 	}

// 	// # Sequencer starts automatically after enabling step, wait for complete:
// 	// while(_getReg(ADC_STEPENABLE) & ADC_ENABLE(analog_pin)): pass
// 	// # Return 12-bit value from the ADC FIFO register:
// 	// return _getReg(ADC_FIFO0DATA) & ADC_FIFO_MASK
// 	res := d.getRegL(BB_ADC_FIFO0DATA) & BB_ADC_FIFO_MASK
// 	fmt.Printf("register output %x\n", res)
// 	return int(res), nil
// }

// // def inVolts(adc_value, bits=12, vRef=1.8):
// //   """ Converts and returns the given ADC value to a voltage according
// //       to the given number of bits and reference voltage. """
// //   return adc_value*(vRef/2**bits)

// // Sets 32 bit Register at address to its current value AND mask.
// func (d *BeagleBoneDriver) andRegL(address uint, mask uint) {
// 	d.setRegL(address, d.getRegL(address)&mask)
// }

// // Sets 32 bit Register at address to its current value OR mask.
// func (d *BeagleBoneDriver) orRegL(address uint, mask uint) {
// 	d.setRegL(address, d.getRegL(address)|mask)
// }

// // Clears mask bits in 32 bit register at given address.
// func (d *BeagleBoneDriver) clearRegL(address uint, mask uint) {
// 	d.andRegL(address, ^mask)
// }

// // Returns 32 bit value at given address
// func (d *BeagleBoneDriver) getRegL(address uint) (result uint) {
// 	return d.memArray[address]
// }

// func (d *BeagleBoneDriver) setRegL(address uint, value uint) {
// 	d.memArray[address] = value
// }

// func (d *BeagleBoneDriver) PinMap() (pinMap HardwarePinMap) {
// 	pinMap = make(HardwarePinMap)

// 	for i, hw := range beaglePins {
// 		names := []string{hw.hwPin}
// 		if hw.hwPin != hw.gpioName {
// 			names = append(names, hw.gpioName)
// 		}
// 		pinMap.add(Pin(i), names, hw.profile)
// 	}

// 	return
// }

// func (d *BeagleBoneDriver) GetModule(name string) Module {
// 	// @tod implement
// 	return nil
// }

// // # _UART_PORT is a wrapper class for pySerial to enable Arduino-like access
// // # to the UART1, UART2, UART4, and UART5 serial ports on the expansion headers:
// // class _UART_PORT(object):
// //   def __init__(self, uart):
// //     assert uart in UART, "*Invalid UART: %s" % uart
// //     self.config = uart
// //     self.baud = 0
// //     self.open = False
// //     self.ser_port = None
// //     self.peek_char = ''

// //   def begin(self, baud, timeout=1):
// //     """ Starts the serial port at the given baud rate. """
// //     # Set proper pinmux to match expansion headers:
// //     tx_pinmux_filename = UART[self.config][1]
// //     tx_pinmux_mode     = UART[self.config][2]+CONF_UART_TX
// //     _pinMux(tx_pinmux_filename, tx_pinmux_mode)

// //     rx_pinmux_filename = UART[self.config][3]
// //     rx_pinmux_mode     = UART[self.config][4]+CONF_UART_RX
// //     _pinMux(rx_pinmux_filename, rx_pinmux_mode)

// //     port = UART[self.config][0]
// //     self.baud = baud
// //     self.ser_port = serial.Serial(port, baud, timeout=timeout)
// //     self.open = True

// //   def end(self):
// //     """ Closes the serial port if open. """
// //     if not(self.open): return
// //     self.flush()
// //     self.ser_port.close()
// //     self.ser_port = None
// //     self.baud = 0
// //     self.open = False

// //   def available(self):
// //     """ Returns the number of bytes currently in the receive buffer. """
// //     return self.ser_port.inWaiting() + len(self.peek_char)

// //   def read(self):
// //     """ Returns first byte of data in the receive buffer or -1 if timeout reached. """
// //     if (self.peek_char):
// //       c = self.peek_char
// //       self.peek_char = ''
// //       return c
// //     byte = self.ser_port.read(1)
// //     return -1 if (byte == None) else byte

// //   def peek(self):
// //     """ Returns the next char from the receive buffer without removing it,
// //         or -1 if no data available. """
// //     if (self.peek_char):
// //       return self.peek_char
// //           if self.available():
// //       self.peek_char = self.ser_port.read(1)
// //       return self.peek_char
// //     return -1

// //   def flush(self):
// //     """ Waits for current write to finish then flushes rx/tx buffers. """
// //     self.ser_port.flush()
// //     self.peek_char = ''

// //   def prints(self, data, base=None):
// //     """ Prints string of given data to the serial port. Returns the number
// //         of bytes written. The optional 'base' argument is used to format the
// //         data per the Arduino serial.print() formatting scheme, see:
// //         http://arduino.cc/en/Serial/Print """
// //     return self.write(self._process(data, base))

// //   def println(self, data, base=None):
// //     """ Prints string of given data to the serial port followed by a
// //         carriage return and line feed. Returns the number of bytes written.
// //         The optional 'base' argument is used to format the data per the Arduino
// //         serial.print() formatting scheme, see: http://arduino.cc/en/Serial/Print """
// //     return self.write(self._process(data, base)+"\r\n")

// //   def write(self, data):
// //     """ Writes given data to serial port. If data is list or string each
// //         element/character is sent sequentially. If data is float it is
// //         converted to an int, if data is int it is sent as a single byte
// //         (least significant if data > 1 byte). Returns the number of bytes
// //         written. """
// //     assert self.open, "*%s not open, call begin() method before writing" %\
// //                       UART[self.config][0]

// //     if (type(data) == float): data = int(data)
// //     if (type(data) == int): data = chr(data & 0xff)

// //     elif ((type(data) == list) or (type(data) == tuple)):
// //       bytes_written = 0
// //       for i in data:
// //         bytes_written += self.write(i)
// //       return bytes_written

// //     else:
// //       # Type not supported by write, e.g. dict; use prints().
// //       return 0

// //     written = self.ser_port.write(data)
// //     # Serial.serial.write() returns None if no bits written, we want 0:
// //     return written if written else 0

// //      def _process(self, data, base):
// //     """ Processes and returns given data per Arduino format specified on
// //         serial.print() page: http://arduino.cc/en/Serial/Print """
// //     if (type(data) == str):
// //       # Can't format if already a string:
// //       return data

// //     if (type(data) is int):
// //       if not (base): base = DEC # Default for ints
// //       if (base == DEC):
// //         return str(data) # e.g. 20 -> "20"
// //       if (base == BIN):
// //         return bin(data)[2:] # e.g. 20 -> "10100"
// //       if (base == OCT):
// //         return oct(data)[1:] # e.g. 20 -> "24"
// //       if (base == HEX):
// //         return hex(data)[2:] # e.g. 20 -> "14"

// //     elif (type(data) is float):
// //       if not (base): base = 2 # Default for floats
// //       if ((base == 0)):
// //         return str(int(data))
// //       if ((type(base) == int) and (base > 0)):
// //         return ("%0." + ("%i" % base) + "f") % data

// //     # If we get here data isn't supported by this formatting scheme,
// //     # just convert to a string and return:
// //     return str(data)

// // # Initialize the global serial port instances:
// // Serial1 = _UART_PORT('UART1')
// // Serial2 = _UART_PORT('UART2')
// // Serial4 = _UART_PORT('UART4')
// // Serial5 = _UART_PORT('UART5')

// func (d *BeagleBoneDriver) analogInit() {
// 	// // """ Initializes the on-board 8ch 12bit ADC. """
// 	// // # Enable ADC module clock, though should already be enabled on
// 	// // # newer Angstrom images:
// 	// d.setRegL(BB_CM_WKUP_ADC_TSC_CLKCTRL, BB_MODULEMODE_ENABLE)
// 	// // _setReg(CM_WKUP_ADC_TSC_CLKCTRL, MODULEMODE_ENABLE)
// 	// // # Wait for enable complete:
// 	// for d.getRegL(BB_CM_WKUP_ADC_TSC_CLKCTRL)&BB_IDLEST_MASK != 0 {
// 	// 	time.Sleep(100 * time.Microsecond)
// 	// }
// 	// // while (_getReg(CM_WKUP_ADC_TSC_CLKCTRL) & IDLEST_MASK): time.sleep(0.1)

// 	// // # Software reset:
// 	// d.setRegL(BB_ADC_SYSCONFIG, BB_ADC_SOFTRESET)
// 	// // _setReg(ADC_SYSCONFIG, ADC_SOFTRESET)
// 	// for d.getRegL(BB_ADC_SYSCONFIG)&BB_ADC_SOFTRESET != 0 {
// 	// }
// 	// // while(_getReg(ADC_SYSCONFIG) & ADC_SOFTRESET): pass

// 	// // # Make sure STEPCONFIG write protect is off:
// 	// d.setRegL(BB_ADC_CTRL, BB_ADC_STEPCONFIG_WRITE_PROTECT_OFF)
// 	// // _setReg(ADC_CTRL, ADC_STEPCONFIG_WRITE_PROTECT_OFF)

// 	// // # Set STEPCONFIG1-STEPCONFIG8 to correspond to ADC inputs 0-7:
// 	// d.setRegL(BB_ADCSTEPCONFIG1, 0<<19)
// 	// d.setRegL(BB_ADCSTEPCONFIG2, 1<<19)
// 	// d.setRegL(BB_ADCSTEPCONFIG3, 2<<19)
// 	// d.setRegL(BB_ADCSTEPCONFIG4, 3<<19)
// 	// d.setRegL(BB_ADCSTEPCONFIG5, 4<<19)
// 	// d.setRegL(BB_ADCSTEPCONFIG6, 5<<19)
// 	// d.setRegL(BB_ADCSTEPCONFIG7, 6<<19)
// 	// d.setRegL(BB_ADCSTEPCONFIG8, 7<<19)

// 	// // for i in xrange(8):
// 	// //   config = SEL_INP('AIN%i' % i)
// 	// //   _setReg(eval('ADCSTEPCONFIG%i' % (i+1)), config)
// 	// // # Now we can enable ADC subsystem, leaving write protect off:

// 	// d.orRegL(BB_ADC_CTRL, BB_TSC_ADC_SS_ENABLE)
// 	// // _orReg(ADC_CTRL, TSC_ADC_SS_ENABLE)
// }

// // def _analog_cleanup():
// //   # Software reset:
// //   _setReg(ADC_SYSCONFIG, ADC_SOFTRESET)
// //   while(_getReg(ADC_SYSCONFIG) & ADC_SOFTRESET): pass
