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
	// @todo allocate the pins

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

func (device *DTI2CDevice) Write(command byte, buffer []byte, numBytes int) (e error) {
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
func (device *DTI2CDevice) ReadByte(command byte) byte {

	device.sendSlaveAddress()

	// i2c_smbus_access(file,I2C_SMBUS_READ,command,I2C_SMBUS_BYTE_DATA,&data)

	data := uint8(0)

	busData := i2c_smbus_ioctl_data{
		read_write: I2C_SMBUS_READ,
		command:    command,
		size:       I2C_SMBUS_BYTE_DATA,
		data:       uintptr(unsafe.Pointer(&data)),
	}

	// fmt.Println("About to Read8 ", module.fd.Fd(), " IOCTL ", I2C_SMBUS, " Command", command)

	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(device.module.fd.Fd()), I2C_SMBUS, uintptr(unsafe.Pointer(&busData)))
	if err != 0 {
		panic(syscall.Errno(err))
	}

	return data
}

func (device *DTI2CDevice) sendSlaveAddress() error {
	// fmt.Println("About to open Bus fd ", module.fd.Fd(), " IOCTL ", I2C_SLAVE, " Arg", address)
	_, _, enum := syscall.Syscall(syscall.SYS_IOCTL, uintptr(device.module.fd.Fd()), I2C_SLAVE, uintptr(device.address))
	if enum != 0 {
		return fmt.Errorf("Could not open I2C bus on module %s", device.module.GetName())
	}
	return nil
}

// // opening bus

// #include <errno.h>
// #include <string.h>
// #include <stdio.h>
// #include <stdlib.h>
// #include <unistd.h>
// #include <linux/i2c-dev.h>
// #include <sys/ioctl.h>
// #include <sys/types.h>
// #include <sys/stat.h>
// #include <fcntl.h>
// int file;
// char *filename = "/dev/i2c-1";
// if ((file = open(filename, O_RDWR)) < 0) {
//     /* ERROR HANDLING: you can check errno to see what went wrong */
//     perror("Failed to open the i2c bus");
//     exit(1);
// }

// initiating comms:
// int addr = 0x48;     // The I2C address of the device
// if (ioctl(file, I2C_SLAVE, addr) < 0) {
//     printf("Failed to acquire bus access and/or talk to slave.\n");
//     /* ERROR HANDLING; you can check errno to see what went wrong */
//     exit(1);
// }

// reading frm device:

// unsigned char buf[10] = {0};

// for (int i = 0; i<4; i++) {
//     // Using I2C Read
//     if (read(file,buf,2) != 2) {
//         /* ERROR HANDLING: i2c transaction failed */
//         printf("Failed to read from the i2c bus: %s.\n", strerror(errno));
//             printf("\n\n");
//     } else {
//         /* Device specific stuff here */
//     }
// }
