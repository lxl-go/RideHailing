package kafkax

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// Message 统一消息体（对标文档 4.7 节 Kafka 消息体规范）
type Message struct {
	Topic     string
	Key       string
	Value     []byte
	Headers   map[string]string
}

// Event 通用事件结构（对标文档统一消息格式）
type Event struct {
	EventType string      `json:"event_type"`
	Timestamp string      `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// NewEvent 创建标准事件（对标文档 4.7 节消息体模板）
func NewEvent(eventType string, data interface{}) Event {
	return Event{
		EventType: eventType,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Data:      data,
	}
}

// Producer Kafka 生产者封装（对标文档生产者模板）
type Producer struct {
	writer *kafka.Writer
}

// ProducerConfig 生产者配置
type ProducerConfig struct {
	Brokers []string
	Topic   string
}

// NewProducer 创建生产者
func NewProducer(cfg ProducerConfig) *Producer {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Brokers...),
		Topic:                  cfg.Topic,
		Balancer:               &kafka.LeastBytes{},
		BatchTimeout:           10 * time.Millisecond,
		RequiredAcks:           kafka.RequireOne,
		AllowAutoTopicCreation: true,
	}
	return &Producer{writer: w}
}

// Send 发送消息
func (p *Producer) Send(ctx context.Context, msg Message) error {
	kafkaMsg := kafka.Message{
		Key:   []byte(msg.Key),
		Value: msg.Value,
	}
	for k, v := range msg.Headers {
		kafkaMsg.Headers = append(kafkaMsg.Headers, kafka.Header{Key: k, Value: []byte(v)})
	}
	return p.writer.WriteMessages(ctx, kafkaMsg)
}

// SendEvent 发送标准事件（对标文档 EventProducer.Publish 模板）
func (p *Producer) SendEvent(ctx context.Context, eventType string, data interface{}) error {
	event := NewEvent(eventType, data)
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event failed: %w", err)
	}
	key := fmt.Sprintf("%d", time.Now().UnixNano())
	return p.Send(ctx, Message{Key: key, Value: payload})
}

// Close 关闭生产者
func (p *Producer) Close() error {
	return p.writer.Close()
}

// Consumer Kafka 消费者封装（对标文档消费者模板）
type Consumer struct {
	reader *kafka.Reader
}

// ConsumerConfig 消费者配置
type ConsumerConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

// NewConsumer 创建消费者
func NewConsumer(cfg ConsumerConfig) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.Brokers,
		Topic:       cfg.Topic,
		GroupID:     cfg.GroupID,
		StartOffset: kafka.LastOffset,
		MinBytes:    1,
		MaxBytes:    10e6,
		MaxWait:     1 * time.Second,
	})
	return &Consumer{reader: r}
}

// Consume 消费消息循环（对标文档 StartConsumer 模板）
func (c *Consumer) Consume(ctx context.Context, handler func(ctx context.Context, msg Message) error) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			kafkaMsg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				continue
			}
			msg := Message{
				Key:   string(kafkaMsg.Key),
				Value: kafkaMsg.Value,
			}
			if err := handler(ctx, msg); err != nil {
				// 记录错误日志，继续消费（对标文档：不阻断消费）
				continue
			}
		}
	}
}

// Close 关闭消费者
func (c *Consumer) Close() error {
	return c.reader.Close()
}
