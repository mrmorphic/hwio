// Support for MCP-23017 I2C port expander.

// Currently only supports basic GPIO (input and output). It does not support interupt features.

package mcp23017

import (
	"fmt"
	"github.com/mrmorphic/hwio"
)

const (
	// This is the default address if pins A2, A1 and A0 are grounded. So device address is base + (A2, A1, A0)
	DEFAULT_BASE_ADDRESS = 0x20

	REG_IODIRA  = 0x00
	REG_IODIRB  = 0x01
	REG_IPOLA   = 0x02
	REG_IPOLB   = 0x03
	REG_GPINENA = 0x04
	REG_GPINENB = 0x05
	REG_DEFVALA = 0x06
	REG_DEFVALB = 0x07
	REG_INTCONA = 0x08
	REG_INTCONB = 0x09
	REG_IOCON   = 0x0a
	REG_GPPUA   = 0x0c
	REG_GPPUB   = 0x0d
	REG_INTFA   = 0x0e
	REG_INTFB   = 0x0f
	REG_INTCAPA = 0x10
	REG_INTCAPB = 0x11
	REG_GPIOA   = 0x12
	REG_GPIOB   = 0x13
	REG_OLATA   = 0x14
	REG_OLATB   = 0x15
)

type MCP23017 struct {
	device hwio.I2CDevice
}

// Create a new isntance, and set it to use Bank 0. The address can either be what is wired on
// (A2,A1,A0) of the physical device, in which case this is added to the base address for the device
// (0x20). Otherwise, you can use 0x20-0x27. Anything else will return an error.
func NewMCP23017(module hwio.I2CModule, address int) (*MCP23017, error) {
	if address < 8 {
		address += DEFAULT_BASE_ADDRESS
	}

	if address < 0x20 || address > 0x27 {
		return nil, fmt.Errorf("Device address %d is invalid for an MCP23017. It must be in the range 0x20-0x27", address)
	}

	device := module.GetDevice(address)
	result := &MCP23017{device: device}

	// set config reg, force BANK=0, SEQOP=0. Note that this only works if already in BANK0, which is default on power-up
	device.WriteByte(REG_IOCON, 0)

	return result, nil
}

// Set direction bits for port A. A 1 bit indicates corresponding pin will be an input,
// A 0 bit indicates it will be an output.
func (d *MCP23017) SetDirA(value byte) error {
	return d.device.WriteByte(REG_IODIRA, value)
}

// Set direction bits for port A. A 1 bit indicates corresponding pin will be an input,
// A 0 bit indicates it will be an output.
func (d *MCP23017) SetDirB(value byte) error {
	return d.device.WriteByte(REG_IODIRB, value)
}

// Read from port A
func (d *MCP23017) GetPortA() (byte, error) {
	return d.device.ReadByte(REG_GPIOA)
}

// Read from port B
func (d *MCP23017) GetPortB() (byte, error) {
	return d.device.ReadByte(REG_GPIOB)
}

// Write to port A
func (d *MCP23017) SetPortA(value byte) error {
	return d.device.WriteByte(REG_GPIOA, value)
}

// Write to port B
func (d *MCP23017) SetPortB(value byte) error {
	return d.device.WriteByte(REG_GPIOB, value)
}

// Set pull-up configuration for port A. If a bit is 1 and the corresponding pin is
// an input, a pull-up resistor of about 100K is enabled. A zero bit indicates no pull-up.
func (d *MCP23017) SetPullupA(value byte) error {
	return d.device.WriteByte(REG_GPPUA, value)
}

// Set pull-up configuration for port B. If a bit is 1 and the corresponding pin is
// an input, a pull-up resistor of about 100K is enabled. A zero bit indicates no pull-up.
func (d *MCP23017) SetPullupB(value byte) error {
	return d.device.WriteByte(REG_GPPUB, value)
}
