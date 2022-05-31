package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	bledevice "github.com/kubeedge/mappers-go/mappers/ble/device"
	modbusdevice "github.com/kubeedge/mappers-go/mappers/modbus/device"
	opcuadevice "github.com/kubeedge/mappers-go/mappers/opcua/device"
	"github.com/kubeedge/mappers-go/pkg/common"
	"github.com/kubeedge/mappers-go/pkg/util/parse"
)

func ResponseOK(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	enc.Encode("ok")
}

func ResponseError(w http.ResponseWriter, err error, status int) {
	if err == nil {
		w.WriteHeader(status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.Encode(err.Error())
}

func StartHttpServer(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ResponseOK(w)
		return
	})
	// device change, reload this device
	mux.HandleFunc("/device", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			ResponseError(w, fmt.Errorf("http method unsupport"), http.StatusBadRequest)
			return
		}
		var data parse.DeviceData
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			ResponseError(w, err, http.StatusBadRequest)
			return
		}
		var model common.DeviceModel
		var deviceInstance *common.DeviceInstance
		if data.DeviceModel != nil {
			model = parse.ParseDeviceModel(data.DeviceModel)
		}
		if data.Device != nil {
			protocol, err := parse.BuildProtocol(data.Device)
			if err != nil {
				ResponseError(w, err, http.StatusInternalServerError)
				return
			}
			deviceInstance, err = parse.ParseDevice(data.Device, &model)
			if err != nil {
				ResponseError(w, err, http.StatusInternalServerError)
				return
			}
			deviceInstance.PProtocol = protocol

			switch deviceInstance.PProtocol.Protocol {
			case common.ProtocolBlueTooth:
				bledevice.UpdateDev(&model, deviceInstance, &protocol)
			case common.ProtocolModbus:
				modbusdevice.UpdateDev(&model, deviceInstance, &protocol)
			case common.ProtocolOnvif:
				// TODO ffmpeg
				//onvifdevice.UpdateDev(&model, deviceInstance, &protocol)
			case common.ProtocolOpcua:
				opcuadevice.UpdateDev(&model, deviceInstance, &protocol)
			default:
				ResponseError(w,
					fmt.Errorf("device %s protocol %s unsupport",
						deviceInstance.Name, deviceInstance.PProtocol.Protocol),
					http.StatusBadRequest)
				return
			}
		}
		ResponseOK(w)
		return
	})
	// TODO does device manager need to use http request to get device twin message?
	return http.ListenAndServe(addr, mux)
}
