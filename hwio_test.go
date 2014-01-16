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

	// test for altenate name of same pin
	p1a, e := GetPin("gpio1")
	if e != nil {
		t.Error(fmt.Sprintf("GetPin('gpio1') should not return an error, returned '%s'"), e)
	}
	if p1a != p1 {
		t.Error(fmt.Sprintf("Expected P1 and gpio1 to be the same pin"))
	}

	// test case insensitivity
	p1b, e := GetPin("GpIo1")
	if e != nil {
		t.Error(fmt.Sprintf("GetPin('GpIo1') should not return an error, returned '%s'"), e)
	}
	if p1b != p1 {
		t.Error(fmt.Sprintf("Expected P1 and GpIo1 to be the same pin"))
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

	pin2, _ := GetPin("p2")
	PinMode(pin2, OUTPUT)
	DigitalWrite(pin2, LOW)

	v := gpio.MockGetPinValue(pin2)
	if v != LOW {
		t.Error("After writing LOW to pin, driver should know this value")
	}

	DigitalWrite(pin2, HIGH)
	v = gpio.MockGetPinValue(pin2)
	if v != HIGH {
		t.Error("After writing HIGH to pin, driver should know this value")
	}
}

func TestDigitalRead(t *testing.T) {
	driver := new(TestDriver)
	SetDriver(driver)

	pin1, _ := GetPin("p1")

	PinMode(pin1, INPUT)
	writePinAndCheck(t, pin1, LOW, driver)
	writePinAndCheck(t, pin1, HIGH, driver)
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

func TestBitManipulation(t *testing.T) {
	v := UInt16FromUInt8(0x45, 0x65)
	if v != 0x4565 {
		t.Error(fmt.Sprintf("UInt16FromUInt8 does not work correctly, expected 0x4565, got %04x", v))
	}
}

func TestCpuInfo(t *testing.T) {
	s := CpuInfo(0, "processor")
	if s != "0" {
		t.Error(fmt.Sprintf("Expected 'processor' property of processor 0 to be 0 from CpuInfo, got '%s'", s))
	}
}

func TestAnalogRead(t *testing.T) {
	SetDriver(new(TestDriver))

	ap1, e := GetPin("p11")
	if e != nil {
		t.Error(fmt.Sprintf("GetPin('p11') should not return an error, returned '%s'", e))
	}

	v, e := AnalogRead(ap1)
	if e != nil {
		t.Error(fmt.Sprintf("After reading from pin %d, got an unexpected error: %s", ap1, e))
	}
	if v != 1 {
		t.Error(fmt.Sprintf("After reading from pin %d, did not get the expected value 1, got %d", ap1, v))
	}

	ap2, _ := GetPin("p12")
	v, _ = AnalogRead(ap2)
	if e != nil {
		t.Error(fmt.Sprintf("After reading from pin %d, got an unexpected error: %s", ap2, e))
	}
	if v != 1000 {
		t.Error(fmt.Sprintf("After reading from pin %d, did not get the expected value 1000, got %d", ap2, v))
	}
}

func TestNoErrorCheck(t *testing.T) {
	SetDriver(new(TestDriver))

	// @todo implement TestNoErrorCheck
}
