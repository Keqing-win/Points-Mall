package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
)

var Ctx = context.Background()
var RdbGOid *redis.Client         //key:GOODId Value:OrderId
var RdbUTid *redis.Client         //key:UserId string:TaskId
var RdbTScore *redis.Client       //key:TaskId string:score
var RdbTGid *redis.Client         //key:TaskId Value:UserId   //唯一任务的热度 //没使用到
var RdbGNumber *redis.Client      //key:GoodID Value:number
var RdbUSIid *redis.Client        //key:UserId  value；SignInId  //方便维护用户连续签到的天数
var RdbUserSignDays *redis.Client //key:rank  score:days member:UserId //方便维护连续签到排行榜
var Z redis.Z

func InitRedis() {
	RdbGOid = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0, // 商品信息存入 DB0.
	})
	RdbUTid = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       1, // 用户完成的任务表 DB1.
	})
	RdbTGid = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       2, //任务被用户完成列表 DB2.
	})
	RdbGNumber = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       3, //商品和数目的列表
	})
	RdbUSIid = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       4, //用户的连续签到天数 DB4
	})
	RdbTScore = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       5, //任务 分数 DB5
	})
	RdbUserSignDays = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       6, //用户的连续签到排行榜 DB6
	})
}
