package amqp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"ride-sharing/shared/contracts"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Publisher handles publishing messages to RabbitMQ
type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewPublisher creates a new AMQP publisher
func NewPublisher(url string) (*Publisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	if err := ch.ExchangeDeclare(
		"ride-sharing",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	return &Publisher{
		conn:    conn,
		channel: ch,
	}, nil
}

// Publish publishes a message to the specified routing key
func (p *Publisher) Publish(ctx context.Context, routingKey string, data interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	msg := contracts.AmqpMessage{
		OwnerID: "system", // Could be service-specific
		Data:    body,
	}

	msgBody, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal AMQP message: %w", err)
	}

	err = p.channel.PublishWithContext(ctx,
		"ride-sharing", // exchange
		routingKey,     // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         msgBody,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Published message to %s", routingKey)
	return nil
}

// Close closes the publisher connection
func (p *Publisher) Close() error {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

// Consumer handles consuming messages from RabbitMQ
type Consumer struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string
}

type Delivery = amqp.Delivery

// NewConsumer creates a new AMQP consumer
func NewConsumer(url, queueName string, routingKeys []string) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare the exchange
	if err := ch.ExchangeDeclare(
		"ride-sharing",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare the queue
	_, err = ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	for _, routingKey := range routingKeys {
		if err := ch.QueueBind(
			queueName,
			routingKey,
			"ride-sharing",
			false,
			nil,
		); err != nil {
			ch.Close()
			conn.Close()
			return nil, fmt.Errorf("failed to bind queue %s to routing key %s: %w", queueName, routingKey, err)
		}
	}

	return &Consumer{
		conn:      conn,
		channel:   ch,
		queueName: queueName,
	}, nil
}

// Consume starts consuming messages and calls the handler for each message
func (c *Consumer) Consume(ctx context.Context, handler func(amqp.Delivery) error) error {
	msgs, err := c.channel.Consume(
		c.queueName, // queue
		"",          // consumer
		false,       // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Printf("Started consuming from queue: %s", c.queueName)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-msgs:
			if err := handler(msg); err != nil {
				log.Printf("Error handling message: %v", err)
				// Nack the message to requeue it
				msg.Nack(false, true)
			} else {
				msg.Ack(false)
			}
		}
	}
}

// Close closes the consumer connection
func (c *Consumer) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
