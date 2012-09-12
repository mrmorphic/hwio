package hwio

// A driver for Raspberry Pi
//
// Things known to work (tested on hardware):
// - digital write on all GPIO pins
// - digital read on all GPIO pins, for modes INPUT, INPUT_PULLUP and INPUT_PULLDOWN
//
// Known issues:
// - digital read with pulldown on GPIO0 and GPIO1 work, but unconnected always behave as being
//   pulled-up. i.e. pull-down setting appears to be ignored.
// - must be run as root, so it has permissions to create mmap
//
// WARNINGS:
// - THIS IS STILL UNDER DEVELOPMENT
// - UNTESTED FEATURES MAY FRY YOUR BOARD
// - ANY CHANGES YOU MAKE TO THIS MAY FRY YOUR BOARD
// Don't say you weren't warned.
// Developed and tested against Occidental0.2, adafruit's distribution.
//
// References:
// - http://elinux.org/RPi_Low-level_peripherals
// - https://projects.drogon.net/raspberry-pi/wiringpi/
// - BCM2835 technical reference
//
// @todo Implement PWM

import (
	"os"
	"syscall"
	"errors"
	"unsafe"
)

// Represents info we need to know about a pin on the Pi.
type RaspberryPiPin struct {
	hwPin     string // This intended for the P8.16 format name (currently unused)
	profile   []Capability
	gpioName  string // This is used for a human readable name
	funcReg   uint
    funcShift uint
	bit       uint   // A single bit in the position of the I/O value on the port
}

func (p RaspberryPiPin) GetName() string {
	return p.gpioName
}

func makePiPin(hwPin string, profile []Capability, gpioName string, funcReg uint, funcShift uint, bit uint) (* RaspberryPiPin) {
	return &RaspberryPiPin{hwPin, profile, gpioName, funcReg, funcShift, bit}
}

var piPins []*RaspberryPiPin
var piGpioProfile []Capability
var piUnusedProfile []Capability

const (
	PI_BCM2708_PERI_BASE = 0x20000000
	PI_GPIO_BASE = PI_BCM2708_PERI_BASE + 0x200000
	// CLOCK_BASE = BCM2708_PERI_BASE + 0x101000
	// GPIO_PWM = BCM2708_PERI_BASE + 0x20C000

	PI_PAGE_SIZE = 4*1024
	PI_BLOCK_SIZE = 4*1024

	// number of 32-bit values in the gpio mmap
	PI_GPIO_MMAP_N_UINT32 = 1024

	PI_GPIO_PORT0_SET_REG = 7
	PI_GPIO_PORT0_CLEAR_REG = 10

	PI_GPIO_PORT0_INPUT_LEVEL = 13

	PI_FUNC_REG_0 = 0
	PI_FUNC_REG_1 = 1
	PI_FUNC_REG_2 = 2

	// registers for pull up/down
	PI_GPPUD = 37
	PI_GPPUDCLK0 = 38

	// values for pull up/down
	PI_PUD_DISABLE = 0
	PI_PUD_PULLDOWN_ENABLE = 1
	PI_PUD_PULLUP_ENABLE = 2
)

