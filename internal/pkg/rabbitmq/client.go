package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/RuLap/trackmus-api/internal/pkg/events"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	log     *slog.Logger
}

func NewClient(url string, log *slog.Logger) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	_, err = channel.QueueDeclare(
		"email_events",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &Client{
		conn:    conn,
		channel: channel,
		log:     log,
	}, nil
}

func (c *Client) PublishEmailEvent(event events.EmailEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = c.channel.PublishWithContext(
		ctx,
		"",
		"email_events",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	c.log.Debug("email event published", "to", event.To, "template", event.Template)
	return nil
}

func (c *Client) ConsumeEmailEvents(ctx context.Context, handler func(events.EmailEvent) error) error {
	msgs, err := c.channel.Consume(
		"email_events",
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to consume messages: %w", err)
	}

	c.log.Info("started consuming email events")

	for {
		select {
		case <-ctx.Done():
			c.log.Info("stopping email events consumer")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				c.log.Warn("email events channel closed")
				return nil
			}

			var event events.EmailEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				c.log.Error("failed to unmarshal email event", "error", err)
				msg.Nack(false, false)
				continue
			}

			if err := handler(event); err != nil {
				c.log.Error("failed to handle email event",
					"error", err,
					"template", event.Template,
					"to", event.To,
				)
				msg.Nack(false, true)
			} else {
				msg.Ack(false)
				c.log.Debug("email event processed successfully",
					"template", event.Template,
					"to", event.To,
				)
			}
		}
	}
}

func (c *Client) Close() error {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			c.log.Error("failed to close channel", "error", err)
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			c.log.Error("failed to close connection", "error", err)
		}
	}
	return nil
}

func (c *Client) HealthCheck() error {
	if c.conn == nil || c.conn.IsClosed() {
		return fmt.Errorf("rabbitmq connection is closed")
	}
	return nil
}
