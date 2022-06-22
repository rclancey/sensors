package tsl2591
/*
reads waveshare ambient light sensor TSL2591
*/

import (
	"errors"
	"log"
	"time"

	"github.com/d2r2/go-i2c"
	"github.com/d2r2/go-logger"
)

func init() {
	logger.ChangePackageLogLevel("i2c", logger.InfoLevel)
}

const (
	INI_PIN = 4

	ADDR = 0x29

	COMMAND_BIT = 0xa0
	// Register (0x00)
	ENABLE_REGISTER = 0x00
	ENABLE_POWERON = 0x01
	ENABLE_POWEROFF = 0x00
	ENABLE_AEN = 0x02
	ENABLE_AIEN = 0x10
	ENABLE_SAI = 0x40
	ENABLE_NPIEN = 0x80

	CONTROL_REGISTER = 0x01
	SRESET = 0x80
	// AGAIN
	LOW_AGAIN           = 0x00 // Low gain (1x)
	MEDIUM_AGAIN        = 0x10 // Medium gain (25x)
	HIGH_AGAIN          = 0x20 // High gain (428x)
	MAX_AGAIN           = 0x30 // Max gain (9876x)
	// ATIME
	ATIME_100MS         = 0x00 // 100 millis //MAX COUNT 36863 
	ATIME_200MS         = 0x01 // 200 millis //MAX COUNT 65535 
	ATIME_300MS         = 0x02 // 300 millis //MAX COUNT 65535 
	ATIME_400MS         = 0x03 // 400 millis //MAX COUNT 65535 
	ATIME_500MS         = 0x04 // 500 millis //MAX COUNT 65535 
	ATIME_600MS         = 0x05 // 600 millis //MAX COUNT 65535 

	AILTL_REGISTER      = 0x04
	AILTH_REGISTER      = 0x05
	AIHTL_REGISTER      = 0x06
	AIHTH_REGISTER      = 0x07
	NPAILTL_REGISTER    = 0x08
	NPAILTH_REGISTER    = 0x09
	NPAIHTL_REGISTER    = 0x0a
	NPAIHTH_REGISTER    = 0x0b

	PERSIST_REGISTER    = 0x0c
	// Bits 3:0
	// 0000          Every ALS cycle generates an interrupt
	// 0001          Any value outside of threshold range
	// 0010          2 consecutive values out of range
	// 0011          3 consecutive values out of range
	// 0100          5 consecutive values out of range
	// 0101          10 consecutive values out of range
	// 0110          15 consecutive values out of range
	// 0111          20 consecutive values out of range
	// 1000          25 consecutive values out of range
	// 1001          30 consecutive values out of range
	// 1010          35 consecutive values out of range
	// 1011          40 consecutive values out of range
	// 1100          45 consecutive values out of range
	// 1101          50 consecutive values out of range
	// 1110          55 consecutive values out of range
	// 1111          60 consecutive values out of range

	ID_REGISTER         = 0x12

	STATUS_REGISTER     = 0x13 // read only

	CHAN0_LOW           = 0x14
	CHAN0_HIGH          = 0x15
	CHAN1_LOW           = 0x16
	CHAN1_HIGH          = 0x14

	//LUX_DF = GA * 53   GA is the Glass Attenuation factor 
	LUX_DF              = 762.0
	// LUX_DF           = 408.0
	MAX_COUNT_100MS     = 36863 // 0x8FFF
	MAX_COUNT           = 65535 // 0xFFFF
)

type TSL2591 struct {
	i2c *i2c.I2C
	address uint8
	id byte
	gain int
	integralTime int
}

func New() (*TSL2591, error) {
	self := &TSL2591{}
	self.address = ADDR
	var err error
	self.i2c, err = i2c.NewI2C(self.address, 1)
	if err != nil {
		return nil, err
	}
	/*
	err = gpio.SetMode(gpio.BCM)
	if err != nil {
		return nil, err
	}
	err = gpio.SetWarnings(false)
	if err != nil {
		return nil, err
	}
	err = gpio.Setup(INI_PIN, gpio.IN)
	if err != nil {
		return nil, err
	}
	*/

	self.id, err = self.ReadByte(ID_REGISTER)
	if err != nil {
		return nil, err
	}
	if self.id != 0x50 {
		log.Fatalf("ID = 0x%x", self.id)
	}

	err = self.Enable()
	if err != nil {
		return nil, err
	}
	err = self.SetGain(MEDIUM_AGAIN)
	if err != nil {
		return nil, err
	}
	err = self.SetIntegralTime(ATIME_100MS)
	if err != nil {
		return nil, err
	}
	err = self.WriteByte(PERSIST_REGISTER, 0x01)
	if err != nil {
		return nil, err
	}
	err = self.Disable()
	if err != nil {
		return nil, err
	}
	return self, nil
}

func (self *TSL2591) ReadByte(addr byte) (byte, error) {
	addr = (COMMAND_BIT | addr) & 0xff
	return self.i2c.ReadRegU8(addr)
}

