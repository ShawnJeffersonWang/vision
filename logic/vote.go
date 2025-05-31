package logic

import (
	"errors"
	"strconv"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"agricultural_vision/constants"
	"agricultural_vision/dao/mysql"
	"agricultural_vision/dao/redis"
	"agricultural_vision/models/request"
)

// 给帖子投票
func VoteForPost(userID int64, p *request.VoteRequest) error {
	zap.L().Debug("VoteForPost",
		zap.Int64("userID", userID),
		zap.Int64("postID", p.PostID),
		zap.Int8("direction", p.Direction),
	)

	// 在mysql中查询postID是否存在
	_, err := mysql.GetPostById(p.PostID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { // 如果查询不到帖子
			return constants.ErrorNoPost
		}
		return err
	}

	// 在redis中投票
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
	comment, err := mysql.GetCommentListByIDs(ids)
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
