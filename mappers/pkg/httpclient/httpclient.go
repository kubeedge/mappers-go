package httpclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/kubeedge/kubeedge/pkg/apis/devices/v1alpha2"
)

var (
	deviceListUrl      = "/apis/devices.kubeedge.io/v1alpha2/namespaces/%s/devices"
	deviceModelListUrl = "/apis/devices.kubeedge.io/v1alpha2/namespaces/%s/devicemodels"
	//deviceUrl          = deviceListUrl + "/%s"
)

func GetDeviceList(addr, ns string) ([]v1alpha2.Device, error) {
	var res v1alpha2.DeviceList
	response, err := resty.New().R().Get(addr + fmt.Sprintf(deviceListUrl, ns))
	if err != nil {
		return nil, err
	}
	if response.RawResponse.StatusCode != http.StatusOK {
		return nil, errors.New(string(response.Body()))
	}
	if err = json.Unmarshal(response.Body(), &res); err != nil {
		return nil, err
	}
	return res.Items, nil
}

func GetDeviceModelList(addr, ns string) ([]v1alpha2.DeviceModel, error) {
	var res v1alpha2.DeviceModelList
	response, err := resty.New().R().Get(addr + fmt.Sprintf(deviceModelListUrl, ns))
	if err != nil {
		return nil, err
	}
	if response.RawResponse.StatusCode != http.StatusOK {
		return nil, errors.New(string(response.Body()))
	}
	if err = json.Unmarshal(response.Body(), &res); err != nil {
		return nil, err
	}
	return res.Items, nil
}
