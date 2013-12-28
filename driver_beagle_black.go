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

// Notes on analog:
//
// echo cape-bone-iio > /sys/devices/bone_capemgr.*/slots    ' once off
// find /sys/ -name '*AIN*':
// /sys/devices/ocp.2/helper.14/AIN0
// /sys/devices/ocp.2/helper.14/AIN1
// /sys/devices/ocp.2/helper.14/AIN2
// /sys/devices/ocp.2/helper.14/AIN3
// /sys/devices/ocp.2/helper.14/AIN4
// /sys/devices/ocp.2/helper.14/AIN5
// /sys/devices/ocp.2/helper.14/AIN6
// /sys/devices/ocp.2/helper.14/AIN7

type BeaglePin struct {
	names   []string // This intended for the P8.16 format name (currently unused)
	modules []string // Names of modules that may allocate this pin

	gpioLogical   int // logical number for GPIO, for pins used by "gpio" module. This is the GPIO port number plus the GPIO pin within the port.
	analogLogical int // analog pin number, for pins used by "analog" module
}

type BeagleBoneBlackDriver struct {
	// all pins understood by the driver
	beaglePins []*BeaglePin

	// a map of module names to module objects, created at initialisation
	modules map[string]Module
}

func (d *BeagleBoneBlackDriver) Init() error {
	d.createPinData()
	d.initialiseModules()

	return nil
}

func (d *BeagleBoneBlackDriver) makePin(names []string, modules []string, gpioLogical int, analogLogical int) *BeaglePin {
	return &BeaglePin{names, modules, gpioLogical, analogLogical}
}

