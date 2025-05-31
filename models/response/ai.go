package response

import "sync"

// 向AI请求体结构体
type Message struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

// 响应体结构体
type Choice struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
}

// 接收AI响应的结构体
type ApiResponse struct {
	Choices []Choice `json:"choices"`
}

// 定义用于保存对话上下文的结构体
type Conversation struct {
	Messages []Message
	Mutex    sync.Mutex // 确保线程安全
}

// ai回答
type AiResponse struct {
	Answer string `json:"answer"` // AI 的回答
}
