# HD44780 over I2C

This provides control of an HD44780-compatible device that is connected via an I2C expander.

# Usage

Import the packages:

	// import the require modules
	import(
		"github.com/mrmorphic/hwio"
		"github.com/mrmorphic/hwio/devices/hd44780"
	)

Initialise by fetching an i2c module from the driver. You can get instances of devices attached to
the bus.

	// Get the i2c module from the driver. i2c is the canonical name. On the BeagleBone, it can also
	// be referred to as i2c2.
	m, e := hwio.GetModule("i2c")

	// Assert that it is an I2C module
	i2c := m.(hwio.I2CModule)


Get the HD44780 device, so you make requests of it:

	// Get a display device on this i2c bus. You need to provide it a device profile. Two profiles
	// are currently implemented, PROFILE_MJKDZ and PROFILE_PCF8574, corresponding to two commonly found
	// port extenders used to interface to LCD displays.
	display := hd44780.NewHD44780(i2c, 0x20, hd44780.PROFILE_MJKDZ)

	// Initialise the display with the size of display you have.
	display.Init(20, 4)

	// The display may not show anything if the backlight is turned off
	display.SetBacklight(true)

	// If you want to see the cursor, you can turn it on
	display.Cursor()

To display things, you can:

	// clear the display
	display.Clear()

	// Set cursor back to (0,0)
	display.Home()

	// Set cursor to a specific column and row (both zero based)
	display.SetCursor(0, 1)  // second line

	// output a single character
	display.Data('A')

	// Use any function that expects a Writer to output to the display. This is because the HD44780 type
	// implements Writer interface.
	fmt.Fprintf(display, "Hi, %s", name)

Note that characters that are output to the device are not necessarily displayed consequetively. In particular wrapping may not work
as you expect. This is because of how the display unit maps it's display buffer to positions on the screen. This is described in
the datasheet for the HD44780 unit.

An alternative way to create the device is to use NewHD44780Extended instead of NewHD44780. This is useful if you have an i2c extender that
does not confirm to the builtin profiles:

	display := NewHD44780Extended(i2c, 0x27,
		0, // en
		1, // rw
		2, // rs
		4, // d4
		5, // d5
		6, // d6
		7, // d7polarity int) *HD44780 {
		3, // backlight,
		hd44780.POSITIVE) // polarity

The pin values are the bit positions for that pin, with 7 being MSB and0 being LSB. The underlying assumption
is that the port extender is 8 bit. This package will not work for 16 bit extenders, for example.

# Notes

This has been tested on an mjkdz i2c expander and a 20x4 character display, and works correctly. Other LCD i2c expanders
may map En, Rw and Rs pins differently, however. If you have issues with a different I2C device, let me know, or submit
a pull request with the configuration that works for you.
