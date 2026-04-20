package service

import (
	"Online-queue-management-system/libs/logger"
	"Online-queue-management-system/services/registration/internal/application/email"
	"Online-queue-management-system/services/registration/internal/application/queue"
	"Online-queue-management-system/services/registration/internal/domain/pending"
	"Online-queue-management-system/services/registration/internal/domain/recovery"
	"Online-queue-management-system/services/registration/internal/infrastructure/security"
	"context"
	crypto "crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type RegistrationService struct {
	repoRedis    PendingRepo
	recoveryRepo RecoveryRepo
	repoPostgres UserRepo
	emailQueue   *queue.EmailQueue
}

func NewRegistrationService(repoRedis PendingRepo, recoveryRepo RecoveryRepo, repoPostgres UserRepo, queue *queue.EmailQueue) *RegistrationService {
	return &RegistrationService{
		repoRedis:    repoRedis,
		recoveryRepo: recoveryRepo,
		repoPostgres: repoPostgres,
		emailQueue:   queue,
	}
}

func (s *RegistrationService) Register(ctx context.Context, req RegisterInput) (RegisterOutput, error) {
	log := logger.From(ctx)
	log.Info("starting registration process for email", "email", req.Email)

	if exists, err := s.repoPostgres.GetUserByEmail(ctx, req.Email); err != nil {
		log.Error("error checking existing user", "email", req.Email, "err", err)
		return RegisterOutput{}, fmt.Errorf("error checking existing user: %w", err)
	} else if exists {
		log.Warn("user with email already exists", "email", req.Email)
		return RegisterOutput{}, errors.New("user with this email already exists")
	}

	registrationID := uuid.NewString()
	code := generateCode()

	hash, err := security.HashPassword(req.Password)
	if err != nil {
		return RegisterOutput{}, err
	}

	pendingItem := pending.PendingRegistration{
		ID:           registrationID,
		Email:        req.Email,
		PasswordHash: hash,
		BusinessName: req.BusinessName,
		BusinessType: req.BusinessType,
		Code:         code,
	}

	log.Info("creating pending registration", "registrationID", pendingItem.ID)
	if err := s.repoRedis.Save(ctx, pendingItem); err != nil {
		log.Error("failed to save pending registration", "registrationID", pendingItem.ID, "err", err)
		return RegisterOutput{}, err
	}
	log.Info("pending registration saved", "registrationID", pendingItem.ID)

	s.emailQueue.Enqueue(email.EmailMessage{
		To:      req.Email,
		Subject: "Код подтверждения",
		Body:    code,
	})

	return RegisterOutput{
		Status:         "pending",
		RegistrationID: registrationID,
	}, nil
}

func (s *RegistrationService) Verify(ctx context.Context, req VerifyInput) error {
	log := logger.From(ctx)
	log.Info("verifying registration", "registrationID", req.RegistrationID)

	pendingItem, err := s.repoRedis.Get(ctx, req.RegistrationID)
	if err != nil {
		log.Error("failed to get pending registration from Redis", "registrationID", req.RegistrationID, "err", err)
		return err
	}

	if pendingItem.Code != req.Code {
		log.Warn("invalid verification code", "registrationID", req.RegistrationID)
		return errors.New("invalid code")
	}

	const maxRetries = 3

	var errRet error
	var slug string
	for i := range maxRetries {
		slug, errRet = generateSlug()
		if errRet != nil {
			return errRet
		}

		pendingItem.ClientSlug = &slug

		errRet = s.repoPostgres.CreateUserWithBusiness(ctx, pendingItem)
		if errRet == nil {
			break
		}

		if isUniqueViolation(errRet, "businesses_registration_slug_uindex") {
			log.Warn("slug collision, retrying", "attempt", i+1)
			continue
		}

		if isUniqueViolation(errRet, "users_login_key") {
			log.Warn("email already exists", "email", pendingItem.Email)
			return errors.New("user with this email already exists")
		}

		return errRet
	}

	if errRet != nil {
		return fmt.Errorf("failed after retries: %w", errRet)
	}

	log.Info("business_admin account is registered")

	if err := s.repoRedis.Delete(ctx, req.RegistrationID); err != nil {
		log.Warn("failed to delete pending registration from Redis", "registrationID", req.RegistrationID, "err", err)
	}

	return nil
}

