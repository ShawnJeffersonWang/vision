package logic

import (
	"errors"
	"strconv"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"agricultural_vision/constants"
	"agricultural_vision/dao/mysql"
	"agricultural_vision/dao/redis"
	"agricultural_vision/models/entity"
	"agricultural_vision/models/request"
	"agricultural_vision/models/response"
)

// 创建评论
func CreateComment(createCommentRequest *request.CreateCommentRequest, userID int64) (*response.CommentResponse, error) {
	// 在mysql中查询postID是否存在
	_, err := mysql.GetPostById(createCommentRequest.PostID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { // 如果查询不到帖子
			return nil, constants.ErrorNoPost
		}
		return nil, err
	}

	// 在mysql中创建评论
	comment := &entity.Comment{
		Content:  createCommentRequest.Content,
		ParentID: createCommentRequest.ParentID,
		RootID:   createCommentRequest.RootID,
		AuthorID: userID,
		PostID:   createCommentRequest.PostID,
	}
	if err := mysql.CreateComment(comment); err != nil {
		return nil, err
	}

	// 在redis中创建评论
	if createCommentRequest.RootID == nil {
		// 创建顶级评论
		err = redis.CreateTopComment(comment.ID, comment.PostID)
	} else {
		// 创建子评论
		err = redis.CreateSonComment(*createCommentRequest.RootID, createCommentRequest.PostID)
	}
	if err != nil {
		return nil, err
	}

	// 查询作者信息
	author, err := mysql.GetUserBriefInfo(userID)
	if err != nil {
		return nil, err
	}

	// 如果是二级以上评论（不展示回复数，展示父评论作者信息）
	if comment.ParentID != nil && *comment.ParentID != *comment.RootID {
		// 查询父评论的作者信息
		parentUserinfo, err := mysql.GetUserBriefInfoByCommentID(*comment.ParentID)
		if err != nil {
			return nil, err
		}
		commentResponse := &response.CommentResponse{
			ID:        comment.ID,
			Content:   comment.Content,
			Author:    author,
			Parent:    parentUserinfo,
			CreatedAt: comment.CreatedAt.Format("2006-01-02 15:04:05"),
			RootID:    *comment.RootID,
			ParentID:  *comment.ParentID,
		}
		return commentResponse, nil
	}

	// 如果是二级评论（不展示回复数）
	if comment.ParentID != nil && *comment.ParentID == *comment.RootID {
		commentResponse := &response.CommentResponse{
			ID:        comment.ID,
			Content:   comment.Content,
			Author:    author,
			CreatedAt: comment.CreatedAt.Format("2006-01-02 15:04:05"),
			RootID:    *comment.RootID,
		}
		return commentResponse, nil
	}

	// 如果是顶级评论（展示回复数）
	var repliesCount int64 = 0
	commentResponse := &response.CommentResponse{
		ID:           comment.ID,
		Content:      comment.Content,
		Author:       author,
		RepliesCount: &repliesCount,
		CreatedAt:    comment.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	return commentResponse, nil
}

// 删除评论
func DeleteComment(commentID int64, userID int64) error {
	// 先从mysql中查找评论
	ids := []string{strconv.Itoa(int(commentID))}
	comment, err := mysql.GetCommentListByIDs(ids)
	if err != nil {
		return err
	}
	if len(comment) == 0 { // 如果未找到此评论
		return constants.ErrorNoComment
	}

	// 校验userID
	if comment[0].AuthorID != userID {
		return constants.ErrorNoPermission
	}

	// 在mysql中删除评论
	if err := mysql.DeleteComment(commentID); err != nil {
		return err
	}

	// 在redis中删除评论
	if err := redis.DeleteComment(commentID, comment[0].PostID, comment[0].RootID); err != nil {
		return err
	}

	return nil
}

// 查询单个帖子的顶级评论
func GetTopCommentList(postID int64, listRequest *request.ListRequest, userID int64) (commentListResponse *response.CommentListResponse, err error) {
	commentListResponse = &response.CommentListResponse{
		Comments: []*response.CommentResponse{},
	}

	//从redis中，根据指定的排序方式和查询数量，查询符合条件的顶级评论id列表
	ids, total, err := redis.GetTopCommentIDsInOrder(listRequest, postID)
	if err != nil {
		return
	}
	commentListResponse.Total = total
	if len(ids) == 0 {
		return
	}

	//根据id列表去数据库查询评论详细信息
	comments, err := mysql.GetCommentListByIDs(ids)
	if err != nil {
		return
	}

	// 查询所有顶级评论的赞成票数——切片
	voteData, err := redis.GetCommentVoteDataByIDs(ids)
	if err != nil {
		return
	}

	// 查询所有顶级评论的子评论数——切片
	commentNum, err := redis.GetSonCommentNumByIDs(ids)

	//将帖子作者及分区信息查询出来填充到帖子中
	for idx, comment := range comments {
		//查询作者简略信息
		userBriefInfo, err := mysql.GetUserBriefInfo(comment.AuthorID)
		if err != nil { // 遇到错误不返回，继续执行后续逻辑
			zap.L().Error("查询作者信息失败", zap.Error(err))
			continue
		}

		//查询当前用户是否点赞了此评论
		liked, err := redis.IsUserLikedComment(strconv.Itoa(int(userID)), strconv.Itoa(int(comment.ID)))
		if err != nil { // 遇到错误不返回，继续执行后续逻辑
			zap.L().Error("查询用户是否点赞失败", zap.Error(err))
			continue
		}

		//封装查询到的信息
		repliesCount := int64(commentNum[idx])
		commentResponse := &response.CommentResponse{
			ID:           comment.ID,
			Content:      comment.Content,
			Author:       userBriefInfo,
			LikeCount:    voteData[idx],
			Liked:        liked,
			RepliesCount: &repliesCount,
			CreatedAt:    comment.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		commentListResponse.Comments = append(commentListResponse.Comments, commentResponse)
	}
	return
}

// 查询单个顶级评论的子评论
func GetSonCommentList(rootID int64, listRequest *request.ListRequest, userID int64) (commentListResponse *response.CommentListResponse, err error) {
	commentListResponse = &response.CommentListResponse{
		Comments: []*response.CommentResponse{},
	}

	// 从mysql中查询子评论
	comments, total, err := mysql.GetSonCommentList(rootID, listRequest.Page, listRequest.Size)
	if err != nil {
		return
	}
	commentListResponse.Total = total
	if len(comments) == 0 {
		return
	}

	// 将子评论的ID提取出来
	commentIDs := make([]string, 0, len(comments))
	for _, comment := range comments {
		commentIDs = append(commentIDs, strconv.FormatInt(comment.ID, 10))
	}

	// 查询所有子评论的赞成票数——切片
	voteData, err := redis.GetCommentVoteDataByIDs(commentIDs)
	if err != nil {
		return
	}

	// 将帖子作者及分区信息查询出来填充到帖子中
	for idx, comment := range comments {
		//查询作者简略信息
		userBriefInfo, err := mysql.GetUserBriefInfo(comment.AuthorID)
		if err != nil {
			zap.L().Error("查询作者信息失败", zap.Error(err))
			continue
		}

		//查询当前用户是否点赞了此评论
		liked, err := redis.IsUserLikedComment(strconv.Itoa(int(userID)), strconv.Itoa(int(comment.ID)))
		if err != nil {
			zap.L().Error("查询用户是否点赞失败", zap.Error(err))
			continue
		}

		//如果是二级以上评论，则需要查询父评论的作者信息
		if *comment.ParentID != *comment.RootID {
			//查询父评论的作者简略信息
			parentUserBriefInfo, err := mysql.GetUserBriefInfoByCommentID(*comment.ParentID)
			if err != nil {
				zap.L().Error("查询父评论作者信息失败", zap.Error(err))
				continue
			}

			commentResponse := &response.CommentResponse{
				ID:        comment.ID,
				Content:   comment.Content,
				Author:    userBriefInfo,
				LikeCount: voteData[idx],
				Liked:     liked,
				Parent:    parentUserBriefInfo,
				CreatedAt: comment.CreatedAt.Format("2006-01-02 15:04:05"),
				RootID:    *comment.RootID,
				ParentID:  *comment.ParentID,
			}

			commentListResponse.Comments = append(commentListResponse.Comments, commentResponse)
			continue
		}

		//如果是二级评论（不展示回复数和父评论作者信息）
		commentResponse := &response.CommentResponse{
			ID:        comment.ID,
			Content:   comment.Content,
			Author:    userBriefInfo,
			LikeCount: voteData[idx],
			Liked:     liked,
			CreatedAt: comment.CreatedAt.Format("2006-01-02 15:04:05"),
			RootID:    *comment.RootID,
		}

		commentListResponse.Comments = append(commentListResponse.Comments, commentResponse)
	}
	return
}

// 查询帖子下的所有评论
func GetCommentList(postID int64, listRequest *request.ListRequest, userID int64) (*response.CommentListResponse, error) {
	commentListResponse := &response.CommentListResponse{
		Comments: []*response.CommentResponse{},
	}

	// 执行业务
	topCommentList, err := GetTopCommentList(postID, listRequest, userID)
	if err != nil {
		zap.L().Error("查询帖子的一级评论失败", zap.Error(err))
		return nil, err
	}
	commentListResponse.Total += topCommentList.Total

	// 遍历一级评论列表
	for _, topComment := range topCommentList.Comments {
		// 封装单个一级评论进响应体
		commentListResponse.Comments = append(commentListResponse.Comments, &response.CommentResponse{
			ID:           topComment.ID,
			Content:      topComment.Content,
			Author:       topComment.Author,
			LikeCount:    topComment.LikeCount,
			Liked:        topComment.Liked,
			RepliesCount: topComment.RepliesCount,
			CreatedAt:    topComment.CreatedAt,
		})

		// 获取单个一级评论的所有二级评论
		sonCommentList, err := GetSonCommentList(topComment.ID, listRequest, userID)
		if err != nil {
			zap.L().Error("查询帖子的二级评论失败", zap.Error(err))
			return nil, err
		}
		commentListResponse.Total += sonCommentList.Total

		// 遍历单个一级评论的子评论列表
		for _, sonComment := range sonCommentList.Comments {
			// 判读是否为二级评论
			if sonComment.Parent == nil {
				// 如果是二级评论
				commentListResponse.Comments = append(commentListResponse.Comments, &response.CommentResponse{
					ID:        sonComment.ID,
					Content:   sonComment.Content,
					Author:    sonComment.Author,
					LikeCount: sonComment.LikeCount,
					Liked:     sonComment.Liked,
					CreatedAt: sonComment.CreatedAt,
					RootID:    topComment.ID,
				})
				continue
			}

			// 如果是二级以上评论
			commentListResponse.Comments = append(commentListResponse.Comments, &response.CommentResponse{
				ID:        sonComment.ID,
				Content:   sonComment.Content,
				Author:    sonComment.Author,
				LikeCount: sonComment.LikeCount,
				Liked:     sonComment.Liked,
				Parent:    sonComment.Parent,
				CreatedAt: sonComment.CreatedAt,
				RootID:    topComment.ID,
				ParentID:  sonComment.ParentID,
			})
		}
	}
	return commentListResponse, nil
}
