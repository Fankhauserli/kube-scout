package event

import (
	"context"
	"fmt"
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

func (r *RabbitMQEventQueue) Consume(ctx context.Context, handler func(context.Context, string) error) error {
	if err := r.Channel.Qos(1, 0, false); err != nil {
		return err
	}
	msgs, err := r.Channel.Consume(r.Queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-msgs:
			if !ok {
				return fmt.Errorf("rabbitmq channel closed")
			}
			if err := handler(ctx, string(d.Body)); err != nil {
				// Retry once; if it already failed once (d.Redelivered == true), drop/dead-letter it
				_ = d.Nack(false, !d.Redelivered)
			} else {
				_ = d.Ack(false)
			}
		}
	}
}
