package publishamqp

import (
	"context"
	"encoding/json"
	"time"

	"github.com/streadway/amqp"
	"github.com/worksinmagic/ytfeed"
)

const (
	DefaultContentType = "application/json"
	DefaultAppID       = "ytfeed"
)

type AMQPPublisher interface {
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	NotifyReturn(c chan amqp.Return) chan amqp.Return
}

type PublishAMQP struct {
	logger    ytfeed.Logger
	channel   AMQPPublisher
	exchange  string
	key       string
	mandatory bool
	immediate bool
	returnCh  chan amqp.Return
}

func (p *PublishAMQP) DataHandler(ctx context.Context, d *ytfeed.Data) {
	rawJSON, err := json.Marshal(d)
	if err != nil {
		p.logger.Errorf("Failed to marshal JSON: %v", err)
		return
	}

	msg := amqp.Publishing{}
	msg.Body = rawJSON
	msg.ContentType = DefaultContentType
	msg.DeliveryMode = amqp.Persistent
	msg.AppId = DefaultAppID
	msg.Timestamp = time.Now()

	err = p.channel.Publish(p.exchange, p.key, p.mandatory, p.immediate, msg)
	if err != nil {
		p.logger.Errorf("Failed to publish data `%s` to AMQP at exchange %s and key %s: %v", string(rawJSON), p.exchange, p.key, err)
		return
	}

	p.logger.Infof("Publish data `%s` to AMQP at exchange %s and key %s", string(rawJSON), p.exchange, p.key)
}

func New(logger ytfeed.Logger, channel AMQPPublisher, exchange, key string, mandatory, immediate bool) (pr *PublishAMQP) {
	pr = &PublishAMQP{}
	pr.logger = logger
	pr.channel = channel
	pr.exchange = exchange
	pr.key = key
	pr.mandatory = mandatory
	pr.immediate = immediate

	pr.returnCh = make(chan amqp.Return, 1)
	pr.returnCh = pr.channel.NotifyReturn(pr.returnCh)

	return
}
