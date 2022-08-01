package dao

import (
	"log"
	"sync"
)

type TaskTable struct {
	Id int64
	//Integral    int32     //一个任务所含的积分
	PublishTime int64  //发表任务的时间戳
	Title       string //任务的标题
	Content     string //任务的内容
	Score       string //任务的分数
	Only        int32  //0表示一次，1表示每日任务
}

type UserTask struct {
	TaskId int64
	//Cancel int32 //0表示取消，1表示完成任务
	UserId int64
}

func (TaskTable) TableName() string {
	return "tasks"
}

type TaskDao struct {
}

//每次只返回一个实例
var (
	taskDao  *TaskDao
	taskOnce sync.Once
)

func NewTaskDao() *TaskDao {

	taskOnce.Do(func() {
		taskDao = &TaskDao{}
	})
	return taskDao
}

func (*TaskDao) InsetFinishTask(userTask UserTask) error {
	if err := Db.Model(userTask).Create(&userTask).Error; err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func (*TaskDao) GetScore(taskId int64) (string, error) {
	var taskTable TaskTable
	if err := Db.Select("score").Where("id", taskId).First(&taskTable).Error; err != nil {
		log.Println(err.Error())
		return "", err
	}
	return taskTable.Score, nil
}

func (*TaskDao) AddTask(task TaskTable) error {
	if err := Db.Create(&task).Error; err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func (*TaskDao) DeleteTask(id int64) error {
	if err := Db.Where("id = ?", id).Delete(&TaskTable{}).Error; err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func (*TaskDao) GetTaskByName(title string) (TaskTable, error) {
	var taskTable TaskTable
	if err := Db.Where("title = ?", title).First(&taskTable).Error; err != nil {
		log.Println(err.Error())
		return taskTable, err
	}
	return taskTable, nil
}

func (*TaskDao) GetTaskById(id int64) (TaskTable, error) {
	var taskTable TaskTable
	if err := Db.Where("id = ?", id).First(&taskTable).Error; err != nil {
		log.Println(err.Error())
		return taskTable, err
	}
	return taskTable, nil
}
