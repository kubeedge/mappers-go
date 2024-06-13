/*
            Atlas 200DK A1 PIN define
  +------+---------+-------++------+---------+-------+
  |  pin |   Name  |voltage||  pin |   Name  |voltage|
  +-----+-----------+------++----+---------+-----+
  |   1  | +3.3v   |  3.3V ||   2  |  +5.0v  |  5.0V |
  |   3  | SDA     |  3.3V ||   4  |  +5.0v  |  5.0V |
  |   5  | SCL     |  3.3V ||   6  | GND     |  -    |
  |   7  | GPIO-0  |  3.3V ||   8  | TXD0    |  3.3V |
  |   9  | GND     |  -    ||  10  | RXD0    |  3.3V |
  |  11  | GPIO-1  |  3.3V ||  12  | NC      |  -    |
  |  13  | NC      |  3.3V ||  14  | GND     |  -    |
  |  15  | GPIO-2  |  3.3V ||  16  | TXD1    |  3.3V |
  |  17  | +3.3v   |  3.3V ||  18  | RXD1    |  3.3V |
  |  19  | SPI-MOSI|  3.3V ||  20  | GND     |  -    |
  |  21  | SPI-MISO|  3.3V ||  22  | NC      |  -    |
  |  23  | SPI-CLK |  3.3V ||  24  | SPI-CS  |  3.3V |
  |  25  | GND     |  -    ||  26  | NC      |  -    |
  |  27  | CAN-H   |  -    ||  28  | CAN-1   |  -    |
  |  29  | GPIO-3  |  3.3V ||  30  | GND     |  -    |
  |  31  | GPIO-4  |  3.3V ||  32  | NC      |  -    |
  |  33  | GPIO-5  |  3.3V ||  34  | GND     |  -    |
  |  35  | GPIO-6  |  3.3V ||  36  | +1.8V   |  1.8V |
  |  37  | GPIO-7  |  3.3V ||  38  | TXD-3559|  3.3V |
  |  39  | GND     |  3.3V ||  40  | RXD-3559|  3.3V |
  +------+---------+-------++------+---------+-------+
gpio 0~1 are directly derived from the Ascend AI processor,
gpio 2 is not  available for user
gpio 3~7 are derived from PCA6416,controlled by i2c
*/

package driver

import (
	"fmt"
	"io/fs"
	"k8s.io/klog/v2"
	"os"
	"syscall"
	"unsafe"
)

// Mode type int8
type Mode uint8

// Pin type int8
type Pin uint8

// State type int8
type State uint8
type i2cMsg struct {
	addr    uint16
	flags   uint16
	len     uint16
	padding uint16
	buf     uintptr
}

type i2cCtrl struct {
	msgs   uintptr
	msgNum uint32
}

const (
	ascendGpio0dir = "/sys/class/gpio/gpio504/direction"
	ascendGpio1dir = "/sys/class/gpio/gpio444/direction"
	ascendgpio0Val = "/sys/class/gpio/gpio504/value"
	ascendgpio1Val = "/sys/class/gpio/gpio444/value"
)
const (
	i2cDeviceName = "/dev/i2c-1"
	i2cRetres     = 0x0701
	i2cTimeOut    = 0x0702
	i2cSlave      = 0x0703
	i2cRDWR       = 0x0707
	i2cmRD        = 0x01

	pca6416SlaveAddr       = 0x20
	pca6416GpioCfgReg      = 0x07
	pca6416GpioPorarityReg = 0x05
	pca6416GpioOutReg      = 0x03
	pca6416GpioInReg       = 0x01

	//GPIO MASK
	gpio3Mask = 0x10
	gpio4Mask = 0x20
	gpio5Mask = 0x40
	gpio6Mask = 0x80
	gpio7Mask = 0x08
)

