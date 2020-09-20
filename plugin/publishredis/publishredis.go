package publishredis

import (
	"context"
	"encoding/json"

	redis "github.com/go-redis/redis/v8"
	"github.com/worksinmagic/ytfeed"
)

type RedisPublisher interface {
	Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd
}

type PublishRedis struct {
	logger  ytfeed.Logger
	client  RedisPublisher
	channel string
	addr    string
}

func (p *PublishRedis) DataHandler(ctx context.Context, d *ytfeed.Data) {
	rawJSON, err := json.Marshal(d)
	if err != nil {
		p.logger.Errorf("Failed to marshal JSON: %v", err)
		return
	}

	err = p.client.Publish(ctx, p.channel, string(rawJSON)).Err()
	if err != nil {
		p.logger.Errorf("Failed to publish data `%s` to Redis at channel %s and address %s: %v", string(rawJSON), p.channel, p.addr, err)
		return
	}

	p.logger.Infof("Publish data `%s` to Redis at channel %s and address %s", string(rawJSON), p.channel, p.addr)
}

func New(logger ytfeed.Logger, channel string, opts *redis.Options) (pr *PublishRedis) {
	pr = &PublishRedis{}
	pr.logger = logger
	pr.client = redis.NewClient(opts)
	pr.channel = channel
	pr.addr = opts.Addr

	return
}
