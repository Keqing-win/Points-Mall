package dao

import (
	"gorm.io/gorm"
	"log"
)

// UserTable User 用户表结构
type UserTable struct {
	Id       int64
	UserName string
	PassWord string
	Score    int64
	Days     int64 //连续签到天数
}

func RankIds() ([]UserTable, error) {
	var ids []UserTable
	if err := Db.Model(UserTable{}).Order("days desc").Limit(20).Find(&ids).Error; err != nil {
		log.Println(err.Error())
		return ids, err
	}
	return ids, nil
}

func UpdateDays(userId int64, days int64) error {
	if err := Db.Model(UserTable{}).Where("id = ?", userId).Update("days", days).Error; err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func DecreaseUserScore(userId int64, number int64) error {
	if err := Db.Model(&UserTable{}).Where("id = ?", userId).Update("score", gorm.Expr("score - ?", number)).Error; err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func GetTableUserByUserName(username string) (UserTable, error) {
	var user UserTable
	if err := Db.Where("user_name = ?", username).First(&user).Error; err != nil {
		log.Println(err.Error())
		return user, err
	}
	return user, nil
}

func GetTableUserById(id int64) (UserTable, error) {
	var user UserTable
	if err := Db.Where("id = ?", id).First(&user).Error; err != nil {
		log.Println(err.Error())
		return user, err
	}
	return user, nil
}

func InsertTableUser(user *UserTable) error {
	if err := Db.Model(UserTable{}).Create(&user).Error; err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func UpdateTableUserScore(userId int64, score string) error {
	if err := Db.Model(UserTable{}).Where("id", userId).Update("score", score).Error; err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}
