package kafka

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"os"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/scram"
	"go.uber.org/zap"
)

// ClickEvent is the payload published to Kafka on every redirect.
type ClickEvent struct {
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	IP          string    `json:"ip"`
	UserAgent   string    `json:"user_agent"`
	Referer     string    `json:"referer"`
	ClickedAt   time.Time `json:"clicked_at"`
}

const TopicClickEvents = "click-events"

// Producer wraps a franz-go client.
type Producer struct {
	client *kgo.Client
	logger *zap.Logger
}

// NewProducer creates a new Kafka producer.
func NewProducer(brokers []string, logger *zap.Logger) (*Producer, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
		kgo.RequiredAcks(kgo.AllISRAcks()),
		kgo.ProducerBatchMaxBytes(1 << 20),
		kgo.RecordDeliveryTimeout(5 * time.Second),
	}

	user := os.Getenv("KAFKA_USERNAME")
	pass := os.Getenv("KAFKA_PASSWORD")

	if user != "" && pass != "" {
		// Redpanda Cloud / Upstash require SASL SCRAM-SHA-256 over TLS
		opts = append(opts,
			kgo.DialTLSConfig(&tls.Config{MinVersion: tls.VersionTLS12}),
			kgo.SASL(scram.Auth{
				User: user,
				Pass: pass,
			}.AsSha256Mechanism()),
		)
		logger.Info("kafka producer: SASL_SSL (SCRAM-SHA-256) enabled",
			zap.Int("user_len", len(user)),
			zap.Int("pass_len", len(pass)),
		)
	} else {
		logger.Info("kafka producer: plain-text mode (no SASL)")
	}

	cl, err := kgo.NewClient(opts...)
	if err != nil {
		logger.Error("kafka producer: failed to create client", zap.Error(err))
		return nil, err
	}
	return &Producer{client: cl, logger: logger}, nil
}

// PublishClick serialises a ClickEvent and produces it asynchronously.
func (p *Producer) PublishClick(event ClickEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		p.logger.Error("kafka: failed to marshal click event", zap.Error(err))
		return
	}

	record := &kgo.Record{
		Topic: TopicClickEvents,
		Key:   []byte(event.ShortCode),
		Value: data,
	}

	p.client.Produce(context.Background(), record, func(r *kgo.Record, err error) {
		if err != nil {
			p.logger.Warn("kafka: failed to deliver click event ("+err.Error()+")",
				zap.String("short_code", event.ShortCode),
			)
		}
	})
}

func (p *Producer) Close() {
	if err := p.client.Flush(context.Background()); err != nil {
		p.logger.Warn("kafka: flush on close failed", zap.Error(err))
	}
	p.client.Close()
}
