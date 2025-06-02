package settings

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

var (
	Conf      = new(AppConfig)
	etcdMutex = concurrency.NewMutex(nil, "config-lock")
)

// EtcdConfig 包含etcd连接信息
type EtcdConfig struct {
	Endpoints   []string
	DialTimeout time.Duration
	Username    string // 可选
	Password    string // 可选
}

// 从etcd加载配置
func LoadConfigFromEtcd(etcdConf EtcdConfig, configKey string) error {
	clientCfg := clientv3.Config{
		Endpoints:   etcdConf.Endpoints,
		DialTimeout: etcdConf.DialTimeout,
		Username:    etcdConf.Username,
		Password:    etcdConf.Password,
	}

	cli, err := clientv3.New(clientCfg)
	if err != nil {
		return fmt.Errorf("创建etcd客户端失败: %w", err)
	}
	defer cli.Close()

	// 使用会话锁保证配置读取原子性
	sess, err := concurrency.NewSession(cli)
	if err != nil {
		return fmt.Errorf("创建etcd会话失败: %w", err)
	}
	defer sess.Close()

	mutex := concurrency.NewMutex(sess, configKey)

	if err := mutex.Lock(context.Background()); err != nil {
		return fmt.Errorf("获取配置锁失败: %w", err)
	}
	defer mutex.Unlock(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	resp, err := cli.Get(ctx, configKey)
	cancel()
	if err != nil {
		return fmt.Errorf("读取etcd配置失败: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return fmt.Errorf("配置键 %s 不存在", configKey)
	}

	var tmpConfig AppConfig
	if err := json.Unmarshal(resp.Kvs[0].Value, &tmpConfig); err != nil {
		return fmt.Errorf("解析配置JSON失败: %w\n原始数据: %s", err, resp.Kvs[0].Value)
	}

	etcdMutex.Lock(context.Background()) // 修复：添加上下文参数
	defer etcdMutex.Unlock(context.Background())
	*Conf = tmpConfig

	go startWatch(cli, configKey)

	return nil
}

// 启动配置变更监听
func startWatch(cli *clientv3.Client, configKey string) {
	rch := cli.Watch(context.Background(), configKey)
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				fmt.Printf("配置已更新（操作类型: %s，键: %s）\n", ev.Type, ev.Kv.Key)
				if err := handleConfigUpdate(ev.Kv.Value); err != nil {
					fmt.Printf("处理配置更新失败: %v\n", err)
				}
			case clientv3.EventTypeDelete:
				fmt.Printf("配置已删除（操作类型: %s，键: %s）\n", ev.Type, ev.Kv.Key)
			}
		}
	}
}

// 处理配置更新
func handleConfigUpdate(data []byte) error {
	etcdMutex.Lock(context.Background()) // 修复：添加上下文参数
	defer etcdMutex.Unlock(context.Background())

	var newConfig AppConfig
	if err := json.Unmarshal(data, &newConfig); err != nil {
		return fmt.Errorf("解析更新的配置失败: %w", err)
	}

	*Conf = newConfig
	fmt.Println("配置已重新加载")
	return nil
}

// 初始化配置
func InitWithEtcd() error {
	etcdConf := EtcdConfig{
		Endpoints:   []string{"etcd-service.agricultural-vision.svc.cluster.local:2379"},
		DialTimeout: 5 * time.Second,
	}

	if err := LoadConfigFromEtcd(etcdConf, "agricultural-vision/config"); err != nil {
		fmt.Printf("etcd加载失败，启用文件回退: %v\n", err)
		if err := loadFromFile(); err != nil {
			return fmt.Errorf("文件加载失败: %w", err)
		}
	}

	return nil
}

// 从本地文件加载配置
func loadFromFile() error {
	viper.SetConfigFile("./conf/config.yaml")
	viper.AutomaticEnv()

	viper.BindEnv("mysql.password", "MYSQL_PASSWORD")
	viper.BindEnv("jwt.secret", "JWT_SECRET")
	viper.BindEnv("ai.api_key", "AI_API_KEY")
	viper.BindEnv("alioss.access_key_id", "OSS_ACCESS_KEY_ID")
	viper.BindEnv("alioss.access_key_secret", "OSS_ACCESS_KEY_SECRET")

	// 修复：使用正确的方法名
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Printf("检测到文件变更: %s\n", e.Name)
		if err := viper.Unmarshal(&Conf); err != nil {
			fmt.Printf("文件解析失败: %v\n", err)
		}
	})

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}
	return viper.Unmarshal(&Conf)
}
