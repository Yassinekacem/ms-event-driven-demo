package events

import (
    "encoding/json"
    "fmt"
    "log"
    
    "ms-notification/database"
    "ms-notification/models"
    
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
    
    _, err = RabbitMQChannel.QueueDeclare(
        QueueName,
        true,
        false,
        false,
        false,
        nil,
    )
    if err != nil {
        return fmt.Errorf("failed to declare queue: %v", err)
    }
    
    err = RabbitMQChannel.QueueBind(
        QueueName,
        RoutingKey,
        ExchangeName,
        false,
        nil,
    )
    if err != nil {
        return fmt.Errorf("failed to bind queue: %v", err)
    }
    
    fmt.Println("Connected to RabbitMQ successfully")
    return nil
}

func StartConsumer() {
    msgs, err := RabbitMQChannel.Consume(
        QueueName,
        "",
        false,
        false,
        false,
        false,
        nil,
    )
    if err != nil {
        log.Fatal("Failed to register consumer:", err)
    }
    
    log.Println("Started consuming status change events...")
    
    forever := make(chan bool)
    
    go func() {
        for d := range msgs {
            var event models.StatusChangeEvent
            err := json.Unmarshal(d.Body, &event)
            if err != nil {
                log.Printf("Failed to unmarshal event: %v", err)
                d.Nack(false, false)
                continue
            }
            
            log.Printf("Received event: %+v", event)
            
            err = storeNotification(event)
            if err != nil {
                log.Printf("Failed to store notification: %v", err)
                d.Nack(false, true)
                continue
            }
            
            d.Ack(false)
            log.Printf("Successfully processed event for user %d", event.UserID)
        }
    }()
    
    <-forever
}

func storeNotification(event models.StatusChangeEvent) error {
    message := fmt.Sprintf("User %s status changed from %s to %s", 
        event.Email, 
        event.OldStatus, 
        event.NewStatus)
    
    query := `INSERT INTO notifications (user_id, user_email, event_type, old_status, new_status, message) 
              VALUES ($1, $2, $3, $4, $5, $6)`
    
    _, err := database.DB.Exec(query, 
        event.UserID, 
        event.Email, 
        "status_change", 
        event.OldStatus, 
        event.NewStatus, 
        message)
    
    return err
}

func CloseRabbitMQ() {
    if RabbitMQChannel != nil {
        RabbitMQChannel.Close()
    }
    if RabbitMQConn != nil {
        RabbitMQConn.Close()
    }
}