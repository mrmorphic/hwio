# Nintendo Wii Nunchuck over I2C

This provides a simple way to access the sensor values of a nunchuck that is connected to an i2c bus on your system.

# Usage

Import the packages:

	// import the require modules
	import(
		"github.com/mrmorphic/hwio"
		"github.com/mrmorphic/hwio/devices/nunchuck"
	)

Initialise by fetching an i2c module from the driver. You can get instances of devices attached to
the bus.

	// Get the i2c module from the driver.
	m, e := hwio.GetModule("i2c")

	// Assert that it is an I2C module
	i2c := m.(hwio.I2CModule)

Get the nunchuck device, so you make requests of it:

	// Get a gyro device on this i2c bus
	controller, e := nunchuck.NewNunchuck(i2c)
	if e != nil {
		return e
	}

To get values from the nunchuck, you need to call ReadSensors, which fetches all data from the device, and then use Get
methods to fetch the properies you want.

	e := controller.ReadSensors()

	// get joystick values (x2 int) from the last call to ReadSensors()
	joyX, joyY := controller.GetJoystick()

	// get accelerometer values (x3 float) from the last call to ReadSensors()
	ax, ay, az := controller.GetAccel()

	// get Z button pressed state (bool) from the last call to ReadSensors()
	zPressed := controller.GetZPressed()

	// get C button pressed state (bool) from the last call to ReadSensors()
	cPressed := controller.GetCPressed()

	// get roll and pitch values (float32) in degrees from the last call to ReadSensors(). Note these are computed from
	// the accelerometer values.
	roll := controller.GetRoll()
	pitch := controller.GetPitch()

The device joystick and accelerometer values are initially calibrated to zero, but you can change these, and will probably need to.
Until the accelerometer values are calibrated, roll and pitch values may not be meaningful.

	// calibrate the center position of the joystick to whatever the last read joystick values were.
	controller.CalibrateJoystick()

	// Set the zero values for the accelerometer to 3 values.
	controller.SetAccelZero(zeroX, zeroY, zeroZ)