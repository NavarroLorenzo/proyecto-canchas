package messaging

import (
	"canchas-api/config"
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

type RabbitMQPublisher interface {
	PublishEvent(event Event) error
	Close() error
}

type Event struct {
	Type      string      `json:"type"`      // "create", "update", "delete"
	Entity    string      `json:"entity"`    // "cancha"
	EntityID  string      `json:"entity_id"` // ID de la cancha
	Data      interface{} `json:"data"`      // Datos completos de la cancha
	Timestamp int64       `json:"timestamp"` // Unix timestamp
}

type rabbitmqPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewRabbitMQPublisher crea una nueva instancia del publicador de RabbitMQ
func NewRabbitMQPublisher() (RabbitMQPublisher, error) {
	conn, err := amqp.Dial(config.AppConfig.RabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declarar el exchange
	err = channel.ExchangeDeclare(
		config.AppConfig.RabbitMQExchange, // name
		"topic",                           // type
		true,                              // durable
		false,                             // auto-deleted
		false,                             // internal
		false,                             // no-wait
		nil,                               // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declarar la cola
	_, err = channel.QueueDeclare(
		config.AppConfig.RabbitMQQueue, // name
		true,                           // durable
		false,                          // delete when unused
		false,                          // exclusive
		false,                          // no-wait
		nil,                            // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = channel.QueueBind(
		config.AppConfig.RabbitMQQueue,    // queue name
		"cancha.*",                        // routing key (cancha.create, cancha.update, cancha.delete)
		config.AppConfig.RabbitMQExchange, // exchange
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	log.Println("RabbitMQ connection established successfully")

	return &rabbitmqPublisher{
		conn:    conn,
		channel: channel,
	}, nil
}

// PublishEvent publica un evento a RabbitMQ
func (p *rabbitmqPublisher) PublishEvent(event Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error marshaling event: %w", err)
	}

	routingKey := fmt.Sprintf("cancha.%s", event.Type)

	err = p.channel.Publish(
		config.AppConfig.RabbitMQExchange, // exchange
		routingKey,                        // routing key
		false,                             // mandatory
		false,                             // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)

	if err != nil {
		return fmt.Errorf("error publishing event: %w", err)
	}

	log.Printf("Event published: %s - %s", routingKey, event.EntityID)
	return nil
}

// Close cierra la conexión con RabbitMQ
func (p *rabbitmqPublisher) Close() error {
	if err := p.channel.Close(); err != nil {
		return err
	}
	if err := p.conn.Close(); err != nil {
		return err
	}
	log.Println("RabbitMQ connection closed")
	return nil
}
