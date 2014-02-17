# TMP-102 I2C

This provides a simple way to access the sensor value of a TMP-102 that is connected to an i2c bus on your system.

# Usage

Import the packages:

	// import the require modules
	import(
		"github.com/mrmorphic/hwio"
		"github.com/mrmorphic/hwio/devices/tmp102"
	)

Initialise by fetching an i2c module from the driver. You can get instances of devices attached to
the bus.

	// Get the i2c module from the driver. This is an example for the Raspberry Pi.
	m, e := hwio.GetModule("i2c")

	// Assert that it is an I2C module
	i2c := m.(I2CModule)

Get the TMP102 device, so you make requests of it:

	// Get a temp device on this i2c bus
	temp := tmp102.NewTMP102(i2c)

Read values from the device whenever you want to:

	// Get the temperature sensor value
	t, e := temp.GetTemp()
