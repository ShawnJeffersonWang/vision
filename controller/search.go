package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"agricultural_vision/constants"
	"agricultural_vision/dao/mysql"
	"agricultural_vision/models/entity"
	"agricultural_vision/models/response"
)

// 关键词搜索
func SearchHandler(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		ResponseError(c, http.StatusBadRequest, constants.CodeEmptyKeyword)
		return
	}

	var results []entity.CropDetail
	query := "%" + keyword + "%" // 模糊匹配
	err := mysql.DB.Where("name LIKE ? OR description LIKE ? OR introduction LIKE ?", query, query, query).
		Find(&results).Error
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}

	// 处理匹配片段并高亮关键词
	var searchResults []response.SearchResponse
	for _, crop := range results {
		var matchedText string
		if strings.Contains(crop.Description, keyword) {
			matchedText = crop.Description
			snippet := extractSnippet(matchedText, keyword)

			searchResults = append(searchResults, response.SearchResponse{
				Id:      crop.Id,
				Name:    crop.Name,
				Snippet: snippet,
			})
		} else if strings.Contains(crop.Introduction, keyword) {
			matchedText = crop.Introduction
			snippet := extractSnippet(matchedText, keyword)

			searchResults = append(searchResults, response.SearchResponse{
				Id:      crop.Id,
				Name:    crop.Name,
				Snippet: snippet,
			})
		}
	}
	ResponseSuccess(c, searchResults)
}

// 提取包含关键词的片段并高亮显示
func extractSnippet(text, keyword string) string {
	if text == "" || keyword == "" {
		return ""
	}

	// 去掉换行符
	text = strings.ReplaceAll(text, "\n", "")

	// 将文本和关键字转换为 rune 切片，防止中文字符索引错误
	runeText := []rune(text)
	runeKeyword := []rune(keyword)
	textLen := len(runeText)
	keywordLen := len(runeKeyword)

	var start int
	for i := 0; i <= textLen-keywordLen; i++ {
		if string(runeText[i:i+keywordLen]) == keyword {
			start = i
			break
		}
	}

	// 如果没有找到关键词，返回原文本
	if start == 0 {
		fmt.Println("未找到关键词:", keyword, "文本:", text) // 调试信息
		return ""
	}

	// 计算 start 和 end 索引
	startIdx := max(0, start-10)
	endIdx := min(textLen, start+keywordLen+10)

	// 确保 startIdx 不大于 endIdx
	if startIdx >= endIdx {
		fmt.Println("startIdx 大于 endIdx，文本:", text) // 调试信息
		return ""
	}

	// 提取片段
	snippet := string(runeText[startIdx:endIdx])

	// 高亮所有匹配的关键字
	highlightedKeyword := `<font color='red'><b>` + keyword + `</b></font>`
	snippet = strings.ReplaceAll(snippet, keyword, highlightedKeyword)

	// **前后加省略号**
	if startIdx > 0 {
		snippet = "..." + snippet
	}
	if endIdx < textLen {
		snippet += "..."
	}

	return snippet
}

// 农作物搜索
func SearchCropHandler(c *gin.Context) {
	cropIDStr := c.Param("crop_id")
	cropID, err := strconv.ParseInt(cropIDStr, 10, 64)
	if err != nil {
		ResponseError(c, http.StatusBadRequest, constants.CodeInvalidParam)
		return
	}

	var crop entity.CropDetail

	err = mysql.DB.Where("id = ?", cropID).Find(&crop).Error
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}
	ResponseSuccess(c, crop)
}