func (d *BeagleBoneBlackDriver) createPinData() {
	d.beaglePins = []*BeaglePin{
		// P8
		d.makePin([]string{"P8.3", "gpmc_ad6", "gpio1_6"}, []string{"gpio"}, 38, 0),
		d.makePin([]string{"P8.4", "gpmc_ad7", "gpio1_7"}, []string{"gpio"}, 39, 0),
		d.makePin([]string{"P8.5", "gpmc_ad2", "gpio1_2"}, []string{"gpio"}, 34, 0),
		d.makePin([]string{"P8.6", "gpmc_ad3", "GPIO1_3"}, []string{"gpio"}, 35, 0),
		d.makePin([]string{"P8.7", "gpmc_advn_ale", "GPIO2_2"}, []string{"gpio"}, 66, 0),
		d.makePin([]string{"P8.8", "gpmc_oen_ren", "GPIO2_3"}, []string{"gpio"}, 67, 0),
		d.makePin([]string{"P8.9", "gpmc_ben0_cle", "GPIO2_5"}, []string{"gpio"}, 69, 0),
		d.makePin([]string{"P8.10", "gpmc_wen", "GPIO2_4"}, []string{"gpio"}, 68, 0),
		d.makePin([]string{"P8.11", "gpmc_ad13", "GPIO1_13"}, []string{"gpio"}, 45, 0),
		d.makePin([]string{"P8.12", "gpmc_ad12", "GPIO1_12"}, []string{"gpio"}, 44, 0),
		d.makePin([]string{"P8.13", "gpmc_ad9", "GPIO0_23"}, []string{"gpio"}, 23, 0),
		d.makePin([]string{"P8.14", "gpmc_ad10", "GPIO0_26"}, []string{"gpio"}, 26, 0),
		d.makePin([]string{"P8.15", "gpmc_ad15", "GPIO1_15"}, []string{"gpio"}, 47, 0),
		d.makePin([]string{"P8.16", "gpmc_ad14", "GPIO1_14"}, []string{"gpio"}, 46, 0),
		d.makePin([]string{"P8.17", "gpmc_ad11", "GPIO0_27"}, []string{"gpio"}, 27, 0),
		d.makePin([]string{"P8.18", "gpmc_clk", "GPIO2_1"}, []string{"gpio"}, 65, 0),
		d.makePin([]string{"P8.19", "gpmc_ad8", "GPIO0_22"}, []string{"gpio"}, 22, 0),
		d.makePin([]string{"P8.20", "gpmc_csn2", "GPIO1_31"}, []string{"gpio"}, 63, 0),
		d.makePin([]string{"P8.21", "gpmc_csn1", "GPIO1_30"}, []string{"gpio"}, 62, 0),
		d.makePin([]string{"P8.22", "gpmc_ad5", "GPIO1_5"}, []string{"gpio"}, 37, 0),
		d.makePin([]string{"P8.23", "gpmc_ad4", "GPIO1_4"}, []string{"gpio"}, 36, 0),
		d.makePin([]string{"P8.24", "gpmc_ad1", "GPIO1_1"}, []string{"gpio"}, 33, 0),
		d.makePin([]string{"P8.25", "gpmc_ad0", "GPIO1_0"}, []string{"gpio"}, 32, 0),
		d.makePin([]string{"P8.26", "gpmc_csn0", "GPIO1_29"}, []string{"gpio"}, 61, 0),
		d.makePin([]string{"P8.27", "lcd_vsync", "GPIO2_22"}, []string{"gpio"}, 86, 0),
		d.makePin([]string{"P8.28", "lcd_pclk", "GPIO2_24"}, []string{"gpio"}, 88, 0),
		d.makePin([]string{"P8.29", "lcd_hsync", "GPIO2_23"}, []string{"gpio"}, 87, 0),
		d.makePin([]string{"P8.30", "lcd_ac_bias_en", "GPIO2_25"}, []string{"gpio"}, 89, 0),
		d.makePin([]string{"P8.31", "lcd_data14", "GPIO0_10"}, []string{"gpio"}, 10, 0),
		d.makePin([]string{"P8.32", "lcd_data15", "GPIO0_11"}, []string{"gpio"}, 11, 0),
		d.makePin([]string{"P8.33", "lcd_data13", "GPIO0_9"}, []string{"gpio"}, 9, 0),
		d.makePin([]string{"P8.34", "lcd_data11", "GPIO2_17"}, []string{"gpio"}, 81, 0),
		d.makePin([]string{"P8.35", "lcd_data12", "GPIO0_8"}, []string{"gpio"}, 8, 0),
		d.makePin([]string{"P8.36", "lcd_data10", "GPIO2_16"}, []string{"gpio"}, 80, 0),
		d.makePin([]string{"P8.37", "lcd_data8", "GPIO2_14"}, []string{"gpio"}, 78, 0),
		d.makePin([]string{"P8.38", "lcd_data9", "GPIO2_15"}, []string{"gpio"}, 79, 0),
		d.makePin([]string{"P8.40", "lcd_data7", "GPIO2_13"}, []string{"gpio"}, 77, 0),
		d.makePin([]string{"P8.41", "lcd_data4", "GPIO2_10"}, []string{"gpio"}, 74, 0),
		d.makePin([]string{"P8.42", "lcd_data5", "GPIO2_11"}, []string{"gpio"}, 75, 0),
		d.makePin([]string{"P8.43", "lcd_data2", "GPIO2_8"}, []string{"gpio"}, 72, 0),
		d.makePin([]string{"P8.44", "lcd_data3", "GPIO2_9"}, []string{"gpio"}, 73, 0),
		d.makePin([]string{"P8.45", "lcd_data0", "GPIO2_6"}, []string{"gpio"}, 70, 0),
		// makePin("P8.46", bbGpioProfile, "GPIO2_7", 2, 7, "lcd_data1", 0),

		// P9
		d.makePin([]string{"P9.11", "gpmc_wait0", "GPIO0_30"}, []string{"gpio"}, 30, 0),
		d.makePin([]string{"P9.12", "gpmc_ben1", "GPIO1_28"}, []string{"gpio"}, 60, 0),
		d.makePin([]string{"P9.13", "gpmc_wpn", "GPIO0_31"}, []string{"gpio"}, 31, 0),
		d.makePin([]string{"P9.14", "gpmc_a2", "GPIO1_18"}, []string{"gpio"}, 50, 0),
		d.makePin([]string{"P9.15", "gpmc_a0", "GPIO1_16"}, []string{"gpio"}, 48, 0),
		d.makePin([]string{"P9.16", "gpmc_a3", "GPIO1_19"}, []string{"gpio"}, 51, 0),
		d.makePin([]string{"P9.17", "spi0_cs0", "GPIO0_5"}, []string{"gpio"}, 5, 0),
		d.makePin([]string{"P9.18", "spi0_d1", "GPIO0_4"}, []string{"gpio"}, 4, 0),
		d.makePin([]string{"P9.19", "uart1_rtsn", "GPIO0_13"}, []string{"gpio"}, 13, 0),
		d.makePin([]string{"P9.20", "uart1_ctsn", "GPIO0_12"}, []string{"gpio"}, 12, 0),
		d.makePin([]string{"P9.21", "spi0_d0", "GPIO0_3"}, []string{"gpio"}, 3, 0),
		d.makePin([]string{"P9.22", "spi0_sclk", "GPIO0_2"}, []string{"gpio"}, 2, 0),
		d.makePin([]string{"P9.23", "gpmc_a1", "GPIO1_17"}, []string{"gpio"}, 49, 0),
		d.makePin([]string{"P9.24", "uart1_txd", "GPIO0_15"}, []string{"gpio"}, 15, 0),
		d.makePin([]string{"P9.25", "mcasp0_ahclkx", "GPIO3_21"}, []string{"gpio"}, 117, 0),
		d.makePin([]string{"P9.26", "uart1_rxd", "GPIO0_14"}, []string{"gpio"}, 14, 0),
		d.makePin([]string{"P9.27", "mcasp0_fsr", "GPIO3_19"}, []string{"gpio"}, 115, 0),
		d.makePin([]string{"P9.28", "mcasp0_ahclkr", "GPIO3_17"}, []string{"gpio"}, 113, 0),
		d.makePin([]string{"P9.29", "mcasp0_fsx", "GPIO3_15"}, []string{"gpio"}, 111, 0),
		d.makePin([]string{"P9.30", "mcasp0_axr0", "GPIO3_16"}, []string{"gpio"}, 112, 0),
		d.makePin([]string{"P9.31", "mcasp0_aclkx", "GPIO3_14"}, []string{"gpio"}, 110, 0),
		d.makePin([]string{"P9.33", "ain4"}, []string{"analog"}, 0, 4),
		d.makePin([]string{"P9.35", "ain6"}, []string{"analog"}, 0, 6),
		d.makePin([]string{"P9.36", "ain5"}, []string{"analog"}, 0, 5),
		d.makePin([]string{"P9.37", "ain2"}, []string{"analog"}, 0, 2),
		d.makePin([]string{"P9.38", "ain3"}, []string{"analog"}, 0, 3),
		d.makePin([]string{"P9.39", "ain0"}, []string{"analog"}, 0, 0),
		d.makePin([]string{"P9.40", "ain1"}, []string{"analog"}, 0, 1),
		d.makePin([]string{"P9.41", "xdma_event_intr1", "GPIO0_20"}, []string{"gpio"}, 20, 0),
		d.makePin([]string{"P9.42", "ecap0_in_pwm0_out", "GPIO0_7"}, []string{"gpio"}, 7, 0),

		// @todo work out what to do with the USR LEDs. These are actually connected to GPIO, but don't work it you treat
		// @todo as GPIO as it used to. Probably wants it's own BBB-specific module.
		// // USR LEDs
		// d.makePin("USR0", bbUsrLedProfile, "USR0", 1, 21, "gpmc_a5", 0),
		// d.makePin("USR1", bbUsrLedProfile, "USR1", 1, 22, "gpmc_a6", 0),
		// d.makePin("USR2", bbUsrLedProfile, "USR2", 1, 23, "gpmc_a7", 0),
		// d.makePin("USR3", bbUsrLedProfile, "USR3", 1, 24, "gpmc_a8", 0),
	}
}

