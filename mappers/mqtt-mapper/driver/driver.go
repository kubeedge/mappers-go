package driver

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kubeedge/mapper-framework/pkg/common"
)

func NewClient(protocol ProtocolConfig) (*CustomizedClient, error) {
	client := &CustomizedClient{
		ProtocolConfig:    protocol,
		deviceMutex:       sync.Mutex{},
		TempMessage:       "",
		DeviceConfigData:  nil,
	}
	return client, nil
}

func (c *CustomizedClient) InitDevice() error {
	configData := &c.ProtocolConfig.ConfigData
	_, operationInfo, _, err := configData.SplitTopic()
	if operationInfo != DEVICEINfO {
		return errors.New("This is not a device config.")
	}
	if err != nil {
		return err
	}
	c.TempMessage = configData.Message
	return nil
}

func (c *CustomizedClient) GetDeviceData(visitor *VisitorConfig) (interface{}, error) {
	configData := &c.ProtocolConfig.ConfigData
	_, operationInfo, _, err := configData.SplitTopic()
	if operationInfo != DEVICEINfO {
		return nil, errors.New("This is not a device config.")
	}
	if err != nil {
		return nil, err
	}
	visitor.ProcessOperation(c.DeviceConfigData)
	return c.DeviceConfigData, nil
}

func (c *CustomizedClient) SetDeviceData(visitor *VisitorConfig) error {
	configData := &c.ProtocolConfig.ConfigData
	_, operationInfo, _, err := configData.SplitTopic()
	if operationInfo == DEVICEINfO {
		return errors.New("This is a device config, not to set device data.")
	}
	if err != nil {
		return err
	}
	visitor.ProcessOperation(c.DeviceConfigData)
	return  nil
}

func (c *CustomizedClient) StopDevice() error {
	updateFieldsByTag(c.DeviceConfigData, map[string]interface{}{
		"status": common.DeviceStatusDisCONN,
		"Status": common.DeviceStatusDisCONN,
	}, "json")
	updateFieldsByTag(c.DeviceConfigData, map[string]interface{}{
		"status": common.DeviceStatusDisCONN,
		"Status": common.DeviceStatusDisCONN,
	}, "yaml")
	updateFieldsByTag(c.DeviceConfigData, map[string]interface{}{
		"status": common.DeviceStatusDisCONN,
		"Status": common.DeviceStatusDisCONN,
	}, "xml")
	return nil
}

func (c *CustomizedClient) GetDeviceStates(visitor *VisitorConfig) (string, error) {
	res, err := visitor.getFieldByTag(c.DeviceConfigData)
	if err != nil {
		return common.DeviceStatusOK, nil
	}
	return res, nil
	
}

/* --------------------------------------------------------------------------------------- */
// The function NewConfigData is a constructor for ConfigData to initialize the structure.
// It returns the ConfigData instance and an error value to handle the validity of the passed parameters.
func NewConfigData(clientID, brokerURL, topic, message, username, password string, connectionTTL time.Duration) (*ConfigData, error) {
	if clientID == "" {
		return nil, errors.New("clientID cannot be empty")
	}
	if brokerURL == "" {
		return nil, errors.New("borkerURL cannot be empty")
	}
	if topic == "" {
		return nil, errors.New("topic cannot be empty")
	}
	if message == "" {
		return nil, errors.New("message cannot be empty")
	}
	if username == "" {
		username = "defaultUser"
	}
	if password == "" {
		password = "defaultPass"
	}
	if connectionTTL == 0 {
		connectionTTL = 30 * time.Second // default timeout of 30 seconds
	}

	return &ConfigData{
		ClientID:      clientID,
		BrokerURL:     brokerURL,
		Topic:         topic,
		Message:       message,
		Username:      username,
		Password:      password,
		ConnectionTTL: connectionTTL,
		LastMessage:   time.Now(), // set last message time to current time
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
	"deviceinfo": DEVICEINfO,
	"setsinglevalue" : SETSINGLEVALUE,
	"getsinglevalue" : GETSINGLEVALUE,
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
		return c.jsonParse()

	case YAML: // yaml
		return c.yamlParse()

	case XML: // xml
		return c.xmlParse()

	default:
		return nil, errors.New("unsupported parse type")
	}
}

