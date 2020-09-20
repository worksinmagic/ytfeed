package config

import (
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	DefaultVersion                       = "v1.0.0"
	DefaultHost                          = ":8123"
	DefaultResubTargetAddr               = "https://pubsubhubbub.appspot.com/subscribe"
	DefaultResubInterval                 = 24 * 3 * time.Hour
	DefaultFileNameTemplate              = "{{.ChannelID}}/{{.PublishedYear}}/{{.PublishedMonth}}/{{.PublishedDay}}/{{.PublishedTimeZone}}/{{.VideoID}}.{{.VideoExtension}}"
	DefaultFormatQuality                 = "720"
	DefaultFormatExtension               = "webm"
	DefaultRedisChannel                  = "ytfeed"
	DefaultStreamSchedulerRetryDelay     = 1 * time.Minute
	DefaultStreamSchedulerMaxRetries     = 5
	DefaultStreamSchedulerWorkerInterval = 1 * time.Minute
	DefaultVideoDownloadMaxRetries       = 5
	DefaultTemporaryFileDir              = "./"
	DefaultAMQPExchange                  = "ytfeed"
	DefaultAMQPKey                       = "schedule"
	DefaultAMQPPublishMandatory          = true
	DefaultAMQPPublishImmediate          = false
	DefaultAMQPExchangeKind              = "topic"
	DefaultAMQPExchangeDurable           = true
	DefaultAMQPExchangeInternal          = false
	DefaultAMQPExchangeAutoDelete        = false
	DefaultAMQPExchangeNoWait            = false

	StorageBackendS3   = "s3"
	StorageBackendGCS  = "gcs"
	StorageBackendDisk = "disk"
	StorageBackendNone = "none"
)

var (
	ErrInvalidStorageBackend = errors.New("invalid storage backend detected")

	ErrInvalidDiskConfig = errors.New("invalid or incomplete disk config")
	ErrInvalidGCSConfig  = errors.New("invalid or incomplete gcs config")
	ErrInvalidS3Config   = errors.New("invalid or incomplete s3 config")
)

func init() {
	viper.SetEnvPrefix("YTFEED")
	viper.AutomaticEnv()

	handleError(viper.BindEnv("version"))
	handleError(viper.BindEnv("host"))
	handleError(viper.BindEnv("youtube_api_key"))
	handleError(viper.BindEnv("verification_token"))

	handleError(viper.BindEnv("resub_interval"))
	handleError(viper.BindEnv("resub_target_addr"))
	handleError(viper.BindEnv("resub_topic"))
	handleError(viper.BindEnv("resub_callback_addr"))

	handleError(viper.BindEnv("s3_endpoint"))
	handleError(viper.BindEnv("s3_access_key_id"))
	handleError(viper.BindEnv("s3_secret_access_key"))
	handleError(viper.BindEnv("s3_bucket_name"))

	handleError(viper.BindEnv("storage_backend"))

	handleError(viper.BindEnv("gcs_credential_json_file_path"))
	handleError(viper.BindEnv("gcs_bucket_name"))

	handleError(viper.BindEnv("disk_directory"))

	handleError(viper.BindEnv("filename_template"))

	handleError(viper.BindEnv("video_format_quality"))
	handleError(viper.BindEnv("video_format_extension"))
	handleError(viper.BindEnv("video_handling_delay"))
	handleError(viper.BindEnv("temporary_file_dir"))
	handleError(viper.BindEnv("video_download_max_retries"))
	handleError(viper.BindEnv("video_download_retry_delay"))

	handleError(viper.BindEnv("redis_addr"))
	handleError(viper.BindEnv("redis_username"))
	handleError(viper.BindEnv("redis_password"))
	handleError(viper.BindEnv("redis_channel"))
	handleError(viper.BindEnv("redis_db"))
	handleError(viper.BindEnv("redis_max_retries"))
	handleError(viper.BindEnv("redis_dial_timeout"))
	handleError(viper.BindEnv("redis_write_timeout"))
	handleError(viper.BindEnv("redis_read_timeout"))
	handleError(viper.BindEnv("redis_pool_size"))
	handleError(viper.BindEnv("redis_min_idle_conns"))
	handleError(viper.BindEnv("redis_max_conn_age"))
	handleError(viper.BindEnv("redis_pool_timeout"))
	handleError(viper.BindEnv("redis_idle_timeout"))
	handleError(viper.BindEnv("redis_idle_check_frequency"))

	handleError(viper.BindEnv("boltdb_path"))
	handleError(viper.BindEnv("stream_scheduler_retry_delay"))
	handleError(viper.BindEnv("stream_scheduler_worker_interval"))
	handleError(viper.BindEnv("stream_scheduler_max_retries"))

	handleError(viper.BindEnv("amqp_dsn"))
	handleError(viper.BindEnv("amqp_exchange"))
	handleError(viper.BindEnv("amqp_key"))
	handleError(viper.BindEnv("amqp_publish_mandatory"))
	handleError(viper.BindEnv("amqp_publish_immediate"))
	handleError(viper.BindEnv("amqp_exchange_kind"))
	handleError(viper.BindEnv("amqp_exchange_durable"))
	handleError(viper.BindEnv("amqp_exchange_internal"))
	handleError(viper.BindEnv("amqp_exchange_auto_delete"))
	handleError(viper.BindEnv("amqp_exchange_no_wait"))

	viper.SetDefault("version", DefaultVersion)
	viper.SetDefault("host", DefaultHost)
	viper.SetDefault("resub_target_addr", DefaultResubTargetAddr)
	viper.SetDefault("resub_interval", DefaultResubInterval)
	viper.SetDefault("filename_template", DefaultFileNameTemplate)
	viper.SetDefault("video_format_quality", DefaultFormatQuality)
	viper.SetDefault("video_format_extension", DefaultFormatExtension)
	viper.SetDefault("redis_channel", DefaultRedisChannel)
	viper.SetDefault("stream_scheduler_retry_delay", DefaultStreamSchedulerRetryDelay)
	viper.SetDefault("stream_scheduler_worker_interval", DefaultStreamSchedulerWorkerInterval)
	viper.SetDefault("stream_scheduler_max_retries", DefaultStreamSchedulerMaxRetries)
	viper.SetDefault("video_download_max_retries", DefaultVideoDownloadMaxRetries)
	viper.SetDefault("temporary_file_dir", DefaultTemporaryFileDir)
	viper.SetDefault("amqp_exchange", DefaultAMQPExchange)
	viper.SetDefault("amqp_key", DefaultAMQPKey)
	viper.SetDefault("amqp_publish_mandatory", DefaultAMQPPublishMandatory)
	viper.SetDefault("amqp_publish_immediate", DefaultAMQPPublishImmediate)
	viper.SetDefault("amqp_exchange_kind", DefaultAMQPExchangeKind)
	viper.SetDefault("amqp_exchange_durable", DefaultAMQPExchangeDurable)
	viper.SetDefault("amqp_exchange_internal", DefaultAMQPExchangeInternal)
	viper.SetDefault("amqp_exchange_auto_delete", DefaultAMQPExchangeAutoDelete)
	viper.SetDefault("amqp_exchange_no_wait", DefaultAMQPExchangeNoWait)
}

