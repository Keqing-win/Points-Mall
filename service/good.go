package service

import "Points_Mall/dao"

type GoodService interface {
	// CreateGood 创建商品
	CreateGood(table *dao.GoodTable) bool
	// ChangeGood 更新商品
	ChangeGood(goodId int64, table *dao.GoodTable) bool
	// DeleteGood 删除商品
	DeleteGood(goodId int64) bool
	// SearchGood 查询商品
	SearchGood(goodId int64) (dao.GoodTable, bool)
	// OrderAction 订单对商品进行操作
	OrderAction(userId, goodId, orderId int64, actionType int32) bool
	// IncreaseGoodNumber 增加商品的数目
	IncreaseGoodNumber(goodId int64, number int32) bool

	CheckUserBalance(userId, goodId int64) (bool, error)
}
