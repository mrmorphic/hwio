package hwio

import (
	"errors"
	"fmt"
	"strings"
)

// This is a module to support the onboard LED functions. While these are actually attached to GPIO pins that
// are not exposed on the expansion headers, we can't use GPIO, as a driver is present that provides ways
// to map what is displayed on the LEDs.
type (
	DTLEDModule struct {
		name        string
		definedPins DTLEDModulePins

		leds map[string]*DTLEDModuleLED
	}

	DTLEDModuleLED struct {
		path           string
		currentTrigger string
	}

	// A map of pin names (e.g. "USR0") to their path e.g. /sys/class/leds/{led}/
	DTLEDModulePins map[string]string
)

func NewDTLEDModule(name string) *DTLEDModule {
	return &DTLEDModule{name: name, leds: make(map[string]*DTLEDModuleLED)}
}

func (m *DTLEDModule) Enable() error {
	return nil
}

func (m *DTLEDModule) Disable() error {
	return nil
}

func (m *DTLEDModule) GetName() string {
	return m.name
}

func (m *DTLEDModule) SetOptions(options map[string]interface{}) error {
	// get the pins
	if p := options["pins"]; p != "" {
		m.definedPins = p.(DTLEDModulePins)

		return nil
	} else {
		return fmt.Errorf("Module '%s' SetOptions() did not get 'pins' value", m.GetName())
	}

}

// Get a LED to manipulate. 'led' must be 0 to 3.
func (m *DTLEDModule) GetLED(led string) (LEDModuleLED, error) {
	led = strings.ToLower(led)

	if ol := m.leds[led]; ol != nil {
		return ol, nil
	}

	if pin := m.definedPins[led]; pin != "" {
		result := &DTLEDModuleLED{}
		result.path = pin
		result.currentTrigger = ""
		m.leds[led] = result
		return result, nil
	} else {
		return nil, fmt.Errorf("GetLED: invalid led '%s'", led)
	}
}

// Set the trigger for the LED. The values come from /sys/class/leds/*/trigger. This tells the driver what should be displayed on the
// LED. The useful values include:
// - none		The LED can be set up programmatic control. If you want to turn a LED on and off yourself, you want
//				this mode.
// - nand-disk	Automatically displays nand disk activity
// - mmc0		Show MMC0 activity.
// - mmc1		Show MMC1 activity. By default, USR3 is configured for mmc1.
// - timer
// - heartbeat	Show a heartbeat for system functioning. By default, USR0 is configured for heartbeat.
// - cpu0		Show CPU activity. By default, USR2 is configured for cpu0.
// For BeagleBone black system defaults (at least for Angstrom are):
// - USR0: heartbeat
// - USR1: mmc0
// - USR2: cpu0
// - USR3: mmc1
// For Raspberry Pi is mmc0.
func (led *DTLEDModuleLED) SetTrigger(trigger string) error {
	led.currentTrigger = trigger
	return WriteStringToFile(led.path+"trigger", trigger)
}

func (led *DTLEDModuleLED) SetOn(on bool) error {
	if led.currentTrigger != "none" {
		return errors.New("LED SetOn requires that the LED trigger has been set to 'none'")
	}

	v := "0"
	if on {
		v = "1"
	}

	return WriteStringToFile(led.path+"brightness", v)
}
