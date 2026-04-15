package config

import (
	"Online-queue-management-system/libs/config"
	"context"
	"time"
)

type EmailSenderConfig struct {
	SMTPHost    string
	SMTPPort    int
	SMTPUser    string
	SMTPPass    string
	SendTimeOut time.Duration
}

type RedisConfig struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

type QueueConfig struct {
	NumWorkers int
	RateLimit  time.Duration
	WrkTimeOut time.Duration
}

type RegistrationConfig struct {
	RegistrationPort string
}

type Config struct {
	RedisCfg       RedisConfig
	EmailSenderCfg EmailSenderConfig
	QueueCfg       QueueConfig
	RegCfg         RegistrationConfig
}

func LoadConfig(ctx context.Context) (*Config, error) {
	//redis config
	redisAddr, err := config.MustGet(ctx, "REDIS_ADDR")
	if err != nil {
		return nil, err
	}

	redisPassword, err := config.MustGet(ctx, "REDIS_PASSWORD")
	if err != nil {
		return nil, err
	}

	redisDB, err := config.GetInt(ctx, "REDIS_DB")
	if err != nil {
		return nil, err
	}

	redisCfg := RedisConfig{
		RedisAddr:     redisAddr,
		RedisPassword: redisPassword,
		RedisDB:       redisDB,
	}
	//registration config
	registrationPort, err := config.MustGet(ctx, "REGISTRATION_PORT")
	if err != nil {
		return nil, err
	}
	regCfg := RegistrationConfig{
		RegistrationPort: registrationPort,
	}

	//emailSender config
	smtpHost, err := config.MustGet(ctx, "SMTP_HOST")
	if err != nil {
		return nil, err
	}

	smtpPort, err := config.GetInt(ctx, "SMTP_PORT")
	if err != nil {
		return nil, err
	}

	smtpUser, err := config.MustGet(ctx, "SMTP_USER")
	if err != nil {
		return nil, err
	}

	smtpPass, err := config.MustGet(ctx, "SMTP_PASS")
	if err != nil {
		return nil, err
	}

	emailTimeOut := config.GetDurationDefault(ctx, "EMAIL_TIMEOUT", 20*time.Second)

	senderCfg := EmailSenderConfig{
		SMTPHost:    smtpHost,
		SMTPPort:    smtpPort,
		SMTPUser:    smtpUser,
		SMTPPass:    smtpPass,
		SendTimeOut: emailTimeOut,
	}

	//queue config
	workers := config.GetIntDefault(ctx, "NUM_WORKERS", 10)

	rateLimit := config.GetDurationDefault(ctx, "RATE_LIMIT", 30*time.Second)

	wrkTimeOut := config.GetDurationDefault(ctx, "WRK_TIMEOUT", 10*time.Second)

	queueCfg := QueueConfig{
		NumWorkers: workers,
		RateLimit:  rateLimit,
		WrkTimeOut: wrkTimeOut,
	}

	return &Config{
		RedisCfg:       redisCfg,
		RegCfg:         regCfg,
		EmailSenderCfg: senderCfg,
		QueueCfg:       queueCfg,
	}, err
}

type DBConfig struct {
	DSN      string
	Host     string
	Port     int
	User     string
	Password string
	SSLMode  string
}

func LoadDBConfig(ctx context.Context) (*DBConfig, error) {
	host, err := config.MustGet(ctx, "DB_HOST")
	if err != nil {
		return nil, err
	}

	port, err := config.GetInt(ctx, "DB_PORT")
	if err != nil {
		return nil, err
	}

	user, err := config.MustGet(ctx, "DB_USER")
	if err != nil {
		return nil, err
	}

	ssl, err := config.MustGet(ctx, "DB_SSLMODE")
	if err != nil {
		return nil, err
	}

	password, err := config.MustGet(ctx, "DB_PASSWORD")
	if err != nil {
		return nil, err
	}

	dsn := config.Get(ctx, "DB_DSN", "")
	return &DBConfig{
		DSN:      dsn,
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		SSLMode:  ssl,
	}, nil

}
