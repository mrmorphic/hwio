package hwio

type DTI2CModule struct {
	name string
}

func NewDTI2CModule(name string) (result *DTI2CModule) {
	result = &DTI2CModule{name}
	return result
}

// enable GPIO module. It doesn't allocate any pins immediately.
func (module *DTI2CModule) Enable() error {
	return nil
}

// disables module and release any pins assigned.
func (module *DTI2CModule) Disable() error {
	return nil
}

func (module *DTI2CModule) GetName() string {
	return module.name
}

func (module *DTI2CModule) SetOptions(map[string]interface{}) {

}
