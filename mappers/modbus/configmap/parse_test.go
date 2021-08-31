package configmap

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kubeedge/mappers-go/mappers/common"
	"github.com/kubeedge/mappers-go/mappers/modbus/globals"
)

func TestParse(t *testing.T) {
	var devices map[string]*globals.ModbusDev
	var models map[string]common.DeviceModel
	var protocols map[string]common.Protocol

	devices = make(map[string]*globals.ModbusDev)
	models = make(map[string]common.DeviceModel)
	protocols = make(map[string]common.Protocol)

	assert.Nil(t, Parse("./configmap_test.json", devices, models, protocols))
	for _, device := range devices {
		var pcc ModbusProtocolCommonConfig
		assert.Nil(t, json.Unmarshal([]byte(device.Instance.PProtocol.ProtocolCommonConfig), &pcc))
		assert.Equal(t, "RS485", pcc.CustomizedValues["serialType"])
	}
}

func TestParseNeg(t *testing.T) {
	var devices map[string]*globals.ModbusDev
	var models map[string]common.DeviceModel
	var protocols map[string]common.Protocol

	devices = make(map[string]*globals.ModbusDev)
	models = make(map[string]common.DeviceModel)
	protocols = make(map[string]common.Protocol)

	assert.NotNil(t, Parse("./configmap_negtest.json", devices, models, protocols))
}
