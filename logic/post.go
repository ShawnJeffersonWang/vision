package logic

import (
	"agricultural_vision/dao"
	"agricultural_vision/models/proto"
	"agricultural_vision/pkg/snowflake"
	"agricultural_vision/service/kafka"
	"errors"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strconv"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"agricultural_vision/constants"
	"agricultural_vision/dao/redis"
	"agricultural_vision/models/entity"
	"agricultural_vision/models/request"
	"agricultural_vision/models/response"
)

// 创建帖子
func CreatePost(createPostRequest *request.CreatePostRequest, authorID int64) (postResponse *response.PostResponse, err error) {
	post := &entity.Post{
		Content:     createPostRequest.Content,
		Image:       createPostRequest.Image,
		AuthorID:    authorID,
		CommunityID: createPostRequest.CommunityID,
	}

	//保存到数据库
	err = dao.CreatePost(post)
	if err != nil {
		return
	}

	//查询作者简略信息
	userBriefInfo, err := dao.GetUserBriefInfo(post.AuthorID)
	if err != nil { // 遇到错误不返回，继续执行后续逻辑
		zap.L().Error("查询作者信息失败", zap.Error(err))
	}

	//查询社区详情
	community, err := dao.GetCommunityById(post.CommunityID)
	if err != nil { // 遇到错误不返回，继续执行后续逻辑
		zap.L().Error("查询社区详情失败", zap.Error(err))
	}

	//封装查询到的信息
	postResponse = &response.PostResponse{
		ID:        post.ID,
		Content:   post.Content,
		Image:     post.Image,
		Author:    *userBriefInfo,
		CreatedAt: post.CreatedAt.Format("2006-01-02 15:04:05"),
		Community: response.CommunityBriefResponse{ID: community.ID, CommunityName: community.CommunityName},
	}

	//保存到redis
	err = redis.CreatePost(post.ID, post.CommunityID)
	return
}

// 创建帖子（改为只发送 Kafka 消息）
func CreatePostAsyncUseJSON(createPostRequest *request.CreatePostRequest, authorID int64) (*response.PostResponse, error) {
	// 1. 生成帖子 ID（提前生成，因为还未写入数据库）
	postID := snowflake.GenID() // 需要使用分布式 ID 生成器

	// 2. 封装 Kafka 消息
	message := kafka.PostCreationMessage{
		MessageID:   uuid.New().String(), // 需要导入 "github.com/google/uuid"
		UserID:      authorID,
		Content:     createPostRequest.Content,
		Image:       createPostRequest.Image,
		CommunityID: createPostRequest.CommunityID,
		CreatedAt:   time.Now().UTC(),
		PostID:      postID, // 需要在 PostCreationMessage 中添加 PostID 字段
	}

	// 3. 发送 Kafka 消息（使用全局生产者）
	if err := kafka.SendPostCreationMessageUseJson(message); err != nil {
		zap.L().Error("CreatePostAsync.SendPostCreationMessage: 发送 kafka 失败", zap.Error(err))
		return nil, err
	}
	zap.L().Info("CreatePostAsync.SendPostCreationMessage: 发送 kafka 成功", zap.Any("message", message))

	// 4. 立即返回（不等待实际创建完成）
	postResponse := &response.PostResponse{
		ID:        postID,
		Content:   createPostRequest.Content,
		Image:     createPostRequest.Image,
		Author:    response.UserBriefResponse{ID: authorID},
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		Community: response.CommunityBriefResponse{ID: createPostRequest.CommunityID},
	}

	return postResponse, nil
}

// logic/post.go
func CreatePostAsync(createPostRequest *request.CreatePostRequest, authorID int64) (*response.PostResponse, error) {
	// 生成帖子 ID
	postID := snowflake.GenID()

	// 封装 Protobuf 消息
	message := &proto.PostCreationMessage{
		MessageId:   uuid.New().String(),
		UserId:      authorID,
		Content:     createPostRequest.Content,
		Image:       createPostRequest.Image,
		CommunityId: createPostRequest.CommunityID,
		CreatedAt: &timestamppb.Timestamp{
			Seconds: time.Now().UTC().Unix(),
			Nanos:   int32(time.Now().UTC().Nanosecond()),
		},
		PostId: postID,
	}

	// 发送 Kafka 消息
	if err := kafka.SendPostCreationMessage(message); err != nil {
		return nil, err
	}

	// 立即返回
	postResponse := &response.PostResponse{
		ID:        postID,
		Content:   createPostRequest.Content,
		Image:     createPostRequest.Image,
		Author:    response.UserBriefResponse{ID: authorID},
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		Community: response.CommunityBriefResponse{ID: createPostRequest.CommunityID},
	}

	return postResponse, nil
}

// 删除帖子
func DeletePost(postID int64, userID int64) error {
	// 从mysql查询帖子
	post, err := dao.GetPostById(postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { // 如果查询不到帖子
			return constants.ErrorNoPost
		}
		return err
	}
	// 校验userID
	if post.AuthorID != userID {
		return constants.ErrorNoPermission
	}
	communityID := post.CommunityID

	// 删除mysql中的帖子
	if err := dao.DeletePost(postID); err != nil {
		return err
	}

	// 删除redis中的帖子
	if err := redis.DeletePost(postID, communityID); err != nil {
		return err
	}

	return nil
}

