package logic

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"go.uber.org/zap"

	"agricultural_vision/constants"
	"agricultural_vision/models/response"
	"agricultural_vision/settings"
)

var userConversations = make(map[int64]map[int64]*response.Conversation) // 使用 map 保存每个用户的不同ai模型的对话历史
var mutex = sync.Mutex{}                                                 // 保护 map 的并发访问

func AiTalk(userInput string, userID, aiModel int64) (aiResponse *response.AiResponse, err error) {
	aiResponse = new(response.AiResponse)

	// 初始化systemContent，用来保存不同ai模型的系统提示
	systemContent := [5]string{
		"",
		settings.Conf.AiConfig.SystemContent1,
		settings.Conf.AiConfig.SystemContent2,
		settings.Conf.AiConfig.SystemContent3,
		settings.Conf.AiConfig.SystemContent4,
	}

	// 获取或创建该用户的对话历史
	mutex.Lock() // 锁住整个 map，确保线程安全
	// 获取该用户的对话历史
	userConversation, exists := userConversations[userID]
	// 如果该用户不存在，则创建一个空的对话历史
	if !exists {
		userConversations[userID] = make(map[int64]*response.Conversation)
		userConversation = userConversations[userID]
	}

	// 获取该用户对应 AI 角色的对话历史
	userAIConversation, exists := userConversation[aiModel]
	if !exists {
		userAIConversation = &response.Conversation{
			Messages: []response.Message{
				{Content: systemContent[aiModel], Role: "system"}, // 每个AI角色的初始设定
			},
		}
		userConversation[aiModel] = userAIConversation
	}
	mutex.Unlock()

	// 将用户输入添加到对应的ai对话历史中
	userAIConversation.Mutex.Lock()
	userAIConversation.Messages = append(userAIConversation.Messages, response.Message{Content: userInput, Role: "user"})
	userAIConversation.Mutex.Unlock()

	// 向智谱清言发送请求
	apiKey := settings.Conf.AiConfig.ApiKey
	apiURL := settings.Conf.AiConfig.ApiUrl

	// 构建请求体
	body := map[string]interface{}{
		"model":    settings.Conf.AiConfig.Model, // 请根据需要选择模型
		"messages": userAIConversation.Messages,
	}

	// 序列化请求体
	jsonData, err := json.Marshal(body)
	if err != nil {
		zap.L().Error("序列化请求体失败", zap.Error(err))
		return
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		zap.L().Error("创建HTTP请求失败", zap.Error(err))
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// 发送请求并获取响应
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		zap.L().Error("发送请求失败", zap.Error(err))
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// 读取响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		zap.L().Error("读取响应体失败", zap.Error(err))
		return
	}

	// 解析 AI 响应
	var apiResponse response.ApiResponse
	err = json.Unmarshal(bodyBytes, &apiResponse)
	if err != nil {
		zap.L().Error("解析AI响应失败", zap.Error(err))
		return
	}

	// 获取 AI 的回答
	if len(apiResponse.Choices) > 0 {
		aiAnswer := apiResponse.Choices[0].Message.Content

		// 将 AI 的回答添加到对话历史中
		userAIConversation.Mutex.Lock()
		userAIConversation.Messages = append(userAIConversation.Messages, response.Message{Content: aiAnswer, Role: "assistant"})
		userAIConversation.Mutex.Unlock()

		// 返回 AI 的回答给前端
		aiResponse.Answer = aiAnswer
		return
	} else {
		return nil, constants.ErrorAiNotAnswer
	}
}
