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
	log "k8s.io/klog"
	"testing"
)

func TestModbus(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Modbus Suite")
}

var _ = BeforeSuite(func() {
	log.Info("Before Suiting Execution")
	err := utils.MakeMapperImages(constants.MakeModbusMapper, constants.CheckModbusMapperImage)
	if err != nil {
		log.Error("Fail to run make mapper image")
		Expect(err).Should(BeNil())
	}
	err = utils.MakeDeviceImage(constants.MakeModbusDevice, constants.CheckModbusDeviceImage)
	if err != nil {
		log.Error("Fail to make device image")
		Expect(err).Should(BeNil())
	}
})

var _ = AfterSuite(func() {
	log.Info("After Suiting Executing")
	err := utils.DeleteMapperImage(constants.DeleteModbusMapperImage)
	if err != nil {
		log.Error("Fail to delete mapper image")
		Expect(err).Should(BeNil())
	}
	err = utils.DeleteDeviceImage(constants.DeleteModbusDevice)
	if err != nil {
		log.Error("Fail to delete device image")
		Expect(err).Should(BeNil())
	}
})
