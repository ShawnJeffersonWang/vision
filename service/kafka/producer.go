package kafka

import (
	"agricultural_vision/constants"
	"agricultural_vision/proto"
	"agricultural_vision/settings"
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	protobuf "google.golang.org/protobuf/proto" // 导入 Protobuf 包
)

var (
	producer *Producer
	//producerOnce sync.Once
)

// Producer 定义 Kafka 生产者结构体
type Producer struct {
	config *settings.KafkaConfig
	writer *kafka.Writer
}

// InitProducer 初始化全局生产者（单例模式）
//func InitProducer() error {
//	producerOnce.Do(func() {
//		conf := settings.Conf.KafkaConfig
//		if !conf.Enabled {
//			return // 未启用则不初始化
//		}
//
//		writer := &kafka.Writer{
//			Addr:         kafka.TCP(conf.Brokers...),
//			Topic:        conf.TopicPostCreation,
//			Balancer:     &kafka.LeastBytes{},
//			MaxAttempts:  conf.RetryMax,
//			WriteTimeout: conf.WriteTimeout,
//			ReadTimeout:  conf.ReadTimeout,
//			RequiredAcks: kafka.RequireOne,
//			Compression:  kafka.Zstd,
//			//Async:        true,  // 启用异步模式
//		}
//
//		producer = &Producer{
//			config: conf,
//			writer: writer,
//		}
//	})
//
//	// 检查是否初始化成功（未启用时 producer 为 nil）
//	if producer != nil && producer.writer == nil {
//		return constants.ErrKafkaNotEnabled
//	}
//	return nil
//}

// Produce 发送消息到 Kafka（实例方法）
func (p *Producer) ProduceUseJson(message PostCreationMessage) error {
	if p == nil || p.writer == nil || !p.config.Enabled {
		return constants.ErrKafkaNotEnabled
	}

	msgBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(message.MessageID),
		Value: msgBytes,
		Time:  message.CreatedAt,
	})
}

// Produce 发送消息到 Kafka（实例方法）
func (p *Producer) Produce(message *proto.PostCreationMessage) error {
	if p == nil || p.writer == nil || !p.config.Enabled {
		return constants.ErrKafkaNotEnabled
	}

	// 使用 proto 包的 Marshal 函数
	msgBytes, err := protobuf.Marshal(message)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(message.MessageId),
		Value: msgBytes,
		Time:  message.CreatedAt.AsTime(),
	})
}

// SendPostCreationMessage 发送消息到 Kafka（全局方法）
func SendPostCreationMessageUseJson(message PostCreationMessage) error {
	if producer == nil {
		return constants.ErrKafkaNotEnabled
	}

	return producer.ProduceUseJson(message)
}

// SendPostCreationMessage 发送消息到 Kafka（全局方法）
//func SendPostCreationMessage(message *proto.PostCreationMessage) error {
//	if producer == nil {
//		return constants.ErrKafkaNotEnabled
//	}
//
//	return producer.Produce(message)
//}

// Close 关闭生产者连接（实例方法）
//func (p *Producer) Close() error {
//	if p == nil || p.writer == nil {
//		return nil
//	}
//	return p.writer.Close()
//}

// CloseProducer 关闭全局生产者连接
func CloseProducer() error {
	if producer == nil {
		return nil
	}
	return producer.Close()
}
