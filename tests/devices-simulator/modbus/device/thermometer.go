/*
Copyright 2021 The KubeEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package device

import (
	"context"
	"encoding/binary"
	"github.com/goburrow/modbus"
	"github.com/pkg/errors"
	"github.com/tbrandon/mbserver"
	log "k8s.io/klog"
	"math/rand"
	"net"
	"time"
)

type thermometer struct {
	handler modbus.ClientHandler
	ctx     context.Context
	cancel  context.CancelFunc
}

func newThermometer(handler modbus.ClientHandler) *thermometer {
	ctx, cancel := context.WithCancel(context.Background())

	return &thermometer{
		handler: handler,
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (t *thermometer) close() {
	if conn, ok := t.handler.(net.Conn); ok {
		_ = conn.Close()
	}

	if t.cancel != nil {
		t.cancel()
	}
}

func changeRegisterValue(register []uint16, address int, value []byte) {
	if len(value)%2 != 0 {
		value = append(value, byte(0))
	}
	for i := 0; i < len(value)/2; i++ {
		register[address+i] = binary.BigEndian.Uint16(value[2*i : 2*i+2])
	}
}

func setDefaultValue(s *mbserver.Server, client modbus.Client) error {
	// set the switch of thermometer is on
	_, err := client.WriteSingleCoil(0, 0xff00)
	if err != nil {
		log.Error("fail to write value to coil register")
		return err
	}

	// set the alarming temperature as 20
	_, err = client.WriteMultipleRegisters(4, 1, []byte{0, 20})
	if err != nil {
		log.Error("fail to write the alarming temperature to holding register")
		return err
	}
	log.Info("set the alarming temperature as 20")

	// set the alarming humidity 60
	_, err = client.WriteMultipleRegisters(7, 2, []byte{0, 0, 0, 60})
	if err != nil {
		log.Error("fail to write the alarming humidity to holding register")
		return err
	}
	log.Info("set the alarming humidity as 60")
	// set the accuracy of temperature measurement  0.5
	changeRegisterValue(s.InputRegisters, 0, ConvertFloat32ToBytes(0.5))
	log.Info("set the accuracy of temperature measurement as 0.5")

	// set the accuracy of humidity measure 3
	changeRegisterValue(s.InputRegisters, 2, ConvertInt8ToBytes(3))
	log.Info("set the accuracy of humidity measurement as 3")

	// set manufacture
	changeRegisterValue(s.InputRegisters, 3, []byte("huawei modbus simulator"))
	log.Info("set manufacture as huawei modbus simulator")

	return nil
}

func (t *thermometer) setValue(s *mbserver.Server, interval time.Duration) error {
	client := modbus.NewClient(t.handler)
	if err := setDefaultValue(s, client); err != nil {
		return err
	}
	ticker := time.NewTicker(interval)

	rand.Seed(time.Now().Unix())

	time1 := time.Now().Unix()

	for {
		switchOn, err := client.ReadCoils(0, 1)
		if err != nil {
			log.Error("fail to read switch value")
			return err
		}
		if switchOn[0] != 1 {
			log.Info("mocking has been stopped")
			select {
			case <-t.ctx.Done():
				return nil
			case <-ticker.C:
			}
			continue
		} else {
			log.Info("mocking is starting")
		}

		// randomly generated temperature value，range as -99 to 99
		temperature := rand.Float64()*198 - 99
		_, err = client.WriteMultipleRegisters(0, 4, ConvertFloat64ToBytes(temperature))
		if err != nil {
			return errors.Wrap(err, "fail to write temperature value into holding register")
		}
		log.Infof("temperature value is %f", temperature)
		// randomly generated humidity value， range as 5 to 95
		humidity := int32(rand.Intn(90) + 5)
		_, err = client.WriteMultipleRegisters(5, 2, ConvertInt32ToBytes(humidity))
		if err != nil {
			return errors.Wrap(err, "fail to write humidity value into holding register")
		}
		log.Infof("humidity value is %d", humidity)
		// whether to trigger high temperature alarm
		temperatureThreshold, err := client.ReadHoldingRegisters(4, 1)
		if err != nil {
			return errors.Wrap(err, "fail to read temperature alarm value")
		}
		log.Infof("the temperature threshold is %d.\n", ConvertBytesToInt16(temperatureThreshold))
		if err != nil {
			return errors.Wrap(err, "fail to read temperature alarm threshold")
		}
		if temperature > float64(ConvertBytesToInt16(temperatureThreshold)) {
			_, err = client.WriteSingleCoil(1, 0xff00)
			if err != nil {
				return errors.Wrap(err, "fail to write temperature alarm into coil register")
			}
			log.Info("add high temperature alarm ++")
		} else {
			_, err = client.WriteSingleCoil(1, 0x0000)
			if err != nil {
				return errors.Wrap(err, "fail to write temperature alarm into coil register")
			}
			log.Info("remove the high temperature alarm --")
		}
		// whether to trigger high humidity alarm
		humidityThreshold, err := client.ReadHoldingRegisters(7, 2)
		if err != nil {
			return errors.Wrap(err, "fail to read humidity alarm value")
		}
		log.Infof("the humidity threshold is %d.\n", humidityThreshold[1])
		if humidity > int32(humidityThreshold[1]) {
			s.DiscreteInputs[0] = 1
			log.Info("add high humidity alarm ++")
		} else {
			s.DiscreteInputs[0] = 0
			log.Info("remove high humidity alarm --")
		}
		// set battery value, drop 1% every second minutes
		time2 := time.Now().Unix()
		var battery = int32(100 - (time2-time1)/120)
		_, err = client.WriteMultipleRegisters(9, 2, ConvertInt32ToBytes(battery))
		if err != nil {
			return errors.Wrap(err, "fail to write to write battery into holding register")
		}
		log.Infof("set battery value as %d", battery)
		select {
		case <-t.ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
