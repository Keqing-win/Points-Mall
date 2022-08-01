package service

import (
	"Points_Mall/dao"
	"log"
	"sync"
)

// Order 封装后返回给api
type Order struct {
	dao.OrderTable
	UserName  string `json:"user_name"`
	ScoreNeed int64  `json:"score_need"` //需要的积分
	GoodName  string `json:"good_name"`  //商品的名字
	Number    int32  `json:"number"`     //商品的数目
	CoverUrl  string `json:"cover_url"`  //图片的url
}

type OrderServiceImpl struct {
	UserService
	GoodService
}

func (o *OrderServiceImpl) ProduceOrder(userId, goodId int64) (Order, bool) {
	//根据商品服务和用户服务获得订单的详细信息
	var order Order
	orderDao := dao.NewOrderDao()
	data, err := orderDao.GetOrderTable(goodId, userId)
	if err != nil {
		log.Println("获取订单信息失败")
		return order, false
	}
	o.createOrderByGoodId(&order, &data, userId, goodId)
	return order, true

}

func (o *OrderServiceImpl) GetOrdersInfo(userId, orderId int64) ([]Order, bool) {
	var orders []Order
	orderDao := dao.NewOrderDao()
	datas, err := orderDao.GetOrderTablesByOrderId(orderId, userId)

	if err != nil {
		log.Println("获取订单信息失败")
		return orders, false
	}
	goodId, err := orderDao.GetGoodId(orderId, userId)
	if err != nil {
		log.Println("获取商品id失败")
		return orders, false
	}
	_ = o.copyVideos(&orders, &datas, userId, goodId)
	return orders, true

}

func (o *OrderServiceImpl) copyVideos(result *[]Order, data *[]dao.OrderTable, userId, goodId int64) error {
	for _, temp := range *data {
		var order Order
		//将video进行组装，添加想要的信息,插入从数据库中查到的数据
		o.createOrderByGoodId(&order, &temp, userId, goodId)
		*result = append(*result, order)
	}
	return nil
}

func (o *OrderServiceImpl) createOrderByGoodId(order *Order, data *dao.OrderTable, userId, goodId int64) {
	//建立携程组
	var wg sync.WaitGroup
	wg.Add(2)

	order.OrderTable = *data
	go func() {
		user := o.GetTableUserById(userId)
		order.UserId = userId
		order.UserName = user.UserName
		wg.Done()
	}()
	go func() {
		good, flag := o.SearchGood(goodId)
		if !flag {
			log.Println("SearchGood err")
			order.GoodId = good.Id
			order.CoverUrl = good.CoverUrl
			order.GoodName = good.GoodName
			order.ScoreNeed = good.ScoreNeed
			wg.Done()
		}
	}()
	wg.Wait()
}
