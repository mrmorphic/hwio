// Implementation of I2C module interface for systems using device tree.

package hwio

// references:
// - http://datko.net/2013/11/03/bbb_i2c/
// - http://elinux.org/Interfacing_with_I2C_Devices
// - http://grokbase.com/t/gg/golang-nuts/1296kz4tkg/go-nuts-how-to-call-linux-kernel-method-i2c-smbus-write-byte-from-go
// - http://learn.adafruit.com/setting-up-io-python-library-on-beaglebone-black/i2c
// - https://bitbucket.org/gmcbay/i2c/src/1235f1776ee749f0eaeb6de69d8804a6dd70d9d5/i2c_bus.go?at=master
// - http://derekmolloy.ie/beaglebone/beaglebone-an-i2c-tutorial-interfacing-to-a-bma180-accelerometer/'

import (
	"fmt"
	"os"
	"sync"
	"syscall"
	"unsafe"
)

// A list of the pins that are allocated when the bus is enabled. We don't need to know what these mean within the
// module. We just need them so we can mark them allocated.
type DTI2CModulePins []Pin

type DTI2CModule struct {
	sync.Mutex

	name        string
	deviceFile  string
	definedPins DTI2CModulePins

	// File used to represent the bus once it's opened
	fd *os.File
}

// Data that is passed to/from ioctl calls
type i2c_smbus_ioctl_data struct {
	read_write uint8
	command    uint8
	size       int
	data       uintptr
}

// Constants used by ioctl, from i2c-dev.h
const (
	I2C_SMBUS_READ           = 1
	I2C_SMBUS_WRITE          = 0
	I2C_SMBUS_BYTE_DATA      = 2
	I2C_SMBUS_I2C_BLOCK_DATA = 8
	I2C_SMBUS_BLOCK_MAX      = 32

	// Talk to bus
	I2C_SMBUS = 0x0720

	// Set bus slave
	I2C_SLAVE = 0x0703
)

func NewDTI2CModule(name string) (result *DTI2CModule) {
	result = &DTI2CModule{name: name}
	return result
}

// Accept options for the I2C module. Expected options include:
// - "device" - a string that identifies the device file, e.g. "/dev/i2c-1".
// - "pins" - an object of type DTI2CModulePins that identifies the pins that will be assigned
//	 when this module is enabled.
func (module *DTI2CModule) SetOptions(options map[string]interface{}) error {
	// get the device
	vd := options["device"]
	if vd == nil {
		return fmt.Errorf("Module '%s' SetOptions() did not get 'device' value", module.GetName())
	}

	module.deviceFile = vd.(string)

	// get the pins
	vp := options["pins"]
	if vp == nil {
		return fmt.Errorf("Module '%s' SetOptions() did not get 'pins' values", module.GetName())
	}

	module.definedPins = vp.(DTI2CModulePins)

	return nil
}

// enable this I2C module
func (module *DTI2CModule) Enable() error {
	// Assign the pins so nothing else can allocate them.
	for _, pin := range module.definedPins {
		// fmt.Printf("assigning pin %d\n", pin)
		AssignPin(pin, module)
	}

	// @todo consider lazily opening the file. Since Enable is called automatically by BBB driver, this
	// @todo file will always be open even if i2c is not used.
	fd, e := os.OpenFile(module.deviceFile, os.O_RDWR, os.ModeExclusive)
	if e != nil {
		return e
	}
	module.fd = fd

	return nil
}

// disables module and release any pins assigned.
func (module *DTI2CModule) Disable() error {
	if e := module.fd.Close(); e != nil {
		return e
	}

	for _, pin := range module.definedPins {
		UnassignPin(pin)
	}

	return nil
}

func (module *DTI2CModule) GetName() string {
	return module.name
}

func (module *DTI2CModule) GetDevice(address int) I2CDevice {
	return NewDTI2CDevice(module, address)
}

type DTI2CDevice struct {
	module  *DTI2CModule
	address int
}

func NewDTI2CDevice(module *DTI2CModule, address int) *DTI2CDevice {
	return &DTI2CDevice{module, address}
}

func (device *DTI2CDevice) Write(command byte, data []byte) (e error) {
	device.module.Lock()
	defer device.module.Unlock()

	device.sendSlaveAddress()

	buffer := make([]byte, len(data)+1)
	buffer[0] = byte(len(data))
	copy(buffer[1:], data)

	//	buffer := make([]byte, numBytes+2)

	busData := i2c_smbus_ioctl_data{
		read_write: I2C_SMBUS_WRITE,
		command:    command,
		size:       I2C_SMBUS_I2C_BLOCK_DATA,
		data:       uintptr(unsafe.Pointer(&buffer[0])),
	}

	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(device.module.fd.Fd()), I2C_SMBUS, uintptr(unsafe.Pointer(&busData)))
	if err != 0 {
		return syscall.Errno(err)
	}

	return nil
}

func (device *DTI2CDevice) Read(command byte, numBytes int) ([]byte, error) {
	device.module.Lock()
	defer device.module.Unlock()

	device.sendSlaveAddress()

	buffer := make([]byte, numBytes+1)
	buffer[0] = byte(numBytes)

	//	buffer := make([]byte, numBytes+2)

	busData := i2c_smbus_ioctl_data{
		read_write: I2C_SMBUS_READ,
		command:    command,
		size:       I2C_SMBUS_I2C_BLOCK_DATA,
		data:       uintptr(unsafe.Pointer(&buffer[0])),
	}

	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(device.module.fd.Fd()), I2C_SMBUS, uintptr(unsafe.Pointer(&busData)))
	if err != 0 {
		return nil, syscall.Errno(err)
	}

	result := make([]byte, numBytes)
	copy(result, buffer[1:])

	return result, nil
}

// Read 1 byte from the bus
func (device *DTI2CDevice) ReadByte(command byte) (byte, error) {
	device.module.Lock()
	defer device.module.Unlock()

	e := device.sendSlaveAddress()
	if e != nil {
		return 0, e
	}

	data := uint8(0)

	busData := i2c_smbus_ioctl_data{
		read_write: I2C_SMBUS_READ,
		command:    command,
		size:       I2C_SMBUS_BYTE_DATA,
		data:       uintptr(unsafe.Pointer(&data)),
	}

	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(device.module.fd.Fd()), I2C_SMBUS, uintptr(unsafe.Pointer(&busData)))
	if err != 0 {
		return 0, syscall.Errno(err)
	}

	return data, nil
}

func (device *DTI2CDevice) WriteByte(command byte, value byte) error {
	device.module.Lock()
	defer device.module.Unlock()

	e := device.sendSlaveAddress()
	if e != nil {
		return e
	}

	busData := i2c_smbus_ioctl_data{
		read_write: I2C_SMBUS_WRITE,
		command:    command,
		size:       I2C_SMBUS_BYTE_DATA,
		data:       uintptr(unsafe.Pointer(&value)),
	}

	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(device.module.fd.Fd()), I2C_SMBUS, uintptr(unsafe.Pointer(&busData)))
	if err != 0 {
		return syscall.Errno(err)
	}

	return nil
}

func (device *DTI2CDevice) sendSlaveAddress() error {
	_, _, enum := syscall.Syscall(syscall.SYS_IOCTL, uintptr(device.module.fd.Fd()), I2C_SLAVE, uintptr(device.address))
	if enum != 0 {
		return fmt.Errorf("Could not open I2C bus on module %s", device.module.GetName())
	}
	return nil
}
