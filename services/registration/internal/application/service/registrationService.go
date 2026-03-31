package service

import (
	"Online-queue-management-system/libs/logger"
	"Online-queue-management-system/services/registration/internal/application/email"
	"Online-queue-management-system/services/registration/internal/domain/pending"
	"Online-queue-management-system/services/registration/internal/infrastructure/security"
	"context"
	"errors"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
)

type RegistrationService struct {
	repoRedis    PendingRepo
	repoPostgres UserRepo
	emailSender  Sender
}

func NewRegistrationService(repoRedis PendingRepo, repoPostgres UserRepo, emailSender Sender) *RegistrationService {
	return &RegistrationService{
		repoRedis:    repoRedis,
		repoPostgres: repoPostgres,
		emailSender:  emailSender,
	}
}

func (s *RegistrationService) Register(ctx context.Context, req RegisterInput) (RegisterOutput, error) {

	log := logger.From(ctx)
	log.Info("starting registration process for email", "email", req.Email)
	// 1. генерим ID
	registrationID := uuid.NewString()

	// 2. генерим код
	code := generateCode()

	// 3. хешируем пароль
	hash, err := security.HashPassword(req.Password)
	if err != nil {
		return RegisterOutput{}, err
	}

	// 4. создаём pending
	pending := pending.PendingRegistration{
		ID:           registrationID,
		Email:        req.Email,
		PasswordHash: hash,
		BusinessName: req.BusinessName,
		BusinessType: req.BusinessType,
		Code:         code,
	}
	log.Info("creating pending registration", "registrationID", pending.ID)
	// 5. сохраняем в Redis
	err = s.repoRedis.Save(ctx, pending)
	if err != nil {
		log.Error("failed to save pending registration", "registrationID", pending.ID, "err", err)
		return RegisterOutput{}, err
	}
	log.Info("pending registration saved", "registrationID", pending.ID)

	msg := email.EmailMessage{
		To:      req.Email,
		Subject: "Код подтверждения для входа в Online Queue",
		Body:    code,
	}
	// 6. отправляем email
	if err := s.emailSender.SendEmail(ctx, msg); err != nil {
		return RegisterOutput{}, fmt.Errorf("failed to send verification email: %w", err)
	}

	// 7. возвращаем ID
	return RegisterOutput{
		Status:         "pending",
		RegistrationID: registrationID,
	}, nil
}

func generateCode() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func (s *RegistrationService) Verify(ctx context.Context, req VerifyInput) error {
	log := logger.From(ctx)
	log.Info("verifying registration", "registrationID", req.RegistrationID)
	// 1. достать из Redis
	pending, err := s.repoRedis.Get(ctx, req.RegistrationID)
	if err != nil {
		log.Error("failed to get pending registration from Redis", "registrationID", req.RegistrationID, "err", err)
		return err
	}

	// 2. проверить код
	if pending.Code != req.Code {
		log.Warn("invalid verification code", "registrationID", req.RegistrationID)
		return errors.New("invalid code")
	}

	// 3. сохранить в Postgres
	err = s.repoPostgres.CreateUserWithBusiness(ctx, pending)
	if err != nil {
		log.Error("failed to create user with business in Postgres", "err", err)
		return err
	}

	// 4. удалить из Redis
	err = s.repoRedis.Delete(ctx, req.RegistrationID)
	if err != nil {
		log.Error("failed to delete pending registration from Redis", "registrationID", req.RegistrationID, "err", err)
		return err
	}

	return nil
}
