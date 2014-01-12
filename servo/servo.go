package servo

import (
	"github.com/mrmorphic/hwio"
)

const (
	// default servo period, in milliseconds
	DEFAULT_SERVO_PERIOD = 20

	// defaults for servo duty, in microseconds
	DEFAULT_DUTY_MIN = 1000
	DEFAULT_DUTY_MAX = 2000
)

type Servo struct {
	PWM     hwio.PWMModule
	Pin     hwio.Pin
	minDuty int // min duty in microseconds
	maxDuty int // max duty in microseconds
}

// Create a new servo and initialise it.
func New(pwm hwio.PWMModule, pin interface{}) (*Servo, error) {
	var p hwio.Pin
	var e error

	switch pt := pin.(type) {
	case hwio.Pin:
		p = pt
	case string:
		p, e = hwio.GetPin(pt)
		if e != nil {
			return nil, e
		}
	}

	result := &Servo{PWM: pwm, Pin: p}

	// enable the servo
	e = pwm.EnablePin(p, true)
	if e != nil {
		return nil, e
	}

	e = result.SetPeriod(DEFAULT_SERVO_PERIOD)
	if e != nil {
		return nil, e
	}

	result.SetRange(DEFAULT_DUTY_MIN, DEFAULT_DUTY_MAX)

	return result, nil
}

// helper function to set the period of each cycle. Servos generally want this to be fixed, typically at 20ms.
// This just sets the underling PWM period, so if you need less than 1 ms you can set that directly on the PWM.
func (servo *Servo) SetPeriod(milliseconds int) error {
	return servo.PWM.SetPeriod(servo.Pin, int64(milliseconds*1000000))
}

// Set the servo to the specified angle, typically 0-180. This sets the duty cycle proportionally between min and max,
// which are defaulted to 1000-2000 microseconds range.
func (servo *Servo) Write(angle int) {
	servo.WriteMicroseconds(hwio.Map(angle, 0, 180, servo.minDuty, servo.maxDuty))
}

// Like the Arduino Servo.writeMicroseconds function. This is really setting the PWM duty directly, so it is possible
// to write values too small or too large for the servo to track.
func (servo *Servo) WriteMicroseconds(ms int) {
	// just pass to the underlying PWM pin.
	servo.PWM.SetDuty(servo.Pin, int64(ms*1000))
}

// Set the minimum and maximum number of microseconds for the servo. Write maps 0-180 to these values.
func (servo *Servo) SetRange(min int, max int) {
	servo.minDuty = min
	servo.maxDuty = max
}
