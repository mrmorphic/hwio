// Support for GY-520 gyroscope.

// Current status:
// - only supports small subset of what the device (MPU-6050) is capable of. In particular, only supports gyroscope,
//   accelerometer and temperature spot data, and has no support for FIFO, interupt or slaves.
// - it is being included as an example of how I2C devices can be added to the hwio package, and hopefully over time this
//   driver will be exended and more devices will be supported.

package gy520

import (
	"fmt"
	"hwio"
)

const (
	// This is the default address. Some devices may also respond to 0x69
	DEVICE_ADDRESS = 0x68

	REG_CONFIG       = 0x1a
	REG_GYRO_CONFIG  = 0x1b
	REG_ACCEL_CONFIG = 0x1c

	// accelerometer sensor registers, read-only
	REG_ACCEL_XOUT_H = 0x3b
	REG_ACCEL_XOUT_L = 0x3c
	REG_ACCEL_YOUT_H = 0x3d
	REG_ACCEL_YOUT_L = 0x3e
	REG_ACCEL_ZOUT_H = 0x3f
	REG_ACCEL_ZOUT_L = 0x40

	// temperature sensor registers, read-only
	REG_TEMP_OUT_H = 0x41
	REG_TEMP_OUT_L = 0x42

	// gyroscope sensor registers, read-only
	REG_GYRO_XOUT_H = 0x43
	REG_GYRO_XOUT_L = 0x44
	REG_GYRO_YOUT_H = 0x45
	REG_GYRO_YOUT_L = 0x46
	REG_GYRO_ZOUT_H = 0x47
	REG_GYRO_ZOUT_L = 0x48

	REG_PWR_MGMT_1 = 0x6b
)

type GY520 struct {
	device hwio.I2CDevice
}

func NewGY520(module hwio.I2CModule) *GY520 {
	device := module.GetDevice(DEVICE_ADDRESS)
	result := &GY520{device: device}

	return result
}

// Wake the device. By default on power on, the device is asleep.
func (g *GY520) Wake() error {
	// get sleep status
	// clear bit
	// write back
	return nil
}

// Put the device back to sleep.
func (g *GY520) Sleep() error {
	// get sleep status
	// clear bit
	// write back
	return nil
}

func (g *GY520) GetGyro() (gyroX int, gyroY int, gyroZ int, e error) {
	buffer, e := g.device.Read(REG_GYRO_XOUT_H, 6)
	if e != nil {
		return 0, 0, 0, e
	}

	//	fmt.Printf("length of received buffer is %d\n", len(buffer))
	//	fmt.Printf("bytes are %02x %02x %02x %02x %02x %02x\n", buffer[0], buffer[1], buffer[2], buffer[3], buffer[4], buffer[5])

	gyroX = int(int16(hwio.UInt16FromUInt8(buffer[0], buffer[1])))
	gyroY = int(int16(hwio.UInt16FromUInt8(buffer[2], buffer[3])))
	gyroZ = int(int16(hwio.UInt16FromUInt8(buffer[4], buffer[5])))

	return gyroX, gyroY, gyroZ, nil
}

func (g *GY520) GetAccel() (accelX int, accelY int, accelZ int, e error) {
	buffer, e := g.device.Read(REG_ACCEL_XOUT_H, 6)
	if e != nil {
		return 0, 0, 0, e
	}

	fmt.Printf("length of received buffer is %d\n", len(buffer))
	fmt.Printf("bytes are %02x %02x %02x %02x %02x %02x\n", buffer[0], buffer[1], buffer[2], buffer[3], buffer[4], buffer[5])

	accelX = int(int16(hwio.UInt16FromUInt8(buffer[0], buffer[1])))
	accelY = int(int16(hwio.UInt16FromUInt8(buffer[2], buffer[3])))
	accelZ = int(int16(hwio.UInt16FromUInt8(buffer[4], buffer[5])))

	return accelX, accelY, accelZ, nil
}

func (g *GY520) GetTemp() (int, error) {
	buffer, e := g.device.Read(REG_TEMP_OUT_H, 2)
	if e != nil {
		return 0, e
	}

	fmt.Printf("length of received buffer is %d\n", len(buffer))
	fmt.Printf("bytes are %02x %02x\n", buffer[0], buffer[1])

	return int(int16(hwio.UInt16FromUInt8(buffer[0], buffer[1]))), nil
}

func (g *GY520) SetAccelSampleRate(rate int) {

}

func (g *GY520) SetGyroSampleRate(rate int) {

}

func (g *GY520) SetTempSampleRate(rate int) {

}
