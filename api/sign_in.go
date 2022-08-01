package api

import (
	"Points_Mall/config"
	"Points_Mall/service"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

type SignIn struct {
	Response
	Msg string
}
type Ranks struct {
	Response
	users []service.User
}

// SingIn POST 用户签到
func SingIn(c *gin.Context) {
	strUserId := c.GetString("userId")
	userId, err := strconv.ParseInt(strUserId, 10, 64)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusOK, Response{StatusCode: 1})
	} else {
		//连续签到还是断签
		strActionType := c.Query("actionType")
		actionType, _ := strconv.ParseInt(strActionType, 10, 64)
		s := service.SignInImpl{}

		if actionType == config.IsCoiled {
			days, flag := s.ConsecutiveCheckIns(userId)
			if !flag {
				c.JSON(http.StatusOK, SignIn{
					Response: Response{StatusCode: 1, StatusMsg: "签到失败哦"}})
			} else {
				c.JSON(http.StatusOK, SignIn{
					Response: Response{StatusCode: 1, StatusMsg: "签到成功"},
					Msg:      fmt.Sprintf("连续签到%d天", days),
				})
			}
		} else {

			flag := s.BreakCheckIns(userId)
			if !flag {
				c.JSON(http.StatusOK, SignIn{
					Response: Response{StatusCode: 1, StatusMsg: "签到失败哦"}})
			} else {
				c.JSON(http.StatusOK, SignIn{
					Response: Response{StatusCode: 1, StatusMsg: "签到成功"},
					Msg:      "断签后第一天签到",
				})
			}

		}

	}
}

func Rank(c *gin.Context) {
	rank, err := strconv.ParseInt(c.Query("rank"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1})
	} else {
		if rank != config.Rank {
			c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "错误的操作"})
		} else {
			s := service.SignInImpl{}
			users, flag := s.GetRank()
			if !flag {
				c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "获取排行榜失败"})
			}
			c.JSON(http.StatusOK, Ranks{
				Response: Response{StatusCode: 0, StatusMsg: "获取排行榜成功"},
				users:    users,
			})
		}
	}

}
