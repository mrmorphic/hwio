# HD44780 over I2C

This provides control of an HD44780-compatible device that is connected via an I2C expander.

# Usage

Import the packages:

	// import the require modules
	import(
		"hwio"
		"hwio/i2c/hd44780"
	)

Initialise by fetching an i2c module from the driver. You can get instances of devices attached to
the bus.

	// Get the i2c module from the driver. This is an example for the BeagleBone Black, which exposes i2c2.
	m, e := hwio.GetModule("i2c2")

	// Assert that it is an I2C module
	i2c := m.(I2CModule)


Get the HD44780 device, so you make requests of it:

	// Get a display device on this i2c bus
	display := hd44780.NewHD44780(i2c)

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

# Notes

This has been tested on an mjkdz i2c expander and a 20x4 character display, and works correctly. Other LCD i2c expanders
may map En, Rw and Rs pins differently, however. If you have issues with a different I2C device, let me know, or submit
a pull request with the configuration that works for you.
