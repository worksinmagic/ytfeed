package publishredis

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	"github.com/worksinmagic/ytfeed"
	"github.com/worksinmagic/ytfeed/mock"
)

func TestPublishRedis(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockLogger(ctrl)
	pub := mock.NewMockRedisPubSub(ctrl)
	channel := "channelname"
	opts := &redis.Options{}
	opts.Addr = "localhost:6789"

	pr := New(logger, channel, opts)
	pr.client = pub

	t.Run("success", func(t *testing.T) {
		pub.EXPECT().Publish(
			gomock.Any(),
			gomock.AssignableToTypeOf("channel"),
			gomock.AssignableToTypeOf("data"),
		).Return(redis.NewIntResult(1, nil))

		logger.EXPECT().Infof(
			gomock.AssignableToTypeOf("published"),
			gomock.AssignableToTypeOf("data"),
			gomock.AssignableToTypeOf("channel"),
			gomock.AssignableToTypeOf("addr"),
		)

		d := &ytfeed.Data{}
		pr.DataHandler(context.TODO(), d)
	})

	t.Run("failed", func(t *testing.T) {
		pub.EXPECT().Publish(
			gomock.Any(),
			gomock.AssignableToTypeOf("channel"),
			gomock.AssignableToTypeOf("data"),
		).Return(redis.NewIntResult(1, fmt.Errorf("error")))

		logger.EXPECT().Errorf(
			gomock.AssignableToTypeOf("published"),
			gomock.AssignableToTypeOf("data"),
			gomock.AssignableToTypeOf("channel"),
			gomock.AssignableToTypeOf("addr"),
			gomock.AssignableToTypeOf(fmt.Errorf("error")),
		)

		d := &ytfeed.Data{}
		pr.DataHandler(context.TODO(), d)
	})
}
