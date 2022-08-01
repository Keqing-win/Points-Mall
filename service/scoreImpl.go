package service

import (
	"Points_Mall/config"
	"Points_Mall/dao"
	"Points_Mall/middlerware/redis"
	"log"
	"strconv"
	"time"
)

type ScoreServiceImpl struct {
	TaskService
}

// Score 封装后返回
type Score struct {
	Title   string //任务的标题
	Content string //任务的内容
	Score   string //任务的分数
	Only    int32  //0表示一次，1表示每日任务
}

func (sc *ScoreServiceImpl) GetScore(userId int64) (int64, error) {
	//先查缓存
	var score int64
	score = 0
	userIdStr := strconv.Itoa(int(userId))
	if n, err := redis.RdbUTid.Exists(redis.Ctx, userIdStr).Result(); n > 0 {
		if err != nil {
			log.Println("redis.RdbUTid.Exists err:", err.Error())
			return 0, err
		} else {
			_, err1 := redis.RdbUTid.SCard(redis.Ctx, userIdStr).Result()
			if err1 != nil {
				log.Println("redis.RdbUTid.SCard err:", err1.Error())
				return 0, err1
			}
			taskIds, err2 := redis.RdbUTid.SMembers(redis.Ctx, userIdStr).Result()
			if err2 != nil {
				log.Println("redis.RdbUTid.SMembers err:", err2.Error())
				return 0, err2
			}
			for _, taskId := range taskIds {
				scoreStr, err3 := redis.RdbTScore.Get(redis.Ctx, taskId).Result()
				if err3 != nil {
					log.Println("redis.RdbTScore.Get err3:", err3.Error())
					return 0, err3
				}
				scoreI, _ := strconv.ParseInt(scoreStr, 10, 64)
				score += scoreI
			}
			return score + 1, nil
		}
	} else {
		//先加入-1 防止删最后一个数据的时候数据库还没更新完出现脏读，或者数据库操作失败造成的脏读
		if _, err := redis.RdbUTid.SAdd(redis.Ctx, userIdStr, config.DefaultRedisValue).Result(); err != nil {
			log.Println("redis.RdbUTid.SAdd err", err.Error())
			redis.RdbUTid.Del(redis.Ctx, userIdStr)
			return 0, err
		}
		//设置过期时间
		_, err := redis.RdbUTid.Expire(redis.Ctx, userIdStr, time.Duration(config.OneMonth)*time.Second).Result()
		if err != nil {
			log.Println("redis.RdbUTid.Expire err", err.Error())
			redis.RdbUTid.Del(redis.Ctx, userIdStr)
			return 0, err
		}
		//没缓存就去数据查
		scoredDao := dao.NewScoreDao()
		scoreTables, err := scoredDao.GetScoreTables(userId)
		if err != nil {
			log.Println("get score tables err", err.Error())
			return 0, err
		}
		log.Println("get score tables suc")
		for _, scoreTable := range scoreTables {
			scoreI, _ := strconv.ParseInt(scoreTable.Score, 10, 64)
			score += scoreI
			//更新数据库
			taskIdStr := strconv.Itoa(int(scoreTable.TaskId))
			if _, err := redis.RdbTScore.Set(redis.Ctx, taskIdStr, scoreTable.Score, time.Duration(config.OneMinute)*time.Second*10).
				Result(); err != nil {
				log.Println("redis.RdbTScore.Set err", err.Error())
				return 0, err
			}
			if _, err := redis.RdbUTid.SAdd(redis.Ctx, userIdStr, taskIdStr, time.Duration(config.OneMonth)*time.Second).Result(); err != nil {
				log.Println("redis.RdbUTid.SAdd err", err.Error())
				return 0, err
			}
		}
		return score, nil
	}
}

func (sc *ScoreServiceImpl) GetScoreRecord(userId int64) ([]Score, bool) {
	var records []Score
	userIdStr := strconv.Itoa(int(userId))
	scoreDao := dao.ScoreDao{}
	if n, err := redis.RdbUTid.Exists(redis.Ctx, userIdStr).Result(); n > 0 {
		if err != nil {
			log.Println("redis.RdbUTid.Exists err:", err.Error())
			return records, false
		} else {
			_, err1 := redis.RdbUTid.SCard(redis.Ctx, userIdStr).Result()
			if err1 != nil {
				log.Println("redis.RdbUTid.SCard err:", err1.Error())
				return records, false
			}
			taskIds, err2 := redis.RdbUTid.SMembers(redis.Ctx, userIdStr).Result()
			if err2 != nil {
				log.Println("redis.RdbUTid.SMembers err:", err2.Error())
				return records, false
			}
			//根据缓存中的taskIds去查信息
			for _, strTaskId := range taskIds {
				taskId, _ := strconv.ParseInt(strTaskId, 10, 64)
				taskTable := sc.GetTaskById(taskId)
				var record Score
				record.Score = taskTable.Score
				record.Title = taskTable.Title
				record.Content = taskTable.Content
				record.Only = taskTable.Only
				records = append(records, record)
			}

			return records, true

		}
	} else {
		//不在缓存中，查数据库,更新缓存
		//先加入-1 防止删最后一个数据的时候数据库还没更新完出现脏读，或者数据库操作失败造成的脏读
		if _, err = redis.RdbUTid.SAdd(redis.Ctx, userIdStr, config.DefaultRedisValue).Result(); err != nil {
			log.Println("redis.RdbUTid.SAdd err", err.Error())
			redis.RdbUTid.Del(redis.Ctx, userIdStr)
			return records, false
		}
		//设置过期时间
		_, err = redis.RdbUTid.Expire(redis.Ctx, userIdStr, time.Duration(config.OneMonth)*time.Second).Result()
		if err != nil {
			log.Println("redis.RdbUTid.Expire err", err.Error())
			redis.RdbUTid.Del(redis.Ctx, userIdStr)
			return records, false
		}
		scoreTables, err := scoreDao.GetScoreTables(userId)
		if err != nil {
			log.Println("获取分数表出错 err:", err.Error())
			return records, false
		}
		for _, score := range scoreTables {
			if _, err = redis.RdbUTid.SAdd(redis.Ctx, userIdStr, score.TaskId).Result(); err != nil {
				log.Println(" redis.RdbUTid.SAdd err", err.Error())
				return records, false
			}

			taskTable := sc.GetTaskById(score.TaskId)
			var record Score
			record.Score = taskTable.Score
			record.Title = taskTable.Title
			record.Content = taskTable.Content
			record.Only = taskTable.Only
			records = append(records, record)

		}
		return records, true

	}

}
func (sc *ScoreServiceImpl) AddScore(userId, taskId int64) bool {
	score := sc.GetTaskScore(taskId)

	if score == "" {
		return false
	}
	scoredDao := dao.NewScoreDao()
	if err := scoredDao.InsertScoreTable(userId, taskId, score); err != nil {
		log.Println("add score err", err.Error())
		return false
	}
	log.Println("add score suc")
	return true

}
