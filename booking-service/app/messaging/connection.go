package messaging

import (
	"fmt"

	"go.uber.org/zap"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Connection управляет соединением с RabbitMQ.
type Connection struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	logger  *zap.Logger
}

// NewConnection создаёт соединение с RabbitMQ и настраивает канал.
func NewConnection(url string, prefetchCount int, logger *zap.Logger) (*Connection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("подключение к RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("открытие канала: %w", err)
	}

	if err := ch.Qos(prefetchCount, 0, false); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("настройка QoS: %w", err)
	}

	logger.Info("подключение к RabbitMQ установлено", zap.String("url", url))

	return &Connection{
		conn:    conn,
		channel: ch,
		logger:  logger,
	}, nil
}

// Channel возвращает AMQP-канал.
func (c *Connection) Channel() *amqp.Channel {
	return c.channel
}

// Close закрывает канал и соединение.
func (c *Connection) Close() error {
	if err := c.channel.Close(); err != nil {
		c.logger.Error("ошибка закрытия канала", zap.Error(err))
	}
	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("ошибка закрытия соединения: %w", err)
	}
	c.logger.Info("соединение с RabbitMQ закрыто")
	return nil
}

// DeclareExchange объявляет exchange для маршрутизации сообщений.
func (c *Connection) DeclareExchange(name string) error {
	return c.channel.ExchangeDeclare(
		name,    // имя exchange
		"topic", // тип (topic позволяет маршрутизацию по routing key)
		true,    // durable
		false,   // auto-delete
		false,   // internal
		false,   // no-wait
		nil,     // arguments
	)
}

// DeclareAndBindQueue объявляет очередь и привязывает её к exchange.
func (c *Connection) DeclareAndBindQueue(queueName, exchangeName, routingKey string) (amqp.Queue, error) {
	q, err := c.channel.QueueDeclare(
		queueName, // имя
		true,      // durable
		false,     // auto-delete
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("объявление очереди %s: %w", queueName, err)
	}

	err = c.channel.QueueBind(
		q.Name,       // queue
		routingKey,   // routing key
		exchangeName, // exchange
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("привязка очереди %s к exchange %s: %w", queueName, exchangeName, err)
	}

	return q, nil
}