// The function parseJSON parses the Message field of the ConfigData (assumed to be a JSON string).
func (c *ConfigData) jsonParse() (map[string]interface{}, error) {
	if c.Message == "" {
		return nil, errors.New("message is empty")
	}

	var jsonMsg map[string]interface{}
	err := json.Unmarshal([]byte(c.Message), &jsonMsg)
	if err != nil {
		return nil, err
	}
	return jsonMsg, nil
}

// The function parseYAML parses the Message field of the ConfigData (assumed to be a YAML string).
func (c *ConfigData)yamlParse() (map[string]interface{}, error) {
	if c.Message == "" {
		return nil, errors.New("message is empty")
	}

	var yamlMsg map[string]interface{}
	err := yaml.Unmarshal([]byte(c.Message), &yamlMsg)
	if err != nil {
		return nil, err
	}
	return yamlMsg, nil
}

// The function xmlParse parses the Message field of the ConfigData (assumed to be a XML string).
func (c *ConfigData)xmlParse() (map[string]interface{}, error) {
	msg := c.Message
	if strings.HasPrefix(msg, "<?xml") {
		end := strings.Index(msg, "?>")
		if end != -1 {
			msg = msg[end+2:]
		}
	}

	var node Node
	err := xml.Unmarshal([]byte(msg), &node)
	if err != nil {
		return nil, err
	}

	xmlMsg := nodeToMap(node)
	var mp map[string]interface{}
	for _, value := range xmlMsg {
		if nestedMap, ok := value.(map[string]interface{}); ok {
			mp = nestedMap
			break
		}
	}
	return mp, err
}

// NewVisitorConfig creates a new instance of VisitorConfig using ConfigData pointer and the result of SplitTopic.
func (c *ConfigData) NewVisitorConfig() (*VisitorConfig, error) {
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
	return &VisitorConfig{
		ProtocolName: "mqtt",
		VisitorConfigData: VisitorConfigData{
			DataType:         "DefaultDataType", 
			ClientID:         clientID,
			DeviceInfo:       deviceInfo,
			OperationInfo:    operationInfo,
			SerializedFormat: serializedFormat,
			ParsedMessage:    parsedMessage,
		},
	}, nil
}

/* --------------------------------------------------------------------------------------- */
// The function ParseMessage parses the Message field according to the incoming type.
// parseType(0: json, 1: yaml, 2: xml)
// The value interface{} represents the parsed structure.
func (v *VisitorConfig) ProcessOperation(deviceConfigData interface{}) error {
	if v.VisitorConfigData.ParsedMessage == nil {
		return errors.New("visitor message is empty")
	}

	if deviceConfigData == nil {
		return errors.New("device message is empty")
	}

	switch v.VisitorConfigData.OperationInfo {
	case DEVICEINfO:  // device config data
		v.updateFullConfig(deviceConfigData)
		return nil
	case UPDATE:  // update the full text according the visitor config and the tag (json, yaml, xml)
		v.updateFullConfig(deviceConfigData)
		return nil
	case SETSINGLEVALUE:  // update the single value according the visitor config and the tag (json, yaml, xml)
		v.updateFieldsByTag(deviceConfigData)
		return nil
	default:
		return errors.New("unsupported operation type")
	}
}

