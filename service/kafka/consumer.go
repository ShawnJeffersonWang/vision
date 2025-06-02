package kafka

import (
	"agricultural_vision/dao/mysql"
	"agricultural_vision/dao/redis"
	"agricultural_vision/models/entity"
	"agricultural_vision/proto"
	"agricultural_vision/settings"
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	protobuf "google.golang.org/protobuf/proto"
	"time"
)

// Consumer Kafka 消费者结构体（带上下文）
type Consumer struct {
	ctx    context.Context
	reader *kafka.Reader
}

// NewConsumer 创建消费者实例（接收上下文）
func NewConsumer(ctx context.Context) (*Consumer, error) {
	conf := settings.Conf.KafkaConfig
	if !conf.Enabled {
		return nil, nil
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     conf.Brokers,
		Topic:       conf.TopicPostCreation,
		GroupID:     conf.GroupPostCreation,
		StartOffset: kafka.FirstOffset,
	})

	return &Consumer{
		ctx:    ctx,
		reader: reader,
	}, nil
}

// Start 启动消费者循环（接收消息处理函数）
func (c *Consumer) StartUseJson(handler func(PostCreationMessage) error) error {
	go func() {
		defer c.reader.Close()

		for {
			select {
			case <-c.ctx.Done(): // 监听上下文取消
				return
			default:
				msg, err := c.reader.ReadMessage(c.ctx)
				if err != nil {
					zap.L().Error("读取 Kafka 消息失败", zap.Error(err))
					time.Sleep(1 * time.Second)
					continue
				}

				var message PostCreationMessage
				if err := json.Unmarshal(msg.Value, &message); err != nil {
					zap.L().Error("解析 Kafka 消息失败", zap.Error(err))
					continue
				}

				// 处理消息（调用外部传入的处理函数）
				if err := handler(message); err != nil {
					zap.L().Error("处理 Kafka 消息失败", zap.Error(err))
				}
			}
		}
	}()

	return nil
}

// Start 启动消费者循环（接收消息处理函数）
func (c *Consumer) Start(handler func(*proto.PostCreationMessage) error) error {
	go func() {
		defer c.reader.Close()

		for {
			select {
			case <-c.ctx.Done(): // 监听上下文取消
				return
			default:
				msg, err := c.reader.ReadMessage(c.ctx)
				if err != nil {
					zap.L().Error("读取 Kafka 消息失败", zap.Error(err))
					zap.L().Debug("Kafka 配置", zap.Any("c.reader.Config()", c.reader.Config()))
					time.Sleep(1 * time.Second)
					continue
				}

				var message proto.PostCreationMessage
				// 使用 proto 包的 Unmarshal 函数
				if err := protobuf.Unmarshal(msg.Value, &message); err != nil {
					zap.L().Error("解析 Kafka 消息失败", zap.Error(err))
					continue
				}

				// 处理消息（调用外部传入的处理函数）
				if err := handler(&message); err != nil {
					zap.L().Error("处理 Kafka 消息失败", zap.Error(err))
				}
			}
		}
	}()

	return nil
}

// ProcessPostCreation 处理帖子创建消息（可导出）
func ProcessPostCreationUseJson(message PostCreationMessage) error {
	zap.L().Info("收到帖子创建消息",
		zap.String("message_id", message.MessageID),
		zap.Int64("user_id", message.UserID))

	// 构建帖子实体
	post := &entity.Post{
		Content:     message.Content,
		Image:       message.Image,
		AuthorID:    message.UserID,
		CommunityID: message.CommunityID,
	}

	// 写入数据库（示例逻辑，需根据实际业务调整）
	if err := mysql.CreatePost(post); err != nil {
		return fmt.Errorf("写入数据库失败: %w", err)
	}

	// 保存到 Redis
	if err := redis.CreatePost(post.ID, post.CommunityID); err != nil {
		zap.L().Error("保存到 Redis 失败", zap.Error(err))
	}

	return nil
}

// ProcessPostCreation 处理帖子创建消息（可导出）
func ProcessPostCreation(message *proto.PostCreationMessage) error {
	zap.L().Info("收到帖子创建消息",
		zap.String("message_id", message.MessageId),
		zap.Int64("user_id", message.UserId))

	// 构建帖子实体
	post := &entity.Post{
		Content:     message.Content,
		Image:       message.Image,
		AuthorID:    message.UserId,
		CommunityID: message.CommunityId,
	}

	// 转换 Protobuf 时间戳为 time.Time
	if message.CreatedAt != nil {
		post.CreatedAt = timestampToTime(message.CreatedAt)
	}

	// 写入数据库
	if err := mysql.CreatePost(post); err != nil {
		return fmt.Errorf("写入数据库失败: %w", err)
	}

	// 保存到 Redis
	if err := redis.CreatePost(post.ID, post.CommunityID); err != nil {
		zap.L().Error("保存到 Redis 失败", zap.Error(err))
	}

	return nil
}

// 辅助函数：将 Protobuf 时间戳转换为 time.Time
func timestampToTime(ts *timestamp.Timestamp) time.Time {
	return time.Unix(ts.Seconds, int64(ts.Nanos)).UTC()
}
