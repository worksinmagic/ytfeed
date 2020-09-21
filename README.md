# ytfeed
[![Build Status](https://travis-ci.org/worksinmagic/ytfeed.svg?branch=master)](https://travis-ci.org/worksinmagic/ytfeed)
[![codecov](https://codecov.io/gh/worksinmagic/ytfeed/branch/master/graph/badge.svg)](https://codecov.io/gh/worksinmagic/ytfeed)
[![GoDoc](https://godoc.org/github.com/worksinmagic/ytfeed?status.svg)](https://godoc.org/github.com/worksinmagic/ytfeed)
[![Go Report Card](https://goreportcard.com/badge/github.com/worksinmagic/ytfeed)](https://goreportcard.com/report/github.com/worksinmagic/ytfeed)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fworksinmagic%2Fytfeed.svg?type=small)](https://app.fossa.io/projects/git%2Bgithub.com%2Fworksinmagic%2Fytfeed?ref=badge_small)
[![Donate](https://img.shields.io/badge/Donate-PayPal-green.svg)](https://paypal.me/didasy)

Automatic Youtube video and stream archiver.

## Note

- You **MUST** install [youtube-dl](https://github.com/ytdl-org/youtube-dl) first and **MUST** be available at `$PATH`.
- You **MUST** install [ffmpeg](https://ffmpeg.org/) too for this to work.
- To run s3 test or complete suite test, you need to run `test_s3.sh` for it to start a Minio server. You can stop it or delete it after testing by running `test_s3_teardown.sh`
- You also need to set `YTFEED_YOUTUBE_API_KEY`, `YTFEED_YOUTUBE_VIDEO_ID`, and `YTFEED_YOUTUBE_VIDEO_URL` to do a complete suite test.
- You can redo failed download by sending a `POST` request with the XML message in the log as the body.
- If you are planning to download live broadcast, the downloaded video will be in the format `mp4` regardless of your extension config.

## Before Using

- You have to run the program first and make sure it could be reached from the outside, then go [here](https://pubsubhubbub.appspot.com/subscribe) to subscribe to a channel.
- If the program got killed because of OOM, you can turn on swap file if you cannot raise the machine's memory. Or you can PR me a better way to handle the upload.

## Configuration

Set configuration by setting up environment variables.

- `YTFEED_YOUTUBE_API_KEY` Youtube API Key, required.
- `YTFEED_VERIFICATION_TOKEN` Verification token used to subscribe and unsubscribe topics, required.
- `YTFEED_RESUB_CALLBACK_ADDR` Callback address to ytfeed, required.
- `YTFEED_RESUB_TARGET_ADDR` The subscription page of pubsubhubbub. Defaulted to [this.](https://pubsubhubbub.appspot.com/subscribe)
- `YTFEED_RESUB_TOPIC` The topic the subscription should subscribe to, for example `https://www.youtube.com/xml/feeds/videos.xml?channel_id=mychannelid`, required.
- `YTFEED_RESUB_INTERVAL` The interval between resubscription, defaulted to 3 days or `72h`.
- `YTFEED_STORAGE_BACKEND` The storage backend, required. Must be one of `disk`, `gcs`, or `s3`.
- `YTFEED_S3_ENDPOINT` The S3 compliant server endpoint, required if `YTFEED_STORAGE_BACKEND` is `s3`.
- `YTFEED_S3_ACCESS_KEY_ID` The access key id for the S3 compliant server, required if `YTFEED_STORAGE_BACKEND` is `s3`.
- `YTFEED_S3_SECRET_ACCESS_KEY` The secret access key id for the S3 compliant server, required if `YTFEED_STORAGE_BACKEND` is `s3`.
- `YTFEED_S3_BUCKET_NAME` The bucket name for the S3 compliant server, required if `YTFEED_STORAGE_BACKEND` is `s3`.
- `YTFEED_GCS_CREDENTIAL_JSON_FILE_PATH` The JSON credential file for GCS, only used if `YTFEED_STORAGE_BACKEND` is `gcs`.
- `YTFEED_GCS_BUCKET_NAME` The bucket name for GCS, required if `YTFEED_STORAGE_BACKEND` is `gcs`.
- `YTFEED_DISK_DIRECTORY` The disk directory path, required if `YTFEED_STORAGE_BACKEND` is `disk`.
- `YTFEED_FILENAME_TEMPLATE` The filename template, defaulted to `{{.ChannelID}}/{{.PublishedYear}}/{{.PublishedMonth}}/{{.PublishedDay}}/{{.PublishedTimeZone}}/{{.VideoID}}.{{.VideoExtension}}`. The usable variables are `.ChannelID`, `.VideoID`, `.Published`, `.Title`, `.PublishedYear`, `.PublishedMonth`, `.PublishedDay`, `.PublishedHour`, `.PublishedMinute`, `.PublishedSecond`, `.PublishedNanosecond`, `.PublishedTimeZone`, `.PublishedTimeZoneOffsetSeconds`, `.VideoQuality`, `.VideoExtension`, and `.Author`.
- `YTFEED_HOST` The host address, defaulted to `:8123`.
- `YTFEED_VIDEO_FORMAT_QUALITY` The quality of the video to download, must be one of `1080`, `720`, `640`, `480`, `360`, `240`, or `144` and defaulted to `720`. 
- `YTFEED_VIDEO_FORMAT_EXTENSION` The extension of the video to download, defaulted to `webm`.
- `YTFEED_VIDEO_DOWNLOAD_RETRY_DELAY` Delay time when retrying, set to activate retries.
- `YTFEED_VIDEO_DOWNLOAD_MAX_RETRIES` Maximum retries before giving up, defaulted to `5`
- `YTFEED_REDIS_ADDR` Redis address, required if you want to publish the data to Redis PubSub.
- `YTFEED_REDIS_USERNAME`
- `YTFEED_REDIS_PASSWORD`   
- `YTFEED_REDIS_CHANNEL` Redis publish channel, defaulted to `ytfeed`.
- `YTFEED_REDIS_DB` 
- `YTFEED_REDIS_MAX_RETRIES`
- `YTFEED_REDIS_DIAL_TIMEOUT`
- `YTFEED_REDIS_WRITE_TIMEOUT`
- `YTFEED_REDIS_READ_TIMEOUT`
- `YTFEED_REDIS_POOL_SIZE`
- `YTFEED_REDIS_MIN_IDLE_CONNS`
- `YTFEED_REDIS_MAX_CONN_AGE`
- `YTFEED_REDIS_POOL_TIMEOUT`
- `YTFEED_REDIS_IDLE_TIMEOUT`
- `YTFEED_REDIS_IDLE_CHECK_FREQUENCY`
- `YTFEED_BOLTDB_PATH` Set this to a file path if you want to activate stream scheduler.
- `YTFEED_STREAM_SCHEDULER_RETRY_DELAY` Retry delay. Defaulted to 1 minute.
- `YTFEED_STREAM_SCHEDULER_MAX_RETRIES` Defaulted to 5 retries.
- `YTFEED_STREAM_SCHEDULER_WORKER_INTERVAL` Defaulted to 1 minute.
- `YTFEED_AMQP_DSN` AMQP DSN, required if you want to publish the data to AMQP broker.
- `YTFEED_AMQP_EXCHANGE` Defaulted to `ytfeed`
- `YTFEED_AMQP_KEY` Defaulted to `schedule`
- `YTFEED_AMQP_PUBLISH_MANDATORY` Defaulted to true
- `YTFEED_AMQP_PUBLISH_IMMEDIATE` Defaulted to false
- `YTFEED_AMQP_EXCHANGE_KIND` Defaulted to `topic`
- `YTFEED_AMQP_EXCHANGE_DURABLE` Defaulted to true
- `YTFEED_AMQP_EXCHANGE_INTERNAL` Defaulted to false
- `YTFEED_AMQP_EXCHANGE_AUTO_DELETE` Defaulted to false 
- `YTFEED_AMQP_EXCHANGE_NO_WAIT` Defaulted to false

## Building

Your ol' plain `go build cmd/ytfeed/main.go`

## Usage

Either create `.env` or use shellscript to set the required environment variables and run the binary like `./ytfeed`
and you'll see log messages if it runs.

## How to Contribute

- Keep your code super simple and clean.
- Add comments to your code if you have to explain what your code is doing.
- Make sure you have Commitizen CLI installed and your commit message must be written through `git cz`.
