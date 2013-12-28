package hwio

// Unit tests for hwio. Each test sets a new TestDriver instance to start with
// same uninitialised state.

import (
	"fmt"
	"testing"
)

// Get the driver's pin map and check for the pins in it. Tests that the
// consumer can determine pin capabilities
func TestPinMap(t *testing.T) {
	SetDriver(new(TestDriver))

	m := GetDefinedPins()

	// P0 should exist
	p0 := m.GetPin(0)
	if p0 == nil {
		t.Error("Pin 0 is expected to be defined")
	}

	// P9 should not exist
	p99 := m.GetPin(99)
	if p99 != nil {
		t.Error("Pin 99 should not exist")
	}
}

func TestGetPin(t *testing.T) {
	SetDriver(new(TestDriver))

	p1, e := GetPin("P1")
	if e != nil {
		t.Error(fmt.Sprintf("GetPin('P1') should not return an error, returned '%s'", e))
	}
	if p1 != 0 {
		t.Error("GetPin('P0') should return 0")
	}

	_, e = GetPin("P99")
	if e == nil {
		t.Error("GetPin('P99') should have returned an error but didn't")
	}
}

func TestPinMode(t *testing.T) {
	SetDriver(new(TestDriver))

	gpio := getMockGPIO(t)

	// Set pin 0 to input. We expect no error as it's GPIO
	e := PinMode(0, INPUT)
	if e != nil {
		t.Error(fmt.Sprintf("GetPin('P1') should not return an error, returned '%s'", e))
	}
	m := gpio.MockGetPinMode(0)
	if m != INPUT {
		t.Error("Pin set to read mode is not set in the driver")
	}

	// Change pin 0 to output. We expect no error in this case either, and the
	// new pin mode takes effect
	e = PinMode(0, OUTPUT)
	m = gpio.MockGetPinMode(0)
	if m != OUTPUT {
		t.Error("Pin changed from read to write mode is not set in the driver")
	}
}

func TestDigitalWrite(t *testing.T) {
	SetDriver(new(TestDriver))

	gpio := getMockGPIO(t)

	PinMode(0, OUTPUT)
	DigitalWrite(0, LOW)

	v := gpio.MockGetPinValue(0)
	if v != LOW {
		t.Error("After writing LOW to pin, driver should know this value")
	}

	DigitalWrite(0, HIGH)
	v = gpio.MockGetPinValue(0)
	if v != HIGH {
		t.Error("After writing HIGH to pin, driver should know this value")
	}
}

func TestDigitalRead(t *testing.T) {
	driver := new(TestDriver)
	SetDriver(driver)

	PinMode(0, INPUT)
	writePinAndCheck(t, 0, LOW, driver)
	writePinAndCheck(t, 0, HIGH, driver)
}

func getMockGPIO(t *testing.T) *testGPIOModule {
	g, e := GetModule("gpio")
	if e != nil {
		t.Error(fmt.Sprintf("Fetching gpio module should not return an error, returned %s", e))
	}
	if g == nil {
		t.Error("Could not get 'gpio' module")
	}

	return g.(*testGPIOModule)
}

func writePinAndCheck(t *testing.T, pin Pin, value int, driver *TestDriver) {
	gpio := getMockGPIO(t)

	gpio.MockSetPinValue(pin, value)
	v, e := DigitalRead(pin)
	if e != nil {
		t.Error("DigitalRead returned an error")
	}
	if v != value {
		t.Error(fmt.Sprintf("After writing %d to driver, DigitalRead method should return this value", value))
	}
}

// func TestAnalogWrite(t *testing.T) {
// 	SetDriver(new(TestDriver))

// 	// @todo implement TestAnalogWrite
// }

// func TestAnalogRead(t *testing.T) {
// 	SetDriver(new(TestDriver))

// 	e := PinMode(6, INPUT_ANALOG)
// 	if e != nil {
// 		t.Error(fmt.Sprintf("PinMode(6) error: %s", e))
// 	} else {
// 		v, e := AnalogRead(6)
// 		if e != nil {
// 			t.Error(fmt.Sprintf("After reading from pin 6, got an unexpected error: %s", e))
// 		}
// 		if v != 1 {
// 			t.Error(fmt.Sprintf("After reading from pin 6, did not get the expected value, got %d", v))
// 		}
// 	}

// 	e = PinMode(7, INPUT_ANALOG)
// 	if e != nil {
// 		t.Error(fmt.Sprintf("PinMode(7) error: %s", e))
// 	} else {
// 		v, e := AnalogRead(7)
// 		if e != nil {
// 			t.Error(fmt.Sprintf("After reading from pin 7, got an unexpected error: %s", e))
// 		}
// 		if v != 1000 {
// 			t.Error(fmt.Sprintf("After reading from pin 7, did not get the expected value, got %d", v))
// 		}

// 	}
// }

func TestNoErrorCheck(t *testing.T) {
	SetDriver(new(TestDriver))

	// @todo implement TestNoErrorCheck
}