// Generic ioctl constants
const (
	iocNone     = 0x0
	iocWrite    = 0x1
	iocRead     = 0x2
	iocNRBits   = 8
	iocTypeBits = 8

	iocSizeBits = 14
	iocDirBits  = 2

	iocNRShift   = 0
	iocTypeShift = iocNRShift + iocNRBits     //8 + 0
	iocSizeShift = iocTypeShift + iocTypeBits //8 + 8
	iocDirShift  = iocSizeShift + iocSizeBits //16 + 14

	iocNRMask   = ((1 << iocNRBits) - 1)
	iocTYPEMask = ((1 << iocTypeBits) - 1)
	iocSizeMask = ((1 << iocSizeBits) - 1)
	iocDirMask  = ((1 << iocDirBits) - 1)
)

const (
	// Input the pin if for input
	Input Mode = iota
	// Output the pin if for output
	Output
)

// State of pin, High / Low
const (
	Low uint8 = iota
	High
)

// SetInPut  Set pin as inputPin
func (pin Pin) SetInPut() {
	setPinMode(pin, Input)
}

// SetOutPut Set pin as Output
func (pin Pin) SetOutPut() {
	setPinMode(pin, Output)
}

// SetHight Set pin Hight
func (pin Pin) SetHight() {
	err := gpioSetValue(pin, High)
	if err != nil {
		klog.Errorf("gpioSetValue fail, pin %v err = %v.", pin, err)
	}
}

// SetLow  Set pin as Low
func (pin Pin) SetLow() {
	err := gpioSetValue(pin, Low)
	if err != nil {
		klog.Errorf("gpioSetValue fail, pin %v err = %v.", pin, err)
	}
}

// Write: Set pin state (high/low)
func (pin Pin) Write(val uint8) {
	WritePin(pin, val)
}

// Read pin state (high/low)
func (pin Pin) Read() uint8 {
	return ReadPin(pin)
}

// WritePin is to write value to pin
func WritePin(pin Pin, val uint8) {
	err := gpioSetValue(pin, val)
	if err != nil {
		klog.Errorf("gpioSetValue fail, pin %v err = %v.", pin, err)
	}
}

// ReadPin is to read value of pin
func ReadPin(pin Pin) uint8 {
	var val uint8
	err := gpioGetValue(pin, &val)
	if err != nil {
		return 0
	}
	return val
}

// setPinMode Spi mode should not be set by this directly, use SpiBegin instead.
func setPinMode(pin Pin, mode Mode) {
	f := uint8(0)
	const in uint8 = 0  // 000
	const out uint8 = 1 // 001

	switch mode {
	case Input:
		f = in
	case Output:
		f = out
	}
	err := gpioSetDirection(pin, f)
	if err != nil {
		klog.Errorf("gpioSetValue fail, pin %v err = %v.", pin, err)
	}
}

// Open  open a pin
func Open() (err error) {
	return nil
}

// Close  close a pin
func Close() (err error) {
	return nil
}
func isPca6416Pin(pin Pin) bool {
	if pin >= 3 && pin <= 7 {
		return true
	}
	return false
}

// IOCTL send ioctl
func IOCTL(f *os.File, flag, data uintptr) error {
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(f.Fd()), flag, uintptr(data))
	if err != 0 {
		return syscall.Errno(err)
	}
	return nil
}

func i2cRead(slave uint8, reg uint8, data *uint8) error {
	regs := []uint8{reg, reg}
	msg := []i2cMsg{
		{
			addr:  uint16(slave),
			flags: 0,
			len:   uint16(1), //the length of reg addr is 1
			buf:   uintptr(unsafe.Pointer(&regs[0])),
		},
		{
			addr:  uint16(slave),
			flags: i2cmRD,
			len:   uint16(1),
			buf:   uintptr(unsafe.Pointer(data)),
		},
	}
	ssmMsg := i2cCtrl{
		msgs:   uintptr(unsafe.Pointer(&msg[0])),
		msgNum: uint32(len(msg)),
	}

	perm := fs.FileMode(0644) //-rw-r--r--
	flag := int(os.O_RDWR | os.O_CREATE | os.O_TRUNC)
	f, err := os.OpenFile(i2cDeviceName, flag, perm)
	defer f.Close()
	if err != nil {
		return err
	}

	err = IOCTL(f, i2cRDWR, uintptr(unsafe.Pointer(&ssmMsg)))
	return err
}