func (self *TSL2591) ReadWord(addr byte) (int, error) {
	addr = (COMMAND_BIT | addr) & 0xff
	val, err := self.i2c.ReadRegU16LE(addr)
	return int(val), err
}

func (self *TSL2591) WriteByte(addr, val byte) error {
	addr = (COMMAND_BIT | addr) & 0xff
	return self.i2c.WriteRegU8(addr, val)
}

func (self *TSL2591) Enable() error {
	return self.WriteByte(ENABLE_REGISTER, ENABLE_AIEN | ENABLE_POWERON | ENABLE_AEN | ENABLE_NPIEN)
}

func (self *TSL2591) Disable() error {
	return self.WriteByte(ENABLE_REGISTER, ENABLE_POWEROFF)
}

func (self *TSL2591) GetGain() (int, error) {
	val, err := self.ReadByte(CONTROL_REGISTER)
	if err != nil {
		return int(val), err
	}
	return int(val) & 0x30, nil
}

var ErrGainParameterError = errors.New("Gain Parameter Error")
func (self *TSL2591) SetGain(val int) error {
	if val == LOW_AGAIN || val == MEDIUM_AGAIN || val == HIGH_AGAIN || val == MAX_AGAIN {
		control, err := self.ReadByte(CONTROL_REGISTER)
		if err != nil {
			return err
		}
		control &= 0xcf
		control |= byte(val)
		err = self.WriteByte(CONTROL_REGISTER, control)
		if err != nil {
			return err
		}
		self.gain = val
	} else {
		return ErrGainParameterError
	}
	return nil
}

func (self *TSL2591) GetIntegralTime() (int, error) {
	control, err := self.ReadByte(CONTROL_REGISTER)
	if err != nil {
		return int(control), err
	}
	return int(control) & 0x07, nil
}

var ErrIntegralTimeParameterError = errors.New("Integral Time Parameter Error")
func (self *TSL2591) SetIntegralTime(val int) error {
	if val & 0x07 < 0x06 {
		control, err := self.ReadByte(CONTROL_REGISTER)
		if err != nil {
			return err
		}
		control &= 0xf8
		control |= byte(val)
		self.WriteByte(CONTROL_REGISTER, control)
		self.integralTime = val
	} else {
		return ErrIntegralTimeParameterError
	}
	return nil
}

func (self *TSL2591) ReadCHAN0() (int, error) {
	return self.ReadWord(CHAN0_LOW)
}

func (self *TSL2591) ReadCHAN1() (int, error) {
	return self.ReadWord(CHAN1_LOW)
}

func (self *TSL2591) ReadFullSpectrum() (int, error) {
	err := self.Enable()
	if err != nil {
		return 0, err
	}
	defer self.Disable()
	ch0, err := self.ReadCHAN0()
	if err != nil {
		return 0, err
	}
	ch1, err := self.ReadCHAN1()
	if err != nil {
		return 0, err
	}
	return (ch1 << 16) | ch0, nil
}

func (self *TSL2591) ReadInfrared() (int, error) {
	err := self.Enable()
	if err != nil {
		return 0, err
	}
	defer self.Disable()
	return self.ReadCHAN0()
}

func (self *TSL2591) ReadVisible() (int, error) {
	err := self.Enable()
	if err != nil {
		return 0, err
	}
	defer self.Disable()
	ch1, err := self.ReadCHAN1()
	if err != nil {
		return 0, err
	}
	ch0, err := self.ReadCHAN0()
	if err != nil {
		return 0, err
	}
	full := (ch1 << 16) | ch0
	return full - ch1, nil
}

var ErrNumericalOverflow = errors.New("Numerical overflow")
func (self *TSL2591) Lux() (int, error) {
	err := self.Enable()
	if err != nil {
		return 0, err
	}
	n := self.integralTime + 2
	for i := 0; i < n; i += 1 {
		time.Sleep(100 * time.Millisecond)
	}
	ch0, err := self.ReadCHAN0()
	if err != nil {
		self.Disable()
		return 0, err
	}
	ch1, err := self.ReadCHAN1()
	if err != nil {
		self.Disable()
		return 0, err
	}
	err = self.Disable()
	if err != nil {
		return 0, err
	}

	err = self.Enable()
	if err != nil {
		return 0, err
	}
	err = self.WriteByte(0xe7, 0x13) // Clear interrupt flag
	if err != nil {
		self.Disable()
		return 0, err
	}
	err = self.Disable()
	if err != nil {
		return 0, err
	}

	atime := 100.0 * float64(self.integralTime) + 100.0

	// Set the maximum sensor counts based on the integration time (atime) setting
	var maxCounts int
	if self.integralTime == ATIME_100MS {
		maxCounts = MAX_COUNT_100MS
	} else {
		maxCounts = MAX_COUNT
	}

	if ch0 >= maxCounts || ch1 >= maxCounts {
		gainT, err := self.GetGain()
		if err != nil {
			return 0, err
		}
		if gainT != LOW_AGAIN {
			gainT = ((gainT >> 4) - 1) << 4
			err = self.SetGain(gainT)
			if err != nil {
				return 0, err
			}
			ch0 = 0
			ch1 = 0
			for ch0 <= 0 && ch1 <= 0 {
				ch0, err = self.ReadCHAN0()
				if err != nil {
					return 0, err
				}
				ch1, err = self.ReadCHAN1()
				if err != nil {
					return 0, err
				}
				time.Sleep(100 * time.Millisecond)
			}
		} else {
			return 0, ErrNumericalOverflow
		}
	}
	var again float64
	switch self.gain {
	case MEDIUM_AGAIN:
		again = 25
	case HIGH_AGAIN:
		again = 428
	case MAX_AGAIN:
		again = 9876
	default:
		again = 1
	}
	cpl := (atime * again) / LUX_DF
	lux1 := float64(ch0  - (2 * ch1)) / cpl
	// lux2 = ((0.6 * channel_0) - (channel_1)) / Cpl
	// This is a two segment lux equation where the first 
	// segment (Lux1) covers fluorescent and incandescent light 
	// and the second segment (Lux2) covers dimmed incandescent light

	if lux1 > 0 {
		return int(lux1), nil
	}
	return 0, nil
}

