package logic

import (
	"errors"
	"strconv"
	"vision/dao"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"vision/constants"
	"vision/dao/redis"
	"vision/models/request"
)

// 给帖子投票
//func VoteForPost(userID int64, p *request.VoteRequest) error {
//	zap.L().Debug("VoteForPost",
//		zap.Int64("userID", userID),
//		zap.Int64("postID", p.PostID),
//		zap.Int8("direction", p.Direction),
//	)
//
//	// 在mysql中查询postID是否存在
//	_, err := dao.GetPostById(p.PostID)
//	if err != nil {
//		if errors.Is(err, gorm.ErrRecordNotFound) { // 如果查询不到帖子
//			return constants.ErrorNoPost
//		}
//		return err
//	}
//
//	// 在redis中投票
//	return redis.VoteForPost(strconv.Itoa(int(userID)), strconv.Itoa(int(p.PostID)), float64(p.Direction))
//}

// VoteForPost 为帖子投票 (修复后：同时写MySQL和Redis)
func VoteForPost(userID int64, p *request.VoteRequest) error {
	zap.L().Debug("VoteForPost",
		zap.Int64("userID", userID),
		zap.Int64("postID", p.PostID),
		zap.Int8("direction", p.Direction),
	)

	// 1. 校验帖子是否存在
	_, err := dao.GetPostById(p.PostID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return constants.ErrorNoPost
		}
		return err
	}

	// 2. 【新增】将投票记录持久化到 MySQL
	// 这一步是关键，没有这一步，GetUserLikedPostList 就查不到数据
	if err := dao.SaveVote(userID, p.PostID, p.Direction); err != nil {
		zap.L().Error("mysql save vote failed", zap.Error(err))
		return err
	}

	// 3. 在 Redis 中投票 (保持原有逻辑，用于计算热度排行等)
	// 注意：如果你的业务完全迁移到MySQL，这步可以去掉；但通常为了高性能排行，Redis还是需要的。
	return redis.VoteForPost(strconv.Itoa(int(userID)), strconv.Itoa(int(p.PostID)), float64(p.Direction))
}

// 给评论投票
func CommentVote(userID int64, p *request.VoteRequest) error {
	zap.L().Debug("CommentVote",
		zap.Int64("userID", userID),
		zap.Int64("commentID", p.CommentID),
		zap.Int8("direction", p.Direction),
	)

	ids := []string{strconv.Itoa(int(p.CommentID))}
	comment, err := dao.GetCommentListByIDs(ids)
	if err != nil {
		return err
	}
	if len(comment) == 0 { // 如果未找到此评论
		return constants.ErrorNoComment
	}

	parentID := comment[0].ParentID
	postID := comment[0].PostID

	// 去redis中投票
	return redis.VoteForComment(strconv.Itoa(int(userID)), strconv.Itoa(int(p.CommentID)), strconv.Itoa(int(postID)), float64(p.Direction), parentID)
}
