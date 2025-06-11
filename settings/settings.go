package settings

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"time"
)

//var Conf = new(AppConfig)

type AppConfig struct {
	Name             string `mapstructure:"name"`
	Mode             string `mapstructure:"mode"`
	Version          string `mapstructure:"version"`
	StartTime        string `mapstructure:"start_time"`
	MachineID        int64  `mapstructure:"machine_id"`
	Port             int    `mapstructure:"port"`
	*LogConfig       `mapstructure:"log"`
	*MySQLConfig     `mapstructure:"mysql"`
	*DragonflyConfig `mapstructure:"dragonfly"`
	*AiConfig        `mapstructure:"ai"`
	*AliossConfig    `mapstructure:"alioss"`
	*JWTConfig       `mapstructure:"jwt"`   // 新增JWT配置
	*KafkaConfig     `mapstructure:"kafka"` // 新增 Kafka 配置
}

type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DB           string `mapstructure:"db"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Password     string `mapstructure:"password"`
	Port         int    `mapstructure:"port"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

type DragonflyConfig struct {
	Host         string `mapstructure:"host"`
	Password     string `mapstructure:"password"`
	Port         int    `mapstructure:"port"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

// 新增：Kafka 配置结构体
type KafkaConfig struct {
	Enabled           bool     `mapstructure:"enabled"`             // 是否启用 Kafka
	Brokers           []string `mapstructure:"brokers"`             // Kafka 地址列表（例如：["kafka:9092"]）
	TopicPostCreation string   `mapstructure:"topic_post_creation"` // 发布帖子主题
	GroupPostCreation string   `mapstructure:"group_post_creation"` // 消费者组 ID
	// 可选：其他配置（如消息重试、超时时间等）
	RetryMax     int           `mapstructure:"retry_max"`     // 最大重试次数
	WriteTimeout time.Duration `mapstructure:"write_timeout"` // 写入超时时间
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`  // 读取超时时间
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
}

type AiConfig struct {
	SystemContent1 string `mapstructure:"system_content1"`
	SystemContent2 string `mapstructure:"system_content2"`
	SystemContent3 string `mapstructure:"system_content3"`
	SystemContent4 string `mapstructure:"system_content4"`
	ApiKey         string `mapstructure:"api_key"`
	ApiUrl         string `mapstructure:"api_url"`
	Model          string `mapstructure:"model"`
}

type AliossConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyId     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	BucketName      string `mapstructure:"bucket_name"`
	UserAvatarPath  string `mapstructure:"user_avatar_path"`
	PostImagePtah   string `mapstructure:"post_image_path"`
}

// JWTConfig JWT配置结构体
type JWTConfig struct {
	Secret             string `mapstructure:"secret"`
	Issuer             string `mapstructure:"issuer"`
	AccessExpireHours  int    `mapstructure:"access_expire_hours"`  // 访问token过期时间（小时）
	RefreshExpireHours int    `mapstructure:"refresh_expire_hours"` // 刷新token过期时间（小时）
}

func LoadConfig() (*MySQLConfig, error) {
	viper.SetConfigFile("./conf/config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	var config MySQLConfig
	if err := viper.Sub("mysql").Unmarshal(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

func Init() error {
	viper.SetConfigFile("./conf/config.yaml")
	viper.AutomaticEnv()

	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println("配置文件已被修改")
		_ = viper.Unmarshal(&Conf)
	})

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("ReadInConfig failed, err: %v", err))
	}

	err = viper.Unmarshal(&Conf)
	if err != nil {
		panic(fmt.Errorf("unmarshal to Conf failed, err:%v", err))
	}
	return err
}
