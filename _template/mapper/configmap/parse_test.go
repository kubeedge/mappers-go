package configmap

import (
	"testing"

	"github.com/kubeedge/mappers-go/pkg/common"

	"github.com/stretchr/testify/assert"

	"github.com/kubeedge/mappers-go/mappers/Template/globals"
)

func TestParse(t *testing.T) {
	var devices map[string]*globals.TemplateDev
	var models map[string]common.DeviceModel
	var protocols map[string]common.Protocol

	devices = make(map[string]*globals.TemplateDev)
	models = make(map[string]common.DeviceModel)
	protocols = make(map[string]common.Protocol)

	assert.Nil(t, Parse("./configmap_test.json", devices, models, protocols))
}

func TestParseNeg(t *testing.T) {
	var devices map[string]*globals.TemplateDev
	var models map[string]common.DeviceModel
	var protocols map[string]common.Protocol

	devices = make(map[string]*globals.TemplateDev)
	models = make(map[string]common.DeviceModel)
	protocols = make(map[string]common.Protocol)

	assert.NotNil(t, Parse("./configmap_negtest.json", devices, models, protocols))
}
