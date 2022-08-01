package dao

import (
	"Points_Mall/config"
	"Points_Mall/middlerware/redis"
	"gorm.io/gorm"
	"log"
	"strconv"
	"sync"
)

type GoodTable struct {
	Id        int64
	ScoreNeed int64  //需要的积分
	GoodName  string //商品的名字
	Number    int32  //商品的数目
	CoverUrl  string //图片的url
}

type GoodDao struct {
}

var (
	goodDao  *GoodDao
	goodOnce sync.Once
)

func NewGoodDao() *GoodDao {
	goodOnce.Do(
		func() {
			goodDao = &GoodDao{}
		})
	return goodDao
}

func (*GoodDao) GetGoodNumber(goodId int64) (int32, error) {
	var goodTable GoodTable
	if err := Db.Select("number").Where("id = ?", goodId).First(&goodTable).Error; err != nil {
		log.Println(err.Error())
		return 0, err
	}
	return goodTable.Number, nil
}

func (*GoodDao) DecreaseNumber(goodId int64, number int32) error {
	if err := Db.Where("good_id", goodId).
		Update("number", gorm.Expr("number - ?", number)).Error; err != nil {
		log.Println(err.Error())
		return err
	}
	return nil

}

func (*GoodDao) IncreaseNumber(goodId int64, number int32) error {
	if err := Db.Where("good_id", goodId).
		Update("number", gorm.Expr("number + ?", number)).Error; err != nil {
		log.Println(err.Error())
		return err
	}

	return nil
}

func (*GoodDao) SelectGood(goodId int64) (GoodTable, error) {
	var goodTable GoodTable
	if err := Db.Where("id = ?", goodId).First(&goodTable).Error; err != nil {
		log.Println(err.Error())
		return goodTable, err
	}
	//更新缓存时间,维持热度
	goodIdStr := strconv.Itoa(int(goodId))
	_, err := redis.RdbGNumber.Expire(redis.Ctx, goodIdStr, config.ExpireTime).Result()
	if err != nil {
		log.Println("redis.RdbGNumber.Expire err:", err.Error())
		//redis出错不影响数据库查询到的数据
		return goodTable, err
	}
	return goodTable, nil
}

func (*GoodDao) UpdateGood(goodId int64, goodTable *GoodTable) error {
	if err := Db.Model(GoodTable{}).Where("id = ?", goodId).Updates(goodTable).Error; err != nil {
		log.Println(err.Error())
		return err
	}
	//更新缓存
	goodIdStr := strconv.Itoa(int(goodTable.Id))
	if _, err := redis.RdbGNumber.Set(redis.Ctx, goodIdStr, goodTable.Number, config.ExpireTime).Result(); err != nil {
		log.Println("redis.RdbGNumber.Set err", err.Error())
		redis.RdbGNumber.Del(redis.Ctx, goodIdStr)
		return err
	}
	return nil
}

func (*GoodDao) InsertGood(goodTable *GoodTable) error {
	if err := Db.Model(GoodTable{}).Create(goodTable).Error; err != nil {
		log.Println(err.Error())
		return err
	}
	goodIdStr := strconv.Itoa(int(goodTable.Id))

	if _, err := redis.RdbGNumber.Set(redis.Ctx, goodIdStr, goodTable.Number, config.ExpireTime).Result(); err != nil {
		log.Println("redis.RdbGNumber.Set err", err.Error())
		redis.RdbGNumber.Del(redis.Ctx, goodIdStr)
		return err
	}
	return nil
}

func (*GoodDao) DeleteGood(goodId int64) error {
	var goodTable GoodTable
	goodTable.Id = goodId
	if err := Db.Delete(&goodTable).Error; err != nil {
		log.Println(err.Error())
		return err
	}
	//删除缓存
	goodIdStr := strconv.Itoa(int(goodId))
	_, err := redis.RdbGNumber.Del(redis.Ctx, goodIdStr).Result()
	if err != nil {
		log.Println("redis.RdbGNumber.Del err", err.Error())
		return err
	}

	return nil
}
