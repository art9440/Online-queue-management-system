// internal/application/queue/email_queue.go
package queue

import (
	"Online-queue-management-system/libs/logger"
	"Online-queue-management-system/services/registration/internal/application/email"
	"context"
	"sync"
	"time"
)

type EmailQueue struct {
	queue   chan email.EmailMessage
	workers int
	sender  *email.EmailSender
	wg      sync.WaitGroup
	closeCh chan struct{}
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewEmailQueue(sender *email.EmailSender, workers int) *EmailQueue {
	ctx, cancel := context.WithCancel(context.Background())

	eq := &EmailQueue{
		queue:   make(chan email.EmailMessage, 10000),
		workers: workers,
		sender:  sender,
		closeCh: make(chan struct{}),
		ctx:     ctx,
		cancel:  cancel,
	}

	// Запускаем воркеров
	for i := range workers {
		eq.wg.Add(1)
		go eq.worker(i)
	}

	return eq
}

func (eq *EmailQueue) Enqueue(msg email.EmailMessage) {
	log := logger.From(eq.ctx)

	select {
	case eq.queue <- msg:
		// Успешно добавили в очередь
		log.Info("email added to queue",
			"to", msg.To,
			"queue_len", len(eq.queue),
			"queue_cap", cap(eq.queue))
	default:
		// Очередь переполнена - нужно мониторить
		log.Warn("email queue is full",
			"to", msg.To,
			"queue_len", len(eq.queue),
			"queue_cap", cap(eq.queue))
	}
}

func (eq *EmailQueue) worker(workerID int) {
	defer eq.wg.Done()

	// Создаем контекст для воркера с ID
	ctx := context.WithValue(eq.ctx, "worker_id", workerID)
	log := logger.From(ctx)

	log.Info("email queue worker started", "worker_id", workerID)

	// Rate limiting для Gmail (10 писем в секунду)
	limiter := time.NewTicker(100 * time.Millisecond)
	defer limiter.Stop()

	// Метрики для воркера
	var processedCount int64
	var failedCount int64

	for {
		select {
		case msg := <-eq.queue:
			<-limiter.C // Ждем разрешения от rate limiter

			log.Info("worker processing email",
				"worker_id", workerID,
				"to", msg.To,
				"subject", msg.Subject,
				"queue_len", len(eq.queue),
				"processed", processedCount)

			// Создаем контекст с таймаутом для отправки конкретного письма
			sendCtx, cancel := context.WithTimeout(ctx, 30*time.Second)

			startTime := time.Now()
			err := eq.sender.SendEmail(sendCtx, msg)
			elapsed := time.Since(startTime)
			cancel()

			if err != nil {
				failedCount++
				log.Error("worker failed to send email",
					"worker_id", workerID,
					"to", msg.To,
					"error", err,
					"duration_ms", elapsed.Milliseconds(),
					"failed_total", failedCount)

			} else {
				processedCount++
				log.Info("worker successfully sent email",
					"worker_id", workerID,
					"to", msg.To,
					"duration_ms", elapsed.Milliseconds(),
					"processed_total", processedCount)
			}

		case <-eq.closeCh:
			log.Info("email queue worker stopping",
				"worker_id", workerID,
				"processed_total", processedCount,
				"failed_total", failedCount)
			return
		}
	}
}

// Shutdown gracefully останавливает очередь
func (eq *EmailQueue) Shutdown() {
	log := logger.From(eq.ctx)
	log.Info("shutting down email queue",
		"pending_emails", len(eq.queue),
		"workers", eq.workers)

	close(eq.closeCh)

	// Ждем завершения всех воркеров с таймаутом
	done := make(chan struct{})
	go func() {
		eq.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info("email queue shutdown complete",
			"final_pending", len(eq.queue))
	case <-time.After(30 * time.Second):
		log.Error("email queue shutdown timeout",
			"pending_emails", len(eq.queue))
		eq.cancel()
	}
}

// GetStats возвращает статистику очереди
func (eq *EmailQueue) GetStats() (queueLen int, queueCap int, workers int) {
	return len(eq.queue), cap(eq.queue), eq.workers
}
