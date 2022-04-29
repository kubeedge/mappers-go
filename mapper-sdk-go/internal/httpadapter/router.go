package httpadapter

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/common"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/httpadapter/response"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/pkg/di"
	"k8s.io/klog/v2"
	"net/http"
)

// RestController the struct of HTTP route
type RestController struct {
	Router         *mux.Router
	reservedRoutes map[string]bool
	dic            *di.Container
}

// NewRestController build a RestController
func NewRestController(r *mux.Router, dic *di.Container) *RestController {
	return &RestController{
		Router:         r,
		reservedRoutes: make(map[string]bool),
		dic:            dic,
	}
}

// InitRestRoutes register the RESTful API
func (c *RestController) InitRestRoutes() {
	klog.V(1).Info("Registering v1 routes...")
	// common
	c.addReservedRoute(common.APIPingRoute, c.Ping).Methods(http.MethodGet)
	//// device command
	c.addReservedRoute(common.APIDeviceWriteCommandByIDRoute, c.WriteCommand).Methods(http.MethodPut)
	c.addReservedRoute(common.APIDeviceReadCommandByIDRoute, c.ReadCommand).Methods(http.MethodGet)
	// callback
	c.addReservedRoute(common.APIDeviceCallbackRoute, c.AddDevice).Methods(http.MethodPost)
	c.addReservedRoute(common.APIDeviceCallbackIDRoute, c.RemoveDevice).Methods(http.MethodDelete)
}

func (c *RestController) addReservedRoute(route string, handler func(http.ResponseWriter, *http.Request)) *mux.Route {
	c.reservedRoutes[route] = true
	return c.Router.HandleFunc(route, handler)
}

func (c *RestController) sendMapperError(
	writer http.ResponseWriter,
	request *http.Request,
	err string,
	API string) {
	correlationID := request.Header.Get(common.CorrelationHeader)
	klog.Error(err, common.CorrelationHeader, correlationID)
	c.sendResponse(writer, request, API, err, response.CodeMapping(common.KindServerError))
}

// sendResponse puts together the response packet for the V2 API
func (c *RestController) sendResponse(
	writer http.ResponseWriter,
	request *http.Request,
	API string,
	response interface{},
	statusCode int) {

	correlationID := request.Header.Get(common.CorrelationHeader)

	writer.Header().Set(common.CorrelationHeader, correlationID)
	writer.Header().Set(common.ContentType, common.ContentTypeJSON)
	writer.WriteHeader(statusCode)

	if response != nil {
		data, err := json.Marshal(response)
		if err != nil {
			klog.Error(fmt.Sprintf("Unable to marshal %s response", API), "error", err.Error(), common.CorrelationHeader, correlationID)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = writer.Write(data)
		if err != nil {
			klog.Error(fmt.Sprintf("Unable to write %s response", API), "error", err.Error(), common.CorrelationHeader, correlationID)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
