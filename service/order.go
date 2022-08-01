package service

type OrderService interface {
	// ProduceOrder 生成订单
	ProduceOrder(userId, goodId int64) (Order, bool)

	// GetOrdersInfo 获取订单的信息
	GetOrdersInfo(userId, orderId int64) ([]Order, bool)
}
