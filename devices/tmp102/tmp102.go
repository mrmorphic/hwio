// Support for TMP-102 temperature sensor.

// Current status:
// - this driver is working as expected.
// - it is being included as an example of how I2C devices can be added to the hwio package, and hopefully over time this
//   driver will be exended and more devices will be supported.

package tmp102

import (
	"github.com/mrmorphic/hwio"
)

const (
	// This is the default address.
	DEVICE_ADDRESS = 0x48
)

type TMP102 struct {
	device hwio.I2CDevice
}

func NewTMP102(module hwio.I2CModule) *TMP102 {
	device := module.GetDevice(DEVICE_ADDRESS)
	result := &TMP102{device: device}

	return result
}

func (t *TMP102) GetTemp() (float32, error) {
	buffer, e := t.device.Read(0x00, 2)
	if e != nil {
		return 0, e
	}
	MSB := buffer[0]
	LSB := buffer[1]

	/* Convert 12bit int using two's compliment */
	/* Credit: http://bildr.org/2011/01/tmp102-arduino/ */
	temp := ((int(MSB) << 8) | int(LSB)) >> 4

	// divide by 16, since lowest 4 bits are fractional.
	return float32(temp) * 0.0625, nil
}
