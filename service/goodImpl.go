package service

import (
	"Points_Mall/config"
	"Points_Mall/dao"
	"Points_Mall/middlerware/rabbitmq"
	"Points_Mall/middlerware/redis"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"
)

type GoodServiceImpl struct {
	UserService
	ScoreService
}

func (g *GoodServiceImpl) IncreaseGoodNumber(goodId int64, number int32) bool {
	var goodDao dao.GoodDao
	if err := goodDao.IncreaseNumber(goodId, number); err != nil {
		log.Println("increase good number err", err.Error())
	}
	//更新缓存
	goodTable, err := goodDao.SelectGood(goodId)
	if err != nil {
		log.Println("search good table err", err.Error())
	}
	log.Println("search good table suc")
	goodIdStr := strconv.Itoa(int(goodId))
	if _, err := redis.RdbGNumber.Set(redis.Ctx, goodIdStr, goodTable.Number, config.ExpireTime).Result(); err != nil {
		log.Println("redis.RdbGNumber.Set err", err.Error())
		redis.RdbGNumber.Del(redis.Ctx, goodIdStr)
		return false
	}
	return true
}

func (g *GoodServiceImpl) CheckUserBalance(userId, goodId int64) (bool, error) {
	//用户的分数
	//直接调score层的接口，因为有缓存要比直接查数据库快
	scroe, err := g.GetScore(userId)
	if err != nil {
		log.Println("获取用户分数失败")
		return false, err
	}
	goodDao := dao.NewGoodDao()
	goodTable, err := goodDao.SelectGood(goodId)
	if err != nil {
		log.Println("获取商品失败")
		return false, err
	}
	if goodTable.ScoreNeed > scroe {
		return true, errors.New("积分不够")
	}
	//对用户的积分进行减少
	flag := g.DecreaseUserScore(userId, goodTable.ScoreNeed)
	if !flag {
		return false, errors.New("减少用户积分失败")
	}
	return true, nil

}

