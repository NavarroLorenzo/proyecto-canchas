package consumers

import (
	"encoding/json"
	"fmt"
	"log"
	"search-api/internal/services"

	"github.com/streadway/amqp"
)

type RabbitConsumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	service services.SearchService
}

// NewRabbitConsumer inicializa la conexión y el canal
func NewRabbitConsumer(rabbitURL string, searchService services.SearchService) (*RabbitConsumer, error) {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open RabbitMQ channel: %w", err)
	}

	return &RabbitConsumer{
		conn:    conn,
		channel: ch,
		service: searchService,
	}, nil
}

// Listen suscribe el consumidor a un exchange y cola
func (r *RabbitConsumer) Listen(exchangeName, queueName string) error {
	err := r.channel.ExchangeDeclare(
		exchangeName, // nombre del exchange
		"topic",      // tipo de exchange
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declaramos la cola donde este servicio escuchará
	q, err := r.channel.QueueDeclare(
		queueName, // nombre de la cola (debería venir del .env → search_queue)
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Enlazamos la cola al exchange (recibe todos los mensajes del tipo “cancha.*”)
	err = r.channel.QueueBind(
		q.Name,
		"cancha.*",   // routing key (pattern de mensajes)
		exchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	// Iniciamos el consumo de mensajes
	msgs, err := r.channel.Consume(
		q.Name, // cola
		"",     // consumer tag
		true,   // auto-ack (true = confirma automáticamente)
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Printf("[RabbitMQ] Listening on queue: %s (exchange: %s)", q.Name, exchangeName)

	// Procesamos los mensajes en una goroutine
	go func() {
		for d := range msgs {
			log.Printf("[RabbitMQ] Received message: %s", d.RoutingKey)

			var event struct {
				Type     string      `json:"type"`
				Entity   string      `json:"entity"`
				EntityID string      `json:"entity_id"`
				Data     interface{} `json:"data"`
			}

			if err := json.Unmarshal(d.Body, &event); err != nil {
				log.Printf("[RabbitMQ] Error decoding message: %v", err)
				continue
			}

			// Solo indexamos si es una cancha
			if event.Entity != "cancha" {
				continue
			}

			switch event.Type {
			case "delete":
				if err := r.service.DeleteCancha(event.EntityID); err != nil {
					log.Printf("[Search] Failed to delete cancha from Solr: %v", err)
				} else {
					log.Printf("[Search] Cancha deleted from Solr: %s", event.EntityID)
				}
			default:
				log.Printf("[Search] Indexing cancha from event: %s", event.Type)
				if err := r.service.IndexCancha(event.Data); err != nil {
					log.Printf("[Search] Failed to index cancha: %v", err)
				} else {
					log.Printf("[Search] Cancha indexed successfully")
				}
			}
		}
	}()

	return nil
}

// Close cierra la conexión y canal
func (r *RabbitConsumer) Close() {
	if r.channel != nil {
		_ = r.channel.Close()
	}
	if r.conn != nil {
		_ = r.conn.Close()
	}
}
