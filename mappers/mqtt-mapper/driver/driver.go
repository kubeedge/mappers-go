package driver

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/kubeedge/mapper-framework/pkg/common"
	"gopkg.in/yaml.v3"
)

func NewClient(protocol ProtocolConfig) (*CustomizedClient, error) {
	client := &CustomizedClient{
		ProtocolConfig:   protocol,
		deviceMutex:      sync.Mutex{},
		DeviceInfo:       "",
		ParsedDeviceInfo: make(map[string]interface{}),
	}
	return client, nil
}

func (c *CustomizedClient) InitDevice() error {
	configData := &c.ProtocolConfig.ConfigData
	_, _, format, _ := configData.SplitTopic()
	c.DeviceInfo = c.ProtocolConfig.ConfigData.Message
	c.ParsedDeviceInfo, _ = c.ParseMessage(format)
	return nil
}

func (c *CustomizedClient) GetDeviceData(visitor *VisitorConfig) (interface{}, error) {

	return nil, nil
}

func (c *CustomizedClient) SetDeviceData(data interface{}, visitor *VisitorConfig) error {
	vPointer := visitor.VisitorConfigData
	vPointer.ModifyVisitorConfigData(c.ParsedDeviceInfo)
	return nil
}

func (c *CustomizedClient) GetDeviceStates() (string, error) {
	// TODO: GetDeviceStates
	return common.DeviceStatusOK, nil
}

/* --------------------------------------------------------------------------------------- */
// The function NewConfigData is a constructor for ConfigData to initialize the structure.
// It returns the ConfigData instance and an error value to handle the validity of the passed parameters.
func NewConfigData(clientID, topic, message string) (*ConfigData, error) {
	if clientID == "" {
		return nil, errors.New("clientID cannot be empty")
	}
	if topic == "" {
		return nil, errors.New("topic cannot be empty")
	}
	if message == "" {
		message = "default message"
	}

	return &ConfigData{
		ClientID: clientID,
		Topic:    topic,
		Message:  message,
	}, nil
}

// The function GetClientID returns the value of the ClientID field and error.
func (c *ConfigData) GetClientID() (string, error) {
	if c.ClientID == "" {
		return "", errors.New("clientID is empty")
	}
	return c.ClientID, nil
}

// The function GetTopic returns the value of the Topic field and error.
func (c *ConfigData) GetTopic() (string, error) {
	if c.Topic == "" {
		return "", errors.New("topic is empty")
	}
	return c.Topic, nil
}

// GetMessage returns the value of the Message field and error.
func (c *ConfigData) GetMessage() (string, error) {
	if c.Message == "" {
		return "", errors.New("message is empty")
	}
	return c.Message, nil
}

// OperationInfoType and SerializedFormatType mappings
var operationTypeMap = map[string]OperationInfoType{
	"update": UPDATE,
}

var serializedFormatMap = map[string]SerializedFormatType{
	"json": JSON,
	"yaml": YAML,
	"xml":  XML,
}

// The function SplitTopic splits the Topic into three parts and returns each.
// OperationInfoType(fulltextmodify: 0, pathmodify: 1, valuemodify: 2)
// SerializedFormatType(json: 0, yaml: 1, xml: 2)
func (c *ConfigData) SplitTopic() (string, OperationInfoType, SerializedFormatType, error) {
	if c.Topic == "" {
		return "", 0, 0, errors.New("topic is empty")
	}

	parts := strings.Split(c.Topic, "/")

	if len(parts) < 3 {
		return "", 0, 0, errors.New("topic format is invalid, must have at least three parts")
	}

	deviceInfo := strings.Join(parts[:len(parts)-2], "/")

	// Get operation type from map
	operationType, exists := operationTypeMap[parts[len(parts)-2]]
	if !exists {
		return "", 0, 0, errors.New("invalid operation type")
	}

	// Get serialized format from map
	serializedFormat, exists := serializedFormatMap[parts[len(parts)-1]]
	if !exists {
		return "", 0, 0, errors.New("invalid serialized format")
	}

	return deviceInfo, operationType, serializedFormat, nil
}

