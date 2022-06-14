// Package controller used to call driver interface to read or write device
package controller

import (
	"k8s.io/klog/v2"

	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/common"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/configmap"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/instancepool"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/pkg/di"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/pkg/models"
)

// SetVisitor write device in thread safe mode
func SetVisitor(instanceID string, twin configmap.Twin, drivers models.ProtocolDriver, mutex *common.Lock, dic *di.Container) error {
	mutex.Lock()
	defer mutex.Unlock()
	deviceIndex := common.DriverPrefix + instanceID + twin.PropertyName
	if twin.PVisitor.PProperty.AccessMode == "ReadOnly" {
		klog.V(4).Info(instanceID + ":" + twin.PropertyName + " is ReadOnly")
		return nil
	}
	if len(twin.Desired.Value) == 0 {
		return nil
	}
	value, err := common.Convert(twin.PVisitor.PProperty.DataType, twin.Desired.Value)
	if err != nil {
		klog.Errorf("Failed to convert value as %s : %v", twin.PVisitor.PProperty.DataType, err)
		return err
	}
	connectInfo := instancepool.ConnectInfoNameFrom(dic.Get)
	err = drivers.WriteDeviceData(value, connectInfo[deviceIndex].ProtocolCommonConfig, connectInfo[deviceIndex].VisitorConfig, connectInfo[deviceIndex].ProtocolConfig)
	if err != nil {
		klog.Errorf("Failed to set %s config: %v", instanceID, err)
		return err
	}
	return nil
}

// GetDeviceData red device data in thread safe mode
func GetDeviceData(instanceID string, twin configmap.Twin, drivers models.ProtocolDriver, mutex *common.Lock, dic *di.Container) (string, error) {
	mutex.Lock()
	defer mutex.Unlock()
	deviceIndex := common.DriverPrefix + instanceID + twin.PropertyName
	connectInfo := instancepool.ConnectInfoNameFrom(dic.Get)
	data, err := drivers.ReadDeviceData(connectInfo[deviceIndex].ProtocolCommonConfig, connectInfo[deviceIndex].VisitorConfig, connectInfo[deviceIndex].ProtocolConfig)
	if err != nil {
		klog.Errorf("Failed to get %s data: %v", instanceID, err)
		return "", err
	}
	sData, err := common.ConvertToString(data)
	if err != nil {
		klog.Errorf("Failed to convert %s %s value as string : %v",instanceID,twin.PropertyName,err)
		return "", err
	}
	if len(sData) > 30{
		klog.V(4).Infof("Get %s : %s ,value is %s......", instanceID, twin.PropertyName, sData[:30])
	}else{
		klog.V(4).Infof("Get %s : %s ,value is %s", instanceID, twin.PropertyName, sData)
	}
	return sData, nil
}

// GetDeviceStatus red device status in thread safe mode
func GetDeviceStatus(instanceID string, twin configmap.Twin, drivers models.ProtocolDriver, mutex *common.Lock, dic *di.Container) string {
	mutex.Lock()
	defer mutex.Unlock()
	deviceIndex := common.DriverPrefix + instanceID + twin.PropertyName
	connectInfo := instancepool.ConnectInfoNameFrom(dic.Get)
	status := drivers.GetDeviceStatus(connectInfo[deviceIndex].ProtocolCommonConfig, connectInfo[deviceIndex].VisitorConfig, connectInfo[deviceIndex].ProtocolConfig)
	if status {
		return common.DEVSTOK
	}
	return common.DEVSTDISCONN
}

// InitDeviceConfig init device when the mapper first run
func InitDeviceConfig(drivers models.ProtocolDriver, dic *di.Container){
	protocolInfo := instancepool.ProtocolNameFrom(dic.Get)
	for _, v := range protocolInfo {
		err := drivers.InitDevice(v.ProtocolCommonConfig)
		if err != nil {
			klog.Errorf("Instance init failed: %s\n",err)
		}
	}
}
