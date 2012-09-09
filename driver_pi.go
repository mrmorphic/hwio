package hwio

// A driver for Raspberry Pi
//
// Things known to work (tested on hardware):
// - nothing yet
//
// WARNINGS:
// - THIS IS STILL UNDER DEVELOPMENT
// - UNTESTED FEATURES MAY FRY YOUR BOARD
// - ANY CHANGES YOU MAKE TO THIS MAY FRY YOUR BOARD
// Don't say you weren't warned.
// Developed and tested against Occidental0.2, adafruit's distribution.

// @todo Implement GPIO output
// @todo Implement GPIO input

import (
	"os"
	"strconv"
	"syscall"
	"errors"
//	"fmt"
//	"time"
)

// Represents info we need to know about a pin on the Pi.
// @todo Determine if 'hwPin' is required
type RaspberryPiPin struct {
	hwPin     string // This intended for the P8.16 format name (currently unused)
	profile   []Capability
	gpioName  string // This is used for a human readable name
	bit       uint   // A single bit in the position of the I/O value on the port
	mode0Name string // mode 0 signal name, used by the muxer
}

func (p RaspberryPiPin) GetName() string {
	return p.gpioName
}

var piPins []*RaspberryPiPin
var piGpioProfile []Capability
var piUnusedProfile []Capability

const
	PI_BCM2708_PERI_BASE = 0x20000000
	PI_GPIO_BASE		(PI_BCM2708_PERI_BASE + 0x200000)
	// CLOCK_BASE		(BCM2708_PERI_BASE + 0x101000)
	// GPIO_PWM		(BCM2708_PERI_BASE + 0x20C000)

	PI_PAGE_SIZE = (4*1024)
	PI_BLOCK_SIZE = (4*1024)


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
		&RaspberryPiPin{"NULL", piUnusedProfile, "", 0, ""}, // 0 - spacer
		&RaspberryPiPin{"3.3V", piUnusedProfile, "", 0, "gpmc_ad6"},
		&RaspberryPiPin{"5V", piUnusedProfile, "", 0, "gpmc_ad7"},
		&RaspberryPiPin{"SDA", piUnusedProfile, "GPIO0", 1 << 0, "gpmc_ad2"}, //also gpio
		&RaspberryPiPin{"DONOTCONNECT1", piUnusedProfile, "", 0, "gpmc_ad3"},
		&RaspberryPiPin{"SCL", piUnusedProfile, "GPIO1", 1 << 1, "gpmc_ad3"}, // also gpio
		&RaspberryPiPin{"GROUND", piUnusedProfile, "", 0, "gpmc_advn_ale"},
		&RaspberryPiPin{"GPIO4", piGpioProfile, "GPIO4", 1 << 4, "gpmc_oen_ren"},
		&RaspberryPiPin{"TXD", piUnusedProfile, "GPIO14", 1 << 14, "gpmc_ben0_cle"},
		&RaspberryPiPin{"DONOTCONNECT2", piUnusedProfile, "", 0, "gpmc_wen"},
		&RaspberryPiPin{"RXD", piUnusedProfile, "GPIO15", 1 << 15, "gpmc_ad13"},
		&RaspberryPiPin{"GPIO17", piGpioProfile, "GPIO17", 1 << 17, "gpmc_ad12"},
		&RaspberryPiPin{"GPIO18", piGpioProfile, "GPIO18", 1 << 18, "gpmc_ad9"}, // also supports PWM
		&RaspberryPiPin{"GPIO21", piGpioProfile, "GPIO21", 1 << 21, "gpmc_ad10"},
		&RaspberryPiPin{"DONOTCONNECT3", piUnusedProfile, "", 0, "gpmc_ad15"},
		&RaspberryPiPin{"GPIO22", piGpioProfile, "GPIO22", 1 << 22, "gpmc_ad14"},
		&RaspberryPiPin{"GPIO23", piGpioProfile, "GPIO23", 1 << 23, "gpmc_ad11"},
		&RaspberryPiPin{"DONOTCONNECT4", piUnusedProfile, "", 0, "gpmc_clk"},
		&RaspberryPiPin{"GPIO24", piGpioProfile, "GPIO24", 1 << 24, "gpmc_ad8"},
		&RaspberryPiPin{"MOSI", piUnusedProfile, "GPIO10", 1 << 10, "gpmc_csn2"},
		&RaspberryPiPin{"DONOTCONNECT5", piUnusedProfile, "", 0, "gpmc_csn1"},
		&RaspberryPiPin{"MISO", piUnusedProfile, "GPIO9", 1 << 9, "gpmc_ad5"},
		&RaspberryPiPin{"GPIO25", piGpioProfile, "GPIO25", 1 << 25, "gpmc_ad4"},
		&RaspberryPiPin{"SCLK", piUnusedProfile, "GPIO11", 1 << 11, "gpmc_ad1"},
		&RaspberryPiPin{"CE0N", piUnusedProfile, "GPIO8", 1 << 8, "gpmc_ad0"},
		&RaspberryPiPin{"DONOTCONNECT6", piUnusedProfile, "", 0, "gpmc_csn0"},
		&RaspberryPiPin{"CE1N", piUnusedProfile, "GPIO7", 1 << 7, "lcd_vsync"},
	}
	piPins = p
}

