package api

import (
	"Points_Mall/dao"
	"Points_Mall/service"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

type Response struct {
	StatusCode int32  `json:"status_code"`
	StatusMsg  string `json:"status_msg,omitempty"`
}

type UserLoginResponse struct {
	Response
	UserId int64  `json:"user_id,omitempty"`
	Token  string `json:"token"`
}

type UserResponse struct {
	Response
	User service.User `json:"user"`
}

func Register(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")

	usi := service.UserServiceImpl{}

	u := usi.GetTableUserByUserName(username)
	if username == u.UserName {
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: 1, StatusMsg: "User already exist"},
		})
	} else {
		newUser := dao.UserTable{
			UserName: username,
			PassWord: service.EnCoder(password),
		}
		if usi.InsertTableUser(&newUser) != true {
			println("Insert Data Fail")
		}
		u := usi.GetTableUserByUserName(username)
		token := service.GenerateToken(username)
		log.Println("注册返回的id: ", u.Id)
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: 0},
			UserId:   u.Id,
			Token:    token,
		})
	}
}

func Login(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")
	encoderPassword := service.EnCoder(password)
	println(encoderPassword)

	usi := service.UserServiceImpl{}

	u := usi.GetTableUserByUserName(username)

	if encoderPassword == u.PassWord {
		token := service.GenerateToken(username)
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: 0},
			UserId:   u.Id,
			Token:    token,
		})
	} else {
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: 1, StatusMsg: "Username or Password Error"},
		})
	}
}

func GetUserInfo(c *gin.Context) {
	userId, err := strconv.ParseInt(c.Query("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1})
	} else {
		usi := service.UserServiceImpl{}
		user, flag := usi.GetUserById(userId)
		if !flag {
			c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "获取用户信息失败"})
		} else {
			c.JSON(http.StatusOK, UserResponse{
				Response: Response{
					StatusCode: 1,
					StatusMsg:  "获取用户信息成功"},
				User: user,
			})
		}
	}
}
