package redis

// redis key 注意使用命名空间
const (
	Prefix = "agricultural_vision:" // 项目 key 前缀

	// 帖子相关
	KeyPostTimeZSet    = "post:time"   // zset; key=post:time, 成员=postID, 分数=发帖时间
	KeyPostScoreZSet   = "post:score"  // zset; key=post:score, 成员=postID, 分数=投票分数（热度排序）
	KeyPostVotedZSetPF = "post:voted:" // zset; key=post:voted:{postID}, 成员=userID, 分数=1(点赞) / -1(踩)
	KeyCommunitySetPF  = "community:"  // set; key=community:{communityID}, 成员=postID（社区帖子集合）

	// 针对高并发单独维护计数的key
	KeyPostCommentNumZSet = "post:comment_num" // zset; key=post:comment_num, 成员=postID, 分数=总评论数
	KeyCommentNumZSet     = "comment:num"      // zset; key=comment:num, 成员=commentID, 分数=子评论数

	// 评论相关
	KeyCommentTimeZSetPF  = "comment:time:"  // zset; key=comment:time:{postID}, 成员=commentID, 分数=评论时间
	KeyCommentScoreZSetPF = "comment:score:" // zset; key=comment:score:{postID}, 成员=commentID, 分数=评论投票分数
	KeyCommentVotedZSetPF = "comment:voted:" // zset; key=comment:voted:{commentID}, 成员=userID, 分数=1(点赞) / -1(踩)

	// 用户点赞相关
	KeyUserLikedPostsSetPF    = "user_liked:posts:"    // set; key=post:voted:{userID}, 成员=postID
	KeyUserLikedCommentsSetPF = "user_liked:comments:" // set; key=post:voted:{userID}, 成员=commentID
)

func getRedisKey(key string) string {
	return Prefix + key
}
