package service

type ScoreService interface {
	// AddScore 添加分数
	AddScore(userId, taskId int64) bool

	// GetScore 通过任务id，用户id(token获取),获取分数，提供给user服务
	GetScore(userId int64) (int64, error)

	// GetScoreRecord 获取分数记录
	GetScoreRecord(userId int64) ([]Score, bool)
}
