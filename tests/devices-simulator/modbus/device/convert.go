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
	"encoding/binary"
	"math"
)

func ConvertInt8ToBytes(i int8) []byte {
	res := make([]byte, 2)
	binary.BigEndian.PutUint16(res, uint16(i))
	return res
}

func ConvertInt16ToBytes(i int16) []byte {
	res := make([]byte, 2)
	binary.BigEndian.PutUint16(res, uint16(i))
	return res
}

func ConvertInt32ToBytes(i int32) []byte {
	res := make([]byte, 4)
	binary.BigEndian.PutUint32(res, uint32(i))
	return res
}

func ConvertInt64ToBytes(i int64) []byte {
	res := make([]byte, 8)
	binary.BigEndian.PutUint64(res, uint64(i))
	return res
}

func ConvertFloat32ToBytes(f float32) []byte {
	res := make([]byte, 4)
	binary.BigEndian.PutUint32(res, math.Float32bits(f))
	return res
}

func ConvertFloat64ToBytes(f float64) []byte {
	res := make([]byte, 8)
	binary.BigEndian.PutUint64(res, math.Float64bits(f))
	return res
}

func ConvertBytesToInt8(bytes []byte) int8 {
	return int8(binary.BigEndian.Uint16(bytes))
}

func ConvertBytesToInt16(bytes []byte) int16 {
	return int16(binary.BigEndian.Uint16(bytes))
}

func ConvertBytesToInt32(bytes []byte) int32 {
	return int32(binary.BigEndian.Uint32(bytes))
}

func ConvertBytesToInt64(bytes []byte) int64 {
	return int64(binary.BigEndian.Uint64(bytes))
}

func ConvertBytesToFloat32(bytes []byte) float32 {
	b := binary.BigEndian.Uint32(bytes)
	return math.Float32frombits(b)
}

func ConvertBytesToFloat64(bytes []byte) float64 {
	b := binary.BigEndian.Uint64(bytes)
	return math.Float64frombits(b)
}
