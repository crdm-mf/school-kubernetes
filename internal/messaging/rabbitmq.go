package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/teko/food-delivery/internal/events"
	"github.com/teko/food-delivery/internal/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	Exchange   = "food.events"
	DeadLetter = "food.dlx"
	DeadQueue  = "food.dead"
)

// Publisher publishes confirmed persistent messages to the topic exchange.
type Publisher struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

// NewPublisher connects to RabbitMQ and declares the shared topology.
func NewPublisher(ctx context.Context, url string, logger *slog.Logger) (*Publisher, error) {
	connection, err := dial(ctx, url, logger)
	if err != nil {
		return nil, err
	}
	channel, err := connection.Channel()
	if err != nil {
		connection.Close()
		return nil, fmt.Errorf("open publisher channel: %w", err)
	}
	if err := declareExchanges(channel); err != nil {
		channel.Close()
		connection.Close()
		return nil, err
	}
	if err := channel.Confirm(false); err != nil {
		channel.Close()
		connection.Close()
		return nil, fmt.Errorf("enable publisher confirms: %w", err)
	}
	return &Publisher{connection: connection, channel: channel}, nil
}

// Publish sends one event and waits for the broker confirmation.
func (p *Publisher) Publish(ctx context.Context, event model.EventEnvelope) error {
	ctx, span := otel.Tracer("github.com/teko/food-delivery/internal/messaging").Start(ctx, "rabbitmq.publish "+event.Type, trace.WithSpanKind(trace.SpanKindProducer))
	defer span.End()
	setEventAttributes(span, event)
	span.SetAttributes(
		attribute.String("messaging.system", "rabbitmq"),
		attribute.String("messaging.destination.name", Exchange),
		attribute.String("messaging.operation.type", "publish"),
		attribute.String("messaging.rabbitmq.exchange", Exchange),
		attribute.String("messaging.rabbitmq.routing_key", events.RoutingKey(event)),
		attribute.String("messaging.message.id", event.ID),
	)
	raw, err := json.Marshal(event)
	if err != nil {
		recordError(span, err)
		return fmt.Errorf("marshal event: %w", err)
	}
	headers := amqp.Table{}
	propagation.TraceContext{}.Inject(ctx, amqpTableCarrier(headers))
	confirmation, err := p.channel.PublishWithDeferredConfirmWithContext(ctx, Exchange, events.RoutingKey(event), false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		MessageId:    event.ID,
		Timestamp:    event.OccurredAt,
		Headers:      headers,
		Body:         raw,
	})
	if err != nil {
		recordError(span, err)
		return fmt.Errorf("publish %s: %w", event.Type, err)
	}
	confirmed, err := confirmation.WaitContext(ctx)
	if err != nil {
		recordError(span, err)
		return fmt.Errorf("wait for publish confirmation: %w", err)
	}
	if !confirmed {
		err := fmt.Errorf("broker rejected event %s", event.ID)
		recordError(span, err)
		return err
	}
	return nil
}

// Close releases the AMQP resources.
func (p *Publisher) Close() error {
	if err := p.channel.Close(); err != nil {
		return fmt.Errorf("close publisher channel: %w", err)
	}
	if err := p.connection.Close(); err != nil {
		return fmt.Errorf("close publisher connection: %w", err)
	}
	return nil
}

// ConsumerConfig describes a durable or ephemeral subscription.
type ConsumerConfig struct {
	Queue      string
	Bindings   []string
	Workers    int
	Prefetch   int
	Exclusive  bool
	AutoDelete bool
}

// Handler processes one event. Returning an error dead-letters the message.
type Handler func(context.Context, model.EventEnvelope) error

// Consume connects, declares the queue and blocks until context cancellation.
func Consume(ctx context.Context, url string, config ConsumerConfig, logger *slog.Logger, handler Handler) error {
	connection, err := dial(ctx, url, logger)
	if err != nil {
		return err
	}
	defer connection.Close()
	channel, err := connection.Channel()
	if err != nil {
		return fmt.Errorf("open consumer channel: %w", err)
	}
	defer channel.Close()
	if err := declareExchanges(channel); err != nil {
		return err
	}
	arguments := amqp.Table{"x-dead-letter-exchange": DeadLetter}
	queue, err := channel.QueueDeclare(config.Queue, !config.Exclusive, config.AutoDelete, config.Exclusive, false, arguments)
	if err != nil {
		return fmt.Errorf("declare queue %s: %w", config.Queue, err)
	}
	for _, binding := range config.Bindings {
		if err := channel.QueueBind(queue.Name, binding, Exchange, false, nil); err != nil {
			return fmt.Errorf("bind queue %s to %s: %w", queue.Name, binding, err)
		}
	}
	if config.Prefetch <= 0 {
		config.Prefetch = 1
	}
	if err := channel.Qos(config.Prefetch, 0, false); err != nil {
		return fmt.Errorf("set qos: %w", err)
	}
	deliveries, err := channel.Consume(queue.Name, "", false, config.Exclusive, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume queue %s: %w", queue.Name, err)
	}
	if config.Workers <= 0 {
		config.Workers = 1
	}

	errCh := make(chan error, config.Workers)
	for worker := 0; worker < config.Workers; worker++ {
		go func(workerID int) {
			for {
				select {
				case <-ctx.Done():
					return
				case delivery, ok := <-deliveries:
					if !ok {
						errCh <- fmt.Errorf("delivery channel closed")
						return
					}
					if err := processDelivery(ctx, config.Queue, delivery, logger, handler); err != nil {
						errCh <- err
						return
					}
				}
			}
		}(worker)
	}

	select {
	case <-ctx.Done():
		return fmt.Errorf("consumer stopped: %w", ctx.Err())
	case err := <-errCh:
		return err
	}
}

