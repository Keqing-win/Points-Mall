package dao

import (
	"log"
	"sync"
)

// ScoreTable Score 分数表
type ScoreTable struct {
	Id     int64
	UserId int64 //所属的用户的id
	TaskId int64
	Score  string
}

type ScoreDao struct {
}

var (
	scoreDao  *ScoreDao
	scoreOnce sync.Once
)

func NewScoreDao() *ScoreDao {
	scoreOnce.Do(func() {
		scoreDao = &ScoreDao{}
	})
	return scoreDao
}

func (*ScoreDao) InsertScoreTable(userId, taskId int64, score string) error {
	scoreTable := ScoreTable{}
	scoreTable.UserId = userId
	scoreTable.TaskId = taskId
	scoreTable.Score = score
	if err := Db.Create(&scoreTable).Error; err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func (*ScoreDao) GetScoreTables(userId int64) ([]ScoreTable, error) {
	var scoreTables []ScoreTable
	if err := Db.Where("user_id", userId).Find(&scoreTables).Error; err != nil {
		log.Println(err.Error())
		return scoreTables, err
	}
	return scoreTables, nil

}
