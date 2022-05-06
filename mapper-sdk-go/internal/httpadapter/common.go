package httpadapter

import (
	"net/http"
	"time"

	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/common"
)

// Ping handles the requests to /ping endpoint. Is used to test if the service is working
// It returns a response as specified by the V1 API swagger in openAPI/common
func (c *RestController) Ping(writer http.ResponseWriter, request *http.Request) {
	response := "This is API " + common.APIVersion + ". Now is " + time.Now().Format(time.UnixDate)
	c.sendResponse(writer, request, common.APIPingRoute, response, http.StatusOK)
}
