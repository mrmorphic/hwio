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

type AnalogModule interface {
	Module

	AnalogRead(pin Pin) (result int, e error)

	// read
	// reference voltage
}

// Interface for I2C implementations
type I2CModule interface {
	Module
	Write(address int, buffer []byte) (e error)
	Read(address int, buffer []byte) (nBytes int, e error)
	Available(address int) (avail bool, e error)
}

type SPIModule interface {
	Module

	// Select the device, and send data to it
	Write(data []byte) (e error)

	// Select the device, and read data from it
	Read([]byte) (nBytes int, e error)
}
