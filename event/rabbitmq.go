package event

import (
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQEventQueue struct {
	conn    *amqp.Connection
	Channel *amqp.Channel
	Queue   amqp.Queue
}

func NewRabbitMQEventQueue(address, queueName, userEnv, passwordEnv string) (*RabbitMQEventQueue, error) {
	user := os.Getenv(userEnv)
	password := os.Getenv(passwordEnv)
	conn, err := amqp.Dial("amqp://" + user + ":" + password + "@" + address)
	if err != nil {
		return nil, err
	}
	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	queue, err := channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &RabbitMQEventQueue{
		conn:    conn,
		Channel: channel,
		Queue:   queue,
	}, nil
}

func (r *RabbitMQEventQueue) Close() {
	r.Channel.Close()
	r.conn.Close()
}
