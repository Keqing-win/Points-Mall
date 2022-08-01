package dao

import (
	"log"
	"sync"
	"time"
)

type OrderTable struct {
	Id        int64
	UserId    int64
	GoodId    int64
	Cancel    int32     //0表示完成，1表示未完成
	OrderTime time.Time //订单的时间戳
}
type OrderDao struct {
}

var (
	orderDao  *OrderDao
	orderOnly sync.Once
)

func NewOrderDao() *OrderDao {
	orderOnly.Do(
		func() {
			orderDao = &OrderDao{}
		})
	return orderDao
}
func (*OrderDao) GetGoodId(orderId, userId int64) (int64, error) {
	var orderTable OrderTable
	if err := Db.Where(map[string]interface{}{"id": orderId, "user_id": userId}).First(&orderTable).Error; err != nil {
		log.Println(err.Error())
		return 0, err
	}
	return orderTable.GoodId, nil
}
func (*OrderDao) GetOrderTablesByOrderId(orderId, userId int64) ([]OrderTable, error) {
	var orderTables []OrderTable
	if err := Db.Where(map[string]interface{}{"id": orderId, "user_id": userId}).Find(&orderTables).Error; err != nil {
		log.Println(err.Error())
		return orderTables, err
	}
	return orderTables, nil
}

func (*OrderDao) GetOrderTables(goodId, userId int64) ([]OrderTable, error) {
	var orderTables []OrderTable
	if err := Db.Where(map[string]interface{}{"good_id": goodId, "user_id": userId}).Find(&orderTables).Error; err != nil {
		log.Println(err.Error())
		return orderTables, err
	}
	return orderTables, nil
}

func (*OrderDao) GetOrderTable(goodId, userId int64) (OrderTable, error) {
	var orderTable OrderTable
	if err := Db.Where(map[string]interface{}{"good_id": goodId, "user_id": userId}).First(&orderTable).Error; err != nil {
		log.Println(err.Error())
		return OrderTable{}, err
	}
	return orderTable, nil
}

func (*OrderDao) GetValueOrderIds(goodId int64) ([]int64, error) {
	var orderIds []int64
	if err := Db.Model(&OrderTable{}).Where("good_id = ?", goodId).Pluck("id", &orderIds).Error; err != nil {
		log.Println(err.Error())
		return orderIds, err
	}
	return orderIds, nil

}

func (*OrderDao) GetOrderInfoByUserId(userId, goodId int64) (OrderTable, error) {
	var orderTable OrderTable
	if err := Db.Where(map[string]interface{}{"user_id": userId, "good_id": goodId}).
		First(&orderTable).Error; err != nil {
		return OrderTable{}, err
	}
	return orderTable, nil
}

func (*OrderDao) GetOrderInfoByOrderId(orderId, goodId int64) (OrderTable, error) {
	var orderTable OrderTable
	if err := Db.Where(map[string]interface{}{"id": orderId, "good_id": goodId}).
		First(&orderTable).Error; err != nil {
		log.Println(err.Error())
		return OrderTable{}, err
	}
	return orderTable, nil
}

func (*OrderDao) InsertOrderTable(order OrderTable) error {
	if err := Db.Model(OrderTable{}).Create(&order).Error; err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}
func (*OrderDao) UpdateOrderStatus(userId, goodId int64, cancel int32) error {
	if err := Db.Model(OrderTable{}).Where(map[string]interface{}{"user_id": userId, "good_id": goodId}).
		Update("cancel", cancel).Error; err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}
