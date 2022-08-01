package api

import (
	"Points_Mall/service"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

type ScoreResp struct {
	Response
	scores []service.Score
}

func AddScore(c *gin.Context) {
	//通过token获取用户id
	userId, err := strconv.ParseInt(c.GetString("userId"), 10, 64)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusOK, Response{StatusCode: 1})
	} else {
		taskI := c.Query("taskId")
		taskId, _ := strconv.ParseInt(taskI, 10, 64)
		s := service.ScoreServiceImpl{}
		flag := s.AddScore(userId, taskId)
		if !flag {
			c.JSON(http.StatusOK, Response{
				StatusCode: 1,
				StatusMsg:  "添加分数失败",
			})
		} else {
			c.JSON(http.StatusOK, Response{
				StatusCode: 0,
				StatusMsg:  "添加分数成功",
			})
		}

	}

}

func GetScoreRecord(c *gin.Context) {
	//通过token获取用户id
	userId, err := strconv.ParseInt(c.GetString("userId"), 10, 64)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusOK, Response{StatusCode: 1})
	} else {
		s := service.ScoreServiceImpl{}
		scores, flag := s.GetScoreRecord(userId)
		if !flag {
			c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "获取分数记录失败"})
		} else {
			c.JSON(http.StatusOK, ScoreResp{
				Response: Response{StatusCode: 0, StatusMsg: "获取分数记录成功"},
				scores:   scores,
			})
		}
	}
}
