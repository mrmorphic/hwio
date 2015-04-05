package hwio

// DTPinConfig represents configuration information for a pin used in a device-tree based driver,
// which typically hold the same details.
type DTPinConfig struct {
	// A list of names that this pin will be known by. They are effectively synonyms; retrieving a pin
	// by one of these names does not automatically cause the pin's behaviour to change.
	names []string

	// Names of modules that may allocate this pin
	modules []string

	// logical number for GPIO, for pins used by "gpio" module. This is the GPIO port number plus the GPIO pin within the port.
	gpioLogical int

	// analog pin number, for pins used by "analog" module
	analogLogical int
}

// Determine if the pin is used by the module
func (c *DTPinConfig) usedBy(module string) bool {
	for _, n := range c.modules {
		if n == module {
			return true
		}
	}
	return false
}