// OrderAction 这里就要考虑高并发了，考虑将商品减少的信息放入消息队列
//对订单进行下单和取消
func (g *GoodServiceImpl) OrderAction(userId, goodId, orderId int64, actionType int32) bool {
	strUserId := strconv.Itoa(int(userId))
	strGoodId := strconv.Itoa(int(goodId))
	strOderId := strconv.Itoa(int(orderId))
	//将要操作数据库orders表的信息打入消息队列RmqOrderAdd或者RmqOrderDel
	//拼接打入信息
	sb := strings.Builder{}
	sb.WriteString(strGoodId)
	sb.WriteString(" ")
	sb.WriteString(strOderId)
	sb.WriteString(" ")
	sb.WriteString(strUserId)
	//第一步:维护redis - RdbGOid  RdbGNumber
	//增加订单
	if actionType == config.ValidOrder {
		if n, err := redis.RdbGOid.Exists(redis.Ctx, strGoodId).Result(); n > 0 {
			if err != nil {
				log.Println("redis.RdbGOid.Exists err", err.Error())
				return false
			}
			//如果加载过就将OrderId加入
			//先查询数量是否正确
			flag, err := redis.RdbGNumber.Exists(redis.Ctx, strGoodId).Result()
			if err != nil {
				log.Println("redis.RdbGNumber.Exists err", err.Error())
				return false
			}
			if flag == config.KeyNotExist {
				//去数据库查
				goodDao := dao.NewGoodDao()
				number, err := goodDao.GetGoodNumber(goodId)
				if err != nil {
					log.Println("get good number err", err.Error())
					return false
				}
				if number <= 0 {
					log.Println("good number empty")
					return false
				} else {
					//商品数量是大于0的
					_, err := redis.RdbGNumber.Set(redis.Ctx, strGoodId, number, config.ExpireTime).Result()
					if err != nil {
						log.Println("redis.RdbGNumber.Set err", err.Error())
						return false
					}
					//当商品数量正确时，将订单消息加入  redis.RdbGOid
					if _, err = redis.RdbGOid.SAdd(redis.Ctx, strGoodId, orderId).Result(); err != nil {
						log.Println("redis.RdbGOid.SAdd err", err.Error())
						return false
					} else {
						//现在redis是正确的数据，客户端一定收到正确的回复，不管数据库操作成功与否，将消息放入消息队列,消息消费成功与否和也不影响操作的成功
						rabbitmq.RmqOrderAdd.Publish(sb.String())
					}
				}
			} else {
				//可以查到数量的缓存的
				//判断数量是否合理
				strNumber, err := redis.RdbGNumber.Get(redis.Ctx, strGoodId).Result()
				if err != nil {
					log.Println("redis.RdbGNumber.Get err ", err.Error())
					return false
				}
				number, _ := strconv.ParseInt(strNumber, 10, 64)
				if number <= 0 {
					log.Println("good number empty")
					return false
				} else {
					//对商品数减一，然后维持热度(更新过期时间)
					_, err := redis.RdbGNumber.Decr(redis.Ctx, strGoodId).Result()
					if err != nil {
						log.Println("redis.RdbGNumber.Decr err", err.Error())
						return false
					}

					strNumber, err = redis.RdbGNumber.Get(redis.Ctx, strGoodId).Result()
					if err != nil {
						log.Println("redis.RdbGNumber.Get err ", err.Error())
						return false
					}
					number, _ := strconv.ParseInt(strNumber, 10, 64)
					_, err = redis.RdbGNumber.Set(redis.Ctx, strGoodId, number, config.ExpireTime).Result()
					if err != nil {
						log.Println("redis.RdbGNumber.Set err:", err.Error())
						return false
					}
					log.Println("缓存商品减1，维持热度成功")
					//当商品数量正确时，将订单消息加入  redis.RdbGOid
					if _, err = redis.RdbGOid.SAdd(redis.Ctx, strGoodId, orderId).Result(); err != nil {
						log.Println("redis.RdbGOid.SAdd err", err.Error())
						return false
					} else {
						//现在redis是正确的数据，客户端一定收到正确的回复，不管数据库操作成功与否，将消息放入消息队列,消息消费成功与否和也不影响操作的成功
						rabbitmq.RmqOrderAdd.Publish(sb.String())
					}
				}

			}

		} else {
			//没查到RdbGOid中的缓存,则维持RdbGOid，key:goodid value:-1(防止脏读) ，然后设置过期时间，最后去订单表中查询，更新缓存
			if _, err := redis.RdbGOid.SAdd(redis.Ctx, strGoodId, config.DefaultRedisValue).Result(); err != nil {
				log.Println("redis.RdbGOid.SAdd err", err.Error())
				return false
			}
			//给键设置过期时间
			_, err = redis.RdbGOid.Expire(redis.Ctx, strGoodId,
				time.Duration(config.OneMonth)*time.Second).Result()
			if err != nil {
				log.Println("设置 redis.RdbGOid 过期时间失败 err", err.Error())
				return false
			}
			//去订单表中查询,现在订单一定是正确的，不会超过商品数量
			orderDao := dao.NewOrderDao()
			orderIds, err := orderDao.GetValueOrderIds(goodId)
			if err != nil {
				log.Println("get value order err", err.Error())
				return false
			}
			for _, orderId := range orderIds {
				if _, err = redis.RdbGOid.SAdd(redis.Ctx, strGoodId, orderId).Result(); err != nil {
					log.Println(" redis.RdbGOid.SAdd err", err.Error())
					//保证和数据库一致
					redis.RdbGOid.Del(redis.Ctx, strGoodId)
					return false
				}
			}
			//查询商品数量是否合适
			flag, err := redis.RdbGNumber.Exists(redis.Ctx, strGoodId).Result()
			if err != nil {
				log.Println("redis.RdbGNumber.Exists err", err.Error())
				return false
			}
			if flag == config.KeyNotExist {
				//去数据库查
				goodDao := dao.NewGoodDao()
				number, err := goodDao.GetGoodNumber(goodId)
				if err != nil {
					log.Println("get good number err", err.Error())
					return false
				}
				if number <= 0 {
					log.Println("good number empty")
					return false
				} else {
					//商品数量是大于0的
					_, err := redis.RdbGNumber.Set(redis.Ctx, strGoodId, number, config.ExpireTime).Result()
					if err != nil {
						log.Println("redis.RdbGNumber.Set err", err.Error())
						return false
					}
					//当商品数量正确时，将订单消息加入  redis.RdbGOid
					if _, err = redis.RdbGOid.SAdd(redis.Ctx, strGoodId, orderId).Result(); err != nil {
						log.Println("redis.RdbGOid.SAdd err", err.Error())
						return false
					} else {
						//现在redis是正确的数据，客户端一定收到正确的回复，不管数据库操作成功与否，将消息放入消息队列,消息消费成功与否和也不影响操作的成功
						rabbitmq.RmqOrderAdd.Publish(sb.String())
					}
				}
			} else {
				//可以查到数量的缓存的
				//判断数量是否合理
				strNumber, err := redis.RdbGNumber.Get(redis.Ctx, strGoodId).Result()
				if err != nil {
					log.Println("redis.RdbGNumber.Get err ", err.Error())
					return false
				}
				number, _ := strconv.ParseInt(strNumber, 10, 64)
				if number <= 0 {
					log.Println("good number empty")
					return false
				} else {
					//对商品数减一，然后维持热度(更新过期时间)
					_, err := redis.RdbGNumber.Decr(redis.Ctx, strGoodId).Result()
					if err != nil {
						log.Println("redis.RdbGNumber.Decr err", err.Error())
						return false
					}

					strNumber, err = redis.RdbGNumber.Get(redis.Ctx, strGoodId).Result()
					if err != nil {
						log.Println("redis.RdbGNumber.Get err ", err.Error())
						return false
					}
					number, _ := strconv.ParseInt(strNumber, 10, 64)
					_, err = redis.RdbGNumber.Set(redis.Ctx, strGoodId, number, config.ExpireTime).Result()
					if err != nil {
						log.Println("redis.RdbGNumber.Set err:", err.Error())
						return false
					}
					log.Println("缓存商品减1，维持热度成功")
					//当商品数量正确时，将订单消息加入  redis.RdbGOid
					if _, err = redis.RdbGOid.SAdd(redis.Ctx, strGoodId, orderId).Result(); err != nil {
						log.Println("redis.RdbGOid.SAdd err", err.Error())
						return false
					} else {
						//现在redis是正确的数据，客户端一定收到正确的回复，不管数据库操作成功与否，将消息放入消息队列,消息消费成功与否和也不影响操作的成功
						rabbitmq.RmqOrderAdd.Publish(sb.String())
					}
				}
			}

		}
		//成功添加订单
		return true

	} else { //取消订单
		if n, err := redis.RdbGOid.Exists(redis.Ctx, strGoodId).Result(); n > 0 {
			if err != nil {
				log.Println("redis.RdbGOid.Exists err", err.Error())
				return false
			}
			//防止出现redis数据不一致情况，当redis删除操作成功，才执行数据库更新操作
			//先判断redis.RdbGNumber
			if n, err = redis.RdbGNumber.Exists(redis.Ctx, strGoodId).Result(); n > 0 {
				if err != nil {
					log.Println("redis.RdbGNumber.Exists err", err.Error())
					return false
				}
				//更新数量
				if _, err := redis.RdbGNumber.Incr(redis.Ctx, strGoodId).Result(); err != nil {
					log.Println("redis.RdbGNumber.Incr err", err.Error())
					return false
				}
				if _, err := redis.RdbGNumber.SRem(redis.Ctx, strGoodId, orderId).Result(); err != nil {
					log.Println("redis.RdbGNumber.SRem err", err.Error())
					return false
				} else {
					//去更新数据库
					rabbitmq.RmqOrderDel.Publish(sb.String())
				}

			} else {
				//没有查到 redis.RdbGNumber缓存 查数据库，重新设置热度
				goodDao := dao.NewGoodDao()
				number, err := goodDao.GetGoodNumber(goodId)
				if err != nil {
					log.Println("get good number err", err.Error())
					return false
				}
				_, err = redis.RdbGNumber.Set(redis.Ctx, strGoodId, number, config.ExpireTime).Result()
				if err != nil {
					log.Println("redis.RdbGNumber.Set err", err.Error())
					return false
				}
				//当商品数量正确时，将订单消息加入  redis.RdbGOid
				if _, err = redis.RdbGOid.SRem(redis.Ctx, strGoodId, orderId).Result(); err != nil {
					log.Println("redis.RdbGOid.SRem err", err.Error())
					return false
				} else {
					//现在redis是正确的数据，客户端一定收到正确的回复，不管数据库操作成功与否，将消息放入消息队列,消息消费成功与否和也不影响操作的成功
					rabbitmq.RmqOrderDel.Publish(sb.String())
				}

			}

		} else {
			//没查到RdbGOid中的缓存,则维持RdbGOid，key:goodid value:-1(防止脏读) ，然后设置过期时间，最后去订单表中查询，更新缓存
			if _, err := redis.RdbGOid.SAdd(redis.Ctx, strGoodId, config.DefaultRedisValue).Result(); err != nil {
				log.Println("redis.RdbGOid.SAdd err", err.Error())
				return false
			}
			//给键设置过期时间
			_, err = redis.RdbGOid.Expire(redis.Ctx, strGoodId,
				time.Duration(config.OneMonth)*time.Second).Result()
			if err != nil {
				log.Println("设置 redis.RdbGOid 过期时间失败 err", err.Error())
				return false
			}
			//去订单表中查询,现在订单一定是正确的，不会超过商品数量
			orderDao := dao.NewOrderDao()
			orderIds, err := orderDao.GetValueOrderIds(goodId)
			if err != nil {
				log.Println("get value order err", err.Error())
				return false
			}
			for _, orderId := range orderIds {
				if _, err = redis.RdbGOid.SAdd(redis.Ctx, strGoodId, orderId).Result(); err != nil {
					log.Println(" redis.RdbGOid.SAdd err", err.Error())
					//保证和数据库一致
					redis.RdbGOid.Del(redis.Ctx, strGoodId)
					return false
				}
			}
			//先判断redis.RdbGNumber
			if n, err = redis.RdbGNumber.Exists(redis.Ctx, strGoodId).Result(); n > 0 {
				if err != nil {
					log.Println("redis.RdbGNumber.Exists err", err.Error())
					return false
				}
				//更新数量
				if _, err := redis.RdbGNumber.Incr(redis.Ctx, strGoodId).Result(); err != nil {
					log.Println("redis.RdbGNumber.Incr err", err.Error())
					return false
				}
				if _, err := redis.RdbGNumber.SRem(redis.Ctx, strGoodId, orderId).Result(); err != nil {
					log.Println("redis.RdbGNumber.SRem err", err.Error())
					return false
				} else {
					//去更新数据库
					rabbitmq.RmqOrderDel.Publish(sb.String())
				}

			} else {
				//没有查到 redis.RdbGNumber缓存 查数据库，重新设置热度
				goodDao := dao.NewGoodDao()
				number, err := goodDao.GetGoodNumber(goodId)
				if err != nil {
					log.Println("get good number err", err.Error())
					return false
				}
				_, err = redis.RdbGNumber.Set(redis.Ctx, strGoodId, number, config.ExpireTime).Result()
				if err != nil {
					log.Println("redis.RdbGNumber.Set err", err.Error())
					return false
				}
				//当商品数量正确时，将订单消息加入  redis.RdbGOid
				if _, err = redis.RdbGOid.SRem(redis.Ctx, strGoodId, orderId).Result(); err != nil {
					log.Println("redis.RdbGOid.SRem err", err.Error())
					return false
				} else {
					//现在redis是正确的数据，客户端一定收到正确的回复，不管数据库操作成功与否，将消息放入消息队列,消息消费成功与否和也不影响操作的成功
					rabbitmq.RmqOrderDel.Publish(sb.String())
				}
			}

		}
		//成功取消订单
		return true

	}

}

func (g *GoodServiceImpl) CreateGood(table *dao.GoodTable) bool {
	var goodDao dao.GoodDao
	if err := goodDao.InsertGood(table); err != nil {
		log.Println("create good table err", err.Error())
		return false
	}

	log.Println("create good table suc")

	return true

}

func (g *GoodServiceImpl) ChangeGood(goodId int64, table *dao.GoodTable) bool {
	var goodDao dao.GoodDao
	if err := goodDao.UpdateGood(goodId, table); err != nil {
		log.Println("update good table err", err.Error())
		return false
	}

	log.Println("update good table suc")

	return true

}

func (g *GoodServiceImpl) DeleteGood(goodId int64) bool {
	var goodDao dao.GoodDao
	if err := goodDao.DeleteGood(goodId); err != nil {
		log.Println("delete good table err", err.Error())
		return false
	}

	log.Println("delete good table suc")

	return true
}

func (g *GoodServiceImpl) SearchGood(goodId int64) (dao.GoodTable, bool) {
	var goodDao dao.GoodDao
	taskTable, err := goodDao.SelectGood(goodId)
	if err != nil {
		log.Println("search good table err", err.Error())
		return taskTable, false
	}
	log.Println("search good table suc")
	return taskTable, true
}
