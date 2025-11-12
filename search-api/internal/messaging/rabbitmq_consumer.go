package messaging

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"search-api/config"
	"search-api/internal/domain"
	"search-api/internal/repositories"
	"time"

	"github.com/streadway/amqp"
)

type Event struct {
	Type     string `json:"type"`
	EntityID string `json:"entity_id"`
}

func StartRabbitMQConsumer(repo repositories.SolrRepository) {
	conn, err := amqp.Dial(config.AppConfig.RabbitMQURL)
	if err != nil {
		log.Fatalf("❌ Error al conectar RabbitMQ: %v", err)
	}
	ch, _ := conn.Channel()
	msgs, _ := ch.Consume(
		config.AppConfig.RabbitMQQueue,
		"",
		true, false, false, false, nil,
	)

	log.Println("📡 Escuchando eventos de RabbitMQ...")

	for msg := range msgs {
		var e Event
		if err := json.Unmarshal(msg.Body, &e); err != nil {
			log.Println("❌ Error parseando evento:", err)
			continue
		}

		switch e.Type {
		case "create", "update":
			syncCancha(repo, e.EntityID)
		case "delete":
			repo.Delete(e.EntityID)
		}
	}
}

func syncCancha(repo repositories.SolrRepository, id string) {
	url := fmt.Sprintf("%s/canchas/%s", config.AppConfig.CanchasAPIURL, id)
	resp, err := http.Get(url)
	if err != nil {
		log.Println("❌ Error al obtener cancha:", err)
		return
	}
	defer resp.Body.Close()

	var cancha domain.CanchaSearch
	if err := json.NewDecoder(resp.Body).Decode(&cancha); err != nil {
		log.Println("❌ Error decodificando cancha:", err)
		return
	}

	cancha.UpdatedAt = time.Now()
	if err := repo.AddOrUpdate(cancha); err != nil {
		log.Println("❌ Error indexando cancha:", err)
		return
	}

	log.Printf("✅ Cancha sincronizada en Solr: %s\n", cancha.ID)
}
