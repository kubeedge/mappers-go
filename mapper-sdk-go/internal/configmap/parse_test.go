package configmap

import (
	"testing"
)

// TestConfigMap get ./configmap_test.json to parse
func TestConfigMap(t *testing.T) {
	var devices map[string]*DeviceInstance
	var models map[string]*DeviceModel
	var protocols map[string]*Protocol

	devices = make(map[string]*DeviceInstance)
	models = make(map[string]*DeviceModel)
	protocols = make(map[string]*Protocol)
	err := Parse("./configmap_test.json", devices, models, protocols,"protocolName")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
}

//TestConfigmapNeg get ./configmap_negtest.json to parse.
//Because the configmap_negtest.json is invalid. Test will fail, and log the reason for error
//func TestConfigmapNeg(t *testing.T) {
//	var devices map[string]*DeviceInstance
//	var models map[string]*DeviceModel
//	var protocols map[string]*Protocol
//
//	devices = make(map[string]*DeviceInstance)
//	models = make(map[string]*DeviceModel)
//	protocols = make(map[string]*Protocol)
//
//	err := Parse("./configmap_negtest.json", devices, models, protocols)
//	if err != nil{
//		t.Log(err)
//		t.FailNow()
//	}
//}
