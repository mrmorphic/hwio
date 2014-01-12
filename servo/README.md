The hwio/servo package contains definitions for driving servo motors using PWM pins. To initialise a servo, you need to first get the PWM
module that the servo is attached. Here is an example of usage:

	import (
		"github.com/mrmorphic/hwio"
		"github.com/mrmorphic/hwio/servo"
	)

	m, e := hwio.GetModule("pwm2")
	if e != nil {
		fmt.Printf("could not get pwm module: %s\n", e)
		return
	}

	pwm := m.(hwio.PWMModule)

	pwm.Enable()

	// create a servo with a named pin. The pin name is passed to GetPin. You can also pass a Pin directly.
	servo, e := servo.New(pwm, "P8.13")

	// Set the servo angle, between 0 and 180 degrees.
	servo.Write(45)

	// Set the duty cycle to a specific number of microseconds
	servo.WriteMicroseconds(1500)

The default values should work for regular servo motors. It assumes servos have a 0-180 degree range, corresponding to
1000-2000 microsecond duty. If your servo has different duty ranges, you can change them:

	// Set duty range of the servo to an 800-2500 microsecond range.
	servo.SetRange(800, 2500)

The PWM and Pin are public properties of the PWM pin, so you can manipulate that directly if required.

Write() and WriteMicroseconds() methods are asynchronous; they set the duty cycle but return immediately before the servo has
moved to that position. This may differ from Arduino implementations.