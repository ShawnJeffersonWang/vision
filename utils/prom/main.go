package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"
)

// KeyValue 表示键值对
type KeyValue struct {
	Key   string
	Value float64
}

// Stats 存储每个key的统计信息
type Stats struct {
	Values []float64
	Sum    float64
	Min    float64
	Max    float64
	Count  int
}

// StatsCollector 统计收集器
type StatsCollector struct {
	mu    sync.RWMutex
	stats map[string]*Stats
}

// NewStatsCollector 创建新的统计收集器
func NewStatsCollector() *StatsCollector {
	return &StatsCollector{
		stats: make(map[string]*Stats),
	}
}

// Add 添加一个键值对到统计中
func (sc *StatsCollector) Add(kv KeyValue) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if _, exists := sc.stats[kv.Key]; !exists {
		sc.stats[kv.Key] = &Stats{
			Values: []float64{},
			Min:    math.MaxFloat64,
			Max:    -math.MaxFloat64,
		}
	}

	stat := sc.stats[kv.Key]
	stat.Values = append(stat.Values, kv.Value)
	stat.Sum += kv.Value
	stat.Count++

	if kv.Value < stat.Min {
		stat.Min = kv.Value
	}
	if kv.Value > stat.Max {
		stat.Max = kv.Value
	}
}

// calculatePercentile 计算分位数
func calculatePercentile(values []float64, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	index := percentile * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sorted[lower]
	}

	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// PrintStats 打印统计结果
func (sc *StatsCollector) PrintStats() {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	fmt.Println("\n========== 统计结果 ==========")

	// 按key排序输出
	keys := make([]string, 0, len(sc.stats))
	for k := range sc.stats {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		stat := sc.stats[key]
		if stat.Count == 0 {
			continue
		}

		avg := stat.Sum / float64(stat.Count)
		p50 := calculatePercentile(stat.Values, 0.5)
		p90 := calculatePercentile(stat.Values, 0.9)
		p99 := calculatePercentile(stat.Values, 0.99)

		fmt.Printf("Key: %s\n", key)
		fmt.Printf("  计数: %d\n", stat.Count)
		fmt.Printf("  最小值: %.2f\n", stat.Min)
		fmt.Printf("  最大值: %.2f\n", stat.Max)
		fmt.Printf("  平均值: %.2f\n", avg)
		fmt.Printf("  总和: %.2f\n", stat.Sum)
		fmt.Printf("  分位数 - P50: %.2f, P90: %.2f, P99: %.2f\n", p50, p90, p99)
		fmt.Println()
	}
}

// generator 生成随机键值对
func generator(ch chan<- KeyValue, done <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(ch)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			fmt.Println("生成器停止")
			return
		case <-ticker.C:
			// 每秒生成一批键值对
			batchSize := rand.Intn(100) + 1
			for i := 0; i < batchSize; i++ {
				key := string('a' + rune(rand.Intn(26)))
				value := rand.Float64()*5.0 + 0.1

				select {
				case ch <- KeyValue{Key: key, Value: value}:
				case <-done:
					fmt.Println("生成器停止")
					return
				}
			}
			fmt.Printf("生成了 %d 个键值对\n", batchSize)
		}
	}
}

// consumer 消费键值对并进行统计
func consumer(id int, ch <-chan KeyValue, collector *StatsCollector, done <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-done:
			fmt.Printf("消费者 %d 停止\n", id)
			return
		case kv, ok := <-ch:
			if !ok {
				fmt.Printf("消费者 %d 通道关闭\n", id)
				return
			}
			collector.Add(kv)
		}
	}
}

func main() {
	// 设置随机数种子
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// 创建通道和统计收集器
	ch := make(chan KeyValue, 1000)
	collector := NewStatsCollector()
	done := make(chan struct{})

	// 创建WaitGroup
	var wg sync.WaitGroup

	// 启动生成器
	wg.Add(1)
	go generator(ch, done, &wg)

	// 启动10个消费者
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go consumer(i, ch, collector, done, &wg)
	}

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT)

	fmt.Println("程序运行中... 按 Ctrl+C 停止")

	// 等待信号
	sig := <-sigChan
	fmt.Printf("\n接收到信号: %v\n", sig)

	// 关闭done通道，通知所有goroutine停止
	close(done)

	// 等待所有goroutine完成
	wg.Wait()

	// 打印统计结果
	collector.PrintStats()
}
