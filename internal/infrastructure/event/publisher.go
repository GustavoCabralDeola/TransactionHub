package event

import (
	"context"
	"log"
	"math/rand"
	"time"
)

// Define o contrato para publicação assíncrona de eventos.
type EventPublisher interface {
	Publish(ctx context.Context, topic string, payload interface{})
}

// publica eventos com retentativas (exponential backoff).
type LogEventPublisher struct {
	maxRetries int
}

func NewLogEventPublisher(maxRetries int) *LogEventPublisher {
	return &LogEventPublisher{
		maxRetries: maxRetries,
	}
}

// Dispara o evento, tenta novamente em caso de erro, dobrando o tempo de espera.
func (p *LogEventPublisher) Publish(ctx context.Context, topic string, payload interface{}) {
	baseDelay := 1 * time.Second

	for attempt := 1; attempt <= p.maxRetries; attempt++ {
		// Simula disparo externo em rede ou mensageria
		err := p.simulateNetworkPublish(topic, payload)
		if err == nil {
			log.Printf("[EventPublisher] Successfully published event to topic '%s': %+v", topic, payload)
			return
		}

		log.Printf("[EventPublisher] Attempt %d/%d failed for topic '%s': %v", attempt, p.maxRetries, topic, err)

		if attempt == p.maxRetries {
			log.Printf("[EventPublisher] Max retries reached. Failed to publish event to topic '%s'.", topic)
			return
		}

		// O tempo de espera dobra a cada erro sucessivo (Exponential Backoff)
		delay := baseDelay * (1 << (attempt - 1))
		log.Printf("[EventPublisher] Retrying in %v...", delay)

		select {
		case <-time.After(delay):
			// continue to next attempt
		case <-ctx.Done():
			log.Printf("[EventPublisher] Context cancelled, aborting event publish for topic '%s'", topic)
			return
		}
	}
}

// Erro forçado para o teste de resiliência
type simulateError struct{}

func (e *simulateError) Error() string { return "simulated network timeout" }

func (p *LogEventPublisher) simulateNetworkPublish(topic string, payload interface{}) error {

	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	if r.Intn(10) < 2 {
		return &simulateError{}
	}
	return nil
}
