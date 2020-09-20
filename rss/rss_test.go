package rss

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/worksinmagic/ytfeed"
	"github.com/worksinmagic/ytfeed/mock"
)

const (
	sampleData = `<entry>
	<id>yt:video:nivpuSG09_E</id>
	<yt:videoId>nivpuSG09_E</yt:videoId>
	<yt:channelId>UCAzsiozXvl0GfoAodNpDSQw</yt:channelId>
	<title>Up- Shania Twain</title>
	<link rel="alternate" href="https://www.youtube.com/watch?v=nivpuSG09_E"/>
	<author>
	 <name>Komari PackoftheFallen</name>
	 <uri>https://www.youtube.com/channel/UCAzsiozXvl0GfoAodNpDSQw</uri>
	</author>
	<published>2012-08-26T01:51:49+00:00</published>
	<updated>2020-03-22T11:53:58.790881024+00:00</updated>
   </entry>`
)

func TestRss(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockLogger(ctrl)
	verificationToken := "token"
	dataHandler := func(ctx context.Context, data *ytfeed.Data) {
		require.NotNil(t, data)
	}

	handler := Handler(context.TODO(), logger, verificationToken, dataHandler)
	require.NotNil(t, handler)

	t.Run("method GET", func(t *testing.T) {
		t.Run("subscribe success", func(t *testing.T) {
			urlVal := url.Values{}
			urlVal.Set("hub.mode", "subscribe")
			urlVal.Set("hub.verify_token", verificationToken)
			urlVal.Set("hub.topic", "https://www.youtube.com/xml/feeds/videos.xml?channel_id=id")
			urlVal.Set("hub.challenge", "mychallenge")
			addr := "http://localhost:8080/?" + urlVal.Encode()
			req, err := http.NewRequest(http.MethodGet, addr, &bytes.Buffer{})
			if err != nil {
				panic(err)
			}
			rec := httptest.NewRecorder()

			logger.EXPECT().Infof(
				gomock.AssignableToTypeOf("subscribe"),
				gomock.AssignableToTypeOf("topic"),
				gomock.AssignableToTypeOf("challenge"),
			)

			handler(rec, req)

			require.Equal(t, http.StatusOK, rec.Result().StatusCode)
			require.Equal(t, "mychallenge", rec.Body.String())
		})

		t.Run("subscribe failed verification token", func(t *testing.T) {
			urlVal := url.Values{}
			urlVal.Set("hub.mode", "subscribe")
			urlVal.Set("hub.verify_token", "wrongtoken")
			urlVal.Set("hub.topic", "https://www.youtube.com/xml/feeds/videos.xml?channel_id=id")
			urlVal.Set("hub.challenge", "mychallenge")
			addr := "http://localhost:8080/?" + urlVal.Encode()
			req, err := http.NewRequest(http.MethodGet, addr, &bytes.Buffer{})
			if err != nil {
				panic(err)
			}
			rec := httptest.NewRecorder()

			handler(rec, req)

			require.Equal(t, http.StatusUnauthorized, rec.Result().StatusCode)
		})

		t.Run("subscribe failed verification token", func(t *testing.T) {
			urlVal := url.Values{}
			urlVal.Set("hub.mode", "subscribe")
			urlVal.Set("hub.verify_token", verificationToken)
			urlVal.Set("hub.topic", "wrongtopic")
			urlVal.Set("hub.challenge", "mychallenge")
			addr := "http://localhost:8080/?" + urlVal.Encode()
			req, err := http.NewRequest(http.MethodGet, addr, &bytes.Buffer{})
			if err != nil {
				panic(err)
			}
			rec := httptest.NewRecorder()

			handler(rec, req)

			require.Equal(t, http.StatusBadRequest, rec.Result().StatusCode)
		})

		t.Run("unsubscribe success", func(t *testing.T) {
			urlVal := url.Values{}
			urlVal.Set("hub.mode", "unsubscribe")
			urlVal.Set("hub.verify_token", verificationToken)
			urlVal.Set("hub.topic", "https://www.youtube.com/xml/feeds/videos.xml?channel_id=id")
			urlVal.Set("hub.challenge", "mychallenge")
			addr := "http://localhost:8080/?" + urlVal.Encode()
			req, err := http.NewRequest(http.MethodGet, addr, &bytes.Buffer{})
			if err != nil {
				panic(err)
			}
			rec := httptest.NewRecorder()

			logger.EXPECT().Infof(
				gomock.AssignableToTypeOf("unsubscribe"),
				gomock.AssignableToTypeOf("topic"),
				gomock.AssignableToTypeOf("challenge"),
			)

			handler(rec, req)

			require.Equal(t, http.StatusOK, rec.Result().StatusCode)
			require.Equal(t, "mychallenge", rec.Body.String())
		})

		t.Run("unsubscribe failed verification token", func(t *testing.T) {
			urlVal := url.Values{}
			urlVal.Set("hub.mode", "unsubscribe")
			urlVal.Set("hub.verify_token", "wrongtoken")
			urlVal.Set("hub.topic", "https://www.youtube.com/xml/feeds/videos.xml?channel_id=id")
			urlVal.Set("hub.challenge", "mychallenge")
			addr := "http://localhost:8080/?" + urlVal.Encode()
			req, err := http.NewRequest(http.MethodGet, addr, &bytes.Buffer{})
			if err != nil {
				panic(err)
			}
			rec := httptest.NewRecorder()

			handler(rec, req)

			require.Equal(t, http.StatusUnauthorized, rec.Result().StatusCode)
		})

		t.Run("unsubscribe failed topic", func(t *testing.T) {
			urlVal := url.Values{}
			urlVal.Set("hub.mode", "unsubscribe")
			urlVal.Set("hub.verify_token", verificationToken)
			urlVal.Set("hub.topic", "wrongtopic")
			urlVal.Set("hub.challenge", "mychallenge")
			addr := "http://localhost:8080/?" + urlVal.Encode()
			req, err := http.NewRequest(http.MethodGet, addr, &bytes.Buffer{})
			if err != nil {
				panic(err)
			}
			rec := httptest.NewRecorder()

			handler(rec, req)

			require.Equal(t, http.StatusBadRequest, rec.Result().StatusCode)
		})

		t.Run("wrong mode", func(t *testing.T) {
			urlVal := url.Values{}
			urlVal.Set("hub.mode", "wrongmode")
			addr := "http://localhost:8080/?" + urlVal.Encode()
			req, err := http.NewRequest(http.MethodGet, addr, &bytes.Buffer{})
			if err != nil {
				panic(err)
			}
			rec := httptest.NewRecorder()

			handler(rec, req)

			require.Equal(t, http.StatusBadRequest, rec.Result().StatusCode)
		})
	})

	t.Run("method POST", func(t *testing.T) {
		t.Run("feed success", func(t *testing.T) {
			body := bytes.NewBufferString(sampleData)
			addr := "http://localhost:8080/"
			req, err := http.NewRequest(http.MethodPost, addr, body)
			if err != nil {
				panic(err)
			}
			rec := httptest.NewRecorder()

			logger.EXPECT().Infof(
				gomock.AssignableToTypeOf("feed data"),
				gomock.AssignableToTypeOf(&ytfeed.Data{}),
			)

			handler(rec, req)

			require.Equal(t, http.StatusCreated, rec.Result().StatusCode)
		})

		t.Run("feed failed invalid XML", func(t *testing.T) {
			body := bytes.NewBufferString(`{"data":"data"}`)
			addr := "http://localhost:8080/"
			req, err := http.NewRequest(http.MethodPost, addr, body)
			if err != nil {
				panic(err)
			}
			rec := httptest.NewRecorder()

			logger.EXPECT().Errorf(
				gomock.AssignableToTypeOf("feed data json unmarshal"),
				gomock.Any(),
			)

			handler(rec, req)

			require.Equal(t, http.StatusBadRequest, rec.Result().StatusCode)
		})
	})

	t.Run("method failed", func(t *testing.T) {
		addr := "http://localhost:8080/"
		req, err := http.NewRequest(http.MethodDelete, addr, &bytes.Buffer{})
		if err != nil {
			panic(err)
		}
		rec := httptest.NewRecorder()

		handler(rec, req)

		require.Equal(t, http.StatusMethodNotAllowed, rec.Result().StatusCode)
	})
}
