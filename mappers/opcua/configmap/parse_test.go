package configmap

import (
	"encoding/json"
	"testing"

	mappercommon "github.com/kubeedge/mappers-go/mappers/common"
	. "github.com/kubeedge/mappers-go/mappers/opcua/globals"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	var devices map[string]*OPCUADev
	var models map[string]mappercommon.DeviceModel
	var protocols map[string]mappercommon.Protocol

	devices = make(map[string]*OPCUADev)
	models = make(map[string]mappercommon.DeviceModel)
	protocols = make(map[string]mappercommon.Protocol)

	assert.Nil(t, Parse("./configmap_test.json", devices, models, protocols))
	for _, device := range devices {
		var pcc ProtocolConfigOPCUA
		assert.Nil(t, json.Unmarshal([]byte(device.Instance.PProtocol.ProtocolConfigs), &pcc))
		assert.Equal(t, pcc.UserName, "testuser")
	}
}
