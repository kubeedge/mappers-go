package configmap

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"strings"

	"k8s.io/klog/v2"

	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/common"
)

// Parse is a method to parse the configmap.
func Parse(path string,
	devices map[string]*DeviceInstance,
	dms map[string]*DeviceModel,
	protocols map[string]*Protocol,
	serviceProtocolName string) error {
	var deviceProfile DeviceProfile
	jsonFile, err := ioutil.ReadFile(path)
	if err != nil {
		err = errors.New("failed to read " + path + " file")
		return err
	}
	//Parse the JSON file and convert it into the data structure of DeviceProfile
	if err = json.Unmarshal(jsonFile, &deviceProfile); err != nil {
		return err
	}
	// loop instIndex : judge whether the configmap definition is correct, and initialize the device instance
	for instIndex := 0; instIndex < len(deviceProfile.DeviceInstances); instIndex++ {
		protocolAssert := true
		instance := deviceProfile.DeviceInstances[instIndex]
		// loop protoIndex : judge whether the device's protocol is correct, and initialize the device protocol
		protoIndex := 0
		for protoIndex = 0; protoIndex < len(deviceProfile.Protocols); protoIndex++ {
			if instance.ProtocolName == deviceProfile.Protocols[protoIndex].Name {
				// Verify that the protocols match
				protocolConfig := make(map[string]interface{})
				err := json.Unmarshal(deviceProfile.Protocols[protoIndex].ProtocolConfigs, &protocolConfig)
				if err != nil {
					err = errors.New("failed to parse " + deviceProfile.Protocols[protoIndex].Name)
					return err
				}
				// If the protocol does not match, the device,model,protocol will not be added
				if strings.ToUpper(protocolConfig["protocolName"].(string)) != strings.ToUpper(serviceProtocolName) {
					klog.Errorf("Failed to add %s , because protocolName should be %s",instance.ID,protocolConfig["protocolName"])
					protocolAssert = false
				}else{
					protocols[deviceProfile.Protocols[protoIndex].Name] = new(Protocol)
					protocols[deviceProfile.Protocols[protoIndex].Name] = &deviceProfile.Protocols[protoIndex]
					instance.PProtocol = deviceProfile.Protocols[protoIndex]
				}
				break
			}
		}
		// The protocol does not match, so the device,model,protocol can not be added
		if !protocolAssert {
			continue
		}
		if protoIndex == len(deviceProfile.Protocols) {
			err = errors.New("protocol mismatch")
			return err
		}
		// loop propertyIndex : find the device model's properties for each device instance's propertyVisitor
		for propertyIndex := 0; propertyIndex < len(instance.PropertyVisitors); propertyIndex++ {
			modelName := instance.PropertyVisitors[propertyIndex].ModelName
			propertyName := instance.PropertyVisitors[propertyIndex].PropertyName
			modelIndex := 0
			// loop modelIndex : find a matching device model, and initialize the device model
			for modelIndex = 0; modelIndex < len(deviceProfile.DeviceModels); modelIndex++ {
				if modelName == deviceProfile.DeviceModels[modelIndex].Name {
					dms[deviceProfile.DeviceModels[modelIndex].Name] = new(DeviceModel)
					dms[deviceProfile.DeviceModels[modelIndex].Name] = &deviceProfile.DeviceModels[modelIndex]
					m := 0
					// loop m :  find a matching device model's properties
					for m = 0; m < len(deviceProfile.DeviceModels[modelIndex].Properties); m++ {
						if propertyName == deviceProfile.DeviceModels[modelIndex].Properties[m].Name {
							instance.PropertyVisitors[propertyIndex].PProperty = deviceProfile.DeviceModels[modelIndex].Properties[m]
							break
						}
					}
					if m == len(deviceProfile.DeviceModels[modelIndex].Properties) {
						err = errors.New("property mismatch")
						return err
					}
					break
				}
			}
			if modelIndex == len(deviceProfile.DeviceModels) {
				err = errors.New("device model mismatch")
				return err
			}
		}
		// loop propertyIndex : find propertyVisitors for each instance's twin
		for propertyIndex := 0; propertyIndex < len(instance.Twins); propertyIndex++ {
			name := instance.Twins[propertyIndex].PropertyName
			l := 0
			// loop l : find a matching propertyName
			for l = 0; l < len(instance.PropertyVisitors); l++ {
				if name == instance.PropertyVisitors[l].PropertyName {
					instance.Twins[propertyIndex].PVisitor = &instance.PropertyVisitors[l]
					break
				}
			}
			if l == len(instance.PropertyVisitors) {
				err = errors.New("propertyVisitor mismatch")
				return err
			}
		}
		// loop propertyIndex : find propertyVisitors for each instance's property
		for propertyIndex := 0; propertyIndex < len(instance.Datas.Properties); propertyIndex++ {
			name := instance.Datas.Properties[propertyIndex].PropertyName
			l := 0
			// loop l : find a matching propertyName
			for l = 0; l < len(instance.PropertyVisitors); l++ {
				if name == instance.PropertyVisitors[l].PropertyName {
					instance.Datas.Properties[propertyIndex].PVisitor = &instance.PropertyVisitors[l]
					break
				}
			}
			if l == len(instance.PropertyVisitors) {
				err = errors.New("propertyVisitor mismatch")
				return err
			}
		}
		devices[instance.ID] = new(DeviceInstance)
		devices[instance.ID] = &instance
		klog.V(4).Infof("Instance:%s Successfully registered", instance.ID)
	}
	return nil
}

