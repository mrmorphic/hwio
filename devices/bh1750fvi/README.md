# BH1750FVI Light sensor I2C

This provides a simple way to access the sensor values of a BH1750FVI light sensor that is connected to an i2c bus on your system.

# Usage

Import the packages:

	// import the require modules
	import(
		"github.com/mrmorphic/hwio"
		"github.com/mrmorphic/hwio/devices/bh1750fvi"
	)

Initialise by fetching an i2c module from the driver. You can get instances of devices attached to
the bus.

	// Get the i2c module from the driver. This is an example for the BeagleBone Black, which exposes i2c2.
	m, e := hwio.GetModule("i2c2")

	// Assert that it is an I2C module
	i2c := m.(I2CModule)

Get the device, so you make requests of it:

	// Get a gyro device on this i2c bus, using the default address (0x23)
	lightSensor := bh1750fvi.NewBH1750FVI(i2c)

	// Or if you want to use the alternate address (0x5c), use:
	lightSensor := bh1750fvi.NewBH1750FVIAddr(i2c, bh1750fvi.DEVICE_ADDRESS_ADDR_HIGH)


Read values from the device whenever you want to:

	value, e := lightSensor.ReadLightLevel(bh1750fvi.ONETIME_HIGH_RES)

The return value is a float32, which is the measure of lux.

The device supports a number of modes, which are declared as constants:

 *	CONTINUOUS_HIGH_RES
 *	CONTINUOUS_HIGH_RES_2
 *	CONTINUOUS_LOW_RES
 *	ONETIME_HIGH_RES
 *	ONETIME_HIGH_RES_2
 *	ONETIME_LOW_RES

The read process takes variable time based on which mode is used. Consult the datasheet for the device
to see what the modes mean and the sampling time.