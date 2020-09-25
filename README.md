# ytfeed
[![Build Status](https://travis-ci.org/worksinmagic/ytfeed.svg?branch=master)](https://travis-ci.org/worksinmagic/ytfeed)
[![codecov](https://codecov.io/gh/worksinmagic/ytfeed/branch/master/graph/badge.svg)](https://codecov.io/gh/worksinmagic/ytfeed)
[![GoDoc](https://godoc.org/github.com/worksinmagic/ytfeed?status.svg)](https://godoc.org/github.com/worksinmagic/ytfeed)
[![Go Report Card](https://goreportcard.com/badge/github.com/worksinmagic/ytfeed)](https://goreportcard.com/report/github.com/worksinmagic/ytfeed)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fworksinmagic%2Fytfeed.svg?type=small)](https://app.fossa.io/projects/git%2Bgithub.com%2Fworksinmagic%2Fytfeed?ref=badge_small)
[![Donate](https://img.shields.io/badge/Donate-PayPal-green.svg)](https://paypal.me/didasy)

Automatic Youtube video and stream archiver.

## Configuration

Set configuration by setting up environment variables.

| Environment Variable                 | Description                                                                                                                                                                                                                                                                                                                                           | Default Value                                                                                                                     | Is Required |
|--------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------|-------------|
|        YTFEED_YOUTUBE_API_KEY        | Youtube Data API Key.                                                                                                                                                                                                                                                                                                                       |                                                                                                                                   | true        |
|       YTFEED_VERIFICATION_TOKEN      | Verification token used to subscribe and unsubscribe topics.                                                                                                                                                                                                                                                                                |                                                                                                                                   | true        |
|      YTFEED_VERIFICATION_SECRET      | Hmac secret used to subscribe and unsubscribe topics.                                                                                                                                                                                                                                                                                       |                                                                                                                                   | true        |
|      YTFEED_RESUB_CALLBACK_ADDR      | Callback address to ytfeed.                                                                                                                                                                                                                                                                                                                 |                                                                                                                                   | true        |
|       YTFEED_RESUB_TARGET_ADDR       | The subscription page of pubsubhubbub.                                                                                                                                                                                                                                                                                                                | `https://pubsubhubbub.appspot.com/subscribe`                                                                                      |             |
|          YTFEED_RESUB_TOPIC          | The topic the subscription should subscribe to, for example `https://www.youtube.com/xml/feeds/videos.xml?channel_id=mychannelid`, can be space separated for multiple topics, required.                                                                                                                                                              |                                                                                                                                   | true        |
|         YTFEED_RESUB_INTERVAL        | The interval between resubscription.                                                                                                                                                                                                                                                                                                                  | `72h`                                                                                                                             |             |
|        YTFEED_STORAGE_BACKEND        | The storage backend, required. Must be one of `disk`, `gcs`, or `s3`.                                                                                                                                                                                                                                                                                 |                                                                                                                                   | true        |
|          YTFEED_S3_ENDPOINT          | The S3 compliant server endpoint, required if `YTFEED_STORAGE_BACKEND` is `s3`.                                                                                                                                                                                                                                                                       |                                                                                                                                   |             |
|        YTFEED_S3_ACCESS_KEY_ID       | The secret access key id for the S3 compliant server, required if  `YTFEED_STORAGE_BACKEND`  is  `s3` .                                                                                                                                                                                                                                               |                                                                                                                                   |             |
|      YTFEED_S3_SECRET_ACCESS_KEY     | The secret access key id for the S3 compliant server, required if  `YTFEED_STORAGE_BACKEND`  is  `s3` .                                                                                                                                                                                                                                               |                                                                                                                                   |             |
|         YTFEED_S3_BUCKET_NAME        | The bucket name for the S3 compliant server, required if  `YTFEED_STORAGE_BACKEND`  is  `s3` .                                                                                                                                                                                                                                                        |                                                                                                                                   |             |
| YTFEED_GCS_CREDENTIAL_JSON_FILE_PATH | The JSON credential file for GCS, only used if  `YTFEED_STORAGE_BACKEND`  is  `gcs` .                                                                                                                                                                                                                                                                 |                                                                                                                                   |             |
|        YTFEED_GCS_BUCKET_NAME        | The bucket name for GCS, required if  `YTFEED_STORAGE_BACKEND`  is  `gcs` .                                                                                                                                                                                                                                                                           |                                                                                                                                   |             |
|         YTFEED_DISK_DIRECTORY        | The disk directory path, required if `YTFEED_STORAGE_BACKEND` is `disk`.                                                                                                                                                                                                                                                                              |                                                                                                                                   |             |
|       YTFEED_FILENAME_TEMPLATE       | The filename template. The usable variables are `.ChannelID`, `.VideoID`, `.Published`, `.Title`, `.PublishedYear`, `.PublishedMonth`, `.PublishedDay`, `.PublishedHour`, `.PublishedMinute`, `.PublishedSecond`, `.PublishedNanosecond`, `.PublishedTimeZone`, `.PublishedTimeZoneOffsetSeconds`, `.VideoQuality`, `.VideoExtension`, and `.Author`. | `{{.ChannelID}}/{{.PublishedYear}}/{{.PublishedMonth}}/{{.PublishedDay}}/{{.PublishedTimeZone}}/{{.VideoID}}.{{.VideoExtension}}` |             |
|              YTFEED_HOST             | The host address.                                                                                                                                                                                                                                                                                                                                     | `:8123`                                                                                                                           |             |
|      YTFEED_VIDEO_FORMAT_QUALITY     | The quality of the video to download, must be one of `1080`, `720`, `640`, `480`, `360`, `240`, or `144`.                                                                                                                                                                                                                                             | `720`                                                                                                                             |             |
|     YTFEED_VIDEO_FORMAT_EXTENSION    | The extension of the video to download.                                                                                                                                                                                                                                                                                                               | `webm`                                                                                                                            |             |
|   YTFEED_VIDEO_DOWNLOAD_RETRY_DELAY  | Delay time when retrying, set to activate retries. Must be Golang time duration string. Example: `5m`                                                                                                                                                                                                                                                 |                                                                                                                                   |             |
|   YTFEED_VIDEO_DOWNLOAD_MAX_RETRIES  | Maximum retries before giving up.                                                                                                                                                                                                                                                                                                                     | `5`                                                                                                                               |             |
|           YTFEED_REDIS_ADDR          | Redis address, required if you want to publish the data to Redis PubSub.                                                                                                                                                                                                                                                                              |                                                                                                                                   |             |
|         YTFEED_REDIS_USERNAME        |                                                                                                                                                                                                                                                                                                                                                       |                                                                                                                                   |             |
|         YTFEED_REDIS_PASSWORD        |                                                                                                                                                                                                                                                                                                                                                       |                                                                                                                                   |             |
|         YTFEED_REDIS_CHANNEL         | Redis publish channel.                                                                                                                                                                                                                                                                                                                                | `ytfeed`                                                                                                                          |             |
|            YTFEED_REDIS_DB           |                                                                                                                                                                                                                                                                                                                                                       |                                                                                                                                   |             |
|       YTFEED_REDIS_MAX_RETRIES       |                                                                                                                                                                                                                                                                                                                                                       |                                                                                                                                   |             |
|       YTFEED_REDIS_DIAL_TIMEOUT      |                                                                                                                                                                                                                                                                                                                                                       |                                                                                                                                   |             |
|      YTFEED_REDIS_WRITE_TIMEOUT      |                                                                                                                                                                                                                                                                                                                                                       |                                                                                                                                   |             |
|       YTFEED_REDIS_READ_TIMEOUT      |                                                                                                                                                                                                                                                                                                                                                       |                                                                                                                                   |             |
|        YTFEED_REDIS_POOL_SIZE        |                                                                                                                                                                                                                                                                                                                                                       |                                                                                                                                   |             |
|      YTFEED_REDIS_MIN_IDLE_CONNS     |                                                                                                                                                                                                                                                                                                                                                       |                                                                                                                                   |             |
|       YTFEED_REDIS_MAX_CONN_AGE      |                                                                                                                                                                                                                                                                                                                                                       |                                                                                                                                   |             |
|       YTFEED_REDIS_POOL_TIMEOUT      |                                                                                                                                                                                                                                                                                                                                                       |                                                                                                                                   |             |
|       YTFEED_REDIS_IDLE_TIMEOUT      |                                                                                                                                                                                                                                                                                                                                                       |                                                                                                                                   |             |
|   YTFEED_REDIS_IDLE_CHECK_FREQUENCY  |                                                                                                                                                                                                                                                                                                                                                       |                                                                                                                                   |             |
|          YTFEED_BOLTDB_PATH          | Set this to a file path if you want to activate stream scheduler.                                                                                                                                                                                                                                                                                     |                                                                                                                                   |             |
|  YTFEED_STREAM_SCHEDULER_RETRY_DELAY | Retry delay of the scheduler.                                                                                                                                                                                                                                                                                                                         | `1m`                                                                                                                              |             |
|            YTFEED_AMQP_DSN           | AMQP DSN, required if you want to publish the data to AMQP broker.                                                                                                                                                                                                                                                                                    |                                                                                                                                   |             |
|         YTFEED_AMQP_EXCHANGE         |                                                                                                                                                                                                                                                                                                                                                       | `ytfeed`                                                                                                                          |             |
|            YTFEED_AMQP_KEY           |                                                                                                                                                                                                                                                                                                                                                       | `schedule`                                                                                                                        |             |
|     YTFEED_AMQP_PUBLISH_MANDATORY    |                                                                                                                                                                                                                                                                                                                                                       | `true`                                                                                                                            |             |
|     YTFEED_AMQP_PUBLISH_IMMEDIATE    |                                                                                                                                                                                                                                                                                                                                                       | `false`                                                                                                                           |             |
|       YTFEED_AMQP_EXCHANGE_KIND      |                                                                                                                                                                                                                                                                                                                                                       | `topic`                                                                                                                           |             |
|     YTFEED_AMQP_EXCHANGE_DURABLE     |                                                                                                                                                                                                                                                                                                                                                       | `true`                                                                                                                            |             |
|     YTFEED_AMQP_EXCHANGE_INTERNAL    |                                                                                                                                                                                                                                                                                                                                                       | `false`                                                                                                                           |             |
|   YTFEED_AMQP_EXCHANGE_AUTO_DELETE   |                                                                                                                                                                                                                                                                                                                                                       | `false`                                                                                                                           |             |
|     YTFEED_AMQP_EXCHANGE_NO_WAIT     |                                                                                                                                                                                                                                                                                                                                                       | `false`                                                                                                                           |             |

Example of fairly common configuration is:

```
export YTFEED_VERIFICATION_TOKEN='sometoken'
export YTFEED_VERIFICATION_SECRET='somesecret' 
export YTFEED_RESUB_TOPIC="https://www.youtube.com/xml/feeds/videos.xml?channel_id=channelid"
export YTFEED_RESUB_CALLBACK_ADDR='https://my.server.addr/path'
export YTFEED_STORAGE_BACKEND='s3'
export YTFEED_S3_ENDPOINT='my.s3.endpoint'
export YTFEED_S3_ACCESS_KEY_ID='acceskeyid'
export YTFEED_S3_SECRET_ACCESS_KEY='secretaccesskey'
export YTFEED_S3_BUCKET_NAME='ytfeed'
export YTFEED_YOUTUBE_API_KEY='myyoutubeapikey'
export YTFEED_RESUB_INTERVAL='24h'

export YTFEED_VIDEO_FORMAT_QUALITY='720'
export YTFEED_VIDEO_FORMAT_EXTENSION='webm'
export YTFEED_VIDEO_DOWNLOAD_RETRY_DELAY='5m'

export YTFEED_BOLTDB_PATH='/home/ytfeed/schedule.db'
export YTFEED_STREAM_SCHEDULER_WORKER_INTERVAL='1m'
```

## Building

Your ol' plain `go build cmd/ytfeed/main.go`

## Note

- To run s3 test or complete suite test, you need to run `test_s3.sh` for it to start a Minio server. You can stop it or delete it after testing by running `test_s3_teardown.sh`
- You also need to set `YTFEED_YOUTUBE_API_KEY`, `YTFEED_YOUTUBE_VIDEO_ID`, and `YTFEED_YOUTUBE_VIDEO_URL` to do a complete suite test.
- You can redo failed download by sending a `POST` request with the XML message in the log as the body.
- If you are planning to download live broadcast, the downloaded video will be in the format `mp4` regardless of your extension config.
- If you send `SIGINT` while ytfeed is downloading video, it will only kills the immediate `youtube-dl` process, you have to make sure `ffmpeg` process `youtube-dl` spawned is also killed and the resulting temporary directory and file removed.
- If the program got killed because of OOM, you can turn on swap file if you cannot raise the machine's memory. Or you can PR me a better way to handle the upload.
- If you want to use stream scheduler, make sure file that is pointed at `YTFEED_BOLTDB_PATH` is already exists. You can create the file using `touch $YTFEED_BOLTDB_PATH` command.

## Before Using

- You **MUST** install [youtube-dl](https://github.com/ytdl-org/youtube-dl) first and **MUST** be available at `$PATH`.
- You **MUST** install [ffmpeg](https://ffmpeg.org/) too for this to work.
- You **MUST*** run the program first and make sure it could be reached from the internet, then go [here](https://pubsubhubbub.appspot.com/subscribe) to subscribe to a channel.
- From the page in the URL above: 
    - Fill `Callback URL` field with the address of your program, for example `http://my.ip.address:8123`. This must be the same as the value of `YTFEED_RESUB_CALLBACK_ADDR`.
    - Fill `Topic URL` field to be the same as your `YTFEED_RESUB_TOPIC`. 
    - Choose `Verify type` field as `synchronous`.
    - Choose `Mode` field as `subscribe`.
    - Fill `Verify token` field to be the same as your `YTFEED_VERIFICATION_TOKEN`.
    - Fill `HMAC secret` field to be the same as your `YTFEED_VERIFICATION_SECRET`.
    - Let `Lease seconds` field to be empty.
    - Press `Do it` button.
    - There should be a log that says something like `INFO[0635] Subscribed to topic https://www.youtube.com/xml/feeds/videos.xml?channel_id=channelid with challenge 4776706698370686416`. 

## Usage

Set the required environment variables and run the binary like `./ytfeed`
and you'll see log messages if it runs.

## How to Contribute

- Keep your code super simple and clean.
- Add comments to your code if you have to explain what your code is doing.
- Make sure you have Commitizen CLI installed and your commit message must be written through `git cz`.
- Pull request to branch other than `development` will be rejected.
