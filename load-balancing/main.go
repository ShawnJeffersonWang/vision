package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// Server 表示一个后端服务器
type Server struct {
	ID          string
	Weight      int32 // 配置的权重
	Current     int32 // 当前有效权重（考虑slow_start）
	StartTime   time.Time
	SlowStart   time.Duration
	Healthy     atomic.Bool
	Connections atomic.Int32
}

// WeightedRandom 加权随机负载均衡器
type WeightedRandom struct {
	servers     []*Server
	totalWeight atomic.Int32
	mu          sync.RWMutex
	rand        *rand.Rand
}

// NewWeightedRandom 创建新的加权随机负载均衡器
func NewWeightedRandom() *WeightedRandom {
	return &WeightedRandom{
		servers: make([]*Server, 0),
		rand:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// AddServer 添加服务器
func (wr *WeightedRandom) AddServer(id string, weight int, slowStart time.Duration) {
	server := &Server{
		ID:        id,
		Weight:    int32(weight),
		Current:   0,
		StartTime: time.Now(),
		SlowStart: slowStart,
	}
	server.Healthy.Store(true)

	wr.mu.Lock()
	wr.servers = append(wr.servers, server)
	wr.mu.Unlock()

	// 如果没有slow_start，直接设置为最大权重
	if slowStart == 0 {
		server.Current = server.Weight
	}
}

// updateWeights 更新所有服务器的权重（考虑slow_start）
func (wr *WeightedRandom) updateWeights() {
	var total int32
	now := time.Now()

	for _, server := range wr.servers {
		if !server.Healthy.Load() {
			server.Current = 0
			continue
		}

		// 计算slow_start期间的权重
		if server.SlowStart > 0 {
			elapsed := now.Sub(server.StartTime)
			if elapsed < server.SlowStart {
				// 线性增长权重
				ratio := float64(elapsed) / float64(server.SlowStart)
				server.Current = int32(float64(server.Weight) * ratio)
				if server.Current == 0 {
					server.Current = 1 // 至少保证有1的权重
				}
			} else {
				server.Current = server.Weight
			}
		}

		total += server.Current
	}

	wr.totalWeight.Store(total)
}

// Select 选择一个服务器
func (wr *WeightedRandom) Select() *Server {
	wr.mu.RLock()
	defer wr.mu.RUnlock()

	if len(wr.servers) == 0 {
		return nil
	}

	// 更新权重
	wr.updateWeights()

	totalWeight := wr.totalWeight.Load()
	if totalWeight <= 0 {
		return nil
	}

	// 生成随机数
	randomWeight := wr.rand.Int31n(totalWeight)

	// 根据权重选择服务器
	var currentWeight int32
	for _, server := range wr.servers {
		if server.Current <= 0 {
			continue
		}
		currentWeight += server.Current
		if randomWeight < currentWeight {
			return server
		}
	}

	return nil
}

// SetHealthy 设置服务器健康状态
func (wr *WeightedRandom) SetHealthy(serverID string, healthy bool) {
	wr.mu.RLock()
	defer wr.mu.RUnlock()

	for _, server := range wr.servers {
		if server.ID == serverID {
			server.Healthy.Store(healthy)
			if healthy {
				// 重置启动时间，重新开始slow_start
				server.StartTime = time.Now()
				if server.SlowStart > 0 {
					server.Current = 0
				}
			}
			break
		}
	}
}

// 演示程序
func main() {
	lb := NewWeightedRandom()

	// 添加服务器，server3 启用 30秒的 slow_start
	lb.AddServer("server1", 5, 0)
	lb.AddServer("server2", 3, 0)
	lb.AddServer("server3", 2, 30*time.Second)

	// 模拟请求分布
	counts := make(map[string]int)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	fmt.Println("开始测试加权随机算法 (server3 有 30s slow_start)...")

	// 运行35秒，观察slow_start效果
	timeout := time.After(35 * time.Second)
	secondsPassed := 0

	for {
		select {
		case <-ticker.C:
			secondsPassed++

			// 每秒发送100个请求
			secondCounts := make(map[string]int)
			for i := 0; i < 100; i++ {
				server := lb.Select()
				if server != nil {
					counts[server.ID]++
					secondCounts[server.ID]++
				}
			}

			// 打印当前秒的分布
			fmt.Printf("\n第 %d 秒 - 请求分布: ", secondsPassed)
			for id, count := range secondCounts {
				fmt.Printf("%s: %d ", id, count)
			}

			// 打印当前权重
			fmt.Printf("\n当前权重: ")
			lb.mu.RLock()
			for _, server := range lb.servers {
				fmt.Printf("%s: %d/%d ", server.ID, server.Current, server.Weight)
			}
			lb.mu.RUnlock()

		case <-timeout:
			fmt.Printf("\n\n测试完成！总请求分布:\n")
			total := 0
			for id, count := range counts {
				total += count
				fmt.Printf("%s: %d 次\n", id, count)
			}
			fmt.Printf("总请求数: %d\n", total)
			return
		}
	}
}
