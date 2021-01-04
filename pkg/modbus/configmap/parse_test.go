package configmap

import (
	"encoding/json"
	"testing"

	"github.com/kubeedge/mappers-go/pkg/common"
  "github.com/kubeedge/mappers-go/pkg/modbus/globals"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	var devices map[string]*ModbusDev
	var models map[string]common.DeviceModel
	var protocols map[string]common.Protocol

	devices = make(map[string]*ModbusDev)
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
	var devices map[string]*ModbusDev
	var models map[string]common.DeviceModel
	var protocols map[string]common.Protocol

	devices = make(map[string]*ModbusDev)
	models = make(map[string]common.DeviceModel)
	protocols = make(map[string]common.Protocol)

	assert.NotNil(t, Parse("./configmap_negtest.json", devices, models, protocols))
}
