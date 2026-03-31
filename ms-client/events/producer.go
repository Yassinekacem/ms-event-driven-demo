package events

import (
    "encoding/json"
    "fmt"
    "log"
    "time"
    
    "ms-client/models"
    
    "github.com/streadway/amqp"
)

var RabbitMQConn *amqp.Connection
var RabbitMQChannel *amqp.Channel

const (
    ExchangeName = "status_change_exchange"
    QueueName    = "status_change_queue"
    RoutingKey   = "status.changed"
)

func ConnectRabbitMQ() error {
    var err error
    RabbitMQConn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
    if err != nil {
        return fmt.Errorf("failed to connect to RabbitMQ: %v", err)
    }
    
    RabbitMQChannel, err = RabbitMQConn.Channel()
    if err != nil {
        return fmt.Errorf("failed to open channel: %v", err)
    }
    
    err = RabbitMQChannel.ExchangeDeclare(
        ExchangeName,
        "topic",
        true,
        false,
        false,
        false,
        nil,
    )
    if err != nil {
        return fmt.Errorf("failed to declare exchange: %v", err)
    }
    
    fmt.Println("Connected to RabbitMQ successfully")
    return nil
}

func PublishStatusChangeEvent(event models.StatusChangeEvent) error {
    event.Timestamp = time.Now().Format(time.RFC3339)
    
    body, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("failed to marshal event: %v", err)
    }
    
    err = RabbitMQChannel.Publish(
        ExchangeName,
        RoutingKey,
        false,
        false,
        amqp.Publishing{
            ContentType:  "application/json",
            Body:         body,
            DeliveryMode: amqp.Persistent,
            Timestamp:    time.Now(),
        },
    )
    
    if err != nil {
        return fmt.Errorf("failed to publish event: %v", err)
    }
    
    log.Printf("Published status change event: %+v", event)
    return nil
}

func CloseRabbitMQ() {
    if RabbitMQChannel != nil {
        RabbitMQChannel.Close()
    }
    if RabbitMQConn != nil {
        RabbitMQConn.Close()
    }
}