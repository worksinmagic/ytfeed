package publishamqp

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/streadway/amqp"
	"github.com/worksinmagic/ytfeed"
	"github.com/worksinmagic/ytfeed/mock"
)

func TestPublishAMQP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockLogger(ctrl)
	pub := mock.NewMockAMQPPublisher(ctrl)
	exchange := "exchange"
	key := "key"
	mandatory := true
	immediate := true
	returnCh := make(chan amqp.Return, 1)

	pub.EXPECT().NotifyReturn(gomock.AssignableToTypeOf(returnCh)).Return(returnCh)
	pr := New(logger, pub, exchange, key, mandatory, immediate)

	t.Run("success", func(t *testing.T) {
		logger.EXPECT().Infof(
			gomock.AssignableToTypeOf("success"),
			gomock.AssignableToTypeOf("json string"),
			gomock.AssignableToTypeOf(exchange),
			gomock.AssignableToTypeOf(key),
		)

		pub.EXPECT().Publish(
			gomock.AssignableToTypeOf(exchange),
			gomock.AssignableToTypeOf(key),
			gomock.AssignableToTypeOf(mandatory),
			gomock.AssignableToTypeOf(immediate),
			gomock.AssignableToTypeOf(amqp.Publishing{}),
		).Return(nil)

		d := &ytfeed.Data{}
		pr.DataHandler(context.TODO(), d)
	})

	t.Run("failed", func(t *testing.T) {
		logger.EXPECT().Errorf(
			gomock.AssignableToTypeOf("error"),
			gomock.AssignableToTypeOf("json string"),
			gomock.AssignableToTypeOf(exchange),
			gomock.AssignableToTypeOf(key),
			gomock.AssignableToTypeOf(fmt.Errorf("error")),
		)

		pub.EXPECT().Publish(
			gomock.AssignableToTypeOf(exchange),
			gomock.AssignableToTypeOf(key),
			gomock.AssignableToTypeOf(mandatory),
			gomock.AssignableToTypeOf(immediate),
			gomock.AssignableToTypeOf(amqp.Publishing{}),
		).Return(fmt.Errorf("error"))

		d := &ytfeed.Data{}
		pr.DataHandler(context.TODO(), d)
	})
}
