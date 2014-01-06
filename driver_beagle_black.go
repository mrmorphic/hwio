package hwio

import (
	"fmt"
	"strings"
)

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
	return d.initialiseModules()
}

func (d *BeagleBoneBlackDriver) makePin(names []string, modules []string, gpioLogical int, analogLogical int) *BeaglePin {
	return &BeaglePin{names, modules, gpioLogical, analogLogical}
}

func (d *BeagleBoneBlackDriver) createPinData() {
	d.beaglePins = []*BeaglePin{
		// P8
		d.makePin([]string{"P8.3", "gpmc_ad6", "gpio1_6"}, []string{"gpio", "emmc2", "preallocated"}, 38, 0), // preassigned via DT in default config
		d.makePin([]string{"P8.4", "gpmc_ad7", "gpio1_7"}, []string{"gpio", "emmc2", "preallocated"}, 39, 0), // preassigned via DT in default config
		d.makePin([]string{"P8.5", "gpmc_ad2", "gpio1_2"}, []string{"gpio", "emmc2", "preallocated"}, 34, 0), // preassigned via DT in default config
		d.makePin([]string{"P8.6", "gpmc_ad3", "gpio1_3"}, []string{"gpio", "emmc2", "preallocated"}, 35, 0), // preassigned via DT in default config
		d.makePin([]string{"P8.7", "gpmc_advn_ale", "gpio2_2"}, []string{"gpio"}, 66, 0),
		d.makePin([]string{"P8.8", "gpmc_oen_ren", "gpio2_3"}, []string{"gpio"}, 67, 0),
		d.makePin([]string{"P8.9", "gpmc_ben0_cle", "gpio2_5"}, []string{"gpio"}, 69, 0),
		d.makePin([]string{"P8.10", "gpmc_wen", "gpio2_4"}, []string{"gpio"}, 68, 0),
		d.makePin([]string{"P8.11", "gpmc_ad13", "gpio1_13"}, []string{"gpio"}, 45, 0),
		d.makePin([]string{"P8.12", "gpmc_ad12", "gpio1_12"}, []string{"gpio"}, 44, 0),
		d.makePin([]string{"P8.13", "gpmc_ad9", "gpio0_23", "ehrpwm2B"}, []string{"gpio", "pwm2"}, 23, 0),
		d.makePin([]string{"P8.14", "gpmc_ad10", "gpio0_26"}, []string{"gpio"}, 26, 0),
		d.makePin([]string{"P8.15", "gpmc_ad15", "gpio1_15"}, []string{"gpio"}, 47, 0),
		d.makePin([]string{"P8.16", "gpmc_ad14", "gpio1_14"}, []string{"gpio"}, 46, 0),
		d.makePin([]string{"P8.17", "gpmc_ad11", "gpio0_27"}, []string{"gpio"}, 27, 0),
		d.makePin([]string{"P8.18", "gpmc_clk", "gpio2_1"}, []string{"gpio"}, 65, 0),
		d.makePin([]string{"P8.19", "gpmc_ad8", "gpio0_22", "ehrpwm2A"}, []string{"gpio", "pwm2"}, 22, 0),
		d.makePin([]string{"P8.20", "gpmc_csn2", "gpio1_31"}, []string{"gpio", "emmc2", "preallocated"}, 63, 0), // preassigned via DT in default config
		d.makePin([]string{"P8.21", "gpmc_csn1", "gpio1_30"}, []string{"gpio", "emmc2", "preallocated"}, 62, 0), // preassigned via DT in default config
		d.makePin([]string{"P8.22", "gpmc_ad5", "gpio1_5"}, []string{"gpio", "emmc2", "preallocated"}, 37, 0),   // preassigned via DT in default config
		d.makePin([]string{"P8.23", "gpmc_ad4", "gpio1_4"}, []string{"gpio", "emmc2", "preallocated"}, 36, 0),   // preassigned via DT in default config
		d.makePin([]string{"P8.24", "gpmc_ad1", "gpio1_1"}, []string{"gpio", "emmc2", "preallocated"}, 33, 0),   // preassigned via DT in default config
		d.makePin([]string{"P8.25", "gpmc_ad0", "gpio1_0"}, []string{"gpio", "emmc2", "preallocated"}, 32, 0),   // preassigned via DT in default config
		d.makePin([]string{"P8.26", "gpmc_csn0", "gpio1_29"}, []string{"gpio"}, 61, 0),
		d.makePin([]string{"P8.27", "lcd_vsync", "gpio2_22"}, []string{"gpio", "hdmi", "preallocated"}, 86, 0),          // preassigned via DT in default config
		d.makePin([]string{"P8.28", "lcd_pclk", "gpio2_24"}, []string{"gpio", "hdmi", "preallocated"}, 88, 0),           // preassigned via DT in default config
		d.makePin([]string{"P8.29", "lcd_hsync", "gpio2_23"}, []string{"gpio", "hdmi", "preallocated"}, 87, 0),          // preassigned via DT in default config
		d.makePin([]string{"P8.30", "lcd_ac_bias_en", "gpio2_25"}, []string{"gpio", "hdmi", "preallocated"}, 89, 0),     // preassigned via DT in default config
		d.makePin([]string{"P8.31", "lcd_data14", "gpio0_10"}, []string{"gpio", "hdmi", "preallocated"}, 10, 0),         // preassigned via DT in default config
		d.makePin([]string{"P8.32", "lcd_data15", "gpio0_11"}, []string{"gpio", "hdmi", "preallocated"}, 11, 0),         // preassigned via DT in default config
		d.makePin([]string{"P8.33", "lcd_data13", "gpio0_9"}, []string{"gpio", "hdmi", "preallocated"}, 9, 0),           // preassigned via DT in default config
		d.makePin([]string{"P8.34", "lcd_data11", "gpio2_17"}, []string{"gpio", "hdmi", "pwm1", "preallocated"}, 81, 0), // preassigned via DT in default config
		d.makePin([]string{"P8.35", "lcd_data12", "gpio0_8"}, []string{"gpio", "hdmi", "preallocated"}, 8, 0),           // preassigned via DT in default config
		d.makePin([]string{"P8.36", "lcd_data10", "gpio2_16"}, []string{"gpio", "hdmi", "pwm1", "preallocated"}, 80, 0), // preassigned via DT in default config
		d.makePin([]string{"P8.37", "lcd_data8", "gpio2_14"}, []string{"gpio", "hdmi", "preallocated"}, 78, 0),          // preassigned via DT in default config
		d.makePin([]string{"P8.38", "lcd_data9", "gpio2_15"}, []string{"gpio", "hdmi", "preallocated"}, 79, 0),          // preassigned via DT in default config
		d.makePin([]string{"P8.40", "lcd_data7", "gpio2_13"}, []string{"gpio", "hdmi", "preallocated"}, 77, 0),          // preassigned via DT in default config
		d.makePin([]string{"P8.41", "lcd_data4", "gpio2_10"}, []string{"gpio", "hdmi", "preallocated"}, 74, 0),          // preassigned via DT in default config
		d.makePin([]string{"P8.42", "lcd_data5", "gpio2_11"}, []string{"gpio", "hdmi", "preallocated"}, 75, 0),          // preassigned via DT in default config
		d.makePin([]string{"P8.43", "lcd_data2", "gpio2_8"}, []string{"gpio", "hdmi", "preallocated"}, 72, 0),           // preassigned via DT in default config
		d.makePin([]string{"P8.44", "lcd_data3", "gpio2_9"}, []string{"gpio", "hdmi", "pwm2", "preallocated"}, 73, 0),   // preassigned via DT in default config
		d.makePin([]string{"P8.45", "lcd_data0", "gpio2_6"}, []string{"gpio", "hdmi", "pwm2", "preallocated"}, 70, 0),   // preassigned via DT in default config
		// makePin("P8.46", bbGpioProfile, "gpio2_7", 2, 7, "lcd_data1", 0),

		// P9
		d.makePin([]string{"P9.11", "gpmc_wait0", "gpio0_30"}, []string{"gpio"}, 30, 0),
		d.makePin([]string{"P9.12", "gpmc_ben1", "gpio1_28"}, []string{"gpio"}, 60, 0),
		d.makePin([]string{"P9.13", "gpmc_wpn", "gpio0_31"}, []string{"gpio"}, 31, 0),
		d.makePin([]string{"P9.14", "gpmc_a2", "gpio1_18"}, []string{"gpio"}, 50, 0),
		d.makePin([]string{"P9.15", "gpmc_a0", "gpio1_16"}, []string{"gpio"}, 48, 0),
		d.makePin([]string{"P9.16", "gpmc_a3", "gpio1_19"}, []string{"gpio"}, 51, 0),
		d.makePin([]string{"P9.17", "spi0_cs0", "gpio0_5"}, []string{"gpio"}, 5, 0),
		d.makePin([]string{"P9.18", "spi0_d1", "gpio0_4"}, []string{"gpio"}, 4, 0),
		d.makePin([]string{"P9.19", "uart1_rtsn", "gpio0_13"}, []string{"gpio", "i2c2"}, 13, 0), // preassigned via DT in default config
		d.makePin([]string{"P9.20", "uart1_ctsn", "gpio0_12"}, []string{"gpio", "i2c2"}, 12, 0), // preassigned via DT in default config
		d.makePin([]string{"P9.21", "spi0_d0", "gpio0_3", "ehrpwm0B"}, []string{"gpio", "pwm0"}, 3, 0),
		d.makePin([]string{"P9.22", "spi0_sclk", "gpio0_2", "ehrpwm0A"}, []string{"gpio", "pwm0"}, 2, 0),
		d.makePin([]string{"P9.23", "gpmc_a1", "gpio1_17"}, []string{"gpio"}, 49, 0),
		d.makePin([]string{"P9.24", "uart1_txd", "gpio0_15"}, []string{"gpio"}, 15, 0),
		d.makePin([]string{"P9.25", "mcasp0_ahclkx", "gpio3_21"}, []string{"gpio", "mcasp0", "preallocated"}, 117, 0), // preassigned via DT in default config
		d.makePin([]string{"P9.26", "uart1_rxd", "gpio0_14"}, []string{"gpio"}, 14, 0),
		d.makePin([]string{"P9.27", "mcasp0_fsr", "gpio3_19"}, []string{"gpio"}, 115, 0),
		d.makePin([]string{"P9.28", "mcasp0_ahclkr", "gpio3_17"}, []string{"gpio", "mcasp0", "preallocated"}, 113, 0),      // preassigned via DT in default config
		d.makePin([]string{"P9.29", "mcasp0_fsx", "gpio3_15"}, []string{"gpio", "mcasp0", "pwm0", "preallocated"}, 111, 0), // preassigned via DT in default config
		d.makePin([]string{"P9.30", "mcasp0_axr0", "gpio3_16"}, []string{"gpio"}, 112, 0),
		d.makePin([]string{"P9.31", "mcasp0_aclkx", "gpio3_14"}, []string{"gpio", "mcasp0", "pwm0", "preallocated"}, 110, 0), // preassigned via DT in default config
		d.makePin([]string{"P9.33", "ain4"}, []string{"analog"}, 0, 4),
		d.makePin([]string{"P9.35", "ain6"}, []string{"analog"}, 0, 6),
		d.makePin([]string{"P9.36", "ain5"}, []string{"analog"}, 0, 5),
		d.makePin([]string{"P9.37", "ain2"}, []string{"analog"}, 0, 2),
		d.makePin([]string{"P9.38", "ain3"}, []string{"analog"}, 0, 3),
		d.makePin([]string{"P9.39", "ain0"}, []string{"analog"}, 0, 0),
		d.makePin([]string{"P9.40", "ain1"}, []string{"analog"}, 0, 1),
		d.makePin([]string{"P9.41", "xdma_event_intr1", "gpio0_20"}, []string{"gpio"}, 20, 0),
		d.makePin([]string{"P9.42", "ecap0_in_pwm0_out", "gpio0_7"}, []string{"gpio"}, 7, 0),

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

	i2c2 := NewDTI2CModule("i2c2")
	e = i2c2.SetOptions(d.getI2C2Options())
	if e != nil {
		return e
	}

	preallocated := newbbbPreassignedModule("preallocated")
	e = preallocated.SetOptions(d.getPreallocatedOptions())
	if e != nil {
		return e
	}

	pwm0 := NewBBPWMModule("pwm0")
	e = pwm0.SetOptions(d.getPWMOptions("pwm0"))
	if e != nil {
		return e
	}

	pwm1 := NewBBPWMModule("pwm1")
	e = pwm1.SetOptions(d.getPWMOptions("pwm1"))
	if e != nil {
		return e
	}
	pwm2 := NewBBPWMModule("pwm2")
	e = pwm2.SetOptions(d.getPWMOptions("pwm2"))
	if e != nil {
		return e
	}

	d.modules["gpio"] = gpio
	d.modules["analog"] = analog
	d.modules["i2c2"] = i2c2
	d.modules["pwm0"] = pwm0
	d.modules["pwm1"] = pwm1
	d.modules["pwm2"] = pwm2

	// alias i2c to i2c2. This is for portability; getting the i2c module on any device should return the default i2c interface,
	// but should not preclude addition of other i2c busses.
	d.modules["i2c"] = i2c2

	// these are the pre-allocated pins
	d.modules["preallocated"] = preallocated

	// initialise by default, which will assign P9.19 and P9.20. This is configured by default in device tree and these pins cannot be assigned.
	i2c2.Enable()
	preallocated.Enable()

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

// Return the i2c options required to initialise that module.
func (d *BeagleBoneBlackDriver) getI2C2Options() map[string]interface{} {
	result := make(map[string]interface{})

	pins := make(DTI2CModulePins, 0)
	p19 := d.getPin("P9.19")
	pins = append(pins, p19)
	p20 := d.getPin("P9.20")
	pins = append(pins, p20)

	result["pins"] = pins

	// this should really look at the device structure to ensure that I2C2 on hardware maps to /dev/i2c1. This confusion seems
	// to happen because of the way the kernel initialises the devices at boot time.
	result["device"] = "/dev/i2c-1"

	return result
}

func (d *BeagleBoneBlackDriver) getPWMOptions(name string) map[string]interface{} {
	result := make(map[string]interface{})

	pins := make(BBPWMModulePinDefMap, 0)
	for i, hw := range d.beaglePins {
		if d.usedBy(hw, name) {
			n := hw.names[0]
			n = strings.Replace(n, ".", "_", -1) // P8.13 => P8_13
			pins[Pin(i)] = &BBPWMModulePinDef{pin: Pin(i), name: n}
		}
	}

	result["pins"] = pins

	return result
}

// internal function to get a Pin. It does not use GetPin because that relies on the driver having already been initialised. This
// method can be called while stil initialising. Only matches names[0], which is the Pn.nn expansion header name.
func (d *BeagleBoneBlackDriver) getPin(name string) Pin {
	for i, hw := range d.beaglePins {
		if hw.names[0] == name {
			return Pin(i)
		}
	}
	return Pin(0)
}

func (d *BeagleBoneBlackDriver) getPreallocatedOptions() map[string]interface{} {
	result := make(map[string]interface{})

	pins := make(PinList, 0)

	// Add all pre-allocated pins to this map (excludes pre-allocated pins picked up by other modules, eg i2c2.)
	for i, hw := range d.beaglePins {
		if d.usedBy(hw, "preallocated") {
			pins = append(pins, Pin(i))
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

// This is a dummy module on the BeagleBone Black. It is passed a list of pre-assigned pins for which
// there is no other module. It covers pins that are defined for HDMI, MMC and mcasp0. On the default
// configuration, these pins are pre-assigned with device tree configuration, so they cannot be assigned
// for gpio (without custom device tree)
type bbbPreassignedModule struct {
	name string
	pins PinList
}

func newbbbPreassignedModule(name string) (result *bbbPreassignedModule) {
	result = &bbbPreassignedModule{name: name}
	return result
}

func (module *bbbPreassignedModule) SetOptions(options map[string]interface{}) error {
	// get the pins
	vp := options["pins"]
	if vp == nil {
		return fmt.Errorf("Module '%s' SetOptions() did not get 'pins' values", module.GetName())
	}

	module.pins = vp.(PinList)

	return nil
}

func (module *bbbPreassignedModule) Enable() error {
	return AssignPins(module.pins, module)
}

func (module *bbbPreassignedModule) Disable() error {
	return UnassignPins(module.pins)
}

func (module *bbbPreassignedModule) GetName() string {
	return module.name
}
