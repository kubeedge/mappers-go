// Package common used to store constants, data conversion functions, timers, etc
package common

// joint the topic like topic := fmt.Sprintf(TopicTwinUpdateDelta, deviceID)
const (
	TopicTwinUpdateDelta = "$hw/events/device/%s/twin/update/delta"
	TopicTwinUpdate      = "$hw/events/device/%s/twin/update"
	TopicStateUpdate     = "$hw/events/device/%s/state/update"
	TopicDataUpdate      = "$ke/events/device/%s/data/update"
	TopicDeviceUpdate    = "$hw/events/node/#"
)

// Device status definition.
const (
	DEVSTOK      = "OK"
	DEVSTDISCONN = "DISCONNECTED"
)

// joint x joint the instancepool like driverName :=  common.DriverPrefix+instanceID+twin.PropertyName
const (
	DriverPrefix = "Driver"
)

const (
	// CorrelationHeader to be added in header
	CorrelationHeader = "X-Correlation-ID"
)

const (
	// APIVersion description API version
	APIVersion = "v1"
	// APIBase to build RESTful API
	APIBase    = "/api/v1"

	// APIDeviceRoute to build RESTful API
	APIDeviceRoute                 = APIBase + "/device"
	// APIDeviceWriteCommandByIDRoute to build read command's RESTful API
	APIDeviceWriteCommandByIDRoute = APIDeviceRoute + "/" + ID + "/{" + IDAndCommand + "}"
	// APIDeviceReadCommandByIDRoute to build write command's RESTful API
	APIDeviceReadCommandByIDRoute  = APIDeviceRoute + "/" + ID + "/{" + ID + "}" + "/{" + Command + "}"
	// APIDeviceCallbackRoute to build update device's RESTful API
	APIDeviceCallbackRoute         = APIBase + "/callback/device"
	// APIDeviceCallbackIDRoute to build update device's RESTful API
	APIDeviceCallbackIDRoute       = APIBase + "/callback/device/id/{id}"

	// APIPingRoute to build ping command's RESTful API
	APIPingRoute = APIBase + "/ping"
)

const (
	// ID to build RESTful API
	ID           = "id"
	// Command to build RESTful API
	Command      = "command"
	// IDAndCommand to build RESTful API
	IDAndCommand = "IdAndCommand"
)

// Constants related to the possible content types supported by the APIs
const (
	ContentType     = "Content-Type"
	ContentTypeJSON = "application/json"
)

// ErrKind define the error's type
type ErrKind string

// Constant Kind identifiers which can be used to label and group errors.
const (
	KindEntityDoesNotExist  ErrKind = "NotFound"
	KindServerError         ErrKind = "UnexpectedServerError"
	KindDuplicateName       ErrKind = "DuplicateName"
	KindInvalidID           ErrKind = "InvalidId"
	KindServiceUnavailable  ErrKind = "ServiceUnavailable"
	KindNotAllowed          ErrKind = "NotAllowed"
	KindServiceLocked       ErrKind = "ServiceLocked"
	KindNotImplemented      ErrKind = "NotImplemented"
	KindRangeNotSatisfiable ErrKind = "RangeNotSatisfiable"
	KindOverflowError       ErrKind = "OverflowError"
	KindNaNError            ErrKind = "NaNError"
)
