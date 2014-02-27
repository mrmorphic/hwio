// An implementation of the HD44780 display communicating through an I2C expander. This is based on the I2C display library for
// Arduino: http://hmario.home.xs4all.nl/arduino/LiquidCrystal_I2C/
//
// The hardware takes the form of an HD44780 that is connected to I2C by a port expander such as an PCF8574, an 8-bit port expander.
// Typically it is connected to the display unit using 4 bits for data, and the other bits for RS, RW, EN and backlight control.

// This has been tested against a mjkdz brand adaptor, and works correctly. In other bits of Arduino code the En, Rw and Rs bits were assigned to
// the port expander differently. The display I tested was a 20x4 unit. Note characters output to the display are not necessarily displayed
// adjacently; you need to understand how the DRAM on the display maps to characters on the display.

package hd44780

import (
	"github.com/mrmorphic/hwio"
)

const (
	// commands
	LCD_CLEARDISPLAY   byte = 0x01
	LCD_RETURNHOME     byte = 0x02
	LCD_ENTRYMODESET   byte = 0x04
	LCD_DISPLAYCONTROL byte = 0x08
	LCD_CURSORSHIFT    byte = 0x10
	LCD_FUNCTIONSET    byte = 0x20
	LCD_SETCGRAMADDR   byte = 0x40
	LCD_SETDDRAMADDR   byte = 0x80

	// flags for display entry mode
	LCD_ENTRYRIGHT          byte = 0x00
	LCD_ENTRYLEFT           byte = 0x02
	LCD_ENTRYSHIFTINCREMENT byte = 0x01
	LCD_ENTRYSHIFTDECREMENT byte = 0x00

	// flags for display on/off control
	LCD_DISPLAYON  byte = 0x04
	LCD_DISPLAYOFF byte = 0x00
	LCD_CURSORON   byte = 0x02
	LCD_CURSOROFF  byte = 0x00
	LCD_BLINKON    byte = 0x01
	LCD_BLINKOFF   byte = 0x00

	// flags for display/cursor shift
	LCD_DISPLAYMOVE byte = 0x08
	LCD_CURSORMOVE  byte = 0x00
	LCD_MOVERIGHT   byte = 0x04
	LCD_MOVELEFT    byte = 0x00

	// flags for function set
	LCD_8BITMODE byte = 0x10
	LCD_4BITMODE byte = 0x00
	LCD_2LINE    byte = 0x08
	LCD_1LINE    byte = 0x00
	LCD_5x10DOTS byte = 0x04
	LCD_5x8DOTS  byte = 0x00

	// flags for backlight control
	//	LCD_BACKLIGHT   byte = 0x00
	//	LCD_NOBACKLIGHT byte = 0x80

	// En byte = 0x10 // B00010000 // Enable bit
	// Rw byte = 0x20 // B00100000 // Read/Write bit
	// Rs byte = 0x40 // B01000000 // Register select bit
	//	// En byte = 0x40 // B00010000 // Enable bit
	//	// Rw byte = 0x20 // B00100000 // Read/Write bit
	//	// Rs byte = 0x10 // B01000000 // Register select bit

	// constants for backlight polarity
	POSITIVE = 0
	NEGATIVE = 1
)

type HD44780 struct {
	device          hwio.I2CDevice
	displayFunction byte
	displayControl  byte
	displayMode     byte
	numLines        int
	backlight       byte

	// the bit masks of the LCD pins on the port extender.
	d7 byte
	d6 byte
	d5 byte
	d4 byte
	bl byte
	en byte
	rs byte
	rw byte

	blPolarity int
}

type I2CExpanderProfile int

const (
	// Profile constants for pre-defined device profiles. See http://forum.arduino.cc/index.php?topic=158312.15

	// mjkdz devices are commonly found in the wild
	PROFILE_MJKDZ I2CExpanderProfile = iota

	// devices based on PCF8574 are also around, but wired a little bit differently.
	PROFILE_PCF8574
)

func NewHD44780(module hwio.I2CModule, address int, profile I2CExpanderProfile) *HD44780 {
	switch profile {
	case PROFILE_MJKDZ:
		return NewHD44780Extended(module, address, 4, 5, 6, 0, 1, 2, 3, 7, NEGATIVE)
	case PROFILE_PCF8574:
		return NewHD44780Extended(module, address, 2, 1, 0, 4, 5, 6, 7, 3, POSITIVE)
	}

	return nil
}

