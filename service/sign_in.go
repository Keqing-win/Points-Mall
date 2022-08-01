package service

type SignInService interface {

	// ConsecutiveCheckIns 连续签到
	ConsecutiveCheckIns(userId int64) (int64, bool)
	// BreakCheckIns 断签后签到
	BreakCheckIns(userId int64) bool
	// GetRank 获取排名
	GetRank() ([]User, bool)
}
