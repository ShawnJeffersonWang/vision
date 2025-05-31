package snowflake

import (
	"time"

	sf "github.com/bwmarrin/snowflake"
)

//雪花算法

var node *sf.Node

// 传入一个开始时间（startTime，格式为 yyyy-MM-dd）和机器ID（machineID）。
// 此函数会初始化雪花算法生成的ID的起始时间和节点
func Init(startTime string, machineID int64) (err error) {
	var st time.Time
	//第一个参数是固定的时间格式模板
	st, err = time.Parse("2006-01-02", startTime)
	if err != nil {
		return
	}
	//初始化开始时间，并转换为毫秒
	sf.Epoch = st.UnixNano() / 1000000
	//初始化节点
	node, err = sf.NewNode(machineID)
	return
}

// 调用函数生成一个新的ID
func GenID() int64 {
	return node.Generate().Int64()
}
