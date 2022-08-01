package service

import (
	"Points_Mall/config"
	"Points_Mall/dao"
	"Points_Mall/middlerware/redis"
	"log"
	"strconv"
	"time"
)

type SignInImpl struct {
	UserService
}

func (si *SignInImpl) GetRank() ([]User, bool) {
	var users []User
	//判断缓存是否失效
	if n, err := redis.RdbUserSignDays.Exists(redis.Ctx, "CheckInsRank").Result(); n > 0 {
		if err != nil {
			log.Println("redis.RdbUserSignDays.Exists err", err.Error())
			return users, false
		}
		//获取前20名
		strUserIds, err := redis.RdbUserSignDays.ZRevRange(redis.Ctx, "CheckInsRank", 0, 19).Result()
		if err != nil {
			log.Println("redis.RdbUserSignDays.ZRevRange err", err.Error())
			return users, false
		}
		for _, strUserId := range strUserIds {
			userId, err := strconv.ParseInt(strUserId, 10, 64)
			if err != nil {
				log.Println(err.Error())
				return users, false
			}
			user, flag := si.GetUserById(userId)
			if !flag {
				log.Println("获取用户信息失败")
				return users, false
			}
			users = append(users, user)
		}
		return users, true
	} else {
		//缓存失效
		if _, err := redis.RdbUserSignDays.Expire(redis.Ctx, "CheckInsRank", config.ExpireTime).Result(); err != nil {
			log.Println("redis.RdbUserSignDays.Expire", err.Error())
			return users, false
		}
		//查数据库
		ids, flag := si.RankIds()
		if !flag {
			log.Println("获取排行榜用户id失败")
			return users, false
		}
		//加入缓存中
		for _, id := range ids {
			strId := strconv.Itoa(int(id.Id))
			z := redis.Z
			z.Member = strId
			z.Score = float64(id.Days)
			if _, err := redis.RdbUserSignDays.ZAdd(redis.Ctx, strId, &z).Result(); err != nil {
				log.Println("redis.RdbUserSignDays.ZAdd", err.Error())
				return users, false
			}
			user, flag := si.GetUserById(id.Id)
			if !flag {
				log.Println("获取用户信息失败")
				return users, false
			}
			users = append(users, user)
		}
		return users, true
	}

}
func (si *SignInImpl) ConsecutiveCheckIns(userId int64) (int64, bool) {
	var signDao dao.SignInDao
	strUserId := strconv.Itoa(int(userId))
	//先查是不是第一次签到
	flag, err := signDao.IsFirstSignIn(userId)
	if err != nil {
		log.Println("这里出问题了")
		return 0, false
	}
	if flag == 0 { //说明是第一次签到，新建缓存
		if _, err := redis.RdbUSIid.SAdd(redis.Ctx, strUserId, config.DefaultRedisValue).Result(); err != nil {
			log.Println("redis.RdbUSIid.SAdd err", err.Error())
			return 0, false
		}
		//设置缓存时间
		if _, err = redis.RdbUSIid.Expire(redis.Ctx, strUserId, time.Duration(config.OneDayOfHours)*time.Second).Result(); err != nil {
			log.Println("redis.RdbUSIid.Expire err", err.Error())
			return 0, false
		}
	}

	//添加签到记录
	err = signDao.UpdateSignInDays(userId)
	if err != nil {
		log.Println("update sign in day err", err.Error())
		return 0, false
	}

	//获取签到的天数
	sIds, err := signDao.SelectSids(userId)
	if err != nil {
		log.Println("获取签到天数 err", err.Error())
		return 0, false
	}
	log.Println("获取天数成功", len(sIds))
	//chess := si.UpdateDays(userId, int64(len(sIds)))
	err = dao.UpdateDays(userId, int64(len(sIds)))
	/*if !chess {
		log.Println("更新用户表的连续签到天数失败")
		return 0, false
	}*/
	if err != nil {
		log.Println("更新用户表的连续签到天数失败", err.Error())
		return 0, false
	}
	//加入缓存,先判断RdbUSIid和RdbUserSignDays是否过期，
	if n, err := redis.RdbUSIid.Exists(redis.Ctx, strUserId).Result(); n > 0 {
		//没有过期
		if err != nil {
			log.Println("redis.RdbUSIid.Exists err:", err.Error())
			return 0, false
		}
		for _, id := range sIds {
			if _, err := redis.RdbUSIid.SAdd(redis.Ctx, strUserId, id).Result(); err != nil {
				log.Println(err.Error())
				//保证和数据库的记录一致
				redis.RdbUSIid.Del(redis.Ctx, strUserId)
				return 0, false
			}

		}

	}
	//过期了
	if _, err := redis.RdbUSIid.SAdd(redis.Ctx, strUserId, config.DefaultRedisValue).Result(); err != nil {
		log.Println("redis.RdbUSIid.SAdd err", err.Error())
		return 0, false
	}
	//设置缓存时间
	if _, err = redis.RdbUSIid.Expire(redis.Ctx, strUserId, time.Duration(config.OneDayOfHours)*time.Second).Result(); err != nil {
		log.Println("redis.RdbUSIid.Expire err", err.Error())
		return 0, false
	}
	for _, id := range sIds {
		if _, err := redis.RdbUSIid.SAdd(redis.Ctx, strUserId, id).Result(); err != nil {
			log.Println(err.Error())
			//保证和数据库的记录一致
			redis.RdbUSIid.Del(redis.Ctx, strUserId)
			return 0, false
		}
	}
	//更新排行榜
	redis.Z.Score = float64(len(sIds))

	redis.Z.Member = userId

	if _, err := redis.RdbUserSignDays.ZAdd(redis.Ctx, "CheckInsRank", &redis.Z).Result(); err != nil {
		log.Println("redis.RdbUserSignDays.ZAdd err", err.Error())
		return 0, false
	}

	if _, err = redis.RdbUserSignDays.Expire(redis.Ctx, "CheckInsRank", config.ExpireTime).Result(); err != nil {
		log.Println("设置过期时间失败", err.Error())
		return 0, false
	}

	return int64(len(sIds)), true

}
func (si *SignInImpl) BreakCheckIns(userId int64) bool {
	//先删除数据库和缓存中的记录
	strUserId := strconv.Itoa(int(userId))
	var signInDao dao.SignInDao
	if n, err := redis.RdbUSIid.Exists(redis.Ctx, strUserId).Result(); n > 0 {
		//缓存存在
		if err != nil {
			log.Println("redis.RdbUSIid.Exists err:", err.Error())
			return false
		}
		if _, err = redis.RdbUSIid.Del(redis.Ctx, strUserId).Result(); err != nil {
			log.Println("删除缓存失败", err.Error())
			return false
		}
		//删数据库
		if err = signInDao.DeleteSignDays(userId); err != nil {
			log.Println("删除连续签到日数失败", err.Error())
			return false
		}

	}
	//缓存过期了，直接删数据库即可
	if err := signInDao.DeleteSignDays(userId); err != nil {
		log.Println("删除连续签到日数失败", err.Error())
		return false
	}
	//更新用户表的连续签到天数
	flag := si.UpdateDays(userId, 0)
	if !flag {
		log.Println("更新用户表连续签到天数失败")
		return false
	}
	//新增记录
	if err := signInDao.UpdateSignInDays(userId); err != nil {
		log.Println("新增签到日数失败", err.Error())
		return false
	}
	//新建缓存
	if _, err := redis.RdbUSIid.SAdd(redis.Ctx, strUserId, config.DefaultRedisValue).Result(); err != nil {
		log.Println("redis.RdbUSIid.SAdd err", err.Error())
		return false
	}
	//设置缓存时间
	if _, err := redis.RdbUSIid.Expire(redis.Ctx, strUserId, time.Duration(config.OneDayOfHours)*time.Second).Result(); err != nil {
		log.Println("redis.RdbUSIid.Expire err", err.Error())
		return false
	}
	//更新排行榜
	redis.Z.Score = 1

	redis.Z.Member = userId

	if _, err := redis.RdbUserSignDays.ZAdd(redis.Ctx, "CheckInsRank", &redis.Z).Result(); err != nil {
		log.Println("redis.RdbUserSignDays.ZAdd err", err.Error())
		return false
	}

	if _, err := redis.RdbUserSignDays.Expire(redis.Ctx, "CheckInsRank", config.ExpireTime).Result(); err != nil {
		log.Println("设置过期时间失败", err.Error())
		return false
	}
	//更新用户表的连续签到天数
	flag = si.UpdateDays(userId, 1)
	if !flag {
		log.Println("更新用户表连续签到天数失败")
		return false
	}

	return true

}