func init() {
	piGpioProfile = []Capability{
		CAP_OUTPUT,
		CAP_INPUT,
		CAP_INPUT_PULLUP,
		CAP_INPUT_PULLDOWN,
	}
	piUnusedProfile = []Capability {
	}

	// The pins are numbered as they are on the connector. This means introducing
	// artificial pins for things like power, to keep the numbering.
	p := []*RaspberryPiPin{
		makePiPin("NULL", piUnusedProfile, "", 0, 0, 0), // 0 - spacer
		makePiPin("3.3V", piUnusedProfile, "", 0, 0, 0),
		makePiPin("5V", piUnusedProfile, "", 0, 0, 0),
		makePiPin("SDA", piGpioProfile, "GPIO0", PI_FUNC_REG_0, 0, 1 << 0), //also gpio
		makePiPin("DONOTCONNECT1", piUnusedProfile, "", 0, 0, 0),
		makePiPin("SCL", piGpioProfile, "GPIO1", PI_FUNC_REG_0, 3, 1 << 1), // also gpio
		makePiPin("GROUND", piUnusedProfile, "", 0, 0, 0),
		makePiPin("GPIO4", piGpioProfile, "GPIO4", PI_FUNC_REG_0, 12, 1 << 4),
		makePiPin("TXD", piGpioProfile, "GPIO14", PI_FUNC_REG_1, 12, 1 << 14),
		makePiPin("DONOTCONNECT2", piUnusedProfile, "", 0, 0, 0),
		makePiPin("RXD", piGpioProfile, "GPIO15", PI_FUNC_REG_1, 15, 1 << 15),
		makePiPin("GPIO17", piGpioProfile, "GPIO17", PI_FUNC_REG_1, 21, 1 << 17),
		makePiPin("GPIO18", piGpioProfile, "GPIO18", PI_FUNC_REG_1, 24, 1 << 18), // also supports PWM
		makePiPin("GPIO21", piGpioProfile, "GPIO21", PI_FUNC_REG_2, 3, 1 << 21),
		makePiPin("DONOTCONNECT3", piUnusedProfile, "", 0, 0, 0),
		makePiPin("GPIO22", piGpioProfile, "GPIO22", PI_FUNC_REG_2, 6, 1 << 22),
		makePiPin("GPIO23", piGpioProfile, "GPIO23", PI_FUNC_REG_2, 9, 1 << 23),
		makePiPin("DONOTCONNECT4", piUnusedProfile, "", 0, 0, 0),
		makePiPin("GPIO24", piGpioProfile, "GPIO24", PI_FUNC_REG_2, 12, 1 << 24),
		makePiPin("MOSI", piGpioProfile, "GPIO10", PI_FUNC_REG_1, 0, 1 << 10),
		makePiPin("DONOTCONNECT5", piUnusedProfile, "", 0, 0, 0),
		makePiPin("MISO", piGpioProfile, "GPIO9", PI_FUNC_REG_0, 27, 1 << 9),
		makePiPin("GPIO25", piGpioProfile, "GPIO25", PI_FUNC_REG_2, 15, 1 << 25),
		makePiPin("SCLK", piGpioProfile, "GPIO11", PI_FUNC_REG_1, 3, 1 << 11),
		makePiPin("CE0N", piGpioProfile, "GPIO8", PI_FUNC_REG_0, 24, 1 << 8),
		makePiPin("DONOTCONNECT6", piUnusedProfile, "", 0, 0, 0),
		makePiPin("CE1N", piGpioProfile, "GPIO7", PI_FUNC_REG_0, 21, 1 << 7),
	}
	piPins = p
}

type RaspberryPiDriver struct {
	// Mapped memory for directly accessing hardware registers
	gpioMmap []byte

	gpioMem *[PI_GPIO_MMAP_N_UINT32]uint
}

