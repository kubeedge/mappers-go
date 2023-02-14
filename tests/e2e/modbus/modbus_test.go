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

package modbus_test

import (
	"github.com/kubeedge/mappers-go/tests/e2e/modbus/constants"
	"github.com/kubeedge/mappers-go/tests/e2e/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "k8s.io/klog/v2"
	"regexp"
	"strconv"
	"time"
)

const readHolding = "readholding"
const errorCode4 = "errorcode4"
const errorTimeout = "errortimeout"
const rtuReadHolding = "rtureadholding"
const writeInt16 = "writeint16"
const readTwoLines = 2
const readSixLines = 6
const readEightLines = 8

var _ = Describe("Modbus TCP Mapper test in E2E scenario", func() {

	Context("Testing TCP Modbus Mapper", func() {

		BeforeEach(func() {
			err := utils.StartDevices(constants.RunModbusTCPDevice, constants.GetModbusDeviceContainerID)
			if err != nil {
				log.Error("Fail to run tcp device")
				Expect(err).Should(BeNil())
			}
		})

		AfterEach(func() {
			err := utils.StopAndDeleteMapper(constants.GetModbusMapperContainerID)
			if err != nil {
				log.Error("Fail to stop and delete modbus mapper container")
				Expect(err).Should(BeNil())
			}
			err = utils.StopAndDeleteDevice(constants.GetModbusDeviceContainerID)
			if err != nil {
				log.Error("Fail to stop ans delete modbus device container")
				Expect(err).Should(BeNil())
			}
		})

		It("test about writing data into holding register - int16", func() {
			// run mapper container
			err := utils.RunMapper(constants.RunTCPModbusMapper, writeInt16, constants.GetModbusMapperContainerID)
			if err != nil {
				log.Error("Fail to run mapper ")
			}
			time.Sleep(5 * time.Second)

			containerLog, _ := utils.ReadDockerLog(constants.GetModbusMapperContainerID, readTwoLines)
			reg1 := regexp.MustCompile(`Get the alarming-temperature value as ([0-9.-]+)`)
			temperatureValue := reg1.FindStringSubmatch(containerLog)

			Expect(temperatureValue[1]).Should(Equal(strconv.Itoa(30)))
		})

		It("test about reading holding value", func() {

			err := utils.RunMapper(constants.RunTCPModbusMapper, readHolding, constants.GetModbusMapperContainerID)
			if err != nil {
				log.Error("Fail to run mapper ")
			}
			time.Sleep(5 * time.Second)

			containerLog, _ := utils.ReadDockerLog(constants.GetModbusMapperContainerID, readTwoLines)
			reg1 := regexp.MustCompile(`Get the temperature value as ([0-9.-]+)`)
			temperatureValue := reg1.FindStringSubmatch(containerLog)

			deviceLog, _ := utils.ReadDockerLog(constants.GetModbusDeviceContainerID, readEightLines)
			reg2 := regexp.MustCompile(`temperature value is ([0-9.-]+)`)
			temperatureValue2 := reg2.FindStringSubmatch(deviceLog)

			Expect(temperatureValue[1]).Should(Equal(temperatureValue2[1]))
		})

	})

	Context("Testing Tcp Modbus Mapper In Negative Cases", func() {
		BeforeEach(func() {
			err := utils.StartDevices(constants.RunModbusTCPErrorDevice, constants.GetModbusDeviceContainerID)
			if err != nil {
				log.Error("Fail to run tcp device")
				Expect(err).Should(BeNil())
			}

		})

		AfterEach(func() {
			err := utils.StopAndDeleteMapper(constants.GetModbusMapperContainerID)
			if err != nil {
				log.Error("Fail to stop and delete modbus mapper container")
				Expect(err).Should(BeNil())
			}
			err = utils.StopAndDeleteDevice(constants.GetModbusDeviceContainerID)
			if err != nil {
				log.Error("Fail to stop ans delete modbus device container")
				Expect(err).Should(BeNil())
			}

		})

		It("test negative case about modbus tcp error code 4 ", func() {
			err := utils.RunMapper(constants.RunTCPModbusMapper, errorCode4, constants.GetModbusMapperContainerID)
			if err != nil {
				log.Error("Fail to run mapper ")
			}
			time.Sleep(5 * time.Second)

			containerLog, err := utils.ReadDockerLog(constants.GetModbusMapperContainerID, readTwoLines)
			if err != nil {
				log.Error("Fail to read container log")
			}
			reg1 := regexp.MustCompile(`exception.*4`)
			result := reg1.FindStringSubmatch(containerLog)
			Expect(len(result)).Should(Equal(1))
		})

		It("test negative case about modbus tcp timeout error", func() {
			err := utils.RunMapper(constants.RunTCPModbusMapper, errorTimeout, constants.GetModbusMapperContainerID)
			if err != nil {
				log.Error("Fail to run mapper ")
			}
			time.Sleep(10 * time.Second)

			containerLog, err := utils.ReadDockerLog(constants.GetModbusMapperContainerID, readSixLines)
			if err != nil {
				log.Error("Fail to read container log")
			}
			reg1 := regexp.MustCompile(`timeout`)
			result := reg1.FindStringSubmatch(containerLog)
			Expect(len(result)).Should(Equal(1))
		})

	})
})

var _ = Describe("Modbus RTU Mapper test in E2E scenario", func() {

	Context("Testing TCP Modbus Mapper", func() {
		BeforeEach(func() {
			err := utils.StartDevices(constants.RunSocat, constants.GetSocat)
			if err != nil {
				log.Error("Fail to run socat")
				Expect(err).Should(BeNil())
			}
			err = utils.StartDevices(constants.RunModbusRTUDevice, constants.GetModbusDeviceContainerID)
			if err != nil {
				log.Error("Fail to run rtu device")
				Expect(err).Should(BeNil())
			}
		})

		AfterEach(func() {
			err := utils.StopAndDeleteMapper(constants.GetModbusMapperContainerID)
			if err != nil {
				log.Error("Fail to stop and delete modbus mapper container")
				Expect(err).Should(BeNil())
			}
			err = utils.StopAndDeleteDevice(constants.GetModbusDeviceContainerID)
			if err != nil {
				log.Error("Fail to stop and delete modbus device container")
				Expect(err).Should(BeNil())
			}
			err = utils.StopAndDeleteDevice(constants.GetSocat)
			if err != nil {
				log.Error("Fail to stop and delete socat container")
			}
		})

		It("test about reading holding value", func() {

			err := utils.RunMapper(constants.RunRTUModbusMapper, rtuReadHolding, constants.GetModbusMapperContainerID)
			if err != nil {
				log.Error("Fail to run mapper")
			}
			time.Sleep(5 * time.Second)

			containerLog, _ := utils.ReadDockerLog(constants.GetModbusMapperContainerID, readTwoLines)
			reg1 := regexp.MustCompile(`Get the temperature value as ([0-9.-]+)`)
			temperatureValue := reg1.FindStringSubmatch(containerLog)

			deviceLog, _ := utils.ReadDockerLog(constants.GetModbusDeviceContainerID, readEightLines)
			reg2 := regexp.MustCompile(`temperature value is ([0-9.-]+)`)
			temperatureValue2 := reg2.FindStringSubmatch(deviceLog)

			Expect(temperatureValue[1]).Should(Equal(temperatureValue2[1]))
		})
	})
})
