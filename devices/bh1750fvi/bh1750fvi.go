// Support for BH1750FVI light sensor.

package bh1750fvi

import (
	"github.com/mrmorphic/hwio"
)

type ReadMode int

type modeConfig struct {
	deviceMode   byte
	sampleTimeMs int
}

const (
	// This is the default address
	DEVICE_ADDRESS_ADDR_LOW  = 0x23 // A0 = L
	DEVICE_ADDRESS_ADDR_HIGH = 0x5c // A0 = H
)

const (
	// These are the modes supported by the device.
	CONTINUOUS_HIGH_RES ReadMode = iota
	CONTINUOUS_HIGH_RES_2
	CONTINUOUS_LOW_RES
	ONETIME_HIGH_RES
	ONETIME_HIGH_RES_2
	ONETIME_LOW_RES
)

var modes map[ReadMode]modeConfig

type BH1750FVI struct {
	device hwio.I2CDevice
}

func init() {
	// Set up modes
	modes = make(map[ReadMode]modeConfig)
	modes[CONTINUOUS_HIGH_RES] = modeConfig{0x10, 180}
	modes[CONTINUOUS_HIGH_RES_2] = modeConfig{0x11, 180}
	modes[CONTINUOUS_LOW_RES] = modeConfig{0x13, 24}
	modes[ONETIME_HIGH_RES] = modeConfig{0x20, 180}
	modes[ONETIME_HIGH_RES_2] = modeConfig{0x21, 180}
	modes[ONETIME_LOW_RES] = modeConfig{0x23, 24}
}

// Create a new device, with i2c address specified. This can be used to access the device
// on a non-standard address, since it has an address bit.
func NewBH1750FVIAddr(module hwio.I2CModule, address int) *BH1750FVI {
	device := module.GetDevice(address)
	result := &BH1750FVI{device: device}

	return result
}

// Create a new device with the default i2c address (A0 is low on the board)
func NewBH1750FVI(module hwio.I2CModule) *BH1750FVI {
	return NewBH1750FVIAddr(module, DEVICE_ADDRESS_ADDR_LOW)
}

// Read the light level in low resolution mode, which is to a 4 lux precision.
func (t *BH1750FVI) ReadLightLevel(mode ReadMode) (float32, error) {
	// Get the settings
	m := modes[mode]

	// send a command to initiate low resolution read. The empty slice indicates there are no additional bytes,
	// just the command
	t.device.Write(m.deviceMode, []byte{})

	// wait for the sampling to be complete, max of 24ms
	hwio.Delay(m.sampleTimeMs)

	// read two bytes
	// @todo verify if this is correct. From Arduino examples I've seen, they use beginTransmission with the address,
	// @todo then requestFrom with the address. However the address 0x23 is also a device register. Need to check this.
	buffer, e := t.device.Read(m.deviceMode, 2)
	if e != nil {
		return 0, e
	}
	MSB := buffer[0]
	LSB := buffer[1]

	/* Convert 12bit int using two's compliment */
	/* Credit: http://bildr.org/2011/01/tmp102-arduino/ */
	level := ((int(MSB) << 8) | int(LSB))

	// divide by 16, since lowest 4 bits are fractional.
	return float32(level), nil
}