func (v *VisitorConfig) updateFullConfig(destDataConfig interface{}) error {
	destValue := reflect.ValueOf(destDataConfig)
	if destValue.Kind() != reflect.Ptr || destValue.Elem().Kind() != reflect.Struct {
		return errors.New("destDataConfig must be a pointer to a struct")
	}

	destValue = destValue.Elem()

	var tagName string
	switch v.VisitorConfigData.SerializedFormat {
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
	if err := updateStructFields(destValue, v.VisitorConfigData.ParsedMessage, tagName); err != nil {
		return err
	}

	return nil
}

func (v *VisitorConfig)updateFieldsByTag(destDataConfig interface{}) error {
	vv := reflect.ValueOf(destDataConfig).Elem()

	var tagName string
	switch v.VisitorConfigData.SerializedFormat {
	case JSON:
		tagName = "json"
	case YAML:
		tagName = "yaml"
	case XML:
		tagName = "xml"
	default:
		return errors.New("unknown serialized format")
	}

	for key, value := range v.VisitorConfigData.ParsedMessage {
		if err := setFieldByTag(vv, key, value, tagName); err != nil {
			return err
		}
	}
	return nil
}

/* --------------------------------------------------------------------------------------- */
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

// Convert XML attributes to map entries
func attrsToMap(attrs []xml.Attr) map[string]interface{} {
	attrMap := make(map[string]interface{})
	for _, attr := range attrs {
		attrMap[attr.Name.Local] = attr.Value
	}
	return attrMap
}

// The function nodeToMap recursively converts XML nodes to map[string]interface{}.
func nodeToMap(node Node) map[string]interface{} {
	xmlMsg := make(map[string]interface{})

	// Process attributes
	if len(node.Attr) > 0 {
		xmlMsg["attributes"] = attrsToMap(node.Attr)
	}

	// If the node has no children, it is a leaf node, apply type conversion.
	if len(node.Nodes) == 0 {
		xmlMsg[node.XMLName.Local] = convertValue(strings.TrimSpace(node.Content))
		return xmlMsg
	}

	// Process child nodes recursively.
	children := make(map[string]interface{})
	for _, child := range node.Nodes {
		childMap := nodeToMap(child)
		if existing, found := children[child.XMLName.Local]; found {
			switch v := existing.(type) {
			case []interface{}:
				children[child.XMLName.Local] = append(v, childMap[child.XMLName.Local])
			default:
				children[child.XMLName.Local] = []interface{}{v, childMap[child.XMLName.Local]}
			}
		} else {
			children[child.XMLName.Local] = childMap[child.XMLName.Local]
		}
	}

	xmlMsg[node.XMLName.Local] = children
	return xmlMsg
}

func setFieldByTag(v reflect.Value, key string, value interface{}, tagName string) error {
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		fieldVal := v.Field(i)

		if field.Tag.Get(tagName) == key {
			val := reflect.ValueOf(value)
			if fieldVal.Type() != val.Type() {
				return fmt.Errorf("type mismatch: cannot assign %s to %s", val.Type(), fieldVal.Type())
			}
			fieldVal.Set(val)
			return nil
		}

		if fieldVal.Kind() == reflect.Struct {
			if err := setFieldByTag(fieldVal, key, value, tagName); err == nil {
				return nil
			}
		}
	}
	return fmt.Errorf("no such field with tag: %s", key)
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

func updateFieldsByTag(s interface{}, updates map[string]interface{}, tagName string) error {
	v := reflect.ValueOf(s).Elem()
	for key, value := range updates {
		if err := setFieldByTag(v, key, value, tagName); err != nil {
			return err
		}
	}
	return nil
}

func (v * VisitorConfig)getFieldByTag(s interface{}) (string, error) {
	vv := reflect.ValueOf(s).Elem()

	var tagName string
	switch v.VisitorConfigData.SerializedFormat {
	case JSON:
		tagName = "json"
	case YAML:
		tagName = "yaml"
	case XML:
		tagName = "xml"
	default:
		return "", errors.New("unknown serialized format")
	}

	res, err := findFieldByTag(vv, "status", tagName)
	if err != nil {
		res, err = findFieldByTag(vv, "Status", tagName)
		if err != nil {
			return "", err
		} else {
			return res, nil
		}
	} else {
		return res, nil
	}
}

func findFieldByTag(v reflect.Value, key string, tagName string) (string, error) {
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		fieldVal := v.Field(i)

		if field.Tag.Get(tagName) == key {
			return fieldVal.String(), nil
		}

		if fieldVal.Kind() == reflect.Struct {
			if value, err := findFieldByTag(fieldVal, key, tagName); err == nil {
				return value, nil
			}
		}
	}
	return "", fmt.Errorf("no such field with tag: %s", key)
}
/* --------------------------------------------------------------------------------------- */