func (s *RegistrationService) ResendCode(ctx context.Context, req ResendInput) error {
	log := logger.From(ctx)
	log.Info("resending verification code", "registrationID", req.RegistrationID)

	pendingItem, err := s.repoRedis.Get(ctx, req.RegistrationID)
	if err != nil {
		log.Error("failed to get pending registration from Redis", "registrationID", req.RegistrationID, "err", err)
		return err
	}

	s.emailQueue.Enqueue(email.EmailMessage{
		To:      pendingItem.Email,
		Subject: "Код подтверждения",
		Body:    pendingItem.Code,
	})

	return nil
}

func (s *RegistrationService) RecoverPassword(ctx context.Context, req PasswordRecoveryInput) (PasswordRecoveryOutput, error) {
	log := logger.From(ctx)
	log.Info("starting password recovery", "email", req.Email)

	exists, err := s.repoPostgres.GetUserByEmail(ctx, req.Email)
	if err != nil {
		log.Error("failed to check user for password recovery", "email", req.Email, "err", err)
		return PasswordRecoveryOutput{}, err
	}

	recoveryID := uuid.NewString()
	code := generateCode()
	recoveryItem := recovery.PasswordRecovery{
		ID:    recoveryID,
		Email: req.Email,
		Code:  code,
	}

	if exists {
		if err := s.recoveryRepo.SaveRecovery(ctx, recoveryItem); err != nil {
			log.Error("failed to save password recovery", "email", req.Email, "recoveryID", recoveryID, "err", err)
			return PasswordRecoveryOutput{}, err
		}

		s.emailQueue.Enqueue(email.EmailMessage{
			To:      req.Email,
			Subject: "Код восстановления пароля",
			Body:    code,
		})
		log.Info("password recovery code queued", "email", req.Email, "recoveryID", recoveryID)
	} else {
		log.Warn("password recovery requested for unknown email", "email", req.Email, "recoveryID", recoveryID)
	}

	return PasswordRecoveryOutput{
		Status:     "password_recovery_pending",
		RecoveryID: recoveryID,
	}, nil
}

func (s *RegistrationService) ConfirmPasswordRecovery(ctx context.Context, req PasswordRecoveryVerifyInput) error {
	log := logger.From(ctx)
	log.Info("confirming password recovery", "recoveryID", req.RecoveryID)

	recoveryItem, err := s.recoveryRepo.GetRecovery(ctx, req.RecoveryID)
	if err != nil {
		log.Warn("password recovery not found", "recoveryID", req.RecoveryID)
		return errors.New("invalid code")
	}

	if recoveryItem.Code != req.Code {
		log.Warn("invalid password recovery code", "recoveryID", req.RecoveryID)
		return errors.New("invalid code")
	}

	temporaryPassword, err := generateTemporaryPassword()
	if err != nil {
		log.Error("failed to generate temporary password", "recoveryID", req.RecoveryID, "err", err)
		return err
	}

	hash, err := security.HashPassword(temporaryPassword)
	if err != nil {
		log.Error("failed to hash temporary password", "recoveryID", req.RecoveryID, "err", err)
		return err
	}

	updated, err := s.repoPostgres.UpdatePasswordByEmail(ctx, recoveryItem.Email, hash)
	if err != nil {
		log.Error("failed to update user password", "email", recoveryItem.Email, "recoveryID", req.RecoveryID, "err", err)
		return err
	}

	if updated {
		s.emailQueue.Enqueue(email.EmailMessage{
			To:      recoveryItem.Email,
			Subject: "Восстановление пароля",
			HTMLBody: fmt.Sprintf(`
				<h2>Восстановление пароля</h2>
				<p>Для вашей учетной записи был выпущен временный пароль:</p>
				<h1 style="font-size: 28px; letter-spacing: 2px;">%s</h1>
				<p>Войдите с этим паролем и сразу смените его на новый.</p>
			`, temporaryPassword),
		})
		log.Info("password recovery completed", "email", recoveryItem.Email, "recoveryID", req.RecoveryID)
	} else {
		log.Warn("password recovery confirmed for unknown email", "email", recoveryItem.Email, "recoveryID", req.RecoveryID)
	}

	if err := s.recoveryRepo.DeleteRecovery(ctx, req.RecoveryID); err != nil {
		log.Warn("failed to delete password recovery", "recoveryID", req.RecoveryID, "err", err)
	}

	return nil
}

func generateCode() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func generateTemporaryPassword() (string, error) {
	raw, err := generateSlug()
	if err != nil {
		return "", err
	}

	return "tmp-" + raw, nil
}

func generateSlug() (string, error) {
	b := make([]byte, 6)

	_, err := crypto.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

func isUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" && pgErr.ConstraintName == constraint
	}
	return false
}
