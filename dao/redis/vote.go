package redis

import (
	"time"

	"github.com/go-redis/redis"

	"agricultural_vision/constants"
)

/*
投票的情况：
direction=1时：
    1、之前没有投过票，现在投赞成票 --> 更新分数和投票记录
    2、之前投反对票，现在改投赞成票 --> 更新分数和投票记录
direction=-1时：
    1、之前没有投过票，现在投反对票 --> 更新分数和投票记录
    2、之前投赞成票，现在改投反对票 --> 更新分数和投票记录
direction=0时：
    1、之前投过赞成票，现在要取消投票 --> 更新分数和投票记录
    2、之前投反对票，现在要取消投票 --> 更新分数和投票记录
无论哪种情况，用本次投票数减去以前投票数即为此时的实际投票数

投票的限制：
	每个帖子自发表之日起一个星期之内允许用户投票，超过一个星期就不允许投票
    1、到期之后将redis中保存的赞成票及反对票存储到mysql表中
	2、到期之后删除 KeyPostVotedZSetPF
*/

// 为帖子投票
func VoteForPost(userID, postID string, direction float64) error {
	// 投票的权重，如果帖子发布时间超过一周，则权重为0.5（减半）
	weight := 1.0

	//判断投票限制
	//利用redis获取帖子发布时间
	//ZScore函数的两个参数：键名和成员名，获取该成员的分数score
	//Val将结果转换为float64类型
	postTime := client.ZScore(getRedisKey(KeyPostTimeZSet), postID).Val()

	//如果帖子发布时间超过一周，则减小权重
	if float64(time.Now().Unix())-postTime > constants.OneWeekInSeconds {
		weight = 0.8
	}

	//更新帖子分数
	//查询当前用户(userID)给当前帖子(postID)的投票记录
	ov := client.ZScore(getRedisKey(KeyPostVotedZSetPF+postID), userID).Val() // 上次投票类型：1 or 0 or -1
	diff := direction - ov                                                    //计算两次投票类型的差值

	//开启事务
	pipeline := client.TxPipeline()

	//给指定的键和成员名增加分数
	pipeline.ZIncrBy(getRedisKey(KeyPostScoreZSet), diff*constants.ScorePerVote*weight, postID)

	//更新用户为该帖子投票的数据
	pipeline.ZAdd(getRedisKey(KeyPostVotedZSetPF+postID), redis.Z{
		Score:  direction, //投票类型
		Member: userID,
	})

	//更新用户的点赞记录
	if direction == 1 {
		pipeline.SAdd(getRedisKey(KeyUserLikedPostsSetPF)+userID, postID)
	} else {
		pipeline.SRem(getRedisKey(KeyUserLikedPostsSetPF)+userID, postID)
	}

	//执行事务
	_, err := pipeline.Exec()
	return err
}

// 为评论投票
func VoteForComment(userID, commentID, postID string, direction float64, parentID *int64) error {
	// 投票的权重，如果一级评论发布时间超过一周，则权重为0.5（减半）
	weight := 1.0

	//判断投票限制
	//利用redis获取帖子发布时间
	//ZScore函数的两个参数：键名和成员名，获取该成员的分数score
	//Val将结果转换为float64类型
	commentTime := client.ZScore(getRedisKey(KeyCommentTimeZSetPF), commentID).Val()

	//如果帖子发布时间超过一周，则权重减半
	if float64(time.Now().Unix())-commentTime > constants.OneWeekInSeconds {
		weight = 0.5
	}

	//更新一级评论分数
	//查询当前用户(userID)给当前一级评论的投票记录
	ov := client.ZScore(getRedisKey(KeyCommentVotedZSetPF+commentID), userID).Val() // 上次投票类型：1 or 0 or -1
	diff := direction - ov                                                          //计算两次投票类型的差值

	//开启事务
	pipeline := client.TxPipeline()

	//给指定的键和成员名增加分数
	if parentID == nil { // 如果是一级评论才更新
		pipeline.ZIncrBy(getRedisKey(KeyCommentScoreZSetPF+postID), diff*constants.ScorePerVote*weight, commentID)
	}

	//更新用户为该评论投票的数据
	pipeline.ZAdd(getRedisKey(KeyCommentVotedZSetPF+commentID), redis.Z{
		Score:  direction, //投票类型
		Member: userID,
	})

	//更新用户点过赞的评论集合
	pipeline.SAdd(getRedisKey(KeyUserLikedCommentsSetPF)+userID, commentID)

	//执行事务
	_, err := pipeline.Exec()
	return err
}

// 查询用户是否点赞过该帖子
func IsUserLikedPost(userID string, postID string) (bool, error) {
	return client.SIsMember(getRedisKey(KeyUserLikedPostsSetPF+userID), postID).Result()
}

// 查询用户是否点赞过该评论
func IsUserLikedComment(userID string, commentID string) (bool, error) {
	return client.SIsMember(getRedisKey(KeyUserLikedCommentsSetPF+userID), commentID).Result()
}
