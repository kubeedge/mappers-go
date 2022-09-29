package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	config := Config{}
	if err := config.Parse(); err != nil {
		t.Log(err)
		t.FailNow()
	}

	assert.Equal(t, "/opt/kubeedge/deviceProfile.json", config.DevInit.Configmap)
}
