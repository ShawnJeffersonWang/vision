package response

// 社区简略信息
type CommunityBriefResponse struct {
	ID            int64  `json:"id" db:"community_id"`
	CommunityName string `json:"name" db:"community_name"`
}
