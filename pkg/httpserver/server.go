package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"k8s.io/klog/v2"

	bledevice "github.com/kubeedge/mappers-go/mappers/ble/device"
	modbusdevice "github.com/kubeedge/mappers-go/mappers/modbus/device"
	opcuadevice "github.com/kubeedge/mappers-go/mappers/opcua/device"
	"github.com/kubeedge/mappers-go/pkg/common"
	"github.com/kubeedge/mappers-go/pkg/util/parse"
)

func ResponseOK(w http.ResponseWriter) {
	Response(w, "ok")
}

func Response(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	_ = enc.Encode(v)
}

func ResponseError(w http.ResponseWriter, err error, status int) {
	if err == nil {
		w.WriteHeader(status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	_ = enc.Encode(err.Error())
}

func StartHTTPServer(addr string) error {
	mux := chi.NewMux()
	mux.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		ResponseOK(w)
	})
	// device change, reload this device
	mux.Post("/device", func(w http.ResponseWriter, r *http.Request) {
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
				modbusdevice.NewDevPanel().UpdateDev(&model, deviceInstance, &protocol)
			case common.ProtocolOnvif:
				// TODO need ffmpeg
				//onvifdevice.UpdateDev(&model, deviceInstance, &protocol)
			case common.ProtocolOpcua:
				opcuadevice.UpdateDev(&model, deviceInstance, &protocol)
			case common.ProtocolCustomized:
				// TODO
				ResponseError(w,
					fmt.Errorf("device %s protocol %s unsupport",
						deviceInstance.Name, deviceInstance.PProtocol.Protocol),
					http.StatusBadRequest)
			default:
				ResponseError(w,
					fmt.Errorf("device %s protocol %s unsupport",
						deviceInstance.Name, deviceInstance.PProtocol.Protocol),
					http.StatusBadRequest)
				return
			}
		}
		ResponseOK(w)
	})
	// get device twin
	mux.Get("/device/{id}/twins", func(w http.ResponseWriter, r *http.Request) {
		deviceID := chi.URLParam(r, "id")
		deviceType := r.URL.Query().Get("type")
		if deviceType == "" {
			ResponseError(w,
				fmt.Errorf("device %s type %s unsupport", deviceID, deviceType),
				http.StatusBadRequest)
		}
		// we can use twin query param to access the property data.
		var res interface{}
		var err error
		switch deviceType {
		case common.ProtocolBlueTooth:
			res, err = bledevice.DealDeviceTwinGet(deviceID, r.URL.Query().Get("twin"))
		case common.ProtocolModbus:
			res, err = modbusdevice.NewDevPanel().DealDeviceTwinGet(deviceID, r.URL.Query().Get("twin"))
		case common.ProtocolOnvif:
			//res, err = onvif.DealDeviceTwinGet(deviceID, r.URL.Query().Get("twin"))
		case common.ProtocolOpcua:
			res, err = opcuadevice.DealDeviceTwinGet(deviceID, r.URL.Query().Get("twin"))
		default:
			ResponseError(w,
				fmt.Errorf("device %s protocol %s unsupport",
					deviceID, deviceType),
				http.StatusBadRequest)
			return
		}
		if err != nil {
			ResponseError(w, err, http.StatusInternalServerError)
			return
		}
		Response(w, res)
	})
	klog.Infof("http server listen on: %s", addr)
	http.ListenAndServe(addr, mux)
	return nil
}
