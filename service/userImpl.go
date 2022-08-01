package service

import (
	"Points_Mall/config"
	"Points_Mall/dao"
	"Points_Mall/middlerware/redis"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"log"
	"strconv"
	"sync"
	"time"
)

type UserServiceImpl struct {
	ScoreService
	SignInService
}

func (u *UserServiceImpl) RankIds() ([]dao.UserTable, bool) {
	userIds, err := dao.RankIds()
	if err != nil {
		log.Println("获取排名用户id失败", err.Error())
		return nil, false
	}
	return userIds, true

}

func (u *UserServiceImpl) UpdateDays(userId int64, days int64) bool {
	err := dao.UpdateDays(userId, days)
	if err != nil {
		log.Println("更新用户连续签到天数失败", err.Error())
		return false
	}
	return true

}

func (u *UserServiceImpl) DecreaseUserScore(userId int64, number int64) bool {
	//score存储的缓存在这个情况下没法使用，只能调用数据库了
	err := dao.DecreaseUserScore(userId, number)
	if err != nil {
		log.Println("减少用户积分失败")
		return false
	}
	return true

}

func (u *UserServiceImpl) GetTableUserByUserName(username string) dao.UserTable {
	tableUser, err := dao.GetTableUserByUserName(username)
	if err != nil {
		log.Println("TableUser not Find")
	}
	log.Println("Find User Suc")
	return tableUser

}

func (u *UserServiceImpl) InsertTableUser(userTable *dao.UserTable) bool {
	if err := dao.InsertTableUser(userTable); err != nil {
		log.Println("Insert TableUser err")
		return false
	}
	log.Println("Insert TableUser Suc")
	return true

}

func (u *UserServiceImpl) GetTableUserById(id int64) dao.UserTable {
	tableUser, err := dao.GetTableUserById(id)
	if err != nil {
		log.Println("TableUser not Find")
	}
	log.Println("Find User Suc")
	return tableUser
}

func (u *UserServiceImpl) GetUserById(id int64) (User, bool) {
	var user User

	userTable, err := dao.GetTableUserById(id)
	if err != nil {
		log.Println("get table user err", err.Error())
		return user, false
	}
	u.createUser(&user, &userTable, id)

	return user, true
}

func (u *UserServiceImpl) copyUsers(result *[]User, data *[]dao.UserTable, userId int64) error {
	for _, temp := range *data {
		var user User
		//将video进行组装，添加想要的信息,插入从数据库中查到的数据
		u.createUser(&user, &temp, userId)
		*result = append(*result, user)
	}
	return nil
}

func (u *UserServiceImpl) createUser(user *User, userTable *dao.UserTable, userId int64) {
	user.Id = userTable.Id
	user.Name = userTable.UserName
	var wg sync.WaitGroup

	wg.Add(3)
	go func() {
		score, err := u.GetScore(userId)
		if err != nil {
			log.Println("获取用户分数失败", err.Error())

		}
		user.Score = score
		wg.Done()
	}()
	go func() {
		var flag bool
		user.Days, flag = u.ConsecutiveCheckIns(userId)
		if !flag {
			log.Println("获取用户连续登录天数失败")
		}
		wg.Done()
	}()
	go func() {
		//获取排名
		if n, err := redis.RdbUserSignDays.Exists(redis.Ctx, "CheckInsRank").Result(); n > 0 {
			//还在缓存中
			if err != nil {
				log.Println("redis.RdbUserSignDays.Exists")
			}
			strUserId := strconv.Itoa(int(userId))
			user.Rank, err = redis.RdbUserSignDays.ZRevRank(redis.Ctx, "CheckInsRank", strUserId).Result()
			if err != nil {
				log.Println("获取用户的排名失败")
			}
		}
		wg.Done()

	}()
	wg.Wait()
}

// GetUsers 提供给连续签到排名的服务
func (u *UserServiceImpl) GetUsers(ids []int64) ([]User, bool) {

	var users []User
	var tableUsers []dao.UserTable
	for _, id := range ids {
		tableUser, err := dao.GetTableUserById(id)
		if err != nil {
			log.Println("get table user err", err.Error())
			return users, false
		}
		tableUsers = append(tableUsers, tableUser)
	}
	for _, id := range ids {
		_ = u.copyUsers(&users, &tableUsers, id)
	}

	return users, true

}

// GenerateToken 根据username生成一个token
func GenerateToken(username string) string {
	u := UserService.GetTableUserByUserName(new(UserServiceImpl), username)
	fmt.Printf("generatetoken: %v\n", u)
	token := NewToken(u)
	println(token)
	return token
}

// NewToken 根据信息创建token
func NewToken(u dao.UserTable) string {
	expiresTime := time.Now().Unix() + int64(config.OneDayOfHours)
	fmt.Printf("expiresTime: %v\n", expiresTime)
	id64 := u.Id
	fmt.Printf("id: %v\n", strconv.FormatInt(id64, 10))
	claims := jwt.StandardClaims{
		Audience:  u.UserName,
		ExpiresAt: expiresTime,
		Id:        strconv.FormatInt(id64, 10),
		IssuedAt:  time.Now().Unix(),
		Issuer:    "pointsmall",
		NotBefore: time.Now().Unix(),
		Subject:   "token",
	}
	var jwtSecret = []byte(config.Secret)
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	if token, err := tokenClaims.SignedString(jwtSecret); err == nil {
		token = "Bearer " + token
		println("generate token success!\n")
		return token
	} else {
		println("generate token fail\n")
		return "fail"
	}
}

// EnCoder 密码加密
func EnCoder(password string) string {
	h := hmac.New(sha256.New, []byte(password))
	sha := hex.EncodeToString(h.Sum(nil))
	fmt.Println("Result: " + sha)
	return sha
}