func i2cWrite(slave uint8, reg uint8, data uint8) error {
	buf := []uint8{reg, data}
	msg := []i2cMsg{
		{
			addr:  uint16(slave),
			flags: 0,
			len:   uint16(len(buf)),
			buf:   uintptr(unsafe.Pointer(&buf[0])),
		},
	}

	ssmMsg := i2cCtrl{
		msgs:   uintptr(unsafe.Pointer(&msg[0])),
		msgNum: uint32(1),
	}

	perm := fs.FileMode(666) // -rw-rw-rw-
	flag := int(os.O_RDWR | os.O_CREATE | os.O_TRUNC)
	f, err := os.OpenFile(i2cDeviceName, flag, perm)
	defer f.Close()
	if err != nil {
		return err
	}
	err = IOCTL(f, i2cRDWR, uintptr(unsafe.Pointer(&ssmMsg)))
	return err
}

func pca6416GpioSetDirection(pin Pin, dir uint8) error {
	var err error
	var data uint8
	var reg uint8
	var slave uint8
	var gpioMask = []uint8{0, 0, 0, gpio3Mask, gpio4Mask, gpio5Mask, gpio6Mask, gpio7Mask}

	if !isPca6416Pin(pin) {
		err = fmt.Errorf("pin number is incorrect,must be 3 to 7")
		return err
	}
	slave = pca6416SlaveAddr
	reg = pca6416GpioCfgReg
	data = 0
	err = i2cRead(slave, reg, &data)
	if err != nil {
		klog.Errorf("pca6416GpioSetDirection read fail, pin %v err = %v.", pin, err)
		return err
	}
	if dir == 0 {
		data |= gpioMask[pin]
	} else {
		data &= ^gpioMask[pin]
	}
	err = i2cWrite(slave, reg, data)
	if err != nil {
		klog.Errorf("pca6416GpioSetDirection write fail pin %v err = %v.", pin, err)
	}
	return err
}
func pca6416GpioSetValue(pin Pin, val uint8) error {
	var err error
	var data uint8
	var reg uint8
	var slave uint8
	var gpioMask = []uint8{0, 0, 0, gpio3Mask, gpio4Mask, gpio5Mask, gpio6Mask, gpio7Mask}

	if !isPca6416Pin(pin) {
		err = fmt.Errorf("pin number is incorrect,must be 3 to 7")
		return err
	}
	slave = pca6416SlaveAddr
	reg = pca6416GpioOutReg
	data = 0
	err = i2cRead(slave, reg, &data)
	if err != nil {
		klog.Errorf("pca6416GpioSetValue read fail, pin %v err = %v.", pin, err)
		return err
	}
	if val == 0 {
		data &= ^gpioMask[pin]
	} else {
		data |= gpioMask[pin]
	}

	err = i2cWrite(slave, reg, data)
	if err != nil {
		klog.Errorf("pca6416GpioSetValue write fail pin %v err = %v.", pin, err)
	}
	return err
}
func pca6416GpioGetValue(pin Pin, val *uint8) error {
	var err error
	var data uint8
	var reg uint8
	var slave uint8
	var gpioMask = []uint8{0, 0, 0, gpio3Mask, gpio4Mask, gpio5Mask, gpio6Mask, gpio7Mask}

	if !isPca6416Pin(pin) {
		err = fmt.Errorf("pin number is incorrect,must be 3 to 7")
		return err
	}
	slave = pca6416SlaveAddr
	reg = pca6416GpioInReg
	data = 0
	err = i2cRead(slave, reg, &data)
	if err != nil {
		klog.Errorf("pca6416GpioSetValue read fail, pin %v err = %v.", pin, err)
		return err
	}
	data &= gpioMask[pin]
	if data > 0 {
		*val = '1'
	} else {
		*val = '0'
	}
	return nil
}

