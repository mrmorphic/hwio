# MCP-23017 I2C Port Expander

This package provides a simple way to connect to the MCP-23017 port expander, a device which exposes 2 8-bit
GPIO ports address via I2C. 

# Usage

Import the packages:

	// import the require modules
	import(
		"github.com/mrmorphic/hwio"
		"github.com/mrmorphic/hwio/devices/mcp23017"
	)

Initialise by fetching an i2c module from the driver. You can get instances of devices attached to
the bus.

	// Get the i2c module from the driver. This is an example for the BeagleBone Black, which exposes i2c2.
	m, e := hwio.GetModule("i2c2")

	// Assert that it is an I2C module
	i2c := m.(hwio.I2CModule)

Get the MCP-23017 device, so you make requests of it:

	// Get the device on this i2c bus. 0 assumes A2, A1 and A0 on the device are grounded.
	expander := mcp23017.NewMCP23017(i2c, 0)

Typically you'll want to get the directions of the pins first:

	expander.SetDirA(0xc0)  // set two high bits of port A to inputs, the rest are outputs.
	expander.SetDirB(0x00)  // set all bits low on port B, to make them all outputs.

The device has configurable pull-resistors of about 100K that can be enabled (disabled by default):

	expander.SetPullupA(0x40)  // set bit 6 of port A to pull-up. It must have been set to an input.

To read raw input values from the ports:

	v, e := expander.GetPortA()
	v2, e = expander.GetPortB()

To write to output ports:

	e := expander.SetPortA(0x80)
	e := expander.SetPortA(0x7b)