// GetConnectInfo is a method to generate link information for each device's property
func GetConnectInfo(
	devices map[string]*DeviceInstance,
	connectInfo map[string]*ConnectInfo) {
	for id, instance := range devices {
		tempID := id
		tempInstance := instance
		for _, visitorV := range tempInstance.PropertyVisitors {
			tempVisitorV := visitorV
			driverName := common.DriverPrefix + tempID + visitorV.PropertyName
			connectInfo[driverName] = &ConnectInfo{
				ProtocolCommonConfig: tempInstance.PProtocol.ProtocolCommonConfig,
				VisitorConfig:        tempVisitorV.VisitorConfig,
				ProtocolConfig:       tempInstance.PProtocol.ProtocolConfigs,
			}
		}
	}
}

// ParseOdd is a method to parse the configmap.
func ParseOdd(path string,
	devices map[string]*DeviceInstance,
	dms map[string]*DeviceModel,
	protocols map[string]*Protocol,
	id string) error {
	var deviceProfile DeviceProfile
	jsonFile, err := ioutil.ReadFile(path)
	if err != nil {
		err = errors.New("failed to read " + path + " file")
		return err
	}
	//Parse the JSON file and convert it into the data structure of DeviceProfile
	if err = json.Unmarshal(jsonFile, &deviceProfile); err != nil {
		return err
	}
	for i := 0; i < len(deviceProfile.DeviceInstances); i++ {
		instance := deviceProfile.DeviceInstances[i]
		if instance.ID == id {
			j := 0
			for j = 0; j < len(deviceProfile.Protocols); j++ {
				if instance.ProtocolName == deviceProfile.Protocols[j].Name {
					instance.PProtocol = deviceProfile.Protocols[j]
					break
				}
			}
			if j == len(deviceProfile.Protocols) {
				err = errors.New("protocol not found")
				return err
			}
			for k := 0; k < len(instance.PropertyVisitors); k++ {
				modelName := instance.PropertyVisitors[k].ModelName
				propertyName := instance.PropertyVisitors[k].PropertyName
				l := 0
				for l = 0; l < len(deviceProfile.DeviceModels); l++ {
					if modelName == deviceProfile.DeviceModels[l].Name {
						m := 0
						for m = 0; m < len(deviceProfile.DeviceModels[l].Properties); m++ {
							if propertyName == deviceProfile.DeviceModels[l].Properties[m].Name {
								instance.PropertyVisitors[k].PProperty = deviceProfile.DeviceModels[l].Properties[m]
								break
							}
						}
						if m == len(deviceProfile.DeviceModels[l].Properties) {
							err = errors.New("property not found")
							return err
						}
						break
					}
				}
				if l == len(deviceProfile.DeviceModels) {
					err = errors.New("device model not found")
					return err
				}
			}
			for k := 0; k < len(instance.Twins); k++ {
				name := instance.Twins[k].PropertyName
				l := 0
				for l = 0; l < len(instance.PropertyVisitors); l++ {
					if name == instance.PropertyVisitors[l].PropertyName {
						instance.Twins[k].PVisitor = &instance.PropertyVisitors[l]
						break
					}
				}
				if l == len(instance.PropertyVisitors) {
					return errors.New("propertyVisitor not found")
				}
			}
			for k := 0; k < len(instance.Datas.Properties); k++ {
				name := instance.Datas.Properties[k].PropertyName
				l := 0
				for l = 0; l < len(instance.PropertyVisitors); l++ {
					if name == instance.PropertyVisitors[l].PropertyName {
						instance.Datas.Properties[k].PVisitor = &instance.PropertyVisitors[l]
						break
					}
				}
				if l == len(instance.PropertyVisitors) {
					return errors.New("propertyVisitor mismatch")
				}
			}
			if _, ok := devices[instance.ID]; !ok {
				devices[instance.ID] = new(DeviceInstance)
				devices[instance.ID] = &instance
			} else {
				return errors.New(instance.ID + " already in the device list")
			}
			for i := 0; i < len(deviceProfile.DeviceModels); i++ {
				if _, ok := dms[deviceProfile.DeviceModels[i].Name]; !ok {
					dms[deviceProfile.DeviceModels[i].Name] = &deviceProfile.DeviceModels[i]
				}
			}
			for i := 0; i < len(deviceProfile.Protocols); i++ {
				protocols[deviceProfile.Protocols[i].Name] = &deviceProfile.Protocols[i]
			}
			return nil
		}
	}

	return errors.New("can't find the device in profile.json")
}
