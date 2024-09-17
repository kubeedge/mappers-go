package driver

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"sync"
	"gopkg.in/yaml.v3"
)

func NewClientWithMessageQueue(protocol ProtocolConfig, queue MessageQueue) (*CustomizedClient, error) {
    return &CustomizedClient{
        ProtocolConfig: protocol,
        deviceMutex:    sync.Mutex{},
        MessageQueue:   queue,
    }, nil
}

// Thread-safe access to ProtocolConfig.
func (client *CustomizedClient) GetProtocolConfig() ProtocolConfig {
	client.deviceMutex.Lock()    
	defer client.deviceMutex.Unlock()  

	return client.ProtocolConfig
}

// Thread-safe setting of ProtocolConfig.
func (client *CustomizedClient) SetProtocolConfig(config ProtocolConfig) {
	client.deviceMutex.Lock()    
	defer client.deviceMutex.Unlock()  

	client.ProtocolConfig = config
}

// Thread-safe publishing of messages to MQTT.
func (client *CustomizedClient) SafePublish(topic string, message interface{}) error {
	client.deviceMutex.Lock()    
	defer client.deviceMutex.Unlock()  

	if client.MessageQueue == nil {
		return errors.New("message queue is not initialized")
	}

	// Calls the Publish method of the message queue
	err := client.MessageQueue.Publish(topic, message)
	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}
	return nil
}

// Thread-safe subscription to MQTT topics
func (client *CustomizedClient) SafeSubscribe(topic string) (interface{}, error) {
	client.deviceMutex.Lock()    
	defer client.deviceMutex.Unlock()  

	if client.MessageQueue == nil {
		return nil, errors.New("message queue is not initialized")
	}

	// Call the Subscribe method of the message queue
	msg, err := client.MessageQueue.Subscribe(topic)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to topic: %v", err)
	}
	return msg, nil
}

// Thread-Safe Unsubscription to MQTT Topics
func (client *CustomizedClient) SafeUnsubscribe(topic string) error {
	client.deviceMutex.Lock()  
	defer client.deviceMutex.Unlock()  

	if client.MessageQueue == nil {
		return errors.New("message queue is not initialized")
	}

	// Call the Unsubscribe method of the message queue
	err := client.MessageQueue.Unsubscribe(topic)
	if err != nil {
		return fmt.Errorf("failed to unsubscribe from topic: %v", err)
	}
	return nil
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
	"fulltextmodify": FULLTEXTMODIFY,
	"pathmodify":     PATHMODIFY,
	"valuemodify":    VALUEMODIFY,
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
func (c *ConfigData) ParseMessage(parseType SerializedFormatType) (interface{}, error) {
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
		convertedMessage, err := convertXMLToJSON(c.Message)
		if err != nil {
			return nil, err
		}
		c.Message = convertedMessage
		return c.parseJSON()

	default:
		return nil, errors.New("unsupported parse type")
	}
}

// The function parseJSON parses the Message field of the ConfigData (assumed to be a JSON string).
func (c *ConfigData) parseJSON() (interface{}, error) {
	if c.Message == "" {
		return nil, errors.New("message is empty")
	}

	var result interface{}
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

// The function convertXMLToJSON converts an XML string to a JSON string.
func convertXMLToJSON(xmlString string) (string, error) {
	xmlData, err := convertXMLToMap(xmlString)
	if err != nil {
		return "", err
	}

	jsonData, err := mapToJSON(xmlData)
	if err != nil {
		return "", err
	}

	return jsonData, nil
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
	wrappedXML := "<root>" + xmlString + "</root>"
	return wrappedXML
}

// Node structure
type Node struct {
	XMLName xml.Name
	Content string     `xml:",chardata"`
	Nodes   []Node     `xml:",any"`
	Attr    []xml.Attr `xml:"-"`
}

// The function nodeToMap recursively converts XML nodes to map[string]interface{}.
func nodeToMap(node Node) map[string]interface{} {
	result := make(map[string]interface{})

	if len(node.Nodes) == 0 {
		// Leaf node
		return map[string]interface{}{node.XMLName.Local: node.Content}
	}

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

/* --------------------------------------------------------------------------------------- */