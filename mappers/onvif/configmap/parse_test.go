package configmap

import (
	"fmt"
	"testing"

	"github.com/kubeedge/mappers-go/mappers/common"
	"github.com/kubeedge/mappers-go/mappers/onvif/globals"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	var devices map[string]*globals.OnvifDev
	var models map[string]common.DeviceModel
	var protocols map[string]common.Protocol

	devices = make(map[string]*globals.OnvifDev)
	models = make(map[string]common.DeviceModel)
	protocols = make(map[string]common.Protocol)

	assert.Nil(t, Parse("./configmap_test.json", devices, models, protocols))
	fmt.Printf("%v", devices)
	fmt.Printf("%v", models)
	fmt.Printf("%v", protocols)

}

func TestParseNeg(t *testing.T) {
	var devices map[string]*globals.OnvifDev
	var models map[string]common.DeviceModel
	var protocols map[string]common.Protocol

	devices = make(map[string]*globals.OnvifDev)
	models = make(map[string]common.DeviceModel)
	protocols = make(map[string]common.Protocol)

	assert.NotNil(t, Parse("./configmap_negtest.json", devices, models, protocols))
}