// The function ParseMessage parses the Message field according to the incoming type.
// parseType(0: json, 1: yaml, 2: xml)
// The value interface{} represents the parsed structure.
func (c *ConfigData) ParseMessage(parseType SerializedFormatType) (map[string]interface{}, error) {
	if c.Message == "" {
		return nil, errors.New("message is empty")
	}

	switch parseType {
	case JSON: // json
		return c.parseJSON()

	case YAML: // yaml
		convertedMessage, err := convertYAMLToJSON(c.Message)
		if err != nil {
			return nil, err
		}
		c.Message = convertedMessage
		return c.parseJSON()

	case XML: // xml
		convertedMessage, err := convertXMLToMap(c.Message)
		if err != nil {
			return nil, err
		}
		// c.Message = convertedMessage
		originalMap := convertedMessage
		var mp map[string]interface{}
		for _, value := range originalMap {
			if nestedMap, ok := value.(map[string]interface{}); ok {
				mp = nestedMap
				break
			}
		}
		return mp, err

	default:
		return nil, errors.New("unsupported parse type")
	}
}

// The function parseJSON parses the Message field of the ConfigData (assumed to be a JSON string).
func (c *ConfigData) parseJSON() (map[string]interface{}, error) {
	if c.Message == "" {
		return nil, errors.New("message is empty")
	}

	var result map[string]interface{}
	err := json.Unmarshal([]byte(c.Message), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// The function ValidateMessage checks if the message content is valid.
func (c *ConfigData) ValidateMessage() error {
	if c.Message == "" {
		return errors.New("message is empty")
	}

	// Example: Check if the message is valid JSON (you can expand for other formats)
	var temp map[string]interface{}
	if err := json.Unmarshal([]byte(c.Message), &temp); err != nil {
		return errors.New("message is not valid JSON")
	}

	return nil
}

// NewVisitorConfigData creates a new instance of VisitorConfigData using ConfigData pointer and the result of SplitTopic.
func (c *ConfigData) NewVisitorConfigData() (*VisitorConfigData, error) {
	// get ClientID
	clientID, err := c.GetClientID()
	if err != nil {
		return nil, err
	}

	// get DeviceInfo, OperationInfo and SerializedFormat
	deviceInfo, operationInfo, serializedFormat, err := c.SplitTopic()
	if err != nil {
		return nil, err
	}

	// get ParsedMessage
	parsedMessage, err := c.ParseMessage(serializedFormat)
	if err != nil {
		return nil, err
	}

	// create
	return &VisitorConfigData{
		DataType:         "string",
		ClientID:         clientID,
		DeviceInfo:       deviceInfo,
		OperationInfo:    operationInfo,
		SerializedFormat: serializedFormat,
		ParsedMessage:    parsedMessage,
	}, nil
}

/* --------------------------------------------------------------------------------------- */
func (v *VisitorConfigData) ModifyVisitorConfigData(destDataConfig interface{}) error {
	destValue := reflect.ValueOf(destDataConfig)
	if destValue.Kind() != reflect.Ptr || destValue.Elem().Kind() != reflect.Struct {
		return errors.New("destDataConfig must be a pointer to a struct")
	}

	destValue = destValue.Elem()

	var tagName string
	switch v.SerializedFormat {
	case JSON:
		tagName = "json"
	case YAML:
		tagName = "yaml"
	case XML:
		tagName = "xml"
	default:
		return errors.New("unknown serialized format")
	}

	// Update the destination struct using JSON tag
	if err := updateStructFields(destValue, v.ParsedMessage, tagName); err != nil {
		return err
	}

	return nil
}

// updateStructFields recursively updates struct fields from the given map using specified tag type
func updateStructFields(structValue reflect.Value, data map[string]interface{}, tagName string) error {
	structType := structValue.Type()

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		fieldType := structType.Field(i)
		tagValue := fieldType.Tag.Get(tagName)

		var value interface{}
		var exists bool

		if tagValue != "" {
			// Attempt to get value using tag
			value, exists = data[tagValue]
		}

		if !exists {
			// Fallback to field name if tag is not found
			tagValue = fieldType.Name
			value, exists = data[tagValue]
		}

		if !exists {
			continue
		}

		// Update the field based on its kind
		if field.Kind() == reflect.Struct {
			nestedData, ok := value.(map[string]interface{})
			if !ok {
				return fmt.Errorf("type mismatch for nested field %s", tagValue)
			}
			if err := updateStructFields(field, nestedData, tagName); err != nil {
				return err
			}
		} else if field.Kind() == reflect.Slice {
			sliceData, ok := value.([]interface{})
			if !ok {
				return fmt.Errorf("type mismatch for slice field %s", tagValue)
			}
			newSlice := reflect.MakeSlice(field.Type(), len(sliceData), len(sliceData))
			for j, item := range sliceData {
				itemValue := reflect.ValueOf(item)
				if newSlice.Index(j).Kind() == itemValue.Kind() {
					newSlice.Index(j).Set(itemValue)
				} else {
					return fmt.Errorf("type mismatch for slice item in field %s", tagValue)
				}
			}
			field.Set(newSlice)
		} else {
			fieldValue := reflect.ValueOf(value)
			if field.Type() == fieldValue.Type() {
				field.Set(fieldValue)
			} else {
				return fmt.Errorf("type mismatch for field %s", tagValue)
			}
		}
	}
	return nil
}

/* --------------------------------------------------------------------------------------- */
// The function ConvertYAMLToJSON converts a YAML string to a JSON string.
func convertYAMLToJSON(yamlString string) (string, error) {
	// Converting a YAML string to a generic map object
	var yamlData map[string]interface{}
	err := yaml.Unmarshal([]byte(yamlString), &yamlData)
	if err != nil {
		return "", err
	}

	// Convert a map object to a JSON string
	jsonData, err := json.Marshal(yamlData)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// The function ConvertXMLToMap converts XML string to map[string]interface{}.
func convertXMLToMap(xmlString string) (map[string]interface{}, error) {
	// Wrap the XML content with <root></root>
	wrappedXML := wrapXMLWithRoot(xmlString)

	var node Node
	err := xml.Unmarshal([]byte(wrappedXML), &node)
	if err != nil {
		return nil, err
	}
	return nodeToMap(node), nil
}

// The function WrapXMLWithRoot wraps XML strings in <root></root> tags.
func wrapXMLWithRoot(xmlString string) string {
	// Remove the XML declaration if it exists
	if strings.HasPrefix(xmlString, "<?xml") {
		end := strings.Index(xmlString, "?>")
		if end != -1 {
			xmlString = xmlString[end+2:]
		}
	}

	// Wrap the remaining XML content with <root></root>
	wrappedXML := xmlString
	return wrappedXML
}

// Node structure
type Node struct {
	XMLName xml.Name
	Content string     `xml:",chardata"`
	Nodes   []Node     `xml:",any"`
	Attr    []xml.Attr `xml:"-"`
}

// convertValue attempts to convert string content to appropriate type.
func convertValue(content string) interface{} {
	if f, err := strconv.ParseFloat(content, 64); err == nil {
		return f
	} else if i, err := strconv.Atoi(content); err == nil {
		return i
	} else if b, err := strconv.ParseBool(content); err == nil {
		return b
	} else {
		return content
	}
}

// The function nodeToMap recursively converts XML nodes to map[string]interface{}.
func nodeToMap(node Node) map[string]interface{} {
	result := make(map[string]interface{})

	// If the node has no children, it is a leaf node, apply type conversion.
	if len(node.Nodes) == 0 {
		return map[string]interface{}{node.XMLName.Local: convertValue(strings.TrimSpace(node.Content))}
	}

	// Process child nodes recursively.
	for _, child := range node.Nodes {
		childMap := nodeToMap(child)
		if existing, found := result[child.XMLName.Local]; found {
			switch v := existing.(type) {
			case []interface{}:
				result[child.XMLName.Local] = append(v, childMap[child.XMLName.Local])
			default:
				result[child.XMLName.Local] = []interface{}{v, childMap[child.XMLName.Local]}
			}
		} else {
			result[child.XMLName.Local] = childMap[child.XMLName.Local]
		}
	}

	return map[string]interface{}{node.XMLName.Local: result}
}

// The function MapToJSON converts map[string]interface{} to JSON string.
func mapToJSON(data map[string]interface{}) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

func StructToJSON(v interface{}) (string, error) {
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

/* --------------------------------------------------------------------------------------- */
