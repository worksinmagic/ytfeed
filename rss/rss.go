package rss

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	xj "github.com/basgys/goxml2json"
	"github.com/worksinmagic/ytfeed"
)

const (
	YoutubeSubscriptionTopicPrefix = "https://www.youtube.com/xml/feeds/videos.xml?channel_id="
)

func Handler(ctx context.Context, logger ytfeed.Logger, verificationToken, hmacSecret string, dataHandlers ...ytfeed.DataHandlerFunc) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		switch req.Method {
		case http.MethodGet:
			mode := req.URL.Query().Get("hub.mode")
			vtoken := req.URL.Query().Get("hub.verify_token")
			topic := req.URL.Query().Get("hub.topic")
			challenge := req.URL.Query().Get("hub.challenge")

			if mode == "unsubscribe" {
				if vtoken != verificationToken {
					w.WriteHeader(http.StatusUnauthorized)
					fmt.Fprintln(w, "UNAUTHORIZED")
					return
				}

				if !IsYoutubeSubscriptionTopic(topic) {
					w.WriteHeader(http.StatusBadRequest)
					fmt.Fprintln(w, "BAD REQUEST")
					return
				}

				logger.Infof("Unsubscribed to topic %s with challenge %s", topic, challenge)

				fmt.Fprint(w, challenge)
				return
			}
			if mode == "subscribe" {
				if vtoken != verificationToken {
					w.WriteHeader(http.StatusUnauthorized)
					fmt.Fprintln(w, "UNAUTHORIZED")
					return
				}

				if !IsYoutubeSubscriptionTopic(topic) {
					w.WriteHeader(http.StatusBadRequest)
					fmt.Fprintln(w, "BAD REQUEST")
					return
				}

				logger.Infof("Subscribed to topic %s with challenge %s", topic, challenge)

				fmt.Fprint(w, challenge)
				return
			}

			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "BAD REQUEST")
			return
		case http.MethodPost:
			tmpRaw, err := ioutil.ReadAll(req.Body)
			if err != nil {
				logger.Errorf("Failed to read body: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "INTERNAL SERVER ERROR: %v", err)
				return
			}
			originalMessage := string(tmpRaw)

			hmacHasher := hmac.New(sha1.New, []byte(hmacSecret))
			var verified bool
			verified, err = VerifyDataFeed(hmacHasher, req.Header.Get("X-Hub-Signature"), string(tmpRaw))
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprintf(w, "INVALID AUTHENTICATED CONTENT DISTRIBUTION SIGNATURE: %v", err)
				return
			}
			if !verified {
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprintf(w, "INVALID AUTHENTICATED CONTENT DISTRIBUTION SIGNATURE: not verified")
				return
			}

			jbuf, err := xj.Convert(bytes.NewReader(tmpRaw))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "INVALID XML INPUT: %v", err)
				return
			}

			data := &ytfeed.Data{}
			err = json.Unmarshal(jbuf.Bytes(), data)
			if err != nil {
				logger.Errorf("Failed to unmarshal JSON: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "FAILED TO UNMARSHAL JSON: %v", err)
				return
			}
			data.OriginalXMLMessage = originalMessage

			logger.Infof("Got subscription data: %s", data)

			for _, d := range dataHandlers {
				go d(ctx, data)
			}

			w.WriteHeader(http.StatusCreated)
			fmt.Fprintln(w, "CREATED")
			return
		default:
			_, err := io.Copy(ioutil.Discard, req.Body)
			if err != nil {
				logger.Errorf("Failed to discard unused request body: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "FAILED TO DISCARD UNUSED REQUEST BODY: %v", err)
				return
			}
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintln(w, "METHOD NOT ALLOWED")
			return
		}
	}
}

func IsYoutubeSubscriptionTopic(topic string) bool {
	return strings.HasPrefix(topic, YoutubeSubscriptionTopicPrefix)
}

func VerifyDataFeed(hmacHasher hash.Hash, hmacHeader, xmlData string) (verified bool, err error) {
	var hmacData []byte
	hmacData, err = hex.DecodeString(strings.ReplaceAll(hmacHeader, "sha1=", ""))
	if err != nil {
		return
	}
	_, err = hmacHasher.Write([]byte(xmlData))
	if err != nil {
		return
	}

	verified = hmac.Equal(hmacData, hmacHasher.Sum(nil))

	return
}