// AscendGpioSetDirection set gpio direction
func AscendGpioSetDirection(pin Pin, dir uint8) error {
	var fileName string
	var direction string
	var err error

	if pin == 0 {
		fileName = ascendGpio0dir
	} else if pin == 1 {
		fileName = ascendGpio1dir
	} else {
		err = fmt.Errorf("pin number is incorrect,must be 0 or 1")
		return err
	}
	direction = "out"
	if dir == 0 {
		direction = "in"
	}
	err = os.WriteFile(fileName, []byte(direction), 0644)
	if err != nil {
		klog.Errorf("os.WriteFile fileName= %v err = %v ", fileName, err)
		return err
	}

	return nil
}

// AscendGpioSetValue set gpio value
func AscendGpioSetValue(pin Pin, val uint8) error {
	var fileName string
	var err error
	if pin == 0 {
		fileName = ascendgpio0Val
	} else if pin == 1 {
		fileName = ascendgpio1Val
	} else {
		err = fmt.Errorf("pin number is incorrect,must be 0 or 1")
		return err
	}
	klog.V(3).Infof("AscendGpioSetValue pin %v val = %v fileName = %v", pin, val, fileName)
	buff := []byte{val + '0'}
	err = os.WriteFile(fileName, buff, 0644)
	if err != nil {
		klog.Errorf("os.WriteFile fileName= %v err = %v ", fileName, err)
	}
	return err
}

// AscendGpioGetValue get gpio direction
func AscendGpioGetValue(pin Pin, val *uint8) error {
	var fileName string
	if pin == 0 {
		fileName = ascendgpio0Val
	} else if pin == 1 {
		fileName = ascendgpio1Val
	} else {
		err := fmt.Errorf("pin number is incorrect,the correct num is must be 0,1")
		return err
	}
	readFile, err := os.ReadFile(fileName)
	*val = readFile[0]
	if err != nil {
		klog.Errorf("AscendGpioGetValue pin %v err = %v.", pin, err)
	}
	klog.V(5).Infof("AscendGpioGetValue pin %v val = %v.", pin, *val)
	return err
}
func isAscendPin(pin Pin) bool {
	if pin == 0 || pin == 1 {
		return true
	}
	return false
}

// set gpio direction ,0-- in ,1--out
func gpioSetDirection(pin Pin, direction uint8) error {
	var result error
	if isAscendPin(pin) {
		result = AscendGpioSetDirection(pin, direction)
	} else {
		result = pca6416GpioSetDirection(pin, direction)
	}

	if nil != result {
		klog.V(3).Infof("gpioSetDirection fail, pin= %v direction= %v result = %v", pin, direction, result)
	} else {
		klog.V(5).Infof("gpioSetDirection ok, pin= %v direction = %v", pin, direction)
	}
	return result
}

func gpioSetValue(pin Pin, val uint8) error {
	var result error

	if isAscendPin(pin) {
		result = AscendGpioSetValue(pin, val)
	} else {
		result = pca6416GpioSetValue(pin, val)
	}
	if nil != result {
		klog.V(3).Infof("gpioSetValue fail, pin= %v val= %v result = %v", pin, val, result)
	} else {
		klog.V(5).Infof("gpioSetValue ok, pin= %v val = %v", pin, val)
	}
	return result
}
func gpioGetValue(pin Pin, val *uint8) error {
	var result error

	if isAscendPin(pin) {
		result = AscendGpioGetValue(pin, val)
	} else {
		result = pca6416GpioGetValue(pin, val)
	}

	if nil != result {
		klog.V(3).Infof("gpioGetValue fail, pin= %v result = %v", pin, result)
	} else {
		klog.V(5).Infof("gpioGetValue ok, pin= %v val = %v", pin, *val)
	}
	return result
}
