package hwio

// Unit tests for hwio. Each test sets a new TestDriver instance to start with
// same uninitialised state,

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

func TestPinMode(t *testing.T) {
	driver := new(TestDriver)
	SetDriver(driver)

	// Set pin 0 to input. We expect no error as it's GPIO
	e := PinMode(0, INPUT)
	m := driver.MockGetPinMode(0)
	if m != INPUT {
		t.Error("Pin set to read mode is not set in the driver")
	}

	// Change pin 0 to output. We expect no error in this case either, and the
	// new pin mode takes effect
	e = PinMode(0, OUTPUT)
	m = driver.MockGetPinMode(0)
	if m != OUTPUT {
		t.Error("Pin changed from read to write mode is not set in the driver")
	}

	// Set pin 1 to an output. We expect failure, as this is a readonly pin
	e = PinMode(1, OUTPUT)
	if e == nil {
		t.Error("Read only pin should not accept OUTPUT as a mode")
	}

	// Set pin 2 to input. We expect failure, as this is a write only pin
	e = PinMode(2, INPUT)
	if e == nil {
		t.Error("Write only pin should not accept INPUT as a mode")
	}
}

func TestDigitalWrite(t *testing.T) {
	driver := new(TestDriver)
	SetDriver(driver)

	PinMode(0, OUTPUT)
	DigitalWrite(0, LOW)

	v := driver.MockGetPinValue(0)
	if v != LOW {
		t.Error("After writing LOW to pin, driver should know this value")
	}

	DigitalWrite(0, HIGH)
	v = driver.MockGetPinValue(0)
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

func writePinAndCheck(t *testing.T, pin Pin, value int, driver *TestDriver) {
	driver.MockSetPinValue(pin, value)
	v, e := DigitalRead(pin)
	if e != nil {
		t.Error("DigitalRead returned an error")
	}
	if v != value {
		t.Error(fmt.Sprintf("After writing %d to driver, DigitalRead method should return this value", value))
	}
}

func TestAnalogWrite(t *testing.T) {
	SetDriver(new(TestDriver))

	// @todo implement TestAnalogWrite
}

func TestAnalogRead(t *testing.T) {
	SetDriver(new(TestDriver))

	// @todo implement TestAnalogRead
}

func TestNoErrorCheck(t *testing.T) {
	SetDriver(new(TestDriver))

	// @todo implement TestNoErrorCheck
}
