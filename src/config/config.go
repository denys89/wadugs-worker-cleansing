package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	AppName    string `envconfig:"APP_NAME" default:"wadugs-worker-cleansing"`
	AppVersion string `envconfig:"APP_VERSION" default:"v1.0.0"`

	// NSQ
	NsqServer           string `envconfig:"NSQ_SERVER" default:"172.31.33.126:3150"`
	MaxInflight         int    `envconfig:"MAX_INFLIGHT" default:"5"`
	NsqConcurrency      int    `envconfig:"NSQ_CONCURRENCY" default:"1"`
	MaxRequeueAttempt   uint16 `envconfig:"MAX_REQUEUE_ATTEMPT" default:"5"`
	TopicName           string `envconfig:"TOPIC_NAME" default:"data-cleansing"`
	ConsumerChannelName string `envconfig:"CONSUMER_CHANNEL_NAME" default:"server-cleansing-consumer-channel"`

	// AWS Configuration
	AWSRegion          string `envconfig:"AWS_REGION" default:"ap-southeast-1"`
	AWSAccessKeyID     string `envconfig:"AWS_ACCESS_KEY_ID" default:"key" required:"true"`
	AWSSecretAccessKey string `envconfig:"AWS_SECRET_ACCESS_KEY" default:"secret" required:"true"`
}

// Get ...
func Get() *Config {
	cfg := Config{}
	envconfig.MustProcess("", &cfg)

	return &cfg
}