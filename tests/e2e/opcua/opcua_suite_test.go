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
	log "k8s.io/klog"
	"testing"
)

func TestOpcua(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Opcua Suite")
}

var _ = BeforeSuite(func() {
	log.Info("Before Suiting Execution")
	err := utils.CopyDirectory(constants.SourceCaPath, constants.DstCaPath)
	Expect(err).Should(BeNil())
	err = utils.MakeMapperImages(constants.MakeOpcuaMapper, constants.CheckOpcuaMapperImage)
	if err != nil {
		log.Error("Fail to run make opcua mapper image")
		Expect(err).Should(BeNil())
	}
	err = utils.MakeDeviceImage(constants.MakeOpcuaDevice, constants.CheckOpcuaDeviceImage)
	if err != nil {
		log.Error("Fail to run make opcua device image")
		Expect(err).Should(BeNil())
	}
})

var _ = AfterSuite(func() {
	log.Info("After Suiting Executing")
	err := utils.RmDirectory(constants.DstCaPath)
	Expect(err).Should(BeNil())
	err = utils.DeleteMapperImage(constants.DeleteOpcuaMapperImage)
	if err != nil {
		log.Error("Fail to delete opcua mapper image")
		Expect(err).Should(BeNil())
	}
	err = utils.DeleteDeviceImage(constants.DeleteOpcuaDevice)
	if err != nil {
		log.Error("Fail to delete opcua device image")
		Expect(err).Should(BeNil())
	}
})