type RaspberryPiDriver struct {
	// Mapped memory for directly accessing hardware registers
	gpioMmap []byte
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

// GPIO:

// Allocate 2 pages - 1 ...

  if ((gpioMem = malloc (BLOCK_SIZE + (PAGE_SIZE-1))) == NULL)
  {
    fprintf (stderr, "wiringPiSetup: malloc failed: %s\n", strerror (errno)) ;
    return -1 ;
  }

// ... presumably to make sure we can round it up to a whole page size

  if (((uint32_t)gpioMem % PAGE_SIZE) != 0)
    gpioMem += PAGE_SIZE - ((uint32_t)gpioMem % PAGE_SIZE) ;

  gpio = (uint32_t *)mmap((caddr_t)gpioMem, BLOCK_SIZE, PROT_READ|PROT_WRITE, MAP_SHARED|MAP_FIXED, fd, GPIO_BASE) ;

  if ((int32_t)gpio < 0)
  {
    fprintf (stderr, "wiringPiSetup: mmap failed: %s\n", strerror (errno)) ;
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
		e := d.pinMux(p.mode0Name, CONF_GPIO_OUTPUT)
		if e != nil {
			return e
		}

		d.clearRegL(p.port+uint(GPIO_OE), p.bit)
	} else {
		pull := CONF_PULL_DISABLE
		// note: pull up/down modes assume that CONF_PULLDOWN resets the pull disable bit
		if mode == INPUT_PULLUP {
			pull = CONF_PULLUP
		} else if mode == INPUT_PULLDOWN {
			pull = CONF_PULLDOWN
		}

		e := d.pinMux(p.mode0Name, CONF_GPIO_INPUT|uint(pull))
		if e != nil {
			return e
		}

//		fmt.Printf("R/W dir reg BEFORE value is %x\n", d.getRegL(p.port+uint(GPIO_OE)))

		d.orRegL(p.port+uint(GPIO_OE), p.bit)
//		fmt.Printf("R/W dir reg AFTER value is %x\n", d.getRegL(p.port+uint(GPIO_OE)))
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

// When we change mode of any pin, we remove the pull up/downs

  pullUpDnControl (pin, PUD_OFF) ;
}

*/

/*
void pullUpDnControlWPi (int pin, int pud)
{
  pin = pinToGpio [pin & 63] ;

  *(gpio + 37) = pud ;
  delayMicroseconds (10) ;
  *(gpio + gpioToPUDCLK [pin]) = 1 << pin ;
  delayMicroseconds (10) ;
  
  *(gpio + 37) = 0 ;
  *(gpio + gpioToPUDCLK [pin]) = 0 ;
}
*/
}

func (d *RaspberryPiDriver) pinMux(mux string, mode uint) error {
	// Uses kernel omap_mux files to set pin modes.
	// There's no simple way to write the control module registers from a 
	// user-level process because it lacks the proper privileges, but it's 
	// easy enough to just use the built-in file-based system and let the 
	// kernel do the work. 
	f, e := os.OpenFile(PINMUX_PATH+mux, os.O_WRONLY|os.O_TRUNC, 0666)
	if e != nil {
		return e
	}

	s := strconv.FormatInt(int64(mode), 16)
//	fmt.Printf("Writing mode %s to mux file %s\n", s, PINMUX_PATH+mux)
	f.WriteString(s)
	return nil
}

