package hwio

import (
	"strings"
)

// Definitions relating to pins.
type PinIOMode int

// The modes for PinMode.
const (
	INPUT PinIOMode = iota
	OUTPUT
	INPUT_PULLUP
	INPUT_PULLDOWN
	INPUT_ANALOG
	// @todo PinIOMode consider an input_analog mode for analog pins. Not an issue on beaglebone,
	// @todo    but could be an issue on devices that support digital and analog input on the same pin,
	// @todo    if such devices exist.
)

// String representation of pin IO mode
func (mode PinIOMode) String() string {
	switch mode {
	case INPUT:
		return "INPUT"
	case OUTPUT:
		return "OUTPUT"
	case INPUT_PULLUP:
		return "INPUT_PULLUP"
	case INPUT_PULLDOWN:
		return "INPUT_PULLDOWN"
	case INPUT_ANALOG:
		return "INPUT_ANALOG"
	}
	return ""
}

// Convenience constants for digital pin values.
const (
	HIGH = 1
	LOW  = 0
)

type Pin int

type PinDef struct {
	pin     Pin      // the pin, also in the map key of HardwarePinMap
	names   []string // a list of names for the pin as defined by driver. There should be at least one. The first is the canonical name.
	modules []string // a list of module names that can use this pin
}

type PinList []Pin

type HardwarePinMap map[Pin]*PinDef

// Add a pin to the map
func (m HardwarePinMap) add(pin Pin, names []string, modules []string) {
	m[pin] = &PinDef{pin, names, modules}
}

// Given a pin number, return it's PinDef, or nil if that pin is not defined in the map
func (m HardwarePinMap) GetPin(pin Pin) *PinDef {
	return m[pin]
}

// Provide a string representation of a logic pin and the capabilties it
// supports.
func (pd *PinDef) String() string {
	s := pd.Names() + "  modules:" + strings.Join(pd.modules, ",")
	return s
}

// From the hwPinRefs, construct a string by appending them together. Not brilliantly efficient,
// but its most for diagnostics anyway.
func (pd *PinDef) Names() string {
	return strings.Join(pd.names, ",")
}

// // Determine if a pin has a particular capability.
// func (pd *PinDef) HasCapability(cap Capability) bool {
// 	//	fmt.Printf("HasCap: checking (%s) has capability %s", pd.String(), cap.String())
// 	for _, v := range pd.capabilities {
// 		if v == cap {
// 			return true
// 		}
// 	}
// 	return false
// }