func (d *BeagleBoneBlackDriver) initialiseModules() error {
	d.modules = make(map[string]Module)

	gpio := NewDTGPIOModule("gpio")
	e := gpio.SetOptions(d.getGPIOOptions())
	if e != nil {
		return e
	}

	analog := NewDTAnalogModule("analog")
	e = analog.SetOptions(d.getAnalogOptions())
	if e != nil {
		return e
	}

	// @todo get the I2C interface working.
	// i2c1 := NewDTI2CModule("i2c1")

	d.modules["gpio"] = gpio
	d.modules["analog"] = analog
	// d.modules["i2c1"] = i2c1

	return nil
}

// Get options for GPIO module, derived from the pin structure
func (d *BeagleBoneBlackDriver) getGPIOOptions() map[string]interface{} {
	result := make(map[string]interface{})

	pins := make(DTGPIOModulePinDefMap)

	// Add the GPIO pins to this map
	for i, hw := range d.beaglePins {
		if d.usedBy(hw, "gpio") {
			pins[Pin(i)] = &DTGPIOModulePinDef{pin: Pin(i), gpioLogical: hw.gpioLogical}
		}
	}
	result["pins"] = pins

	return result
}

// Get options for analog module, derived from the pin structure
func (d *BeagleBoneBlackDriver) getAnalogOptions() map[string]interface{} {
	result := make(map[string]interface{})

	pins := make(DTAnalogModulePinDefMap)

	// Add the GPIO pins to this map
	for i, hw := range d.beaglePins {
		if d.usedBy(hw, "analog") {
			pins[Pin(i)] = &DTAnalogModulePinDef{pin: Pin(i), analogLogical: hw.analogLogical}
		}
	}
	result["pins"] = pins

	return result
}

// Determine if the pin is used by the module
func (d *BeagleBoneBlackDriver) usedBy(pinDef *BeaglePin, module string) bool {
	for _, n := range pinDef.modules {
		if n == module {
			return true
		}
	}
	return false
}

func (d *BeagleBoneBlackDriver) GetModules() map[string]Module {
	return d.modules
}

func (d *BeagleBoneBlackDriver) Close() {
	// Disable all the modules
	for _, module := range d.modules {
		module.Disable()
	}
}

func (d *BeagleBoneBlackDriver) PinMap() (pinMap HardwarePinMap) {
	pinMap = make(HardwarePinMap)

	for i, hw := range d.beaglePins {
		pinMap.add(Pin(i), hw.names, hw.modules)
	}

	return
}
