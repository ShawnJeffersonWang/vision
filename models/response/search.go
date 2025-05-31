package response

type SearchResponse struct {
	Id      int64  `json:"id"`
	Name    string `json:"name"`
	Snippet string `json:"snippet"` // 片段
}
