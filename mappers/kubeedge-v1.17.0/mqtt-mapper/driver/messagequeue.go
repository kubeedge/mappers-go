package driver

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// NewMqttMessageQueueWithAuth creates and initializes an MqttMessageQueue instance with authentication.
func NewMqttMessageQueueWithAuth(broker string, clientID string, username string, password string) (*MqttMessageQueue, error) {
    opts := mqtt.NewClientOptions().
        AddBroker(broker).
        SetClientID(clientID).
        SetUsername(username).  
        SetPassword(password).  
        SetConnectTimeout(10 * time.Second)  

    client := mqtt.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        return nil, token.Error()
    }

    return &MqttMessageQueue{
        client: client,
    }, nil
}


// Publish is used to publish a message to the specified topic, the type of the message is interface{} in order to support multiple message formats.
func (mq *MqttMessageQueue) Publish(topic string, message interface{}) error {
    payload, ok := message.(string)
    if !ok {
        return fmt.Errorf("message must be a string")
    }
    token := mq.client.Publish(topic, 0, false, payload)
    token.Wait()
    return token.Error()
}

// Subscribe subscribes to the specified topic, return the received message.
func (mq *MqttMessageQueue) Subscribe(topic string) (interface{}, error) {
    ch := make(chan mqtt.Message)
    token := mq.client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
        ch <- msg
    })
    token.Wait()
    if token.Error() != nil {
        return nil, token.Error()
    }
    message := <-ch
    return string(message.Payload()), nil
}

// Unsubscribe unsubscribes to the specified topic.
func (mq *MqttMessageQueue) Unsubscribe(topic string) error {
    token := mq.client.Unsubscribe(topic)
    token.Wait()
    return token.Error()
}
