package redis

import (
	"errors"
	"strconv"
	"time"

	"github.com/go-redis/redis"

	"agricultural_vision/constants"
	"agricultural_vision/models/request"
)

// 创建顶级评论
func CreateTopComment(commentID, postID int64) error {
	postIDStr := strconv.FormatInt(postID, 10)
	commentIDStr := strconv.FormatInt(commentID, 10)

	//开启事务
	pipeline := client.TxPipeline()

	//在redis中更新评论时间
	pipeline.ZAdd(getRedisKey(KeyCommentTimeZSetPF)+postIDStr, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: commentIDStr,
	})

	//在redis中更新评论分数
	pipeline.ZAdd(getRedisKey(KeyCommentScoreZSetPF)+postIDStr, redis.Z{
		Score:  float64(time.Now().Unix()), // 默认分数不是0，而是当前的时间戳（这样总分数可以结合投票数和时间）
		Member: commentIDStr,
	})

	//在redis中更新评论数（累计+1）
	pipeline.ZIncrBy(getRedisKey(KeyPostCommentNumZSet), 1, postIDStr)

	//初始化二级评论数为0
	pipeline.ZAdd(getRedisKey(KeyCommentNumZSet), redis.Z{
		Score:  0,
		Member: commentIDStr,
	})

	_, err := pipeline.Exec()
	return err
}

// 创建子评论
func CreateSonComment(rootID, postID int64) error {
	rootIDStr := strconv.FormatInt(rootID, 10)
	postIDStr := strconv.FormatInt(postID, 10)

	// 在redis中更新评论数（累计+1）
	client.ZIncrBy(getRedisKey(KeyPostCommentNumZSet), 1, postIDStr)

	// 在redis中更新子评论数（累计+1）
	err := client.ZIncrBy(getRedisKey(KeyCommentNumZSet), 1, rootIDStr).Err()

	return err
}

// 删除评论
func DeleteComment(commentID, postID int64, rootID *int64) error {
	pipeline := client.TxPipeline()

	postIDStr := strconv.FormatInt(postID, 10)
	commentIDStr := strconv.FormatInt(commentID, 10)

	// 1. 从帖子评论时间集合中删除该评论（一级评论才有效）
	pipeline.ZRem(getRedisKey(KeyCommentTimeZSetPF+postIDStr), commentIDStr)

	// 2. 从评论分数集合中删除该评论（一级评论才有效）
	pipeline.ZRem(getRedisKey(KeyCommentScoreZSetPF+postIDStr), commentIDStr)

	// 3. 删除该评论的点赞记录
	pipeline.Del(getRedisKey(KeyCommentVotedZSetPF + commentIDStr))

	// 4. 从评论数集合中删除该评论（一级评论才有效）
	pipeline.ZRem(getRedisKey(KeyCommentNumZSet), commentIDStr)

	// 5. 如果是顶级评论，减少帖子总评论数（减少：顶级评论的子评论数+1）
	if rootID == nil {
		// 查找该评论的子评论数
		sonCommentNum, err := GetSonCommentNumByIDs([]string{commentIDStr})
		if err != nil {
			return err
		}
		if len(sonCommentNum) == 0 {
			return constants.ErrorNoResult
		}

		// 减少帖子总评论数
		pipeline.ZIncrBy(getRedisKey(KeyPostCommentNumZSet), -1-sonCommentNum[0], postIDStr)
	} else {
		// 6. 如果是子评论，减少帖子总评论数并且减少父评论的子评论数
		rootIDStr := strconv.FormatInt(*rootID, 10)
		pipeline.ZIncrBy(getRedisKey(KeyPostCommentNumZSet), -1, postIDStr)
		pipeline.ZIncrBy(getRedisKey(KeyCommentNumZSet), -1, rootIDStr)
	}

	// 执行 Redis 事务
	_, err := pipeline.Exec()
	return err
}

// 根据排序方式和索引范围，查询顶级评论id列表
func GetTopCommentIDsInOrder(p *request.ListRequest, postID int64) ([]string, int64, error) {
	//从redis中获取id
	//1.根据用户请求中携带的order参数（排序方式）确定要查询的redis key
	key := getRedisKey(KeyCommentTimeZSetPF + strconv.Itoa(int(postID)))
	if p.Order == constants.OrderScore {
		key = getRedisKey(KeyCommentScoreZSetPF) + strconv.Itoa(int(postID))
	}

	return getIDsFormKey(key, p.Page, p.Size)
}

// 根据ids列表查询每条评论的赞成票数据
func GetCommentVoteDataByIDs(ids []string) (data []int64, err error) {
	// 使用 pipeline 批量执行 Redis 命令
	pipeline := client.Pipeline()
	var cmds []redis.Cmder

	// 将所有 ZCount 命令添加到 pipeline 中
	for _, id := range ids {
		key := getRedisKey(KeyCommentVotedZSetPF + id)
		// 统计该帖子的赞成票数（1代表赞成票）
		cmds = append(cmds, pipeline.ZCount(key, "1", "1"))
	}

	// 执行所有命令
	_, err = pipeline.Exec()
	if err != nil {
		return nil, err
	}

	// 处理返回的结果
	for _, cmd := range cmds {
		// 类型断言为 redis.IntCmd
		voteCmd := cmd.(*redis.IntCmd)
		data = append(data, voteCmd.Val()) // 获取赞成票数
	}

	return data, nil
}

// 根据帖子id列表查询帖子的评论数
func GetCommentNumByIDs(ids []string) ([]float64, error) {
	// 使用 pipeline 批量执行 Redis 命令
	pipeline := client.Pipeline()

	// 用于存储命令
	cmds := make([]*redis.FloatCmd, len(ids))

	// 将所有 ZScore 命令添加到 pipeline
	for i, id := range ids {
		key := getRedisKey(KeyPostCommentNumZSet)
		cmds[i] = pipeline.ZScore(key, id)
	}

	// 执行 pipeline
	_, err := pipeline.Exec()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	// 处理结果
	nums := make([]float64, len(ids))
	for i, cmd := range cmds {
		if err := cmd.Err(); errors.Is(err, redis.Nil) {
			nums[i] = 0 // 特殊值表示未找到，避免误判 0
		} else {
			nums[i] = cmd.Val()
		}
	}

	return nums, nil
}

// 根据顶级评论id列表查询顶级评论的子评论数
func GetSonCommentNumByIDs(ids []string) (nums []float64, err error) {
	// 使用 pipeline 批量执行 Redis 命令
	pipeline := client.Pipeline()

	// 创建一个切片用于保存所有 ZScore 命令
	var cmds []redis.Cmder

	// 将所有 ZScore 命令添加到 pipeline 中
	for _, id := range ids {
		key := getRedisKey(KeyCommentNumZSet)
		cmds = append(cmds, pipeline.ZScore(key, id))
	}

	// 执行所有命令
	_, err = pipeline.Exec()
	if err != nil {
		return nil, err
	}

	// 处理结果
	for _, cmd := range cmds {
		scoreCmd := cmd.(*redis.FloatCmd)   // 类型断言为 FloatCmd
		nums = append(nums, scoreCmd.Val()) // 获取评论数
	}

	return nums, nil
}
