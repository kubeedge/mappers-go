package configmap

import (
	"encoding/json"
	"github.com/kubeedge/mappers-go/pkg/ble/globals"
	"github.com/kubeedge/mappers-go/pkg/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParse(t *testing.T) {
	var devices map[string]*globals.BluetoothDev
	var models map[string]common.DeviceModel
	var protocols map[string]common.Protocol

	devices = make(map[string]*globals.BluetoothDev)
	models = make(map[string]common.DeviceModel)
	protocols = make(map[string]common.Protocol)

	assert.Nil(t, Parse("./configmap_test.json", devices, models, protocols))
	for _, device := range devices {
		var bpc BluetoothProtocolConfig
		assert.Nil(t, json.Unmarshal([]byte(device.Instance.PProtocol.ProtocolConfigs), &bpc))
		assert.Equal(t, "A4:C1:38:1A:49:90", bpc.MacAddress)
	}
}
