package config

import (
	libconfig "Online-queue-management-system/libs/config"
	"context"
	"time"
)

type RegistrationConfig struct {
	RegistrationPort string
}

type Config struct {
	RedisCfg       libconfig.RedisConfig
	EmailSenderCfg libconfig.EmailSenderConfig
	QueueCfg       libconfig.QueueConfig
	RegCfg         RegistrationConfig
	AppEnv         string
}

func LoadConfig(ctx context.Context) (*Config, error) {
	//redis config
	redisAddr, err := libconfig.MustGet(ctx, "REDIS_ADDR")
	if err != nil {
		return nil, err
	}

	redisPassword, err := libconfig.MustGet(ctx, "REDIS_PASSWORD")
	if err != nil {
		return nil, err
	}

	redisDB, err := libconfig.GetInt(ctx, "REDIS_DB")
	if err != nil {
		return nil, err
	}

	redisCfg := libconfig.RedisConfig{
		RedisAddr:     redisAddr,
		RedisPassword: redisPassword,
		RedisDB:       redisDB,
	}
	//registration config
	registrationPort, err := libconfig.MustGet(ctx, "REGISTRATION_PORT")
	if err != nil {
		return nil, err
	}
	regCfg := RegistrationConfig{
		RegistrationPort: registrationPort,
	}

	//emailSender config
	smtpHost, err := libconfig.MustGet(ctx, "SMTP_HOST")
	if err != nil {
		return nil, err
	}

	smtpPort, err := libconfig.GetInt(ctx, "SMTP_PORT")
	if err != nil {
		return nil, err
	}

	smtpUser, err := libconfig.MustGet(ctx, "SMTP_USER")
	if err != nil {
		return nil, err
	}

	smtpPass, err := libconfig.MustGet(ctx, "SMTP_PASS")
	if err != nil {
		return nil, err
	}

	emailTimeOut := libconfig.GetDurationDefault(ctx, "EMAIL_TIMEOUT", 20*time.Second)

	senderCfg := libconfig.EmailSenderConfig{
		SMTPHost:    smtpHost,
		SMTPPort:    smtpPort,
		SMTPUser:    smtpUser,
		SMTPPass:    smtpPass,
		SendTimeOut: emailTimeOut,
	}

	//queue config
	workers := libconfig.GetIntDefault(ctx, "NUM_WORKERS", 10)

	rateLimit := libconfig.GetDurationDefault(ctx, "RATE_LIMIT", 30*time.Second)

	wrkTimeOut := libconfig.GetDurationDefault(ctx, "WRK_TIMEOUT", 10*time.Second)

	queueCfg := libconfig.QueueConfig{
		NumWorkers: workers,
		RateLimit:  rateLimit,
		WrkTimeOut: wrkTimeOut,
	}

	appEnv := libconfig.Get(ctx, "APP_ENV", "")

	return &Config{
		RedisCfg:       redisCfg,
		RegCfg:         regCfg,
		EmailSenderCfg: senderCfg,
		QueueCfg:       queueCfg,
		AppEnv:         appEnv,
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
	host, err := libconfig.MustGet(ctx, "DB_HOST")
	if err != nil {
		return nil, err
	}

	port, err := libconfig.GetInt(ctx, "DB_PORT")
	if err != nil {
		return nil, err
	}

	user, err := libconfig.MustGet(ctx, "DB_USER")
	if err != nil {
		return nil, err
	}

	ssl, err := libconfig.MustGet(ctx, "DB_SSLMODE")
	if err != nil {
		return nil, err
	}

	password, err := libconfig.MustGet(ctx, "DB_PASSWORD")
	if err != nil {
		return nil, err
	}

	dsn := libconfig.Get(ctx, "DB_DSN", "")
	return &DBConfig{
		DSN:      dsn,
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		SSLMode:  ssl,
	}, nil

}
