package dao

import (
	"Points_Mall/config"
	"gorm.io/gorm"
	"log"
	"sync"
)

type SignInTable struct {
	Id     int64
	UserId int64
	Cancel int32 //0连续，1不连续
	//Days   int64 //连续签到天数,可以不需要，用redis记录即可
}

type SignInDao struct {
}

var (
	signInDap  *SignInDao
	signInOnce sync.Once
)

func NewSignInDao() *SignInDao {
	signInOnce.Do(func() {
		signInDap = &SignInDao{}
	})
	return signInDap
}

func (*SignInDao) SelectSids(userId int64) ([]int64, error) {
	var sIds []int64
	if err := Db.Model(SignInTable{}).Where("user_id = ?", userId).Pluck("id", &sIds).Error; err != nil {
		return sIds, err
	}
	return sIds, nil
}

func (*SignInDao) IsFirstSignIn(userId int64) (int64, error) {
	var signIn SignInTable
	if err := Db.Select("id").Where("user_id", userId).First(&signIn).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		log.Println(err.Error())
		return -1, err
	}
	return signIn.Id, nil
}

func (*SignInDao) UpdateSignInDays(userId int64) error {
	if err := Db.Model(SignInTable{}).
		Create(map[string]interface{}{"user_id": userId, "cancel": config.IsCoiled}).Error; err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func (*SignInDao) DeleteSignDays(userId int64) error {
	if err := Db.Where("user_id", userId).Delete(&SignInTable{}).Error; err != nil {
		log.Println(err.Error())
		return err
	}
	return nil

}