func handleError(err error) {
	if err != nil {
		log.Fatalf("Failed to process configuration: %v", err)
	}
}

type Validator interface {
	Struct(interface{}) error
}

type Configuration struct {
	validator Validator `validate:"required"`

	YoutubeAPIKey     string `validate:"required"`
	VerificationToken string `validate:"required"`

	ResubTargetAddr   string        `validate:"required"`
	ResubTopic        string        `validate:"required"`
	ResubCallbackAddr string        `validate:"required"`
	ResubInterval     time.Duration `validate:"required"`

	S3Endpoint        string `validate:""`
	S3AccessKeyID     string `validate:""`
	S3SecretAccessKey string `validate:""`
	S3BucketName      string `validate:""`

	GCSCredentialJSONFilePath string `validate:""`
	GCSBucketName             string `validate:""`

	DiskDirectory string `validate:""`

	StorageBackend   string `validate:"required,oneof=gcs s3 disk none"`
	FileNameTemplate string `validate:"required"`
	Host             string `validate:"required"`
	Version          string `validate:"required"`

	VideoFormatQuality      string        `validate:"required,oneof=1080 720 640 480 360 240 144"`
	VideoFormatExtension    string        `validate:"required,oneof=mp4 webm mkv"`
	VideoDownloadMaxRetries int           `validate:"required,min=0"`
	VideoDownloadRetryDelay time.Duration `validate:"omitempty,min=1"`
	TemporaryFileDir        string        `validate:"required,dir"`

	RedisAddr               string        `validate:"omitempty,hostname_port"`
	RedisUsername           string        `validate:""`
	RedisPassword           string        `validate:""`
	RedisChannel            string        `validate:"required"`
	RedisDB                 int           `validate:"omitempty,min=0"`
	RedisMaxRetries         int           `validate:"omitempty,min=0"`
	RedisDialTimeout        time.Duration `validate:""`
	RedisWriteTimeout       time.Duration `validate:""`
	RedisReadTimeout        time.Duration `validate:""`
	RedisPoolSize           int           `validate:"omitempty,min=1"`
	RedisMinIdleConns       int           `validate:"omitempty,min=0"`
	RedisMaxConnAge         time.Duration `validate:""`
	RedisPoolTimeout        time.Duration `validate:""`
	RedisIdleTimeout        time.Duration `validate:""`
	RedisIdleCheckFrequency time.Duration `validate:""`

	BoltDBPath                    string        `validate:"omitempty,file"`
	StreamSchedulerRetryDelay     time.Duration `validate:"required,min=0"`
	StreamSchedulerWorkerInterval time.Duration `validate:"required,min=1000000000"`
	StreamSchedulerMaxRetries     int           `validate:""`

	AMQPDSN                string `validate:""`
	AMQPExchange           string `validate:"required"`
	AMQPKey                string `validate:"required"`
	AMQPPublishMandatory   bool   `validate:""`
	AMQPPublishImmediate   bool   `validate:""`
	AMQPExchangeKind       string `validate:"required,oneof=direct fanout topic headers"`
	AMQPExchangeDurable    bool   `validate:""`
	AMQPExchangeAutoDelete bool   `validate:""`
	AMQPExchangeInternal   bool   `validate:""`
	AMQPExchangeNoWait     bool   `validate:""`
}

