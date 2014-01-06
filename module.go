// Defines generic types and behaviours for modules. A given hardware platform will typically support one or more
// modules.

package hwio

// Generic interface type for all modules.
type Module interface {
	// Set parameters require to initialise the module. Generally should be called before Enable() is called,
	// but this may depend on the module.
	SetOptions(map[string]interface{}) error

	// enables the module for use.
	Enable() error

	// disables module and releases pins.
	Disable() error

	// Return the module name so it can be used for error reporting
	GetName() string
}

type GPIOModule interface {
	Module

	PinMode(pin Pin, mode PinIOMode) (e error)
	DigitalWrite(pin Pin, value int) (e error)
	DigitalRead(pin Pin) (result int, e error)
}

type PWMModule interface {
	Module

	EnablePin(pin Pin, enabled bool) error

	// Set the period of this pin, in nanoseconds
	SetPeriod(pin Pin, ns int64) error

	// Set the duty time, the amount of time during each period that that output is HIGH.
	SetDuty(pin Pin, ns int64) error
}

type AnalogModule interface {
	Module

	AnalogRead(pin Pin) (result int, e error)

	// read
	// reference voltage
}

// Interface for I2C implementations. Assumes that this device is the only bus master, so initiates all transactions. An I2C module
// supports exactly one i2c bus, so for systems with multiple i2c busses, the driver will create an instance for each accessible
// i2c bus.
type I2CModule interface {
	Module

	GetDevice(address int) I2CDevice
}

// An object that represents a device on a bus. Once an i2c module has been enabled, you can use GetDevice to get an instance
// of i2c device. You can then talk to the device directly with the supported operations.
type I2CDevice interface {
	// Read a single byte from a register on the device.
	ReadByte(command byte) (byte, error)

	// Write a single byte to a register on the device.
	WriteByte(command byte, value byte) error

	// Read one or more bytes from the selected slave.
	Read(command byte, numBytes int) ([]byte, error)

	// Write one or more bytes to the selected slave.
	Write(command byte, buffer []byte) (e error)
}

// Interface for SPI implementations
type SPIModule interface {
	Module

	// Select the device, and send data to it
	Write(slaveSelect int, data []byte) (e error)

	// Select the device, and read data from it
	Read(slaveSelect int, data []byte) (nBytes int, e error)
}
