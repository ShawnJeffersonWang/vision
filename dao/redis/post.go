package redis

import (
	"strconv"
	"time"

	"github.com/go-redis/redis"

	"agricultural_vision/constants"
	"agricultural_vision/models/request"
)

// 新建帖子
func CreatePost(postID int64, communityID int64) error {
	postIDStr := strconv.FormatInt(postID, 10)
	communityIDStr := strconv.FormatInt(communityID, 10)

	//开启事务
	pipeline := client.TxPipeline()

	//在redis中更新帖子创建时间
	pipeline.ZAdd(getRedisKey(KeyPostTimeZSet), redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: postIDStr,
	})

	//在redis中更新帖子分数
	pipeline.ZAdd(getRedisKey(KeyPostScoreZSet), redis.Z{
		Score:  float64(time.Now().Unix()), // 默认分数不是0，而是当前的时间戳（这样总分数可以结合投票数和时间）
		Member: postIDStr,
	})

	//在redis中更新帖子和社区关系
	pipeline.SAdd(getRedisKey(KeyCommunitySetPF)+communityIDStr, postIDStr)

	//初始化帖子评论数
	pipeline.ZAdd(getRedisKey(KeyPostCommentNumZSet), redis.Z{
		Score:  0,
		Member: postIDStr,
	})

	_, err := pipeline.Exec()
	return err
}

// 删除帖子
func DeletePost(postID int64, communityID int64) error {
	postIDStr := strconv.FormatInt(postID, 10)
	communityIDStr := strconv.FormatInt(communityID, 10)

	// 开启事务
	pipeline := client.TxPipeline()

	// 查询出此帖子所有一级评论的id
	commentIDs, _ := client.ZRange(getRedisKey(KeyCommentTimeZSetPF+postIDStr), 0, -1).Result()

	// 从时间排序集合删除
	pipeline.ZRem(getRedisKey(KeyPostTimeZSet), postIDStr) // 从 zSet 中移除指定成员

	// 从热度排序集合删除
	pipeline.ZRem(getRedisKey(KeyPostScoreZSet), postIDStr)

	// 删除帖子点赞记录
	pipeline.Del(getRedisKey(KeyPostVotedZSetPF + postIDStr)) // 删除整个 key

	// 从社区帖子集合删除
	pipeline.SRem(getRedisKey(KeyCommunitySetPF+communityIDStr), postIDStr) // 从 set 中移除指定成员

	// 删除帖子评论数记录
	pipeline.ZRem(getRedisKey(KeyPostCommentNumZSet), postIDStr)

	// 删除该帖子下的所有评论时间记录
	pipeline.Del(getRedisKey(KeyCommentTimeZSetPF + postIDStr))

	// 删除该帖子下的所有评论投票记录
	pipeline.Del(getRedisKey(KeyCommentScoreZSetPF + postIDStr))

	// 处理该帖子下所有评论的删除
	if len(commentIDs) > 0 {
		// 删除该帖子下所有评论的点赞记录，拼凑出需要删除的key的列表
		keysToDelete := make([]string, len(commentIDs))
		for i, commentID := range commentIDs {
			keysToDelete[i] = getRedisKey(KeyCommentVotedZSetPF + commentID)
		}
		pipeline.Del(keysToDelete...) // 一次删除多个 key，提高性能

		// 删除所有评论的子评论数记录，根据commentId列表删除
		interfaceIDs := make([]interface{}, len(commentIDs))
		for i, id := range commentIDs {
			interfaceIDs[i] = id
		}
		pipeline.ZRem(getRedisKey(KeyCommentNumZSet), interfaceIDs...)
	}

	// 执行事务
	_, err := pipeline.Exec()
	return err
}

// 根据键名和索引，分页id列表，返回列表和总数（工具函数）
func getIDsFormKey(key string, page, size int64) ([]string, int64, error) {
	// 查询总数
	totalCount, err := client.ZCard(key).Result()
	if err != nil {
		return nil, 0, err
	}

	// 进行分页查询
	start := (page - 1) * size
	end := start + size - 1

	// ZRevRange 按分数从大到小查询指定数量的元素
	result, err := client.ZRevRange(key, start, end).Result()
	return result, totalCount, err
}

// 根据排序方式和索引，查询id列表
func GetPostIDsInOrder(p *request.ListRequest) ([]string, int64, error) {
	// 根据用户请求中携带的 order 参数（排序方式）确定要查询的 redis key
	key := getRedisKey(KeyPostTimeZSet)
	if p.Order == constants.OrderScore {
		key = getRedisKey(KeyPostScoreZSet)
	}

	return getIDsFormKey(key, p.Page, p.Size)
}

