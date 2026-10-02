package email

import (
	"Online-queue-management-system/libs/config"
	"Online-queue-management-system/libs/logger"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

// Outgoing emails live in a Redis Stream instead of process memory (12-factor VI):
// a restart or crash of the service does not lose queued emails, and several replicas
// share one queue through a consumer group, so every email is delivered by exactly one worker.
const (
	StreamKey     = "email:outbox"
	DeadStreamKey = "email:dead"
	GroupName     = "email-senders"

	streamMaxLen = 100_000
	// Short block keeps shutdown fast and lets workers check for abandoned emails every second.
	readBlock = time.Second
	// Consumers idle longer than this with no pending emails belong to dead processes.
	staleConsumerIdle = time.Hour
)

const (
	fieldTo       = "to"
	fieldSubject  = "subject"
	fieldBody     = "body"
	fieldHTMLBody = "html_body"
)

type EmailQueue struct {
	client   *goredis.Client
	sender   Sender
	consumer string

	workers     int
	rateLimit   time.Duration
	timeOut     time.Duration
	retryAfter  time.Duration
	maxAttempts int64

	wg     sync.WaitGroup
	cancel context.CancelFunc
	ctx    context.Context
}

// NewEmailQueue creates the consumer group (if needed) and starts the workers.
// Workers stop when ctx is cancelled or Shutdown is called.
func NewEmailQueue(ctx context.Context, client *goredis.Client, sender Sender, cfg config.QueueConfig) (*EmailQueue, error) {
	err := client.XGroupCreateMkStream(ctx, StreamKey, GroupName, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return nil, fmt.Errorf("create email consumer group: %w", err)
	}

	removeStaleConsumers(ctx, client)

	workerCtx, cancel := context.WithCancel(ctx)

	eq := &EmailQueue{
		client:      client,
		sender:      sender,
		consumer:    consumerName(),
		workers:     cfg.NumWorkers,
		rateLimit:   cfg.RateLimit,
		timeOut:     cfg.WrkTimeOut,
		retryAfter:  cfg.RetryAfter,
		maxAttempts: int64(cfg.MaxAttempts),
		cancel:      cancel,
		ctx:         workerCtx,
	}

	for i := range cfg.NumWorkers {
		eq.wg.Add(1)
		go eq.worker(i)
	}

	return eq, nil
}

// Enqueue durably stores the email in Redis; it is sent asynchronously by a worker.
func (eq *EmailQueue) Enqueue(ctx context.Context, msg EmailMessage) error {
	id, err := eq.client.XAdd(ctx, &goredis.XAddArgs{
		Stream: StreamKey,
		MaxLen: streamMaxLen,
		Approx: true,
		Values: map[string]any{
			fieldTo:       msg.To,
			fieldSubject:  msg.Subject,
			fieldBody:     msg.Body,
			fieldHTMLBody: msg.HTMLBody,
		},
	}).Result()
	if err != nil {
		return fmt.Errorf("enqueue email: %w", err)
	}

	logger.From(ctx).Info("email added to queue", "to", msg.To, "message_id", id)
	return nil
}

func (eq *EmailQueue) worker(workerID int) {
	defer eq.wg.Done()

	log := logger.From(eq.ctx).With("worker_id", workerID, "consumer", eq.consumer)
	log.Info("email queue worker started")

	// Rate limiting for the SMTP provider (Gmail).
	limiter := time.NewTicker(eq.rateLimit)
	defer limiter.Stop()

	for {
		msg, ok, err := eq.next()
		if eq.ctx.Err() != nil {
			log.Info("email queue worker stopping")
			return
		}
		if err != nil {
			log.Error("failed to read email queue", "err", err)
			eq.sleep(time.Second)
			continue
		}
		if !ok {
			continue
		}

		select {
		case <-limiter.C:
		case <-eq.ctx.Done():
			// Not acknowledged: stays pending and is reclaimed after restart.
			log.Info("email queue worker stopping")
			return
		}

		eq.process(log, msg)
	}
}

// next returns an email to send: first one abandoned by a crashed or failed worker,
// otherwise a new one (blocking up to readBlock).
func (eq *EmailQueue) next() (goredis.XMessage, bool, error) {
	claimed, _, err := eq.client.XAutoClaim(eq.ctx, &goredis.XAutoClaimArgs{
		Stream:   StreamKey,
		Group:    GroupName,
		Consumer: eq.consumer,
		MinIdle:  eq.retryAfter,
		Start:    "0-0",
		Count:    1,
	}).Result()
	if err != nil {
		return goredis.XMessage{}, false, err
	}
	if len(claimed) > 0 {
		return claimed[0], true, nil
	}

	streams, err := eq.client.XReadGroup(eq.ctx, &goredis.XReadGroupArgs{
		Group:    GroupName,
		Consumer: eq.consumer,
		Streams:  []string{StreamKey, ">"},
		Count:    1,
		Block:    readBlock,
	}).Result()
	if errors.Is(err, goredis.Nil) {
		return goredis.XMessage{}, false, nil
	}
	if err != nil {
		return goredis.XMessage{}, false, err
	}
	if len(streams) == 0 || len(streams[0].Messages) == 0 {
		return goredis.XMessage{}, false, nil
	}

	return streams[0].Messages[0], true, nil
}

func (eq *EmailQueue) process(log *slog.Logger, raw goredis.XMessage) {
	msg := messageFromValues(raw.Values)
	log = log.With("message_id", raw.ID, "to", msg.To)

	// The current email is finished even during shutdown (12-factor IX), bounded by the timeout.
	sendCtx, cancel := context.WithTimeout(context.WithoutCancel(eq.ctx), eq.timeOut)
	start := time.Now()
	err := eq.sender.SendEmail(sendCtx, msg)
	cancel()

	if err == nil {
		eq.ack(log, raw.ID)
		log.Info("email sent", "duration_ms", time.Since(start).Milliseconds())
		return
	}

	attempts := eq.attempts(raw.ID)
	log.Error("failed to send email", "err", err, "attempt", attempts, "max_attempts", eq.maxAttempts)

	if attempts < eq.maxAttempts {
		// Left unacknowledged: XAUTOCLAIM retries it after retryAfter.
		return
	}

	values := raw.Values
	values["error"] = err.Error()
	values["source_id"] = raw.ID
	if err := eq.client.XAdd(context.WithoutCancel(eq.ctx), &goredis.XAddArgs{
		Stream: DeadStreamKey,
		MaxLen: streamMaxLen,
		Approx: true,
		Values: values,
	}).Err(); err != nil {
		log.Error("failed to move email to dead letter stream", "err", err)
		return
	}
	eq.ack(log, raw.ID)
	log.Warn("email moved to dead letter stream", "stream", DeadStreamKey)
}

// attempts returns how many times the message has been delivered to a worker.
func (eq *EmailQueue) attempts(id string) int64 {
	pending, err := eq.client.XPendingExt(context.WithoutCancel(eq.ctx), &goredis.XPendingExtArgs{
		Stream: StreamKey,
		Group:  GroupName,
		Start:  id,
		End:    id,
		Count:  1,
	}).Result()
	if err != nil || len(pending) == 0 {
		return 1
	}
	return pending[0].RetryCount
}

func (eq *EmailQueue) ack(log *slog.Logger, id string) {
	ctx := context.WithoutCancel(eq.ctx)
	_, err := eq.client.TxPipelined(ctx, func(pipe goredis.Pipeliner) error {
		pipe.XAck(ctx, StreamKey, GroupName, id)
		pipe.XDel(ctx, StreamKey, id)
		return nil
	})
	if err != nil {
		log.Error("failed to acknowledge email", "err", err)
	}
}

func (eq *EmailQueue) sleep(d time.Duration) {
	select {
	case <-time.After(d):
	case <-eq.ctx.Done():
	}
}

// Shutdown stops the workers and waits for in-flight emails to finish.
func (eq *EmailQueue) Shutdown() {
	log := logger.From(eq.ctx)
	log.Info("shutting down email queue", "workers", eq.workers)

	eq.cancel()

	done := make(chan struct{})
	go func() {
		eq.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info("email queue shutdown complete")
	case <-time.After(eq.timeOut + 5*time.Second):
		log.Error("email queue shutdown timeout")
	}
}

func messageFromValues(values map[string]any) EmailMessage {
	str := func(key string) string {
		s, _ := values[key].(string)
		return s
	}
	return EmailMessage{
		To:       str(fieldTo),
		Subject:  str(fieldSubject),
		Body:     str(fieldBody),
		HTMLBody: str(fieldHTMLBody),
	}
}

// removeStaleConsumers drops consumers of previous processes that have nothing pending;
// every restart registers a new consumer name, so without this the group only grows.
func removeStaleConsumers(ctx context.Context, client *goredis.Client) {
	consumers, err := client.XInfoConsumers(ctx, StreamKey, GroupName).Result()
	if err != nil {
		logger.From(ctx).Warn("failed to list email consumers", "err", err)
		return
	}
	for _, c := range consumers {
		if c.Pending == 0 && c.Idle > staleConsumerIdle {
			_ = client.XGroupDelConsumer(ctx, StreamKey, GroupName, c.Name).Err()
		}
	}
}

// consumerName is unique per process, so replicas are distinct members of the consumer group.
func consumerName() string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "registration"
	}
	return host + "-" + uuid.NewString()[:8]
}
