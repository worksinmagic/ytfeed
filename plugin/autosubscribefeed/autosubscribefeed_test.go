package autosubscribefeed

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/worksinmagic/ytfeed/mock"
)

func TestSubscriber(t *testing.T) {
	wrongTopic := "wrongtopic"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			panic(err)
		}

		if r.Form.Get(HubTopic) == wrongTopic {
			w.WriteHeader(http.StatusBadRequest)
		}

		fmt.Fprintln(w, "OK")
	}))
	defer ts.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockLogger(ctrl)
	verificationToken := "mytoken"
	secret := "mysecret"
	targetAddr := ts.URL
	topics := []string{"mytopic", "yourtopic"}
	callbackAddr := "http://localhost:9876"
	resubInterval := 100 * time.Millisecond

	t.Run("success", func(t *testing.T) {
		s := New(logger, verificationToken, secret, targetAddr, callbackAddr, topics, resubInterval)
		require.NotNil(t, s)

		customHTTPClient := &http.Client{}
		s.SetHTTPClient(customHTTPClient)

		ctx, cancel := context.WithTimeout(context.TODO(), 200*time.Millisecond)
		defer cancel()

		// do this twice because we have two topics
		logger.EXPECT().Infof(
			gomock.AssignableToTypeOf("resubscribed"),
			gomock.AssignableToTypeOf("topic"),
			gomock.AssignableToTypeOf("callback address"),
		)
		logger.EXPECT().Infof(
			gomock.AssignableToTypeOf("resubscribed"),
			gomock.AssignableToTypeOf("topic"),
			gomock.AssignableToTypeOf("callback address"),
		)

		err := s.Subscribe(ctx)
		require.NoError(t, err)
	})

	t.Run("failed", func(t *testing.T) {
		topics := []string{wrongTopic}
		s := New(logger, verificationToken, secret, targetAddr, callbackAddr, topics, resubInterval)
		require.NotNil(t, s)

		customHTTPClient := &http.Client{}
		s.SetHTTPClient(customHTTPClient)

		ctx, cancel := context.WithTimeout(context.TODO(), 150*time.Millisecond)
		defer cancel()

		logger.EXPECT().Errorf(
			gomock.AssignableToTypeOf("failed"),
			gomock.Any(),
		)

		err := s.Subscribe(ctx)
		require.Error(t, err)
	})
}
