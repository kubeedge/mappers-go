// Package response used to implement the responses structure
package response

import (
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/common"
)

// BaseResponse the base response struct of all request
// all response's payload should contain BaseResponse
type BaseResponse struct {
	Version    string
	RequestID  string `json:"requestId,omitempty"`
	Message    string `json:"message,omitempty"`
	StatusCode int    `json:"statusCode"`
}

// ReadCommandResponse the response struct of read command
type ReadCommandResponse struct {
	BaseResponse
	DeviceID     string
	PropertyName string
	Value        string
}

// WriteCommandResponse the response struct of write command
type WriteCommandResponse struct {
	BaseResponse
	DeviceID     string
	PropertyName string
	Status       string
}

// UpdateDeviceResponse the response struct of update device
type UpdateDeviceResponse struct {
	BaseResponse
	DeviceID  string
	Operation string
	Status    string
}

// NewBaseResponse build the base message
func NewBaseResponse(requestID string, message string, statusCode int) BaseResponse {
	return BaseResponse{
		Version:    common.APIVersion,
		RequestID:  requestID,
		Message:    message,
		StatusCode: statusCode,
	}
}

// NewReadCommandResponse build the read command message
func NewReadCommandResponse(response BaseResponse, deviceID, propertyName, value string) ReadCommandResponse {
	return ReadCommandResponse{
		response,
		deviceID,
		propertyName,
		value,
	}
}

// NewWriteCommandResponse build the write command message
func NewWriteCommandResponse(response BaseResponse, deviceID, propertyName, status string) WriteCommandResponse {
	return WriteCommandResponse{
		response,
		deviceID,
		propertyName,
		status,
	}
}

// NewUpdateDeviceResponse build the update device message
func NewUpdateDeviceResponse(response BaseResponse, deviceID, operation, status string) UpdateDeviceResponse {
	return UpdateDeviceResponse{
		response,
		deviceID,
		operation,
		status,
	}
}