func New() (c *Configuration) {
	c = &Configuration{}
	c.validator = validator.New()

	c.YoutubeAPIKey = viper.GetString("youtube_api_key")
	c.VerificationToken = viper.GetString("verification_token")

	c.ResubCallbackAddr = viper.GetString("resub_callback_addr")
	c.ResubTargetAddr = viper.GetString("resub_target_addr")
	c.ResubTopic = viper.GetString("resub_topic")
	c.ResubInterval = viper.GetDuration("resub_interval")

	c.S3Endpoint = viper.GetString("s3_endpoint")
	c.S3AccessKeyID = viper.GetString("s3_access_key_id")
	c.S3SecretAccessKey = viper.GetString("s3_secret_access_key")
	c.S3BucketName = viper.GetString("s3_bucket_name")

	c.GCSCredentialJSONFilePath = viper.GetString("gcs_credential_json_file_path")
	c.GCSBucketName = viper.GetString("gcs_bucket_name")

	c.DiskDirectory = viper.GetString("disk_directory")

	c.Host = viper.GetString("host")
	c.Version = viper.GetString("version")
	c.StorageBackend = viper.GetString("storage_backend")
	c.FileNameTemplate = viper.GetString("filename_template")

	c.VideoFormatQuality = viper.GetString("video_format_quality")
	c.VideoFormatExtension = viper.GetString("video_format_extension")
	c.VideoDownloadMaxRetries = viper.GetInt("video_download_max_retries")
	c.VideoDownloadRetryDelay = viper.GetDuration("video_download_retry_delay")
	c.TemporaryFileDir = viper.GetString("temporary_file_dir")

	c.RedisAddr = viper.GetString("redis_addr")
	c.RedisUsername = viper.GetString("redis_username")
	c.RedisPassword = viper.GetString("redis_password")
	c.RedisChannel = viper.GetString("redis_channel")
	c.RedisDB = viper.GetInt("redis_db")
	c.RedisMaxRetries = viper.GetInt("redis_max_retries")
	c.RedisDialTimeout = viper.GetDuration("redis_dial_timeout")
	c.RedisWriteTimeout = viper.GetDuration("redis_write_timeout")
	c.RedisReadTimeout = viper.GetDuration("redis_read_timeout")
	c.RedisPoolSize = viper.GetInt("redis_pool_size")
	c.RedisMinIdleConns = viper.GetInt("redis_min_idle_conns")
	c.RedisMaxConnAge = viper.GetDuration("redis_max_conn_age")
	c.RedisPoolTimeout = viper.GetDuration("redis_pool_timeout")
	c.RedisIdleTimeout = viper.GetDuration("redis_idle_timeout")
	c.RedisIdleCheckFrequency = viper.GetDuration("redis_idle_check_frequency")

	c.BoltDBPath = viper.GetString("boltdb_path")
	c.StreamSchedulerRetryDelay = viper.GetDuration("stream_scheduler_retry_delay")
	c.StreamSchedulerWorkerInterval = viper.GetDuration("stream_scheduler_worker_interval")
	c.StreamSchedulerMaxRetries = viper.GetInt("stream_scheduler_max_retries")

	c.AMQPDSN = viper.GetString("amqp_dsn")
	c.AMQPExchange = viper.GetString("amqp_exchange")
	c.AMQPKey = viper.GetString("amqp_key")
	c.AMQPPublishMandatory = viper.GetBool("amqp_publish_mandatory")
	c.AMQPPublishImmediate = viper.GetBool("amqp_publish_immediate")
	c.AMQPExchangeKind = viper.GetString("amqp_exchange_kind")
	c.AMQPExchangeDurable = viper.GetBool("amqp_exchange_durable")
	c.AMQPExchangeAutoDelete = viper.GetBool("amqp_exchange_auto_delete")
	c.AMQPExchangeInternal = viper.GetBool("amqp_exchange_internal")
	c.AMQPExchangeNoWait = viper.GetBool("amqp_exchange_no_wait")

	return
}

func (c *Configuration) Validate() (err error) {
	switch c.StorageBackend {
	case StorageBackendDisk:
		if c.DiskDirectory == "" {
			return ErrInvalidDiskConfig
		}
	case StorageBackendGCS:
		if c.GCSBucketName == "" {
			return ErrInvalidGCSConfig
		}
	case StorageBackendS3:
		if c.S3AccessKeyID == "" || c.S3BucketName == "" || c.S3Endpoint == "" || c.S3SecretAccessKey == "" {
			return ErrInvalidS3Config
		}
	case StorageBackendNone:
	default:
		return ErrInvalidStorageBackend
	}

	err = c.validator.Struct(c)

	return
}
