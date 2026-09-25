package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/IBM/sarama"
)

const (
	brokers = "localhost:9092"
	groupID = "go-analytics-consumer"
	topic   = "user-events"
)

// ConsumerGroupHandler implements sarama.ConsumerGroupHandler
type ConsumerGroupHandler struct {
	ready chan bool
	mu    sync.Mutex
}

// Setup runs at the beginning of a new session, before ConsumeClaim
func (h *ConsumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Log assigned partitions for observability
	log.Printf("REBALANCE: Session setup. MemberID=%s GenerationID=%d", session.MemberID(), session.GenerationID())
	for topic, partitions := range session.Claims() {
		log.Printf("REBALANCE: Assigned topic=%s partitions=%v", topic, partitions)
	}

	close(h.ready)
	return nil
}

// Cleanup runs at the end of a session, once all ConsumeClaim goroutines have exited
func (h *ConsumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	log.Printf("REBALANCE: Session cleanup. MemberID=%s GenerationID=%d",
		session.MemberID(), session.GenerationID())
	return nil
}

// ConsumeClaim must start a consumer loop for each Claim
func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				log.Printf("Message channel closed for topic=%s partition=%d",
					claim.Topic(), claim.Partition())
				return nil
			}

			// Process message
			var event map[string]interface{}
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("Failed to unmarshal message offset=%d: %v", msg.Offset, err)
				// Mark as processed anyway to avoid poison pill
				session.MarkMessage(msg, "")
				continue
			}

			log.Printf("PROCESSED topic=%s partition=%d offset=%d key=%s",
				msg.Topic, msg.Partition, msg.Offset, string(msg.Key))

			// Commit offset after successful processing
			session.MarkMessage(msg, "")

		case <-session.Context().Done():
			// Rebalance or shutdown signaled
			log.Printf("Session context done for topic=%s partition=%d",
				claim.Topic(), claim.Partition())
			return nil
		}
	}
}

func main() {
	config := sarama.NewConfig()
	config.Version = sarama.V3_5_0_0
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
		sarama.NewBalanceStrategySticky(), // Sticky: minimizes partition movement on rebalance
	}
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Session.Timeout = 10 * time.Second
	config.Consumer.Group.Heartbeat.Interval = 3 * time.Second

	consumerGroup, err := sarama.NewConsumerGroup([]string{brokers}, groupID, config)
	if err != nil {
		log.Fatalf("Failed to create consumer group: %v", err)
	}
	defer consumerGroup.Close()

	handler := &ConsumerGroupHandler{ready: make(chan bool)}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle SIGINT/SIGTERM for graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signals
		log.Println("Shutdown signal received, cancelling context...")
		cancel()
	}()

	for {
		// Consume blocks until rebalance or shutdown
		if err := consumerGroup.Consume(ctx, []string{topic}, handler); err != nil {
			log.Printf("Consume error: %v", err)
		}

		if ctx.Err() != nil {
			log.Println("Context cancelled, exiting consume loop")
			return
		}

		// Reset ready channel for next rebalance
		handler.ready = make(chan bool)
	}
}
