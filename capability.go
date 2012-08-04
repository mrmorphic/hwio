package hwio

// Definitions for capabilities.
import (
	"strings"
)

//Define a generic way to represent pin capabilities.
type Capability int

const (
	CAP_INPUT          Capability = iota // digital input
	CAP_OUTPUT                           // digital output
	CAP_INPUT_PULLUP                     // digital input with pull up
	CAP_INPUT_PULLDOWN                   // digital input with pull down
	CAP_PWM                              // "analog" output using pwm
	CAP_ANALOG_IN                        // analog input using A/D converter
)

// This represents a set of capabilities that a pin may have. There may be multiple pins on a device that have identical
// capability set.
type CapabilitySet []Capability

func (c Capability) String() string {
	switch c {
	case CAP_INPUT:
		return "input"
	case CAP_OUTPUT:
		return "output"
	case CAP_INPUT_PULLUP:
		return "input_pullup"
	case CAP_INPUT_PULLDOWN:
		return "input_pulldown"
	case CAP_PWM:
		return "pwm"
	case CAP_ANALOG_IN:
		return "analog_in"
	}
	return ""
}

func (cs CapabilitySet) String() string {
	s := []string{}
	for _, c := range cs {
		s = append(s, Capability(c).String())
	}
	return strings.Join(s, ",")
}
