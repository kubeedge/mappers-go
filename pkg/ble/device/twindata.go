package device

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/currantlabs/ble"
	"github.com/kubeedge/kubeedge/cloud/pkg/apis/devices/v1alpha2"
	"github.com/kubeedge/mappers-go/pkg/ble/configmap"
	"github.com/kubeedge/mappers-go/pkg/ble/driver"
	"github.com/kubeedge/mappers-go/pkg/ble/globals"
	"github.com/kubeedge/mappers-go/pkg/common"
	"k8s.io/klog/v2"
)

// TwinData is the timer structure for getting twin/data.
type TwinData struct {
	BluetoothClient        *driver.BluetoothClient
	Name                   string
	Type                   string
	BluetoothVisitorConfig configmap.BluetoothVisitorConfig
	Result                 string
	Topic                  string
}

// Run timer function.
func (td *TwinData) Run() {
	uuid := ble.MustParse(td.BluetoothVisitorConfig.CharacteristicUUID)
	if p, err := td.BluetoothClient.Client.DiscoverProfile(true); err == nil {
		if u := p.Find(ble.NewCharacteristic(uuid)); u != nil {
			c := u.(*ble.Characteristic)
			// If this Characteristic suports notifications and there's a CCCD
			// Then subscribe to it
			if (c.Property&ble.CharNotify) != 0 && c.CCCD != nil {
				if err := td.BluetoothClient.Client.Subscribe(c, false, td.handlerPublisher()); err != nil {
					klog.Error(err)
				}
			}
		}
	}
}

func (td *TwinData) handlerPublisher() func(req []byte) {
	return func(req []byte) {
		td.Result = fmt.Sprintf("%f", td.ConvertReadData(req))
		// construct payload
		var payload []byte
		var err error
		if strings.Contains(td.Topic, "$hw") {
			if payload, err = common.CreateMessageTwinUpdate(td.Name, td.Type, td.Result); err != nil {
				klog.Error("Create message twin update failed")
				return
			}
		} else {
			if payload, err = common.CreateMessageData(td.Name, td.Type, td.Result); err != nil {
				klog.Error("Create message data failed")
				return
			}
		}
		if err = globals.MqttClient.Publish(td.Topic, payload); err != nil {
			klog.Error(err)
		}

		klog.V(2).Infof("Update value: %s, topic: %s", td.Result, td.Topic)
	}
}

//ConvertReadData is the function responsible to convert the data read from the device into meaningful data
func (td *TwinData) ConvertReadData(data []byte) float64 {
	var intermediateResult uint64
	var initialValue []byte
	var initialStringValue = ""
	if td.BluetoothVisitorConfig.DataConvert.StartIndex <= td.BluetoothVisitorConfig.DataConvert.EndIndex {
		for index := td.BluetoothVisitorConfig.DataConvert.StartIndex; index <= td.BluetoothVisitorConfig.DataConvert.EndIndex; index++ {
			initialValue = append(initialValue, data[index])
		}
	} else {
		for index := td.BluetoothVisitorConfig.DataConvert.StartIndex; index >= td.BluetoothVisitorConfig.DataConvert.EndIndex; index-- {
			initialValue = append(initialValue, data[index])
		}
	}
	for _, value := range initialValue {
		initialStringValue = initialStringValue + strconv.Itoa(int(value))
	}
	initialByteValue, _ := strconv.ParseUint(initialStringValue, 10, 64)
	if td.BluetoothVisitorConfig.DataConvert.ShiftLeft != 0 {
		intermediateResult = initialByteValue << td.BluetoothVisitorConfig.DataConvert.ShiftLeft
	} else if td.BluetoothVisitorConfig.DataConvert.ShiftRight != 0 {
		intermediateResult = initialByteValue >> td.BluetoothVisitorConfig.DataConvert.ShiftRight
	} else {
		intermediateResult = initialByteValue
	}
	finalResult := float64(intermediateResult)
	for _, orderOfOperations := range td.BluetoothVisitorConfig.DataConvert.OrderOfOperations {
		if strings.ToUpper(orderOfOperations.OperationType) == strings.ToUpper(string(v1alpha2.BluetoothAdd)) {
			finalResult = finalResult + orderOfOperations.OperationValue
		} else if strings.ToUpper(orderOfOperations.OperationType) == strings.ToUpper(string(v1alpha2.BluetoothSubtract)) {
			finalResult = finalResult - orderOfOperations.OperationValue
		} else if strings.ToUpper(orderOfOperations.OperationType) == strings.ToUpper(string(v1alpha2.BluetoothMultiply)) {
			finalResult = finalResult * orderOfOperations.OperationValue
		} else if strings.ToUpper(orderOfOperations.OperationType) == strings.ToUpper(string(v1alpha2.BluetoothDivide)) {
			finalResult = finalResult / orderOfOperations.OperationValue
		}
	}
	return finalResult
}
