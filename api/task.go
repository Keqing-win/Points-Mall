package api

import (
	"Points_Mall/config"
	"Points_Mall/dao"
	"Points_Mall/service"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"time"
)

type TaskResponse struct {
	Response
	Task dao.TaskTable `json:"task"`
}

type Task struct {
	//Integral    int32     `json:"integral"`    //一个任务所含的积分
	//PublishTime time.Time `form:"publishTime" json:"publishTime" binding:"required"` //发表任务的时间戳
	Title   string `form:"title" json:"title" binding:"required"`      //任务的标题
	Content string `form:"content" json:"content" binding:"required" ` //任务的内容
	Score   string ` form:"score" json:"score" binding:"required"`
	//Cancel      int32     `json:"cancel"`      //0表示取消，1表示完成任务
	Only int32 `form:"only" json:"only" binding:"required"` //0表示一次，1表示每日任务
}

func FinishTask(c *gin.Context) {
	userId, err := strconv.ParseInt(c.GetString("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1})
	}
	finishStr := c.Query("finish")
	taskId, _ := strconv.ParseInt(c.Query("taskId"), 10, 64)
	finish, _ := strconv.ParseInt(finishStr, 10, 64)
	if finish != config.Finish {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "错误的操作，请重试"})
	}
	t := service.TaskServiceImpl{}
	flag := t.FinishTask(userId, taskId)
	if !flag {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "操作失败"})
	}
	c.JSON(http.StatusOK, Response{StatusCode: 0, StatusMsg: "任务完成"})

}

func AddTask(c *gin.Context) {
	var task Task
	if err := c.ShouldBind(&task); err != nil {
		log.Println("ShouldBind err:", err.Error())
		c.JSON(http.StatusOK, Response{
			StatusCode: 1,
			StatusMsg:  "数据处理出错",
		})
	} else {
		t := service.TaskServiceImpl{}
		var taskTable dao.TaskTable
		taskTable.Title = task.Title
		taskTable.Content = task.Content
		taskTable.PublishTime = time.Now().Unix()
		taskTable.Only = task.Only
		taskTable.Score = task.Score
		if taskTable.Only == config.OnlyDay {
			flag := t.AddTaskLimit(taskTable)
			if !flag {
				c.JSON(http.StatusOK, Response{
					StatusCode: 1,
					StatusMsg:  "数据处理出错",
				})
			} else {
				c.JSON(http.StatusOK, TaskResponse{
					Response: Response{
						StatusCode: 0,
						StatusMsg:  "任务到时清除",
					},
				})
			}

		} else {
			if taskTable.Only == config.IsDaily {
				flag := t.AddTask(taskTable)
				if !flag {
					c.JSON(http.StatusOK, Response{
						StatusCode: 1,
						StatusMsg:  "数据处理出错",
					})
				} else {
					c.JSON(http.StatusOK, Response{
						StatusCode: 0,
						StatusMsg:  "添加每日任务成功",
					})
				}

			} else {
				c.JSON(http.StatusOK, Response{
					StatusCode: 1,
					StatusMsg:  "请输入正确的任务类型",
				})
			}
		}
	}

}

func DeleteTask(c *gin.Context) {
	taskId := c.Query("taskId")
	t := service.TaskServiceImpl{}
	id, err := strconv.ParseInt(taskId, 10, 64)
	if err != nil {
		log.Println(err.Error())
	}
	flag := t.DeleteTask(id)
	if !flag {
		c.JSON(http.StatusOK, Response{
			StatusCode: 1,
			StatusMsg:  "删除任务失败",
		})
	} else {
		c.JSON(http.StatusOK, Response{
			StatusCode: 0,
			StatusMsg:  "删除任务成功",
		})
	}

}