func NewHD44780Extended(module hwio.I2CModule, address int, en int, rw int, rs int, d4 int, d5 int, d6 int, d7 int, bl int, polarity int) *HD44780 {
	device := module.GetDevice(address)
	result := &HD44780{
		device:     device,
		d7:         1 << uint16(d7),
		d6:         1 << uint16(d6),
		d5:         1 << uint16(d5),
		d4:         1 << uint16(d4),
		bl:         1 << uint16(bl),
		en:         1 << uint16(en),
		rs:         1 << uint16(rs),
		rw:         1 << uint16(rw),
		blPolarity: polarity}

	return result
}

func (display *HD44780) Init(cols int, lines int) {
	display.displayFunction = LCD_4BITMODE | LCD_1LINE | LCD_5x8DOTS

	if lines > 1 {
		display.displayFunction |= LCD_2LINE
	}
	display.numLines = lines

	// for some 1 line displays you can select a 10 pixel high font
	// if (dotsize != 0) && (lines == 1) {
	// 	_displayfunction |= LCD_5x10DOTS
	// }

	// SEE PAGE 45/46 FOR INITIALIZATION SPECIFICATION!
	// according to datasheet, we need at least 40ms after power rises above 2.7V
	// before sending commands. Arduino can turn on way befer 4.5V so we'll wait 50
	hwio.DelayMicroseconds(50000)

	// Now we pull both RS and R/W low to begin commands
	display.backlight = display.bl
	display.expanderWrite(display.backlight) // reset expander and turn backlight off (Bit 8 =1)
	hwio.Delay(1000)

	//put the LCD into 4 bit mode
	// this is according to the hitachi HD44780 datasheet
	// figure 24, pg 46

	// we start in 8bit mode, try to set 4 bit mode
	display.write4bits(0x03, 0)
	hwio.DelayMicroseconds(4500) // wait min 4.1ms

	// // second try
	display.write4bits(0x03, 0)
	hwio.DelayMicroseconds(4500) // wait min 4.1ms

	// // third go!
	display.write4bits(0x03, 0)
	hwio.DelayMicroseconds(150)

	// // finally, set to 4-bit interface
	display.write4bits(0x02, 0)

	// set # lines, font size, etc.
	display.Command(LCD_FUNCTIONSET | display.displayFunction)

	// turn the display on with no cursor or blinking default
	display.displayControl = LCD_DISPLAYON | LCD_CURSOROFF | LCD_BLINKOFF
	display.Display()

	// clear it off
	display.Clear()

	// Initialize to default text direction (for roman languages)
	display.displayMode = LCD_ENTRYLEFT | LCD_ENTRYSHIFTDECREMENT

	// set the entry mode
	display.Command(LCD_ENTRYMODESET | display.displayMode)

	display.Home()
}

func (display *HD44780) Clear() {
	display.Command(LCD_CLEARDISPLAY) // clear display, set cursor position to zero
	hwio.DelayMicroseconds(2000)      // this command takes a long time!
}

func (display *HD44780) Home() {
	display.Command(LCD_RETURNHOME) // set cursor position to zero
	hwio.DelayMicroseconds(2000)    // this command takes a long time!
}

func (display *HD44780) SetCursor(col int, row int) {
	rowOffsets := []byte{0x00, 0x40, 0x14, 0x54}
	if row > display.numLines {
		row = display.numLines - 1 // we count rows starting w/0
	}
	display.Command(LCD_SETDDRAMADDR | (byte(col) + rowOffsets[row]))
}

// Turn the display on/off (quickly)
func (display *HD44780) NoDisplay() {
	display.displayControl &= ^LCD_DISPLAYON
	display.Command(LCD_DISPLAYCONTROL | display.displayControl)
}

func (display *HD44780) Display() {
	display.displayControl |= LCD_DISPLAYON
	display.Command(LCD_DISPLAYCONTROL | display.displayControl)
}

