package service

import "Points_Mall/dao"

type TaskService interface {
	/*
		上层api使用
	*/

	// AddTask 添加任务
	AddTask(table dao.TaskTable) bool

	// GetTaskById GetTask 获取任务
	GetTaskById(id int64) dao.TaskTable

	//DeleteTask 删除任务
	DeleteTask(id int64) bool

	// AddTaskLimit 删除时限到的任务
	AddTaskLimit(taskTable dao.TaskTable) bool

	// FinishTask 完成任务
	FinishTask(taskId, userId int64) bool

	/*
		其他服务调用
	*/

	// GetTask 通过任务名称获取这个任务
	GetTask(taskName string) dao.TaskTable

	// GetTaskScore 通过任务的id获取分数
	GetTaskScore(taskId int64) string
}
