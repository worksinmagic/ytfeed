package autosubscribefeed

import (
	"context"
	"net/http"
	"net/url"
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
)

var (
	ErrFailedToSubscribeFeed = errors.New("failed to subscribe to feed")
)

type Subscriber struct {
	resubInterval     time.Duration
	targetAddr        string
	topic             string
	callbackAddr      string
	verificationToken string

	logger ytfeed.Logger
	client *http.Client
}

func New(logger ytfeed.Logger, verificationToken, targetAddr, topic, callbackAddr string, resubInterval time.Duration) (s *Subscriber) {
	s = &Subscriber{}
	s.resubInterval = resubInterval
	s.targetAddr = targetAddr
	s.callbackAddr = callbackAddr
	s.verificationToken = verificationToken
	s.topic = topic
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

func (s *Subscriber) subscribe() (err error) {
	data := url.Values{}
	data.Set(HubTopic, s.topic)
	data.Set(HubCallback, s.callbackAddr)
	data.Set(HubMode, DefaultHubMode)
	data.Set(HubVerify, DefaultHubVerify)
	data.Set(HubVerificationToken, s.verificationToken)

	var resp *http.Response
	resp, err = s.client.PostForm(s.targetAddr, data)
	if err != nil {
		return
	}
	if resp.StatusCode >= http.StatusBadRequest {
		err = errors.Wrapf(ErrFailedToSubscribeFeed, "HTTP status %d", resp.StatusCode)
		return
	}

	s.logger.Infof("Resubscribed to topic %s with callback address %s", s.topic, s.callbackAddr)

	return
}
