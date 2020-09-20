package ytfeed

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	mainytfeed "github.com/worksinmagic/ytfeed"
	"github.com/worksinmagic/ytfeed/config"
	"github.com/worksinmagic/ytfeed/health"
	"github.com/worksinmagic/ytfeed/plugin/autosubscribefeed"
	"github.com/worksinmagic/ytfeed/plugin/disk"
	"github.com/worksinmagic/ytfeed/plugin/gcs"
	"github.com/worksinmagic/ytfeed/plugin/publishamqp"
	"github.com/worksinmagic/ytfeed/plugin/publishredis"
	"github.com/worksinmagic/ytfeed/plugin/s3"
	"github.com/worksinmagic/ytfeed/plugin/savevideo"
	"github.com/worksinmagic/ytfeed/plugin/streamschedule"
	"github.com/worksinmagic/ytfeed/rss"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func Run(ctx context.Context, logger mainytfeed.Logger) (err error) {
	// declare dependencies
	cfg := config.New()
	err = cfg.Validate()
	if err != nil {
		err = errors.Wrap(err, "failed to validate configuration")
		return
	}

	yts, err := youtube.NewService(ctx, option.WithAPIKey(cfg.YoutubeAPIKey))
	if err != nil {
		err = errors.Wrap(err, "failed to create new YouTube service")
		return
	}

	// declare data handlers
	dataHandlers := make([]mainytfeed.DataHandlerFunc, 0, 3)

	var streamScheduler *streamschedule.StreamSchedule
	if cfg.BoltDBPath != "" {
		streamScheduler, err = streamschedule.New(logger, cfg.BoltDBPath, cfg.ResubCallbackAddr, cfg.StreamSchedulerRetryDelay, cfg.StreamSchedulerWorkerInterval, cfg.StreamSchedulerMaxRetries)
		if err != nil {
			err = errors.Wrap(err, "failed to create stream scheduler service")
			return
		}
		defer func(streamScheduler *streamschedule.StreamSchedule) {
			err := streamScheduler.CloseDatabase()
			if err != nil {
				logger.Errorf("Failed to close stream scheduler database: %v", err)
			}
		}(streamScheduler)
	}

	var dataSaver savevideo.DataSaver
	switch cfg.StorageBackend {
	case config.StorageBackendS3:
		dataSaver, err = s3.New(cfg.S3Endpoint, cfg.S3AccessKeyID, cfg.S3SecretAccessKey, cfg.S3BucketName, s3.UseSSL)
		if err != nil {
			err = errors.Wrap(err, "failed to create new S3 data saver service")
			return
		}
	case config.StorageBackendGCS:
		dataSaver, err = gcs.New(cfg.GCSBucketName, cfg.GCSCredentialJSONFilePath, nil)
		if err != nil {
			err = errors.Wrap(err, "failed to create new GCS data saver service")
			return
		}
	case config.StorageBackendDisk:
		dataSaver, err = disk.New(cfg.DiskDirectory)
		if err != nil {
			err = errors.Wrap(err, "failed to create new disk data saver service")
			return
		}
	case config.StorageBackendNone:
		// No backend, don't save video
	default:
		err = errors.Wrap(err, "storage backend type is invalid")
		return
	}

	var saveVideo *savevideo.SaveVideo
	if dataSaver != nil {
		saveVideo, err = savevideo.New(
			logger, yts.Videos, dataSaver,
			cfg.TemporaryFileDir, cfg.FileNameTemplate, cfg.VideoFormatQuality, cfg.VideoFormatExtension,
		)
		if err != nil {
			err = errors.Wrap(err, "failed to initialize savevideo")
			return
		}
	}
	if saveVideo != nil && streamScheduler != nil {
		saveVideo.SetStreamScheduler(streamScheduler)
	}
	if saveVideo != nil && cfg.VideoDownloadRetryDelay > 0 && cfg.VideoDownloadMaxRetries > 0 {
		saveVideo.SetRetries(cfg.VideoDownloadRetryDelay, cfg.VideoDownloadMaxRetries)
	}
	if saveVideo != nil {
		dataHandlers = append(dataHandlers, saveVideo.DataHandler)
	}

	if cfg.RedisAddr != "" {
		opts := &redis.Options{}
		opts.Addr = cfg.RedisAddr
		opts.DB = cfg.RedisDB
		opts.DialTimeout = cfg.RedisDialTimeout
		opts.IdleCheckFrequency = cfg.RedisIdleCheckFrequency
		opts.IdleTimeout = cfg.RedisIdleTimeout
		opts.MaxConnAge = cfg.RedisMaxConnAge
		opts.MaxRetries = cfg.RedisMaxRetries
		opts.MinIdleConns = cfg.RedisMinIdleConns
		opts.Password = cfg.RedisPassword
		opts.PoolSize = cfg.RedisPoolSize
		opts.PoolTimeout = cfg.RedisPoolTimeout
		opts.ReadTimeout = cfg.RedisReadTimeout
		opts.Username = cfg.RedisUsername
		opts.WriteTimeout = cfg.RedisWriteTimeout
		redisClient := publishredis.New(logger, cfg.RedisChannel, opts)

		dataHandlers = append(dataHandlers, redisClient.DataHandler)
	}

	if cfg.AMQPDSN != "" {
		var conn *amqp.Connection
		conn, err = amqp.Dial(cfg.AMQPDSN)
		if err != nil {
			err = errors.Wrap(err, "failed to initialize publishamqp")
			return
		}
		defer conn.Close()

		var amqpChannel *amqp.Channel
		amqpChannel, err = conn.Channel()
		if err != nil {
			err = errors.Wrap(err, "failed to create amqp channel")
			return
		}
		defer amqpChannel.Close()

		err = amqpChannel.Confirm(cfg.AMQPExchangeNoWait)
		if err != nil {
			err = errors.Wrap(err, "failed to set amqp broker to confirm mode")
			return
		}

		err = amqpChannel.ExchangeDeclare(
			cfg.AMQPExchange,
			cfg.AMQPExchangeKind,
			cfg.AMQPExchangeDurable,
			cfg.AMQPExchangeAutoDelete,
			cfg.AMQPExchangeInternal,
			cfg.AMQPExchangeNoWait,
			nil,
		)
		if err != nil {
			err = errors.Wrap(err, "failed to declare amqp exchange")
			return
		}

		pa := publishamqp.New(
			logger,
			amqpChannel,
			cfg.AMQPExchange,
			cfg.AMQPKey,
			cfg.AMQPPublishMandatory,
			cfg.AMQPPublishImmediate,
		)

		dataHandlers = append(dataHandlers, pa.DataHandler)
	}

	// run workers
	subscriber := autosubscribefeed.New(logger, cfg.VerificationToken, cfg.ResubTargetAddr, cfg.ResubTopic, cfg.ResubCallbackAddr, cfg.ResubInterval)
	go func(ctx context.Context, subscriber *autosubscribefeed.Subscriber) {
		err := subscriber.Subscribe(ctx)
		if err != nil {
			err = errors.Wrap(err, "resubscriber worker exited with error")
			logger.Errorln(err)
			return
		}
	}(ctx, subscriber)

	if streamScheduler != nil {
		go func(ctx context.Context, streamScheduler *streamschedule.StreamSchedule) {
			err := streamScheduler.RunWorker(ctx)
			if err != nil {
				err = errors.Wrap(err, "stream scheduler worker exited with error")
				logger.Errorln(err)
				return
			}
		}(ctx, streamScheduler)
	}

	// declare handler functions
	http.HandleFunc("/health", health.Handler)
	http.HandleFunc("/", rss.Handler(ctx, logger, cfg.VerificationToken, dataHandlers...))

	// listen
	errCh := make(chan error, 1)
	go func(errCh chan<- error) {
		logger.Infof("Server is listening at %s", cfg.Host)
		errCh <- http.ListenAndServe(cfg.Host, nil)
	}(errCh)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
	case err = <-errCh:
		err = errors.Wrap(err, "unexpected server error")
		return
	case <-quit:
	}

	return
}
