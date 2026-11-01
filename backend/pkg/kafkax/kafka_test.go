package kafkax

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestKafkaMemoryBrokerFallback(t *testing.T) {
	topic := "test_im_messages"
	producer := NewProducer(nil) // runs in memory broker mode
	defer producer.Close()

	consumer := NewConsumer(nil, "test_group", topic)
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(3)

	received := make([]string, 0)
	var mu sync.Mutex

	consumer.Start(ctx, func(msg Message) error {
		mu.Lock()
		received = append(received, string(msg.Value))
		mu.Unlock()
		wg.Done()
		return nil
	})

	// Give consumer a moment to attach
	time.Sleep(50 * time.Millisecond)

	// Publish 3 messages
	msgs := []string{"msg_1", "msg_2", "msg_3"}
	for _, m := range msgs {
		err := producer.Publish(ctx, topic, []byte("key"), []byte(m))
		if err != nil {
			t.Fatalf("failed to publish: %v", err)
		}
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for messages to be consumed")
	}

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 3 {
		t.Fatalf("expected 3 messages, got %d: %v", len(received), received)
	}
}
