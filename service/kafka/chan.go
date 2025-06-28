package kafka

import (
	"agricultural_vision/constants"
	"agricultural_vision/models/proto"
	"agricultural_vision/settings"
	"context"
	"errors"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	protobuf "google.golang.org/protobuf/proto"
	"sync"
)

// kafka/producer.go

var (
	messageQueue chan *proto.PostCreationMessage
	producerOnce sync.Once
)

// InitProducer 初始化全局生产者和消息队列
func InitProducer() error {
	producerOnce.Do(func() {
		conf := settings.Conf.KafkaConfig
		if !conf.Enabled {
			return
		}

		// 创建缓冲通道作为消息队列
		messageQueue = make(chan *proto.PostCreationMessage, 1000)

		writer := &kafka.Writer{
			Addr:         kafka.TCP(conf.Brokers...),
			Topic:        conf.TopicPostCreation,
			Balancer:     &kafka.LeastBytes{},
			MaxAttempts:  conf.RetryMax,
			WriteTimeout: conf.WriteTimeout,
			ReadTimeout:  conf.ReadTimeout,
			RequiredAcks: kafka.RequireOne,
			Compression:  kafka.Zstd,
		}

		producer = &Producer{
			config: conf,
			writer: writer,
		}

		go producer.startWorker()
	})

	if producer != nil && producer.writer == nil {
		return constants.ErrKafkaNotEnabled
	}
	return nil
}

// startWorker 启动工作协程处理消息队列
func (p *Producer) startWorker() {
	defer p.writer.Close()

	for msg := range messageQueue {
		msgBytes, err := protobuf.Marshal(msg)
		if err != nil {
			zap.L().Error("Protobuf 序列化失败", zap.Error(err))
			continue
		}

		err = p.writer.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte(msg.MessageId),
			Value: msgBytes,
			Time:  msg.CreatedAt.AsTime(),
		})

		if err != nil {
			zap.L().Error("发送 Kafka 消息失败", zap.Error(err))
			// 可添加重试逻辑或消息回退机制
		}
	}
}

// SendPostCreationMessage 将消息放入队列而非直接发送
func SendPostCreationMessage(message *proto.PostCreationMessage) error {
	if producer == nil || messageQueue == nil {
		return constants.ErrKafkaNotEnabled
	}

	// 非阻塞发送：如果队列满则记录错误并继续
	select {
	case messageQueue <- message:
		return nil
	default:
		zap.L().Warn("Kafka 消息队列已满，丢弃消息", zap.String("message_id", message.MessageId))
		return errors.New("kafka message queue is full")
	}
}

// Close 关闭生产者和消息队列
func (p *Producer) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}

	// 关闭消息队列
	if messageQueue != nil {
		close(messageQueue)
	}

	return p.writer.Close()
}