// 根据ids列表查询每篇帖子的投赞成票的数据
func GetPostVoteDataByIDs(ids []string) (data []int64, err error) {
	// 使用 pipeline 批量执行 Redis 命令
	pipeline := client.Pipeline()
	var cmds []redis.Cmder

	// 将所有 ZCount 命令添加到 pipeline 中
	for _, id := range ids {
		key := getRedisKey(KeyPostVotedZSetPF + id)
		// 统计该帖子的赞成票数
		cmds = append(cmds, pipeline.ZCount(key, "1", "1"))
	}

	// 执行所有命令
	_, err = pipeline.Exec()
	if err != nil {
		return nil, err
	}

	// 处理结果
	for _, cmd := range cmds {
		// 类型断言为 redis.IntCmd
		voteCmd := cmd.(*redis.IntCmd)
		data = append(data, voteCmd.Val()) // 获取票数并保存
	}

	return data, nil
}

// 根据社区id查询该社区下的帖子id列表
func GetCommunityPostIDsInOrder(p *request.ListRequest, communityID int64) (ids []string, total int64, err error) {
	//根据指定的排序方式，确定要操作的redis中的key
	//orderKey指定排序方式的键名，按时间排序则是KeyPostTimeZSet，按分数排序则是KeyPostScoreZSet
	orderKey := getRedisKey(KeyPostTimeZSet)
	if p.Order == constants.OrderScore {
		orderKey = getRedisKey(KeyPostScoreZSet)
	}

	//从KeyCommunitySetPF中查询该社区下的帖子id列表，根据id列表去KeyPostTimeZSet或KeyPostScoreZSet中去查询时间或分数
	//也就是查询交集，将查询到的内容（帖子postID和对应的时间或分数）保存到新的自定义的key中

	//社区的key
	communityKey := getRedisKey(KeyCommunitySetPF + strconv.Itoa(int(communityID)))

	//自定义新key，用来存储两表交集的，值为postID和对应的时间或分数，表示此社区分类下的帖子和时间/分数
	key := orderKey + strconv.Itoa(int(communityID))

	// 使用 pipeline 批量执行 Redis 命令
	pipeline := client.Pipeline()

	//通过 ZInterStore 对有序集合communityKey和orderKey进行交集运算，结果存储到key中
	pipeline.ZInterStore(key, redis.ZStore{
		Aggregate: "MAX", //表示交集的分数取较大的值，如果将一个普通的set（无序集合）与一个zSet（有序集合）一起参与ZInterStore操作，Redis会自动将set视为一个所有成员分数为1的特殊zSet
	}, communityKey, orderKey)

	pipeline.Expire(key, 60*time.Second) // 设置超时时间

	_, err = pipeline.Exec()
	if err != nil {
		return
	}

	//查询指定索引范围的id列表
	return getIDsFormKey(key, p.Page, p.Size)
}

// 根据用户id查询用户点赞过的帖子id列表
func GetUserLikeIDsInOrder(userID int64, listRequest *request.ListRequest) ([]string, int64, error) {
	// 获取 user_liked:posts:{userID} 中的所有帖子ID
	likedPostsKey := getRedisKey(KeyUserLikedPostsSetPF) + strconv.Itoa(int(userID))

	// 获取有效的帖子ID
	validPostIDsKey := getRedisKey(KeyPostTimeZSet)

	// 获取 zSet 中所有的有效帖子ID
	validPostIDs, err := client.ZRange(validPostIDsKey, 0, -1).Result()
	if err != nil {
		return nil, 0, err
	}

	// 存储交集的帖子ID
	var intersection []string

	// 循环检查有效帖子ID是否在用户点赞的帖子集合中
	for _, postID := range validPostIDs {
		// 检查用户是否点赞了该帖子
		isMember, err := client.SIsMember(likedPostsKey, postID).Result()
		if err != nil {
			return nil, 0, err
		}
		if isMember {
			// 如果点赞过，添加到交集结果中
			intersection = append(intersection, postID)
		}
	}

	// 获取交集结果的总数
	totalCount := int64(len(intersection))

	// 如果分页请求存在，按分页要求返回数据
	start := (listRequest.Page - 1) * listRequest.Size
	end := start + listRequest.Size

	// 确保分页范围不超过交集结果的长度
	if start >= totalCount {
		return nil, totalCount, nil // 页码超出范围，返回空切片
	}
	if end > totalCount {
		end = totalCount
	}

	// 返回分页后的结果
	paginatedPosts := intersection[start:end]

	return paginatedPosts, totalCount, nil
}