func processDelivery(ctx context.Context, queue string, delivery amqp.Delivery, logger *slog.Logger, handler Handler) error {
	messageCtx := propagation.TraceContext{}.Extract(ctx, amqpTableCarrier(delivery.Headers))
	messageCtx, span := otel.Tracer("github.com/teko/food-delivery/internal/messaging").Start(messageCtx, "rabbitmq.consume", trace.WithSpanKind(trace.SpanKindConsumer))
	defer span.End()
	span.SetAttributes(
		attribute.String("messaging.system", "rabbitmq"),
		attribute.String("messaging.destination.name", queue),
		attribute.String("messaging.operation.type", "process"),
		attribute.String("messaging.rabbitmq.exchange", Exchange),
		attribute.String("messaging.rabbitmq.queue", queue),
		attribute.String("messaging.rabbitmq.routing_key", delivery.RoutingKey),
	)
	if delivery.MessageId != "" {
		span.SetAttributes(attribute.String("messaging.message.id", delivery.MessageId))
	}

	var event model.EventEnvelope
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		recordError(span, err)
		logger.Error("invalid event moved to DLQ", "queue", queue, "error", err)
		if nackErr := delivery.Nack(false, false); nackErr != nil {
			return fmt.Errorf("nack invalid message: %w", nackErr)
		}
		return nil
	}
	span.SetName("rabbitmq.consume " + event.Type)
	setEventAttributes(span, event)
	if err := handler(messageCtx, event); err != nil {
		if errors.Is(err, context.Canceled) {
			recordError(span, err)
			if nackErr := delivery.Nack(false, true); nackErr != nil {
				return fmt.Errorf("requeue interrupted event: %w", nackErr)
			}
			return nil
		}
		recordError(span, err)
		logger.Error("event handler failed", "event_id", event.ID, "event_type", event.Type, "error", err)
		if nackErr := delivery.Nack(false, false); nackErr != nil {
			return fmt.Errorf("nack event: %w", nackErr)
		}
		return nil
	}
	if err := delivery.Ack(false); err != nil {
		recordError(span, err)
		return fmt.Errorf("ack event: %w", err)
	}
	return nil
}

func setEventAttributes(span trace.Span, event model.EventEnvelope) {
	span.SetAttributes(
		attribute.String("food_delivery.event.type", event.Type),
		attribute.String("food_delivery.event.id", event.ID),
		attribute.String("food_delivery.correlation_id", event.CorrelationID),
		attribute.String("food_delivery.causation_id", event.CausationID),
	)
}

func recordError(span trace.Span, err error) {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

type amqpTableCarrier amqp.Table

func (c amqpTableCarrier) Get(key string) string {
	value, ok := c[key]
	if !ok {
		return ""
	}
	switch value := value.(type) {
	case string:
		return value
	case []byte:
		return string(value)
	default:
		return fmt.Sprint(value)
	}
}

func (c amqpTableCarrier) Set(key, value string) {
	c[key] = value
}

func (c amqpTableCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for key := range c {
		keys = append(keys, key)
	}
	return keys
}

var _ propagation.TextMapCarrier = amqpTableCarrier{}

func declareExchanges(channel *amqp.Channel) error {
	if err := channel.ExchangeDeclare(Exchange, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare event exchange: %w", err)
	}
	if err := channel.ExchangeDeclare(DeadLetter, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare dead letter exchange: %w", err)
	}
	if _, err := channel.QueueDeclare(DeadQueue, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare dead letter queue: %w", err)
	}
	if err := channel.QueueBind(DeadQueue, "#", DeadLetter, false, nil); err != nil {
		return fmt.Errorf("bind dead letter queue: %w", err)
	}
	return nil
}

func dial(ctx context.Context, url string, logger *slog.Logger) (*amqp.Connection, error) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for attempt := 1; ; attempt++ {
		connection, err := amqp.DialConfig(url, amqp.Config{Heartbeat: 10 * time.Second, Dial: amqp.DefaultDial(5 * time.Second)})
		if err == nil {
			return connection, nil
		}
		logger.Warn("waiting for RabbitMQ", "attempt", attempt, "error", err)
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("connect to RabbitMQ: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}
