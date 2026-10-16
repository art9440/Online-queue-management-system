package polling

import (
	"Online-queue-management-system/services/bot/internal/application/service"
	redisrepo "Online-queue-management-system/services/bot/internal/infrastructure/redis"
	"Online-queue-management-system/services/bot/internal/infrastructure/telegram"
	"context"
	"maps"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
)

const testLockTTL = 150 * time.Millisecond

// Two bot replicas: only the leader polls Telegram, every update is handled once,
// and when the leader stops the other replica continues from the saved offset.
func TestPoller_SingleLeaderAndHandover(t *testing.T) {
	server := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	telegramAPI := newFakeTelegram()
	handler := &recordingHandler{}

	startReplica := func(name string) context.CancelFunc {
		ctx, cancel := context.WithCancel(context.Background())
		poller := New(
			telegramAPI.clientFor(name),
			handler,
			redisrepo.NewLeaderLock(client, name, testLockTTL),
			redisrepo.NewOffsetStore(client),
			0,
			10*time.Millisecond,
			testLockTTL,
		)
		done := make(chan struct{})
		go func() { poller.Run(ctx); close(done) }()
		return func() { cancel(); <-done }
	}

	stops := map[string]context.CancelFunc{
		"a": startReplica("a"),
		"b": startReplica("b"),
	}

	telegramAPI.push(1, 2, 3)
	waitFor(t, func() bool { return handler.count() == 3 })

	// Stop the leader; the follower must take over.
	leader := telegramAPI.onlyCaller(t)
	stops[leader]()
	delete(stops, leader)
	defer func() {
		for _, stop := range stops {
			stop()
		}
	}()

	telegramAPI.push(4, 5)
	waitFor(t, func() bool { return handler.count() == 5 })

	if telegramAPI.maxConcurrent() != 1 {
		t.Fatalf("getUpdates was called concurrently by %d replicas", telegramAPI.maxConcurrent())
	}
	for chatID, n := range handler.snapshot() {
		if n != 1 {
			t.Fatalf("update for chat %d handled %d times", chatID, n)
		}
	}
	if offset, _ := redisrepo.NewOffsetStore(client).Load(context.Background()); offset != 6 {
		t.Fatalf("saved offset = %d, want 6", offset)
	}
}

type fakeTelegram struct {
	mu      sync.Mutex
	updates []telegram.Update
	callers map[string]bool
	active  int
	maxSeen int
}

func newFakeTelegram() *fakeTelegram {
	return &fakeTelegram{callers: map[string]bool{}}
}

func (f *fakeTelegram) push(ids ...int64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, id := range ids {
		f.updates = append(f.updates, telegram.Update{
			ID:      id,
			Message: &telegram.Message{Chat: telegram.Chat{ID: id}, Text: "/start"},
		})
	}
}

func (f *fakeTelegram) clientFor(name string) UpdatesClient {
	return &namedClient{api: f, name: name}
}

func (f *fakeTelegram) onlyCaller(t *testing.T) string {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.callers) != 1 {
		t.Fatalf("expected exactly one polling replica, got %v", f.callers)
	}
	for name := range f.callers {
		return name
	}
	return ""
}

func (f *fakeTelegram) maxConcurrent() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.maxSeen
}

type namedClient struct {
	api  *fakeTelegram
	name string
}

// GetUpdates behaves like Telegram: returns updates with id >= offset, long-polls briefly otherwise.
func (c *namedClient) GetUpdates(ctx context.Context, offset int64, _ int) ([]telegram.Update, error) {
	f := c.api
	f.mu.Lock()
	f.callers[c.name] = true
	f.active++
	if f.active > f.maxSeen {
		f.maxSeen = f.active
	}
	var result []telegram.Update
	for _, u := range f.updates {
		if u.ID >= offset {
			result = append(result, u)
		}
	}
	f.mu.Unlock()

	if len(result) == 0 {
		select {
		case <-ctx.Done():
		case <-time.After(20 * time.Millisecond):
		}
	}

	f.mu.Lock()
	f.active--
	f.mu.Unlock()
	return result, nil
}

type recordingHandler struct {
	mu    sync.Mutex
	chats map[int64]int
}

func (h *recordingHandler) HandleMessage(_ context.Context, msg service.IncomingMessage) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.chats == nil {
		h.chats = map[int64]int{}
	}
	h.chats[msg.ChatID]++
	return nil
}

func (h *recordingHandler) count() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.chats)
}

func (h *recordingHandler) snapshot() map[int64]int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return maps.Clone(h.chats)
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition not met in time")
}