func (self *TSL2591) SetInterrupThreshold(high, low int) error {
	err := self.Enable()
	if err != nil {
		return err
	}
	defer self.Disable()
	err = self.WriteByte(AILTL_REGISTER, byte(low & 0xff))
	if err != nil {
		return err
	}
	err = self.WriteByte(AILTH_REGISTER, byte(low >> 8))
	if err != nil {
		return err
	}
	err = self.WriteByte(AIHTL_REGISTER, byte(high & 255))
	if err != nil {
		return err
	}
	err = self.WriteByte(AIHTH_REGISTER, byte(high >> 8))
	if err != nil {
		return err
	}
	err = self.WriteByte(NPAILTL_REGISTER, 0)
	if err != nil {
		return err
	}
	err = self.WriteByte(NPAILTH_REGISTER, 0)
	if err != nil {
		return err
	}
	err = self.WriteByte(NPAIHTL_REGISTER, 0xff)
	if err != nil {
		return err
	}
	err = self.WriteByte(NPAIHTH_REGISTER, 0xff)
	if err != nil {
		return err
	}
	return nil
}

func (self *TSL2591) SetLuxInterrupt(setLow, setHigh int) error {
	atime := 100 * float64(self.integralTime) + 100
	var again float64
	switch self.gain {
	case MEDIUM_AGAIN:
		again = 25
	case HIGH_AGAIN:
		again = 428
	case MAX_AGAIN:
		again = 9876
	default:
		again = 1
	}
	cpl := (atime * again) / LUX_DF
	ch1, err := self.ReadCHAN1()
	if err != nil {
		return err
	}
	setHigh = int(cpl * float64(setHigh)) + 2 * ch1 - 1
	setLow = int(cpl * float64(setLow)) + 2 * ch1 + 1
	err = self.Enable()
	if err != nil {
		return err
	}
	defer self.Disable()
	err = self.WriteByte(AILTL_REGISTER, byte(setLow & 0xff))
	if err != nil {
		return err
	}
	err = self.WriteByte(AILTH_REGISTER, byte(setLow >> 8))
	if err != nil {
		return err
	}
	err = self.WriteByte(AIHTL_REGISTER, byte(setHigh & 0xff))
	if err != nil {
		return err
	}
	err = self.WriteByte(AIHTH_REGISTER, byte(setHigh >> 8))
	if err != nil {
		return err
	}
	err = self.WriteByte(NPAILTL_REGISTER, 0)
	if err != nil {
		return err
	}
	err = self.WriteByte(NPAILTH_REGISTER, 0)
	if err != nil {
		return err
	}
	err = self.WriteByte(NPAIHTL_REGISTER, 0xff)
	if err != nil {
		return err
	}
	err = self.WriteByte(NPAIHTH_REGISTER, 0xff)
	if err != nil {
		return err
	}
	return nil
}

type SensorData struct {
	Lux          int `json:"lux"`
	Infrared     int `json:"infrared"`
	Visible      int `json:"visible"`
	FullSpectrum int `json:"full_spectrum"`
}

func (self *TSL2591) ReadSensorData() (*SensorData, error) {
	data := &SensorData{}
	var err error
	data.Lux, err = self.Lux()
	if err != nil {
		return data, err
	}
	err = self.SetLuxInterrupt(50, 200)
	if err != nil {
		return data, err
	}
	data.Infrared, err = self.ReadInfrared()
	if err != nil {
		return data, err
	}
	data.Visible, err = self.ReadVisible()
	if err != nil {
		return data, err
	}
	data.FullSpectrum, err = self.ReadFullSpectrum()
	if err != nil {
		return data, err
	}
	return data, nil
}
