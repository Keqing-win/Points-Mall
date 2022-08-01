package service

import (
	"Points_Mall/config"
	"Points_Mall/dao"
	"Points_Mall/middlerware/redis"
	"log"
	"strconv"
	"time"
)

type TaskServiceImpl struct {
	ScoreService
}

func (t *TaskServiceImpl) FinishTask(userId, taskId int64) bool {
	//将记录加入到 UserTask 表中
	//先获取任务的分数

	task := dao.NewTaskDao()
	scoreSolo, err := task.GetScore(taskId)
	socrel, _ := strconv.ParseInt(scoreSolo, 10, 64)
	if err != nil {
		log.Println("获取单个任务分数失败", err.Error())
	}
	var userTask dao.UserTask
	userTask.TaskId = taskId
	userTask.UserId = userId
	err = task.InsetFinishTask(userTask)
	if err != nil {
		log.Println("更新UserTask表 err", err.Error())
		return false
	}

	//更新用户表中的分数
	score, err := t.GetScore(userId)
	if err != nil {
		log.Println("获取用户总分数失败", err.Error())
		return false
	}
	scoreTotal := score + socrel
	scoreStr := strconv.Itoa(int(scoreTotal))
	err = dao.UpdateTableUserScore(userId, scoreStr)
	if err != nil {
		log.Println("更改用户总分数失败", err.Error())
		return false
	}
	return true

}

func (t *TaskServiceImpl) AddTask(taskTable dao.TaskTable) bool {
	task := dao.NewTaskDao()
	//先加入数据库
	if err := task.AddTask(taskTable); err != nil {
		log.Println("add task err")
		return false
	}
	//加入缓存
	taskIdStr := strconv.Itoa(int(taskTable.Id))
	//过期时间设置为10分钟
	redis.RdbTScore.Set(redis.Ctx, taskIdStr, taskTable.Score, time.Duration(config.OneMinute)*time.Second*10)
	log.Println("add task Suc")
	return true
}

// AddTaskLimit 加入数据库,一条时限一天
func (t *TaskServiceImpl) AddTaskLimit(taskTable dao.TaskTable) bool {
	task := dao.NewTaskDao()
	if err := task.AddTask(taskTable); err != nil {
		log.Println("add task err")
		return false
	}
	log.Println("add task Suc")
	timer := time.NewTimer(24 * time.Hour)
	<-timer.C
	taskLimit, err := task.GetTaskByName(taskTable.Title)
	if err != nil {
		log.Println("get task err")
		return false
	}
	log.Println("get task Suc")
	err = task.DeleteTask(taskLimit.Id)
	if err != nil {
		log.Println("delete limit task err")
		return false
	}
	log.Println("delete limit task Suc")
	return true

}

func (t *TaskServiceImpl) GetTask(taskName string) dao.TaskTable {
	task := dao.NewTaskDao()
	taskTable, err := task.GetTaskByName(taskName)
	if err != nil {
		log.Println("get task err")
		return taskTable
	}
	log.Println("get task Suc")
	return taskTable
}

func (t *TaskServiceImpl) GetTaskById(id int64) dao.TaskTable {
	task := dao.NewTaskDao()
	taskTable, err := task.GetTaskById(id)
	if err != nil {
		log.Println("get task err")
		return taskTable
	}
	log.Println("get task Suc")
	return taskTable
}

func (t *TaskServiceImpl) DeleteTask(id int64) bool {
	task := dao.NewTaskDao()
	err := task.DeleteTask(id)
	if err != nil {
		log.Println("delete limit task err")
		return false
	}
	log.Println("delete limit task Suc")
	return true
}

func (t *TaskServiceImpl) GetTaskScore(taskId int64) string {
	task := dao.NewTaskDao()
	score, err := task.GetScore(taskId)
	if err != nil {
		log.Println("get task score err")
		return ""
	}
	return score
}
