package service

import "Points_Mall/dao"

type UserService interface {
	/*

		上层api使用

	*/
	// GetTableUserByUserName  根据用户名，获取一个用户对象
	GetTableUserByUserName(username string) dao.UserTable

	// InsertTableUser 将用户对象放入sql
	InsertTableUser(userTable *dao.UserTable) bool

	// GetTableUserById 根据Id，获取一个用户对象
	GetTableUserById(id int64) dao.UserTable

	/*
		其它服务使用
	*/

	// GetUserById 根据id获取一个详细的用户
	GetUserById(id int64) (User, bool)

	// GetUsers 获取全部的用户信息
	GetUsers(ids []int64) ([]User, bool)

	copyUsers(result *[]User, data *[]dao.UserTable, userId int64) error

	createUser(user *User, userTable *dao.UserTable, userId int64)

	// DecreaseUserScore 减少用户的积分
	DecreaseUserScore(userId int64, number int64) bool

	// RankIds 获取连续签到的排名id
	RankIds() ([]dao.UserTable, bool)

	// UpdateDays 更新用户连续签到的天数
	UpdateDays(userId int64, days int64) bool
}
type User struct {
	Id    int64  `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Score int64  `json:"score"`
	Rank  int64  `json:"rank"`
	Days  int64  `json:"days"`
}
