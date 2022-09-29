/*
Copyright 2020 The KubeEdge Authors.

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

// This application needs physical devices.
// Please edit by demand for testing.

package modbus

import (
	"fmt"
	"os"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/json"

	"k8s.io/utils/pointer"

	v12 "k8s.io/api/core/v1"

	"github.com/kubeedge/kubeedge/pkg/apis/devices/v1alpha2"
	"github.com/kubeedge/mappers-go/pkg/util/parse"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestName(t *testing.T) {
	res := parse.DeviceData{
		Device: &v1alpha2.Device{
			TypeMeta: v1.TypeMeta{
				Kind:       "Device",
				APIVersion: "devices.kubeedge.io/v1alpha2",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      "iyrqt090eh4u6vn68tbr",
				Namespace: "dec-public",
			},
			Spec: v1alpha2.DeviceSpec{
				DeviceModelRef: &v12.LocalObjectReference{Name: "iyrqt090eh4u6vn68tbr"},
				Protocol: v1alpha2.ProtocolConfig{
					Modbus: &v1alpha2.ProtocolConfigModbus{
						SlaveID: pointer.Int64(1),
					},
					Common: &v1alpha2.ProtocolConfigCommon{
						COM: &v1alpha2.ProtocolConfigCOM{
							SerialPort: "/dev/ttyUSB1",
							BaudRate:   9600,
							DataBits:   8,
							Parity:     "none",
							StopBits:   1,
						},
					},
				},
				PropertyVisitors: []v1alpha2.DevicePropertyVisitor{
					{
						PropertyName: "twin11",
						ReportCycle:  1000000000,
						CollectCycle: 1000000000,
						VisitorConfig: v1alpha2.VisitorConfig{
							Modbus: &v1alpha2.VisitorConfigModbus{
								Register:       "InputRegister",
								Offset:         pointer.Int64(2),
								Limit:          pointer.Int64(1),
								Scale:          1,
								IsSwap:         true,
								IsRegisterSwap: true,
							},
						},
					},
				},
				Data:         v1alpha2.DeviceData{},
				NodeSelector: nil,
			},
			Status: v1alpha2.DeviceStatus{},
		},
		DeviceModel: &v1alpha2.DeviceModel{
			TypeMeta: v1.TypeMeta{
				Kind:       "DeviceModel",
				APIVersion: "devices.kubeedge.io/v1alpha2",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      "iyrqt090eh4u6vn68tbr",
				Namespace: "dec-public",
			},
			Spec: v1alpha2.DeviceModelSpec{
				Properties: []v1alpha2.DeviceProperty{{
					Name:        "twin11",
					Description: "",
					Type:        v1alpha2.PropertyType{},
				}},
			},
		},
	}
	marshal, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(marshal))
}

func tdriver() {
	var modbusrtu ModbusRTU

	modbusrtu.SerialName = "/dev/ttyS0"
	modbusrtu.BaudRate = 9600
	modbusrtu.DataBits = 8
	modbusrtu.StopBits = 1
	modbusrtu.SlaveID = 1
	modbusrtu.Parity = "N"
	modbusrtu.Timeout = 2 * time.Second

	client, err := NewClient(modbusrtu)
	if err != nil {
		fmt.Println("New client error")
		os.Exit(1)
	}

	results, err := client.Set("DiscreteInputRegister", 2, 1)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(results)
	results, err = client.Set("CoilRegister", 2, 1)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(results)
	os.Exit(0)
}

func main() {
	/*
		var modbustcp ModbusTCP

		modbustcp.DeviceIp = "192.168.56.1"
		modbustcp.TcpPort = "502"
		modbustcp.SlaveId = 0x1
		client := NewClient(modbustcp)
		if client == nil {
			fmt.Println("New client error")
			os.Exit(1)
		}
		fmt.Println("status: ", client.GetStatus())

		results, err := client.Client.ReadDiscreteInputs(0, 1)
		if err != nil {
			fmt.Println("Read error: ", err)
			os.Exit(1)
		}
		fmt.Println("result: ", results)
	*/
	tdriver()
	os.Exit(0)
}
