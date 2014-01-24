// Interface for Nintendo Wii nunchucks over I2C.
// With reference to WiiChuck for Arduino.

package nunchuck

import (
	"fmt"
	"github.com/mrmorphic/hwio"
	"math"
)

const (
	DEVICE_ADDRESS = 0x52

	DEFAULT_JOYSTICK_ZERO_X = 124
	DEFAULT_JOYSTICK_ZERO_Y = 132

	DEFAULT_ACCEL_ZEROX = 0.0
	DEFAULT_ACCEL_ZEROY = 0.0
	DEFAULT_ACCEL_ZEROZ = 0.0

	RADIUS = 210
)

type Nunchuck struct {
	device     hwio.I2CDevice
	zeroJoyX   int
	zeroJoyY   int
	zeroAccelX float32
	zeroAccelY float32
	zeroAccelZ float32

	lastJoyX     int
	lastJoyY     int
	lastAccelX   float32
	lastAccelY   float32
	lastAccelZ   float32
	lastZPressed bool
	lastCPressed bool
}

func NewNunchuck(module hwio.I2CModule) (*Nunchuck, error) {
	device := module.GetDevice(DEVICE_ADDRESS)
	n := &Nunchuck{device: device}

	n.SetJoystickZero(DEFAULT_JOYSTICK_ZERO_X, DEFAULT_JOYSTICK_ZERO_Y)
	n.SetAccelZero(DEFAULT_ACCEL_ZEROX, DEFAULT_ACCEL_ZEROY, DEFAULT_ACCEL_ZEROZ)

	// instead of the common 0x40 -> 0x00 initialization, we
	// use 0xF0 -> 0x55 followed by 0xFB -> 0x00.
	// this lets us use 3rd party nunchucks (like cheap $4 ebay ones)
	// while still letting us use official oness.
	// see http://www.arduino.cc/cgi-bin/yabb2/YaBB.pl?num=1264805255

	e := device.WriteByte(0xF0, 0x55) // first config register
	if e != nil {
		return nil, e
	}

	hwio.Delay(1)

	e = device.WriteByte(0xFB, 0x00) // second config register
	if e != nil {
		return nil, e
	}

	return n, nil
}

// Read all sensor values from the nunchuck and reads them into the internal state of the nunchuck instance.
// Use Get methods to retrieve sensor values since last call of ReadSensors.
func (n *Nunchuck) ReadSensors() error {
	// Get bytes from the sensor, packed into 6 bytes.
	bytes, e := n.device.Read(0, 6)
	if e != nil {
		return e
	}

	if len(bytes) < 6 {
		return fmt.Errorf("Error getting nunchuck data, expected 6 bytes but got %d", len(bytes))
	}

	// Split out the packet into the n.last* variables.

	// bytes[0] and bytes[1] are joystick X and Y respectively
	n.lastJoyX = int(bytes[0])
	n.lastJoyY = int(bytes[1])

	// bytes[2] - bytes[4] are accelX, accelY and accelZ most significant byte respectively. LSB are in bytes[5]
	ax := int(int8(bytes[2]<<2)) | int((bytes[5]>>2)&3)
	ay := int(int8(bytes[3]<<2)) | int((bytes[5]>>4)&3)
	az := int(int8(bytes[4]<<2)) | int((bytes[5]>>6)&3)

	n.lastAccelX = float32(ax) - n.zeroAccelX
	n.lastAccelY = float32(ay) - n.zeroAccelY
	n.lastAccelZ = float32(az) - n.zeroAccelZ

	n.lastZPressed = true
	if bytes[5]&1 > 0 {
		n.lastZPressed = false
	}

	n.lastCPressed = true
	if bytes[5]&2 > 0 {
		n.lastCPressed = false
	}

	return nil
}

// Calibrate the joystick to the most recently read values.
func (n *Nunchuck) CalibrateJoystick() {
	n.SetJoystickZero(n.lastJoyX, n.lastJoyY)
}

// Calibrate the joystick to explicit values.
func (n *Nunchuck) SetJoystickZero(x int, y int) {
	n.zeroJoyX = x
	n.zeroJoyY = y
}

func (n *Nunchuck) SetAccelZero(x float32, y float32, z float32) {
	n.zeroAccelX = x
	n.zeroAccelY = y
	n.zeroAccelZ = z
}

func (n *Nunchuck) GetJoystick() (x int, y int) {
	return n.lastJoyX, n.lastJoyY
}

func (n *Nunchuck) GetAccel() (ax float32, ay float32, az float32) {
	return n.lastAccelX, n.lastAccelY, n.lastAccelZ
}

func (n *Nunchuck) GetZPressed() bool {
	return n.lastZPressed
}

func (n *Nunchuck) GetCPressed() bool {
	return n.lastCPressed
}

// Read roll in degrees, computed from accelerometer
func (n *Nunchuck) GetRoll() float32 {
	return float32(math.Atan2(float64(n.lastAccelX), float64(n.lastAccelZ)) / math.Pi * 180.0)
}

// Read pitch in degrees, computed from accelerometer
func (n *Nunchuck) GetPitch() float32 {
	return float32(math.Acos(float64(n.lastAccelY)/RADIUS) / math.Pi * 180.0)
}
