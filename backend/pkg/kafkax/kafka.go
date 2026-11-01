package kafkax

import (
	"context"
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/segmentio/kafka-go"
)

// Message encapsulates a message consumed from or published to Kafka
type Message struct {
	Topic     string
	Key       []byte
	Value     []byte
	Offset    int64
	Partition int
	Time      time.Time
}

// Config configures Kafka connection and behaviors
type Config struct {
	Brokers []string
	GroupID string
}

// Producer publishes messages to Kafka topics
type Producer struct {
	writer   *kafka.Writer
	brokers  []string
	memQueue *MemoryBroker
	closed   atomic.Bool
}

// NewProducer creates a Kafka producer. If brokers is empty, runs in mock memory mode.
func NewProducer(brokers []string) *Producer {
	if len(brokers) == 0 {
		return &Producer{
			memQueue: GetGlobalMemoryBroker(),
		}
	}

	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 3 * time.Second,
		ReadTimeout:  3 * time.Second,
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
		Async:        false,
	}

	return &Producer{
		writer:   w,
		brokers:  brokers,
		memQueue: GetGlobalMemoryBroker(),
	}
}

// Publish writes a message to the specified topic
func (p *Producer) Publish(ctx context.Context, topic string, key, value []byte) error {
	if p.closed.Load() {
		return errors.New("producer is closed")
	}

	if p.writer != nil {
		err := p.writer.WriteMessages(ctx, kafka.Message{
			Topic: topic,
			Key:   key,
			Value: value,
			Time:  time.Now(),
		})
		if err == nil {
			return nil
		}
		log.Printf("[Kafka Producer Warning] Failed to publish to %s (%v), falling back to memory queue", topic, err)
	}

	// Fallback to in-memory broker
	p.memQueue.Publish(topic, Message{
		Topic: topic,
		Key:   key,
		Value: value,
		Time:  time.Now(),
	})
	return nil
}

// Close closes the producer connection
func (p *Producer) Close() error {
	if p.closed.Swap(true) {
		return nil
	}
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}

// Consumer consumes messages from a Kafka topic
type Consumer struct {
	reader   *kafka.Reader
	topic    string
	groupID  string
	memQueue *MemoryBroker
	cancel   context.CancelFunc
	closed   atomic.Bool
}

// NewConsumer creates a Kafka consumer for a topic
func NewConsumer(brokers []string, groupID, topic string) *Consumer {
	c := &Consumer{
		topic:    topic,
		groupID:  groupID,
		memQueue: GetGlobalMemoryBroker(),
	}

	if len(brokers) > 0 {
		r := kafka.NewReader(kafka.ReaderConfig{
			Brokers:        brokers,
			GroupID:        groupID,
			Topic:          topic,
			MinBytes:       10e3, // 10KB
			MaxBytes:       10e6, // 10MB
			CommitInterval: time.Second,
			StartOffset:    kafka.LastOffset,
		})
		c.reader = r
	}

	return c
}

// Start consumes messages and invokes the handler
func (c *Consumer) Start(ctx context.Context, handler func(msg Message) error) {
	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	if c.reader != nil {
		go func() {
			defer c.reader.Close()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					m, err := c.reader.ReadMessage(ctx)
					if err != nil {
						if errors.Is(err, context.Canceled) || c.closed.Load() {
							return
						}
						// If read fails (e.g. broker offline), fall back to memory channel
						time.Sleep(500 * time.Millisecond)
						continue
					}

					msg := Message{
						Topic:     m.Topic,
						Key:       m.Key,
						Value:     m.Value,
						Offset:    m.Offset,
						Partition: m.Partition,
						Time:      m.Time,
					}
					if err := handler(msg); err != nil {
						log.Printf("[Kafka Consumer] Handler error on topic %s: %v", c.topic, err)
					}
				}
			}
		}()
	}

	// Also subscribe to in-memory queue for hybrid/fallback events
	go func() {
		subCh := c.memQueue.Subscribe(c.topic)
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-subCh:
				if !ok {
					return
				}
				if err := handler(msg); err != nil {
					log.Printf("[Memory Consumer] Handler error on topic %s: %v", c.topic, err)
				}
			}
		}
	}()
}

// Close stops the consumer
func (c *Consumer) Close() error {
	if c.closed.Swap(true) {
		return nil
	}
	if c.cancel != nil {
		c.cancel()
	}
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}

// ---------------------------------------------------------------------
// In-Memory Message Broker (Clean Fallback & Zero-Dependency Unit Tests)
// ---------------------------------------------------------------------

type MemoryBroker struct {
	subscribers map[string][]chan Message
	mu          sync.RWMutex
}

var (
	globalMemBroker *MemoryBroker
	globalMemOnce   sync.Once
)

func GetGlobalMemoryBroker() *MemoryBroker {
	globalMemOnce.Do(func() {
		globalMemBroker = &MemoryBroker{
			subscribers: make(map[string][]chan Message),
		}
	})
	return globalMemBroker
}

func (mb *MemoryBroker) Publish(topic string, msg Message) {
	mb.mu.RLock()
	subs, ok := mb.subscribers[topic]
	if !ok || len(subs) == 0 {
		mb.mu.RUnlock()
		return
	}

	targets := make([]chan Message, len(subs))
	copy(targets, subs)
	mb.mu.RUnlock()

	for _, ch := range targets {
		select {
		case ch <- msg:
		default:
			// Non-blocking drop if consumer buffer full
		}
	}
}

func (mb *MemoryBroker) Subscribe(topic string) <-chan Message {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	ch := make(chan Message, 256)
	mb.subscribers[topic] = append(mb.subscribers[topic], ch)
	return ch
}