func (d *RaspberryPiDriver) DigitalWrite(pin Pin, value int) (e error) {
	p := piPins[pin]
	if value == 0 {
		d.gpioMem[p.clrAddr] = p.bit
	} else {
		d.gpioMem[p.setAddr] = p.bit
	}
	return nil
}

func (d *RaspberryPiDriver) DigitalRead(pin Pin) (value int, e error) {
	p := piPins[pin]
	reg := d.getRegL(p.port+GPIO_DATAIN)
	//	fmt.Printf("\nraw in: %x (checking bit %d)\n", reg, p.bit)
	if (reg & p.bit) != 0 {
		return HIGH, nil
	}
	return LOW, nil
/*

int digitalReadWPi (int pin)
{
  int gpioPin ;

  pin &= 63 ;

  gpioPin = pinToGpio [pin] ;

  if ((*(gpio + gpioToGPLEV [gpioPin]) & (1 << gpioPin)) != 0)
    return HIGH ;
  else
    return LOW ;
}
*/
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

// Sets 32 bit Register at address to its current value AND mask.
func (d *RaspberryPiDriver) andRegL(address uint, mask uint) {
	d.setRegL(address, d.getRegL(address)&mask)
}

// Sets 32 bit Register at address to its current value OR mask.
func (d *RaspberryPiDriver) orRegL(address uint, mask uint) {
	d.setRegL(address, d.getRegL(address)|mask)
}

// Clears mask bits in 32 bit register at given address.
func (d *RaspberryPiDriver) clearRegL(address uint, mask uint) {
	d.andRegL(address, ^mask)
}

// Returns unpacked 32 bit register value starting from address. Integers
// are little endian on BeagleBone
func (d *RaspberryPiDriver) getRegL(address uint) (result uint) {
	result = uint(d.mmap[address])
	result |= uint(d.mmap[address+1])<<8
	result |= uint(d.mmap[address+2])<<16
	result |= uint(d.mmap[address+3])<<24
	return result
}

func (d *RaspberryPiDriver) setRegL(address uint, value uint) {
	d.mmap[address] = byte(value & 0xff)
	d.mmap[address+1] = byte((value >> 8) & 0xff)
	d.mmap[address+2] = byte((value >> 16) & 0xff)
	d.mmap[address+3] = byte((value >> 24) & 0xff)
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

#define	NUM_PINS	17

#define	WPI_MODE_PINS		 0
#define	WPI_MODE_GPIO		 1
#define	WPI_MODE_GPIO_SYS	 2
#define	WPI_MODE_PIFACE		 3

#define	INPUT		 0
#define	OUTPUT		 1
#define	PWM_OUTPUT	 2

#define	LOW		 0
#define	HIGH		 1

#define	PUD_OFF		 0
#define	PUD_DOWN	 1
#define	PUD_UP		 2

extern void (*pinMode)          (int pin, int mode) ;
extern void (*pullUpDnControl)  (int pin, int pud) ;
extern void (*digitalWrite)     (int pin, int value) ;
extern void (*pwmWrite)         (int pin, int value) ;
extern int  (*digitalRead)      (int pin) ;

// Interrupts

extern int  (*waitForInterrupt) (int pin, int mS) ;

// Schedulling priority

extern int piHiPri (int pri) ;


void (*pinMode)          (int pin, int mode) ;
void (*pullUpDnControl)  (int pin, int pud) ;
void (*digitalWrite)     (int pin, int value) ;
void (*pwmWrite)         (int pin, int value) ;
int  (*digitalRead)      (int pin) ;
int  (*waitForInterrupt) (int pin, int mS) ;

// Port function select bits

#define	FSEL_INPT		0b000
#define	FSEL_OUTP		0b001
#define	FSEL_ALT0		0b100
#define	FSEL_ALT0		0b100
#define	FSEL_ALT1		0b101
#define	FSEL_ALT2		0b110
#define	FSEL_ALT3		0b111
#define	FSEL_ALT4		0b011
#define	FSEL_ALT5		0b010


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

// The BCM2835 has 54 GPIO pins.
//	BCM2835 data sheet, Page 90 onwards.
//	There are 6 control registers, each control the functions of a block
//	of 10 pins.
//	Each control register has 10 sets of 3 bits per GPIO pin:
//
//	000 = GPIO Pin X is an input
//	001 = GPIO Pin X is an output
//	100 = GPIO Pin X takes alternate function 0
//	101 = GPIO Pin X takes alternate function 1
//	110 = GPIO Pin X takes alternate function 2
//	111 = GPIO Pin X takes alternate function 3
//	011 = GPIO Pin X takes alternate function 4
//	010 = GPIO Pin X takes alternate function 5
//
// So the 3 bits for port X are:
//	X / 10 + ((X % 10) * 3)

// sysFds:
//	Map a file descriptor from the /sys/class/gpio/gpioX/value

static int sysFds [64] ;

// Doing it the Arduino way with lookup tables...
//	Yes, it's probably more innefficient than all the bit-twidling, but it
//	does tend to make it all a bit clearer. At least to me!

// pinToGpio:
//	Take a Wiring pin (0 through X) and re-map it to the BCM_GPIO pin

static int pinToGpio [64] =
{
  17, 18, 21, 22, 23, 24, 25, 4,	// From the Original Wiki - GPIO 0 through 7
   0,  1,				// I2C  - SDA0, SCL0
   8,  7,				// SPI  - CE1, CE0
  10,  9, 11, 				// SPI  - MOSI, MISO, SCLK
  14, 15,				// UART - Tx, Rx

// Padding:

          -1, -1, -1,-1,-1,-1,-1, -1, -1, -1, -1, -1, -1, -1, -1,	// ... 31
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,	// ... 47
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,	// ... 63
} ;

// gpioToGPFSEL:
//	Map a BCM_GPIO pin to it's control port. (GPFSEL 0-5)

static uint8_t gpioToGPFSEL [] =
{
  0,0,0,0,0,0,0,0,0,0,
  1,1,1,1,1,1,1,1,1,1,
  2,2,2,2,2,2,2,2,2,2,
  3,3,3,3,3,3,3,3,3,3,
  4,4,4,4,4,4,4,4,4,4,
  5,5,5,5,5,5,5,5,5,5,
} ;

// gpioToShift
//	Define the shift up for the 3 bits per pin in each GPFSEL port

static uint8_t gpioToShift [] =
{
  0,3,6,9,12,15,18,21,24,27,
  0,3,6,9,12,15,18,21,24,27,
  0,3,6,9,12,15,18,21,24,27,
  0,3,6,9,12,15,18,21,24,27,
  0,3,6,9,12,15,18,21,24,27,
} ;

// gpioToGPSET:
//	(Word) offset to the GPIO Set registers for each GPIO pin

static uint8_t gpioToGPSET [] =
{
   7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
   8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
} ;

// gpioToGPCLR:
//	(Word) offset to the GPIO Clear registers for each GPIO pin

static uint8_t gpioToGPCLR [] =
{
  10,10,10,10,10,10,10,10,10,10,10,10,10,10,10,10,10,10,10,10,10,10,10,10,10,10,10,10,10,10,10,10,
  11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,
} ;

// gpioToGPLEV:
//	(Word) offset to the GPIO Input level registers for each GPIO pin

static uint8_t gpioToGPLEV [] =
{
  13,13,13,13,13,13,13,13,13,13,13,13,13,13,13,13,13,13,13,13,13,13,13,13,13,13,13,13,13,13,13,13,
  14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,
} ;

// gpioToPUDCLK
//	(Word) offset to the Pull Up Down Clock regsiter

static uint8_t gpioToPUDCLK [] =
{
  38,38,38,38,38,38,38,38,38,38,38,38,38,38,38,38,38,38,38,38,38,38,38,38,38,38,38,38,38,38,38,38,
  39,39,39,39,39,39,39,39,39,39,39,39,39,39,39,39,39,39,39,39,39,39,39,39,39,39,39,39,39,39,39,39,
} ;

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