// Turns the underline cursor on/off
func (display *HD44780) NoCursor() {
	display.displayControl &= ^LCD_CURSORON
	display.Command(LCD_DISPLAYCONTROL | display.displayControl)
}
func (display *HD44780) Cursor() {
	display.displayControl |= LCD_CURSORON
	display.Command(LCD_DISPLAYCONTROL | display.displayControl)
}

// Turn on and off the blinking cursor
func (display *HD44780) NoBlink() {
	display.displayControl &= ^LCD_BLINKON
	display.Command(LCD_DISPLAYCONTROL | display.displayControl)
}
func (display *HD44780) Blink() {
	display.displayControl |= LCD_BLINKON
	display.Command(LCD_DISPLAYCONTROL | display.displayControl)
}

// These commands scroll the display without changing the RAM
func (display *HD44780) ScrollDisplayLeft() {
	display.Command(LCD_CURSORSHIFT | LCD_DISPLAYMOVE | LCD_MOVELEFT)
}

func (display *HD44780) ScrollDisplayRight() {
	display.Command(LCD_CURSORSHIFT | LCD_DISPLAYMOVE | LCD_MOVERIGHT)
}

// This is for text that flows Left to Right
func (display *HD44780) LeftToRight() {
	display.displayMode |= LCD_ENTRYLEFT
	display.Command(LCD_ENTRYMODESET | display.displayMode)
}

// This is for text that flows Right to Left
func (display *HD44780) RightToLeft() {
	display.displayMode &= ^LCD_ENTRYLEFT
	display.Command(LCD_ENTRYMODESET | display.displayMode)
}

// This will 'right justify' text from the cursor
func (display *HD44780) Autoscroll() {
	display.displayMode |= LCD_ENTRYSHIFTINCREMENT
	display.Command(LCD_ENTRYMODESET | display.displayMode)
}

// This will 'left justify' text from the cursor
func (display *HD44780) NoAutoscroll() {
	display.displayMode &= ^LCD_ENTRYSHIFTINCREMENT
	display.Command(LCD_ENTRYMODESET | display.displayMode)
}

// // Allows us to fill the first 8 CGRAM locations
// // with custom characters
// func (display *HD44780) createChar(location byte, charmap byte[]) {
// 	location &= 0x7; // we only have 8 locations 0-7
// 	display.Command(LCD_SETCGRAMADDR | (location << 3));
// 	for i := 0; i < 8; i++ {
// 		write(charmap[i]);
// 	}
// }

func (display *HD44780) SetBacklight(on bool) {
	if on {
		display.backlight = display.bl
	} else {
		display.backlight = 0
	}
	display.expanderWrite(0)
}

func (display *HD44780) Command(command byte) {
	display.send(command, 0)
}

func (display *HD44780) Data(data byte) {
	display.send(data, display.rs)
}

func (display *HD44780) send(data byte, mode byte) {
	highnib := data >> 4
	lownib := data & 0x0F
	display.write4bits(highnib, mode)
	display.write4bits(lownib, mode)
}

// write 4 bits to the port extender. The low 4 bits of data are mapped to the d7-d4 pins on the device,
// so you cannot OR other control bits to the data. Mode is provided for that.
func (display *HD44780) write4bits(data byte, mode byte) {
	// map the 4 low bits of data into d
	var d byte = 0
	if data&0x08 != 0 {
		d |= display.d7
	}
	if data&0x04 != 0 {
		d |= display.d6
	}
	if data&0x02 != 0 {
		d |= display.d5
	}
	if data&0x01 != 0 {
		d |= display.d4
	}
	display.expanderWrite(d | mode)
	display.pulseEnable(d | mode)
}

// Write a byte to the port expander. The bits are already assumed to be in the right positions for
// the device profile.
func (display *HD44780) expanderWrite(data byte) {
	display.device.WriteByte(data|display.backlight, 0)
}

func (display *HD44780) pulseEnable(data byte) {
	display.expanderWrite(data | display.en) // En high
	hwio.DelayMicroseconds(1)                // enable pulse must be >450ns

	display.expanderWrite(data & ^display.en) // En low
	hwio.DelayMicroseconds(50)                // commands need > 37us to settle
}

func (display *HD44780) Write(p []byte) (n int, err error) {
	for _, b := range p {
		display.Data(b)
	}
	return len(p), nil
}
