/*
            Atlas 200DK PIN define
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

type Mode uint8
type Pin uint8
type State uint8
type i2c_msg struct {
	addr      uint16
	flags     uint16
	len       uint16
	__padding uint16
	buf       uintptr
}

type i2c_ctrl struct {
	msgs    uintptr
	msg_num uint32
}

const (
	ascend_gpio_0_dir = "/sys/class/gpio/gpio504/direction"
	ascend_gpio_1_dir = "/sys/class/gpio/gpio444/direction"
	ascend_gpio_0_val = "/sys/class/gpio/gpio504/value"
	ascend_gpio_1_val = "/sys/class/gpio/gpio444/value"
)
const (
	i2c_device_name = "/dev/i2c-1"
	i2c_retres      = 0x0701
	i2c_timeout     = 0x0702
	i2c_slave       = 0x0703
	i2c_rdwr        = 0x0707
	i2c_m_rd        = 0x01

	pca6416_slave_addr        = 0x20
	pca6416_gpio_cfg_reg      = 0x07
	pca6416_gpio_porarity_reg = 0x05
	pca6416_gpio_out_reg      = 0x03
	pca6416_gpio_in_reg       = 0x01

	//GPIO MASK

	gpio3_mask = 0x10
	gpio4_mask = 0x20
	gpio5_mask = 0x40
	gpio6_mask = 0x80
	gpio7_mask = 0x08
)

// Generic ioctl constants
const (
	IOC_NONE  = 0x0
	IOC_WRITE = 0x1
	IOC_READ  = 0x2

	IOC_NRBITS   = 8
	IOC_TYPEBITS = 8

	IOC_SIZEBITS = 14
	IOC_DIRBITS  = 2

	IOC_NRSHIFT   = 0
	IOC_TYPESHIFT = IOC_NRSHIFT + IOC_NRBITS     //8 + 0
	IOC_SIZESHIFT = IOC_TYPESHIFT + IOC_TYPEBITS //8 + 8
	IOC_DIRSHIFT  = IOC_SIZESHIFT + IOC_SIZEBITS //16 + 14

	IOC_NRMASK   = ((1 << IOC_NRBITS) - 1)
	IOC_TYPEMASK = ((1 << IOC_TYPEBITS) - 1)
	IOC_SIZEMASK = ((1 << IOC_SIZEBITS) - 1)
	IOC_DIRMASK  = ((1 << IOC_DIRBITS) - 1)
)

// Some useful additional ioctl constanst
const (
	IOC_IN        = IOC_WRITE << IOC_DIRSHIFT
	IOC_OUT       = IOC_READ << IOC_DIRSHIFT
	IOC_INOUT     = (IOC_WRITE | IOC_READ) << IOC_DIRSHIFT
	IOCSIZE_MASK  = IOC_SIZEMASK << IOC_SIZESHIFT
	IOCSIZE_SHIFT = IOC_SIZESHIFT
)

const (
	Input Mode = iota
	Output
)

// State of pin, High / Low
const (
	Low uint8 = iota
	High
)

// setInput: Set pin as InputPin
func (pin Pin) SetInPut() {
	setPinMode(pin, Input)
}

// setOutput: Set pin as Output
func (pin Pin) SetOutPut() {
	setPinMode(pin, Output)
}

// setHight: Set pin Hight
func (pin Pin) SetHight() {
	gpioSetValue(pin, High)
}

// setLow: Set pin as Low
func (pin Pin) SetLow() {
	gpioSetValue(pin, Low)
}

// Write: Set pin state (high/low)
func (pin Pin) Write(val uint8) {
	WritePin(pin, val)
}

// Read pin state (high/low)
func (pin Pin) Read() uint8 {
	return ReadPin(pin)
}

func WritePin(pin Pin, val uint8) {
	gpioSetValue(pin, val)
}

func ReadPin(pin Pin) uint8 {
	var val uint8
	err := gpioGetValue(pin, &val)
	if err != nil {
		return 0
	}
	return val
}

// Spi mode should not be set by this directly, use SpiBegin instead.
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
	gpioSetDirection(pin, f)
}

func Open() (err error) {
	return nil
}
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
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(f.Fd()), 0x0707, uintptr(data))
	if err != 0 {
		return syscall.Errno(err)
	}
	return nil
}

func i2c_read(slave uint8, reg uint8, data *uint8) error {
	regs := []uint8{reg, reg}
	msg := []i2c_msg{
		{
			addr:  uint16(slave),
			flags: 0,
			len:   uint16(1), //the length of reg addr is 1
			buf:   uintptr(unsafe.Pointer(&regs[0])),
		},
		{
			addr:  uint16(slave),
			flags: i2c_m_rd,
			len:   uint16(1),
			buf:   uintptr(unsafe.Pointer(data)),
		},
	}

	ssm_msg := i2c_ctrl{
		msgs:    uintptr(unsafe.Pointer(&msg[0])),
		msg_num: uint32(len(msg)),
	}

	perm := fs.FileMode(0644) //--w----r--
	flag := int(os.O_RDWR | os.O_CREATE | os.O_TRUNC)
	f, err := os.OpenFile(i2c_device_name, flag, perm)
	defer f.Close()
	if err != nil {
		return err
	}

	err = IOCTL(f, i2c_rdwr, uintptr(unsafe.Pointer(&ssm_msg)))
	return err
}

func i2c_write(slave uint8, reg uint8, data uint8) error {
	buf := []uint8{reg, data}
	msg := []i2c_msg{
		{
			addr:  uint16(slave),
			flags: 0,
			len:   uint16(len(buf)),
			buf:   uintptr(unsafe.Pointer(&buf[0])),
		},
	}

	ssm_msg := i2c_ctrl{
		msgs:    uintptr(unsafe.Pointer(&msg[0])),
		msg_num: uint32(1),
	}

	perm := fs.FileMode(666)
	flag := int(os.O_RDWR | os.O_CREATE | os.O_TRUNC)
	f, err := os.OpenFile(i2c_device_name, flag, perm)
	defer f.Close()
	if err != nil {
		return err
	}
	err = IOCTL(f, i2c_rdwr, uintptr(unsafe.Pointer(&ssm_msg)))
	return err
}

func pca6416GpioSetDirection(pin Pin, dir uint8) error {
	var err error
	var data uint8
	var reg uint8
	var slave uint8
	var gpio_mask = []uint8{0, 0, 0, gpio3_mask, gpio4_mask, gpio5_mask, gpio6_mask, gpio7_mask}

	if false == isPca6416Pin(pin) {
		err = fmt.Errorf("pin number is incorrect,must be 3 to 7")
		return err
	}
	slave = pca6416_slave_addr
	reg = pca6416_gpio_cfg_reg
	data = 0
	err = i2c_read(slave, reg, &data)
	if err != nil {
		klog.Errorf("pca6416GpioSetDirection read fail, pin %v err = %v.", pin, err)
		return err
	}
	if dir == 0 {
		data |= gpio_mask[pin]
	} else {
		data &= ^gpio_mask[pin]
	}
	err = i2c_write(slave, reg, data)
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
	var gpio_mask = []uint8{0, 0, 0, gpio3_mask, gpio4_mask, gpio5_mask, gpio6_mask, gpio7_mask}

	if false == isPca6416Pin(pin) {
		err = fmt.Errorf("pin number is incorrect,must be 3 to 7")
		return err
	}
	slave = pca6416_slave_addr
	reg = pca6416_gpio_out_reg
	data = 0
	err = i2c_read(slave, reg, &data)
	if err != nil {
		klog.Errorf("pca6416GpioSetValue read fail, pin %v err = %v.", pin, err)
		return err
	}
	if val == 0 {
		data &= ^gpio_mask[pin]
	} else {
		data |= gpio_mask[pin]
	}

	err = i2c_write(slave, reg, data)
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
	var gpio_mask = []uint8{0, 0, 0, gpio3_mask, gpio4_mask, gpio5_mask, gpio6_mask, gpio7_mask}

	if false == isPca6416Pin(pin) {
		err = fmt.Errorf("pin number is incorrect,must be 3 to 7")
		return err
	}
	slave = pca6416_slave_addr
	reg = pca6416_gpio_in_reg
	data = 0
	err = i2c_read(slave, reg, &data)
	if err != nil {
		klog.Errorf("pca6416GpioSetValue read fail, pin %v err = %v.", pin, err)
		return err
	}
	data &= gpio_mask[pin]
	if data > 0 {
		*val = 1
	} else {
		*val = 0
	}
	return nil
}

func AscendGpioSetDirection(pin Pin, dir uint8) error {
	var fileName string
	var direction string
	var err error

	if pin == 0 {
		fileName = ascend_gpio_0_dir
	} else if pin == 1 {
		fileName = ascend_gpio_1_dir
	} else {
		err = fmt.Errorf("pin number is incorrect,must be 0 or 1")
		return err
	}
	direction = "out"
	if dir == 0 {
		direction = "in"
	}
	err = os.WriteFile(fileName, []byte(direction), 0777)
	if err != nil {
		klog.Errorf("os.WriteFile fileName= %v err = %v ", fileName, err)
		return err
	}

	return nil
}

func AscendGpioSetValue(pin Pin, val uint8) error {
	var fileName string
	var err error
	if pin == 0 {
		fileName = ascend_gpio_0_val
	} else if pin == 1 {
		fileName = ascend_gpio_1_val
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

func AscendGpioGetValue(pin Pin, val *uint8) error {
	var fileName string
	if pin == 0 {
		fileName = ascend_gpio_0_val
	} else if pin == 1 {
		fileName = ascend_gpio_1_val
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
	klog.V(3).Infof("gpioSetDirection pin %v val = %v", pin, direction)
	if true == isAscendPin(pin) {
		return AscendGpioSetDirection(pin, direction)
	} else {
		return pca6416GpioSetDirection(pin, direction)
	}
}

func gpioSetValue(pin Pin, val uint8) error {
	klog.V(3).Infof("gpioSetValue pin %v val = %v", pin, val)
	if true == isAscendPin(pin) {
		return AscendGpioSetValue(pin, val)
	} else {
		return pca6416GpioSetValue(pin, val)
	}
}
func gpioGetValue(pin Pin, val *uint8) error {
	if true == isAscendPin(pin) {
		return AscendGpioGetValue(pin, val)
	} else {
		return pca6416GpioGetValue(pin, val)
	}
}

func Gpio_test() {
	var value uint8 = 0
	var pin Pin
	var result error

	for pin = 1; pin <= 8; pin++ {
		// value = 0
		// result = gpioSetValue(pin, value)
		// fmt.Printf("\n gpioSetValue pin %v val = %v,result = %v", pin, value, result)
		// time.Sleep(10 * (time.Millisecond))
		// result = gpioGetValue(pin, &value)
		// fmt.Printf("\n gpioGetValue pin %v val = %v, result = %v", pin, value, result)
		// time.Sleep(10 * (time.Millisecond))
		fmt.Printf("\r\n\r\n gpioSetDirection pin %v Direction = %v ", pin, 1)
		result = gpioSetDirection(pin, 1)
		result = result
		//fmt.Printf("\r\n result result =  %v ", result)

		//fmt.Printf("\r\n\r\n gpioSetDirection pin %v Direction = %v ", 4, 1)
		//result = gpioSetDirection(4, 1)
		////fmt.Printf("\r\n result result =  %v ", result)
		//
		//fmt.Printf("\r\n\r\n gpioSetDirection pin %v Direction = %v ", 5, 1)
		//result = gpioSetDirection(5, 1)
		////fmt.Printf("\r\n result result =  %v ", result)

		value = 0
		fmt.Printf("\r\n\r\n gpioSetValue pin %v val = %v", pin, value)
		result = gpioSetValue(pin, value)
		//fmt.Printf("\r\n result result =  %v ", result)
		time.Sleep(10 * (time.Millisecond))

		result = gpioGetValue(pin, &value)
		fmt.Printf("\r\n \r\n gpioGetValue pin %v val = %v", pin, value)
		fmt.Printf("\r\n result result =  %v ", result)
		time.Sleep(10 * (time.Millisecond))

		value = 1
		fmt.Printf("\r\n\r\n gpioSetValue pin %v val = %v", pin, value)
		result = gpioSetValue(pin, value)
		//fmt.Printf("\r\n result result =  %v ", result)
		time.Sleep(10 * (time.Millisecond))

		result = gpioGetValue(pin, &value)
		fmt.Printf("\r\n \r\n gpioGetValue pin %v val = %v", pin, value)
		fmt.Printf("\r\n result result =  %v ", result)
		time.Sleep(10 * (time.Millisecond))

		fmt.Printf("\r\n ")
	}
}
