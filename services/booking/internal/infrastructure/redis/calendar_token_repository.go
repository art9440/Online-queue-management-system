package redis

import (
	"Online-queue-management-system/services/booking/internal/domain"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const (
	oauthStatePrefix = "google-calendar:oauth-state:"
	tokenPrefix      = "google-calendar:token:"
	oauthStateTTL    = 10 * time.Minute
)

type CalendarTokenRepository struct {
	client *goredis.Client
}

func NewCalendarTokenRepository(client *goredis.Client) *CalendarTokenRepository {
	return &CalendarTokenRepository{client: client}
}

func (r *CalendarTokenRepository) SaveOAuthState(ctx context.Context, state string, userID int64) error {
	return r.client.Set(ctx, oauthStatePrefix+state, userID, oauthStateTTL).Err()
}

func (r *CalendarTokenRepository) GetOAuthState(ctx context.Context, state string) (int64, error) {
	userID, err := r.client.Get(ctx, oauthStatePrefix+state).Int64()
	if errors.Is(err, goredis.Nil) {
		return 0, domain.ErrGoogleCalendarNotLinked
	}
	return userID, err
}

func (r *CalendarTokenRepository) DeleteOAuthState(ctx context.Context, state string) error {
	return r.client.Del(ctx, oauthStatePrefix+state).Err()
}

func (r *CalendarTokenRepository) SaveToken(ctx context.Context, userID int64, token domain.GoogleCalendarToken) error {
	raw, err := json.Marshal(token)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, tokenKey(userID), raw, 0).Err()
}

func (r *CalendarTokenRepository) GetToken(ctx context.Context, userID int64) (domain.GoogleCalendarToken, error) {
	raw, err := r.client.Get(ctx, tokenKey(userID)).Result()
	if errors.Is(err, goredis.Nil) {
		return domain.GoogleCalendarToken{}, domain.ErrGoogleCalendarNotLinked
	}
	if err != nil {
		return domain.GoogleCalendarToken{}, err
	}

	var token domain.GoogleCalendarToken
	if err := json.Unmarshal([]byte(raw), &token); err != nil {
		return domain.GoogleCalendarToken{}, fmt.Errorf("decode google calendar token: %w", err)
	}

	return token, nil
}

func tokenKey(userID int64) string {
	return tokenPrefix + strconv.FormatInt(userID, 10)
}
