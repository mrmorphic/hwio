package hwio

import "fmt"

// This is a dummy module for devices that have pins that are pre-assigned, but not to any of the supported
// modules on the device. It is passed a list of these pre-assigned pins. e.g. on BeagleBone Black, it covers
// pins that are defined for HDMI, MMC and mcasp0. On the default configuration, these pins are pre-assigned
// with device tree configuration, so they cannot be assigned for gpio (without custom device tree)
type PreassignedModule struct {
	name string
	pins PinList
}

func NewPreassignedModule(name string) (result *PreassignedModule) {
	result = &PreassignedModule{name: name}
	return result
}

func (module *PreassignedModule) SetOptions(options map[string]interface{}) error {
	// get the pins
	vp := options["pins"]
	if vp == nil {
		return fmt.Errorf("Module '%s' SetOptions() did not get 'pins' values", module.GetName())
	}
	module.pins = vp.(PinList)

	return nil
}

func (module *PreassignedModule) Enable() error {
	return AssignPins(module.pins, module)
}

func (module *PreassignedModule) Disable() error {
	return UnassignPins(module.pins)
}

func (module *PreassignedModule) GetName() string {
	return module.name
}
