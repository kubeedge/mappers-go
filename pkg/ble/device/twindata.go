package device

import (
	"fmt"
	"github.com/currantlabs/ble"
	"github.com/kubeedge/kubeedge/cloud/pkg/apis/devices/v1alpha2"
	"github.com/kubeedge/mappers-go/pkg/ble/configmap"
	"github.com/kubeedge/mappers-go/pkg/ble/driver"
	"github.com/kubeedge/mappers-go/pkg/ble/globals"
	"github.com/kubeedge/mappers-go/pkg/common"
	"k8s.io/klog/v2"
	"strconv"
	"strings"
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
	klog.V(2).Infof("[Device:sub:%s] ", uuid.String())
	if p, err := td.BluetoothClient.Client.DiscoverProfile(true); err == nil {
		if u := p.Find(ble.NewCharacteristic(uuid)); u != nil {
			c := u.(*ble.Characteristic)
			// If this Characteristic suports notifications and there's a CCCD
			// Then subscribe to it
			if (c.Property&ble.CharNotify) != 0 && c.CCCD != nil {
				if err := td.BluetoothClient.Client.Subscribe(c, false, td.handlerPublisher(td.BluetoothVisitorConfig.DataConvert)); err != nil {
					klog.Error(err)
				}
			}
		}
	}
}

func (td *TwinData) handlerPublisher(dataConvert configmap.DataConvert) func(req []byte) {
	return func(req []byte) {
		td.Result = fmt.Sprintf("%f", ConvertReadData(req, dataConvert))
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
func ConvertReadData(data []byte, dataConvert configmap.DataConvert) float64 {
	var intermediateResult uint64
	var initialValue []byte
	var initialStringValue = ""
	if dataConvert.StartIndex <= dataConvert.EndIndex {
		for index := dataConvert.StartIndex; index <= dataConvert.EndIndex; index++ {
			initialValue = append(initialValue, data[index])
		}
	} else {
		for index := dataConvert.StartIndex; index >= dataConvert.EndIndex; index-- {
			initialValue = append(initialValue, data[index])
		}
	}
	for _, value := range initialValue {
		initialStringValue = initialStringValue + strconv.Itoa(int(value))
	}
	initialByteValue, _ := strconv.ParseUint(initialStringValue, 16, 16)

	if dataConvert.ShiftLeft != 0 {
		intermediateResult = initialByteValue << dataConvert.ShiftLeft
	} else if dataConvert.ShiftRight != 0 {
		intermediateResult = initialByteValue >> dataConvert.ShiftRight
	}
	finalResult := float64(intermediateResult)
	for _, orderOfOperations := range dataConvert.OrderOfOperations {
		if orderOfOperations.OperationType == strings.ToUpper(string(v1alpha2.BluetoothAdd)) {
			finalResult = finalResult + orderOfOperations.OperationValue
		} else if orderOfOperations.OperationType == strings.ToUpper(string(v1alpha2.BluetoothSubtract)) {
			finalResult = finalResult - orderOfOperations.OperationValue
		} else if orderOfOperations.OperationType == strings.ToUpper(string(v1alpha2.BluetoothMultiply)) {
			finalResult = finalResult * orderOfOperations.OperationValue
		} else if orderOfOperations.OperationType == strings.ToUpper(string(v1alpha2.BluetoothDivide)) {
			finalResult = finalResult / orderOfOperations.OperationValue
		}
	}
	return finalResult
}