// 根据id列表查询帖子列表，并封装响应数据
func GetPostListByIDs(ids []string, userID int64) (postResponses []*response.PostResponse, err error) {
	//调用此函数前，已经对ids进行判断，不为空

	//根据id列表去数据库查询帖子详细信息
	posts, err := dao.GetPostListByIDs(ids)
	if err != nil {
		return
	}

	//查询所有帖子的赞成票数——切片
	/*创建帖子时，此key的默认值为0，所以voteData一定不为空，不会出现空指针异常*/
	voteData, err := redis.GetPostVoteDataByIDs(ids)
	if err != nil {
		return
	}

	// 查询所有帖子的评论数——切片
	/*创建帖子时，此key的默认值为0，所以voteData一定不为空，不会出现空指针异常*/
	commentNum, err := redis.GetCommentNumByIDs(ids)
	if err != nil {
		return
	}

	//将帖子作者及分区信息查询出来填充到帖子中
	for idx, post := range posts {
		//查询作者简略信息
		userBriefInfo, err := dao.GetUserBriefInfo(post.AuthorID)
		if err != nil { // 遇到错误不返回，继续执行后续逻辑
			zap.L().Error("查询作者信息失败", zap.Error(err))
			continue
		}

		//查询社区详情
		community, err := dao.GetCommunityById(post.CommunityID)
		if err != nil { // 遇到错误不返回，继续执行后续逻辑
			zap.L().Error("查询社区详情失败", zap.Error(err))
			continue
		}

		// 查询当前用户是否点赞了此帖子
		liked := false
		if userID != 0 { // 如果用户已登录，则查询用户是否点赞了此帖子
			liked, err = redis.IsUserLikedPost(strconv.Itoa(int(userID)), strconv.Itoa(int(post.ID)))
			if err != nil { // 遇到错误不返回，继续执行后续逻辑
				zap.L().Error("查询用户是否点赞失败", zap.Error(err))
				continue
			}
		}

		//封装查询到的信息
		postResponse := &response.PostResponse{
			ID:           post.ID,
			Content:      post.Content,
			Image:        post.Image,
			Author:       *userBriefInfo,
			LikeCount:    voteData[idx],
			Liked:        liked,
			CommentCount: int64(commentNum[idx]),
			CreatedAt:    post.CreatedAt.Format("2006-01-02 15:04:05"),
			Community:    response.CommunityBriefResponse{ID: community.ID, CommunityName: community.CommunityName},
		}

		postResponses = append(postResponses, postResponse)
	}
	return
}

// 查询帖子列表，并按照指定方式排序
func GetPostList(p *request.ListRequest, userID int64) (postListResponse *response.PostListResponse, err error) {
	postListResponse = &response.PostListResponse{
		Posts: []*response.PostResponse{},
	}

	// 从redis中，根据指定的排序方式和查询数量，查询符合条件的id列表
	ids, total, err := redis.GetPostIDsInOrder(p)
	if err != nil {
		return
	}
	postListResponse.Total = total
	if len(ids) == 0 {
		return
	}

	// 根据id列表查询帖子列表，并封装响应数据
	postListResponse.Posts, err = GetPostListByIDs(ids, userID)
	return
}

// 查询该社区下的帖子列表，并按指定方式排序
func GetCommunityPostList(listRequest *request.ListRequest, communityID int64, userID int64) (postListResponse *response.PostListResponse, err error) {
	postListResponse = &response.PostListResponse{
		Posts: []*response.PostResponse{},
	}

	//从redis中，根据指定的排序方式和查询数量，查询符合条件的分页后的id列表
	ids, total, err := redis.GetCommunityPostIDsInOrder(listRequest, communityID)
	if err != nil {
		return
	}
	postListResponse.Total = total
	if len(ids) == 0 {
		return
	}

	// 根据id列表查询帖子列表，并封装响应数据
	postListResponse.Posts, err = GetPostListByIDs(ids, userID)
	return
}

// 获取用户发布的帖子列表
func GetUserPostList(userID int64, listRequest *request.ListRequest) (postListResponse *response.PostListResponse, err error) {
	postListResponse = &response.PostListResponse{
		Posts: []*response.PostResponse{},
	}

	// 查询该用户的所有帖子
	posts, total, err := dao.GetPostListByUserID(userID, listRequest.Page, listRequest.Size)
	if err != nil {
		return
	}
	postListResponse.Total = total
	if len(posts) == 0 {
		return
	}

	// 拼凑帖子id列表
	ids := make([]string, len(posts))
	for idx, post := range posts {
		ids[idx] = strconv.Itoa(int(post.ID))
	}

	// 根据id列表查询帖子列表，并封装响应数据
	postListResponse.Posts, err = GetPostListByIDs(ids, userID)
	return
}

// 获取用户点赞的帖子列表
func GetUserLikedPostList(userID int64, listRequest *request.ListRequest) (postListResponse *response.PostListResponse, err error) {
	postListResponse = &response.PostListResponse{
		Posts: []*response.PostResponse{},
	}

	// 从redis中查询用户点赞的帖子id列表
	ids, total, err := redis.GetUserLikeIDsInOrder(userID, listRequest)
	if err != nil {
		return
	}
	postListResponse.Total = total
	if len(ids) == 0 {
		return
	}

	// 根据id列表查询帖子列表，并封装响应数据
	postListResponse.Posts, err = GetPostListByIDs(ids, userID)
	return
}
