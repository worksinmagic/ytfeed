package autosubscribefeed

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/worksinmagic/ytfeed"
)

const (
	DefaultTimeout   = 30 * time.Second
	DefaultHubMode   = "subscribe"
	DefaultHubVerify = "sync"

	HubTopic             = "hub.topic"
	HubCallback          = "hub.callback"
	HubMode              = "hub.mode"
	HubVerify            = "hub.verify"
	HubVerificationToken = "hub.verify_token"
	HubSecret            = "hub.secret"

	ErrResubscribeFormat = "failed to resubscribe for topic %s with error '%v'"
)

var (
	ErrFailedToSubscribeFeed = errors.New("failed to subscribe to feed")
)

type Subscriber struct {
	resubInterval     time.Duration
	targetAddr        string
	topics            []string
	callbackAddr      string
	verificationToken string
	hmacSecret        string

	logger ytfeed.Logger
	client *http.Client
}

func New(logger ytfeed.Logger, verificationToken, hmacSecret, targetAddr, callbackAddr string, topics []string, resubInterval time.Duration) (s *Subscriber) {
	s = &Subscriber{}
	s.resubInterval = resubInterval
	s.targetAddr = targetAddr
	s.callbackAddr = callbackAddr
	s.verificationToken = verificationToken
	s.hmacSecret = hmacSecret
	s.topics = topics
	s.client = &http.Client{}
	s.client.Timeout = DefaultTimeout
	s.logger = logger

	return
}

func (s *Subscriber) SetHTTPClient(c *http.Client) {
	s.client = c
}

func (s *Subscriber) Subscribe(ctx context.Context) (err error) {
	for {
		select {
		case <-time.After(s.resubInterval):
			err = s.subscribe()
			if err != nil {
				s.logger.Errorf("Failed to resubscribe feed: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

type ErrorSub struct {
	Topic string
	Err   error
}

func (s *Subscriber) subscribe() (err error) {
	failedReqs := make([]ErrorSub, 0, 8)

	for _, topic := range s.topics {
		data := url.Values{}
		data.Set(HubTopic, topic)
		data.Set(HubCallback, s.callbackAddr)
		data.Set(HubMode, DefaultHubMode)
		data.Set(HubVerify, DefaultHubVerify)
		data.Set(HubVerificationToken, s.verificationToken)
		data.Set(HubSecret, s.hmacSecret)

		var resp *http.Response
		resp, err = s.client.PostForm(s.targetAddr, data)
		if err != nil {
			failedReqs = append(failedReqs, ErrorSub{
				Topic: topic,
				Err:   err,
			})
			continue
		}
		if resp.StatusCode >= http.StatusBadRequest {
			err = errors.Wrapf(ErrFailedToSubscribeFeed, "HTTP status %d", resp.StatusCode)
			failedReqs = append(failedReqs, ErrorSub{
				Topic: topic,
				Err:   err,
			})
			continue
		}

		s.logger.Infof("Resubscribed to topic %s with callback address %s", topic, s.callbackAddr)
	}

	if len(failedReqs) > 0 {
		errMessages := make([]string, 0, len(failedReqs))
		for _, f := range failedReqs {
			errMessages = append(errMessages, fmt.Sprintf(ErrResubscribeFormat, f.Topic, f.Err))
		}

		err = errors.New(strings.Join(errMessages, ","))
	}

	return
}
