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

package opcua_test

import (
	"github.com/kubeedge/mappers-go/tests/e2e/opcua/constants"
	"github.com/kubeedge/mappers-go/tests/e2e/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "k8s.io/klog/v2"
	"regexp"
	"strconv"
	"time"
)

const readFloat = "readfloat"
const writeInt = "writeint"
const readOneLine = 1
const readTwoLines = 2
const readSixLines = 6

var _ = Describe("Opcua", func() {
	Context("Testing Opcua Mapper", func() {

		BeforeEach(func() {
			err := utils.StartDevices(constants.RunOpcuaDevice, constants.GetOpcuaDeviceContainerID)
			if err != nil {
				log.Error("Fail to run tcp device")
				Expect(err).Should(BeNil())
			}
			time.Sleep(1 * time.Second)
		})
		AfterEach(func() {
			err := utils.StopAndDeleteMapper(constants.GetOpcuaMapperContainerID)
			if err != nil {
				log.Error("Fail to stop and delete opcua mapper container")
				Expect(err).Should(BeNil())
			}
			err = utils.StopAndDeleteDevice(constants.GetOpcuaDeviceContainerID)
			if err != nil {
				log.Error("Fail to stop ans delete opcua device container")
				Expect(err).Should(BeNil())
			}
		})


		It("test about reading float value", func() {

			err := utils.RunMapper(constants.RunOpcuaMapper, readFloat, constants.GetOpcuaMapperContainerID)
			if err != nil {
				log.Error("Fail to run mapper ")
			}
			time.Sleep(5 * time.Second)

			containerLog, _ := utils.ReadDockerLog(constants.GetOpcuaMapperContainerID, readOneLine)
			reg1 := regexp.MustCompile(` Update value: ([0-9.-E+/-]+)`)
			temperatureValue := reg1.FindStringSubmatch(containerLog)

			deviceLog, _ := utils.ReadDockerLog(constants.GetOpcuaDeviceContainerID, readSixLines)
			reg2 := regexp.MustCompile(`temperature value is  ([0-9.-]+)`)
			temperatureValue2 := reg2.FindStringSubmatch(deviceLog)
			v, _ := strconv.ParseFloat(temperatureValue2[1], 64)
			num := strconv.FormatFloat(v, 'E', -1, 32)

			Expect(temperatureValue[1]).Should(Equal(num))
		})
		It("test case about writing int data into device", func() {
			err := utils.RunMapper(constants.RunOpcuaMapper, writeInt, constants.GetOpcuaMapperContainerID)
			if err != nil {
				log.Error("Fail to run mapper", err)
			}
			time.Sleep(5 * time.Second)
			containerLog, _ := utils.ReadDockerLog(constants.GetOpcuaMapperContainerID, readOneLine)
			reg1 := regexp.MustCompile(`Update value: ([0-9]+)`)
			temperatureThreshold := reg1.FindStringSubmatch(containerLog)
			expectTemperature := "50"
			Expect(temperatureThreshold[1]).Should(Equal(expectTemperature))
		})

	})

	Context("Test Case About Opcua Device Disconnect", func() {

		BeforeEach(func() {
			err := utils.StartDevices(constants.RunOpcuaDevice, constants.GetOpcuaDeviceContainerID)
			if err != nil {
				log.Error("Fail to run tcp device")
				Expect(err).Should(BeNil())
			}
			time.Sleep(1 * time.Second)
		})
		AfterEach(func() {
			err := utils.StopAndDeleteMapper(constants.GetOpcuaMapperContainerID)
			if err != nil {
				log.Error("Fail to stop and delete opcua mapper container")
				Expect(err).Should(BeNil())
			}

		})

		It("Test Case About Opcua Device Disconnect", func() {
			err := utils.RunMapper(constants.RunOpcuaMapper, readFloat, constants.GetOpcuaMapperContainerID)
			if err != nil {
				log.Error("Fail to run mapper ")
			}
			time.Sleep(2 * time.Second)
			err = utils.StopAndDeleteDevice(constants.GetOpcuaDeviceContainerID)
			if err != nil {
				log.Error("Fail to stop ans delete modbus device container")
				Expect(err).Should(BeNil())
			}
			time.Sleep(15 * time.Second)
			containerLog, _ := utils.ReadDockerLog(constants.GetOpcuaMapperContainerID, readTwoLines)
			target := "Get register failed"
			reg1 := regexp.MustCompile(target)
			s := reg1.FindString(containerLog)
			Expect(s).Should(Equal(target))
		})
	})
})
