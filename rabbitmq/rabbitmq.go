package rabbitmq

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Client wraps the RabbitMQ connection
type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  Config
}

// RabbitMQClient is an alias for Client for backward compatibility
type RabbitMQClient = Client

// Config holds RabbitMQ configuration
type Config struct {
	URL         string
	Exchange    string
	QueuePrefix string
}

// NewClient creates a new RabbitMQ client
func NewClient(cfg Config) (*Client, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange
	err = channel.ExchangeDeclare(
		cfg.Exchange, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	return &Client{
		conn:    conn,
		channel: channel,
		config:  cfg,
	}, nil
}

// Publish publishes a message to the exchange
func (c *Client) Publish(ctx context.Context, routingKey string, body []byte) error {
	return c.channel.Publish(
		c.config.Exchange, // exchange
		routingKey,        // routing key
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
}

// Consume starts consuming messages from a queue
func (c *Client) Consume(queueName, routingKey string) (<-chan amqp.Delivery, error) {
	fullQueueName := fmt.Sprintf("%s.%s", c.config.QueuePrefix, queueName)

	// Declare queue
	q, err := c.channel.QueueDeclare(
		fullQueueName, // name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = c.channel.QueueBind(
		q.Name,            // queue name
		routingKey,        // routing key
		c.config.Exchange, // exchange
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	// Start consuming
	msgs, err := c.channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register consumer: %w", err)
	}

	return msgs, nil
}

// Close closes the RabbitMQ connection
func (c *Client) Close() error {
	if err := c.channel.Close(); err != nil {
		return err
	}
	return c.conn.Close()
}

// HealthCheck performs a health check
func (c *Client) HealthCheck() error {
	if c.conn.IsClosed() {
		return fmt.Errorf("connection is closed")
	}
	return nil
}

// NewRabbitMQClient creates a new RabbitMQ client (wrapper for NewClient for backward compatibility)
func NewRabbitMQClient(url string) (*RabbitMQClient, error) {
	cfg := Config{
		URL:         url,
		Exchange:    "",
		QueuePrefix: "",
	}
	return NewClient(cfg)
}

// DeclareExchange declares an exchange
func (c *Client) DeclareExchange(name, kind string) error {
	return c.channel.ExchangeDeclare(
		name,  // name
		kind,  // type
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
}

// DeclareQueue declares a queue
func (c *Client) DeclareQueue(name string) error {
	_, err := c.channel.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	return err
}

// BindQueue binds a queue to an exchange
func (c *Client) BindQueue(queueName, routingKey, exchangeName string) error {
	return c.channel.QueueBind(
		queueName,    // queue name
		routingKey,   // routing key
		exchangeName, // exchange
		false,        // no-wait
		nil,          // arguments
	)
}
