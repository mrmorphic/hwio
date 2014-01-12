# GY-520 (MPU-6050) I2C

This provides a simple way to access the sensor values of a GY-520 that is connected to an i2c bus on your system.

# Usage

Import the packages:

	// import the require modules
	import(
		"github.com/mrmorphic/hwio"
		"github.com/mrmorphic/hwio/devices/gy520"
	)

Initialise by fetching an i2c module from the driver. You can get instances of devices attached to
the bus.

	// Get the i2c module from the driver. This is an example for the BeagleBone Black, which exposes i2c2.
	m, e := hwio.GetModule("i2c2")

	// Assert that it is an I2C module
	i2c := m.(I2CModule)

Get the GY520 device, so you make requests of it:

	// Get a gyro device on this i2c bus
	gyro := gy520.NewGY520(i2c)

	// gyro is asleep by default, to save power
	gyro.Wake()

Read values from the device whenever you want to:

	// Get the gyroscope x, y and z sensor values
	gx, gy, gz, e := gyro.GetGyro()

	// Get the accelerometer x, y and z sensor values
	ax, ay, az, e := gyro.GetAccel()

	// Get the temperature
	temp, e := gyro.GetTemp()

Note that you will need to calibrate your device to make sense of the values coming out.