func (d *RaspberryPiDriver) Init() error {
	// Set up the memory mapped file giving us access to hardware registers
	file, e := os.OpenFile("/dev/mem", os.O_RDWR|os.O_APPEND, 0)
	if e != nil {
		return e
	}
	mmap, e := syscall.Mmap(int(file.Fd()), PI_GPIO_BASE, PI_BLOCK_SIZE, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if e != nil {
		return e
	}
	d.gpioMmap = mmap
	d.gpioMem = (*[PI_GPIO_MMAP_N_UINT32] uint) (unsafe.Pointer(&mmap[0]))

	return nil

/*
int wiringPiSetup (void)
{
  int      fd ;
  uint8_t *gpioMem, *pwmMem, *clkMem ;
  struct timeval tv ;

// Open the master /dev/memory device

  if ((fd = open ("/dev/mem", O_RDWR | O_SYNC) ) < 0)
  {
    fprintf (stderr, "wiringPiSetup: Unable to open /dev/mem: %s\n", strerror (errno)) ;
    return -1 ;
  }

// PWM

  if ((pwmMem = malloc (BLOCK_SIZE + (PAGE_SIZE-1))) == NULL)
  {
    fprintf (stderr, "wiringPiSetup: pwmMem malloc failed: %s\n", strerror (errno)) ;
    return -1 ;
  }

  if (((uint32_t)pwmMem % PAGE_SIZE) != 0)
    pwmMem += PAGE_SIZE - ((uint32_t)pwmMem % PAGE_SIZE) ;

  pwm = (uint32_t *)mmap(pwmMem, BLOCK_SIZE, PROT_READ|PROT_WRITE, MAP_SHARED|MAP_FIXED, fd, GPIO_PWM) ;

  if ((int32_t)pwm < 0)
  {
    fprintf (stderr, "wiringPiSetup: mmap failed (pwm): %s\n", strerror (errno)) ;
    return -1 ;
  }
 
// Clock control (needed for PWM)

  if ((clkMem = malloc (BLOCK_SIZE + (PAGE_SIZE-1))) == NULL)
  {
    fprintf (stderr, "wiringPiSetup: clkMem malloc failed: %s\n", strerror (errno)) ;
    return -1 ;
  }

  if (((uint32_t)clkMem % PAGE_SIZE) != 0)
    clkMem += PAGE_SIZE - ((uint32_t)clkMem % PAGE_SIZE) ;

  clk = (uint32_t *)mmap(clkMem, BLOCK_SIZE, PROT_READ|PROT_WRITE, MAP_SHARED|MAP_FIXED, fd, CLOCK_BASE) ;

  if ((int32_t)clk < 0)
  {
    fprintf (stderr, "wiringPiSetup: mmap failed (clk): %s\n", strerror (errno)) ;
    return -1 ;
  }
 
}
*/
}

func (d *RaspberryPiDriver) Close() {
	syscall.Munmap(d.gpioMmap)
}

func (d *RaspberryPiDriver) PinMode(pin Pin, mode PinIOMode) error {
	p := piPins[pin]

	if mode == OUTPUT {
		d.gpioMem[p.funcReg] = (d.gpioMem[p.funcReg] & ^(7 << p.funcShift)) | 1<<p.funcShift
	} else {
		// set pin as input
		d.gpioMem[p.funcReg] = (d.gpioMem[p.funcReg] & ^(7 << p.funcShift))

		// set pull up/down as appropriate. Write the mode to PI_PUD register, and clock it in
		// with PI_PUD_CLK_REG
		pull := PI_PUD_DISABLE
		if mode == INPUT_PULLUP {
			pull = PI_PUD_PULLUP_ENABLE
		} else if mode == INPUT_PULLDOWN {
			pull = PI_PUD_PULLDOWN_ENABLE
		}

		d.gpioMem[PI_GPPUD] = uint(pull)
		DelayMicroseconds(25)

		d.gpioMem[PI_GPPUDCLK0] = p.bit
		DelayMicroseconds(25)

		d.gpioMem[PI_GPPUD] = 0
		d.gpioMem[PI_GPPUDCLK0] = 0
	}
	return nil

/*
void pinModeGpio (int pin, int mode)
{
  static int pwmRunning  = FALSE ;
  int fSel, shift, alt ;

  pin &= 63 ;

  fSel    = gpioToGPFSEL [pin] ;
  shift   = gpioToShift  [pin] ;

  if (mode == INPUT)
    *(gpio + fSel) = (*(gpio + fSel) & ~(7 << shift)) ; // Sets bits to zero = input
  else if (mode == OUTPUT)
    *(gpio + fSel) = (*(gpio + fSel) & ~(7 << shift)) | (1 << shift) ;
  else if (mode == PWM_OUTPUT)
  {
    if ((alt = gpioToPwmALT [pin]) == 0)	// Not a PWM pin
      return ;

// Set pin to PWM mode

    *(gpio + fSel) = (*(gpio + fSel) & ~(7 << shift)) | (alt << shift) ;

// We didn't initialise the PWM hardware at setup time - because it's possible that
//	something else is using the PWM - e.g. the Audio systems! So if we use PWM
//	here, then we're assuming that nothing else is, otherwise things are going
//	to sound a bit funny...

    if (!pwmRunning)
    {

//	Gert/Doms Values
      *(clk + PWMCLK_DIV)  = 0x5A000000 | (32<<12) ;	// set pwm div to 32 (19.2/3 = 600KHz)
      *(clk + PWMCLK_CNTL) = 0x5A000011 ;		// Source=osc and enable
      digitalWrite (pin, LOW) ;
      *(pwm + PWM_CONTROL) = 0 ;			// Disable PWM
      delayMicroseconds (10) ;
      *(pwm + PWM0_RANGE) = 0x400 ;
      delayMicroseconds (10) ;
      *(pwm + PWM1_RANGE) = 0x400 ;
      delayMicroseconds (10) ;

// Enable PWMs

      *(pwm + PWM0_DATA) = 512 ;
      *(pwm + PWM1_DATA) = 512 ;

      *(pwm + PWM_CONTROL) = PWM0_ENABLE | PWM1_ENABLE ;
    }

  }

*/
}

func (d *RaspberryPiDriver) DigitalWrite(pin Pin, value int) (e error) {
	p := piPins[pin]
	if value == 0 {
		d.gpioMem[PI_GPIO_PORT0_CLEAR_REG] = p.bit
	} else {
		d.gpioMem[PI_GPIO_PORT0_SET_REG] = p.bit
	}
	return nil
}

func (d *RaspberryPiDriver) DigitalRead(pin Pin) (value int, e error) {
	p := piPins[pin]
	reg := d.gpioMem[PI_GPIO_PORT0_INPUT_LEVEL]
	if (reg & p.bit) != 0 {
		return HIGH, nil
	}
	return LOW, nil
}

func (d *RaspberryPiDriver) AnalogWrite(pin Pin, value int) (e error) {
	return nil
/*void pwmWriteWPi (int pin, int value)
{
  int port, gpioPin ;

  gpioPin = pinToGpio [pin & 63] ;
  port    = gpioToPwmPort [gpioPin] ;

  *(pwm + port) = value & 0x3FF ;
}*/
}


func (d *RaspberryPiDriver) AnalogRead(pin Pin) (value int, e error) {
	return 0, errors.New("Analog input is not supported")
}

func (d *RaspberryPiDriver) PinMap() (pinMap HardwarePinMap) {
	pinMap = make(HardwarePinMap)

	for i, hw := range piPins {
		names := []string{hw.hwPin}
		if hw.hwPin != hw.gpioName {
			names = append(names, hw.gpioName)
		}
		pinMap.add(Pin(i), names, hw.profile)
	}

	return
}

/*****

// Interrupts

extern int  (*waitForInterrupt) (int pin, int mS) ;

// Schedulling priority

extern int piHiPri (int pri) ;

// PWM

#define	PWM_CONTROL 0
#define	PWM_STATUS  1
#define	PWM0_RANGE  4
#define	PWM0_DATA   5
#define	PWM1_RANGE  8
#define	PWM1_DATA   9

#define	PWMCLK_CNTL	40
#define	PWMCLK_DIV	41

#define	PWM1_MS_MODE    0x8000  // Run in MS mode
#define	PWM1_USEFIFO    0x2000  // Data from FIFO
#define	PWM1_REVPOLAR   0x1000  // Reverse polarity
#define	PWM1_OFFSTATE   0x0800  // Ouput Off state
#define	PWM1_REPEATFF   0x0400  // Repeat last value if FIFO empty
#define	PWM1_SERIAL     0x0200  // Run in serial mode
#define	PWM1_ENABLE     0x0100  // Channel Enable

#define	PWM0_MS_MODE    0x0080  // Run in MS mode
#define	PWM0_USEFIFO    0x0020  // Data from FIFO
#define	PWM0_REVPOLAR   0x0010  // Reverse polarity
#define	PWM0_OFFSTATE   0x0008  // Ouput Off state
#define	PWM0_REPEATFF   0x0004  // Repeat last value if FIFO empty
#define	PWM0_SERIAL     0x0002  // Run in serial mode
#define	PWM0_ENABLE     0x0001  // Channel Enable


// Locals to hold pointers to the hardware

static volatile uint32_t *gpio ;
static volatile uint32_t *pwm ;
static volatile uint32_t *clk ;

// gpioToPwmALT
//	the ALT value to put a GPIO pin into PWM mode

static uint8_t gpioToPwmALT [] =
{
          0,         0,         0,         0,         0,         0,         0,         0,	//  0 ->  7
          0,         0,         0,         0, FSEL_ALT0, FSEL_ALT0,         0,         0, 	//  8 -> 15
          0,         0, FSEL_ALT5, FSEL_ALT5,         0,         0,         0,         0, 	// 16 -> 23
          0,         0,         0,         0,         0,         0,         0,         0,	// 24 -> 31
          0,         0,         0,         0,         0,         0,         0,         0,	// 32 -> 39
  FSEL_ALT0, FSEL_ALT0,         0,         0,         0, FSEL_ALT0,         0,         0,	// 40 -> 47
          0,         0,         0,         0,         0,         0,         0,         0,	// 48 -> 55
          0,         0,         0,         0,         0,         0,         0,         0,	// 56 -> 63
} ;

static uint8_t gpioToPwmPort [] =
{
          0,         0,         0,         0,         0,         0,         0,         0,	//  0 ->  7
          0,         0,         0,         0, PWM0_DATA, PWM1_DATA,         0,         0, 	//  8 -> 15
          0,         0, PWM0_DATA, PWM1_DATA,         0,         0,         0,         0, 	// 16 -> 23
          0,         0,         0,         0,         0,         0,         0,         0,	// 24 -> 31
          0,         0,         0,         0,         0,         0,         0,         0,	// 32 -> 39
  PWM0_DATA, PWM1_DATA,         0,         0,         0, PWM1_DATA,         0,         0,	// 40 -> 47
          0,         0,         0,         0,         0,         0,         0,         0,	// 48 -> 55
          0,         0,         0,         0,         0,         0,         0,         0,	// 56 -> 63

} ;


int waitForInterruptWPi (int pin, int mS)
{
  return waitForInterruptSys (pinToGpio [pin & 63], mS) ;
}

*****/

