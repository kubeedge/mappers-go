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
	"github.com/tbrandon/mbserver"
	"time"
)

const readDiscreteRegisterCode = 2
const writeHoldingRegisterCode = 6

func ReadDiscreteRegisterError(s *mbserver.Server, frame mbserver.Framer) ([]byte, *mbserver.Exception) {
	return frame.GetData()[0:4], &mbserver.SlaveDeviceFailure
}

func WriteHoldingRegisterError(s *mbserver.Server, frame mbserver.Framer) ([]byte, *mbserver.Exception) {
	time.Sleep(10 * time.Second)
	return []byte{}, &mbserver.Success
}

func InsertServerError(s *mbserver.Server) {
	s.RegisterFunctionHandler(readDiscreteRegisterCode, ReadDiscreteRegisterError)
	s.RegisterFunctionHandler(writeHoldingRegisterCode, WriteHoldingRegisterError)
}
