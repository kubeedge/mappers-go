package common

import (
	"encoding/json"
	"errors"
	"strconv"
)

// Convert string to other types
func Convert(valueType string, value string) (result interface{}, err error) {
	switch valueType {
	case "int":
		return strconv.ParseInt(value, 10, 64)
	case "float":
		return strconv.ParseFloat(value, 32)
	case "double":
		return strconv.ParseFloat(value, 64)
	case "boolean":
		return strconv.ParseBool(value)
	case "string":
		return value, nil
	default:
		return nil, errors.New("convert failed")
	}
}

// ConvertToString other types to string
func ConvertToString(value interface{}) (string, error) {
	var result string
	if value == nil {
		return result, nil
	}
	switch value.(type) {
	case float64:
		ft := value.(float64)
		result = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		result = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		result = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		result = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		result = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		result = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		result = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		result = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		result = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		result = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		result = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		result = strconv.FormatUint(it, 10)
	case string:
		result = value.(string)
	case []byte:
		result = string(value.([]byte))
	default:
		newValue, err := json.Marshal(value)
		if err != nil {
			return "", err
		}
		result = string(newValue)
	}
	return result, nil
}
