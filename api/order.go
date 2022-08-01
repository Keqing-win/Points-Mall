package api

import (
	"Points_Mall/service"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

type OrderResp struct {
	Response
	order  service.Order
	orders []service.Order
}

// ProduceOrder POST
func ProduceOrder(c *gin.Context) {
	strUserId := c.Query("userId")
	userId, err := strconv.ParseInt(strUserId, 10, 64)

	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusOK, Response{StatusCode: 1})
	} else {
		strGoodId := c.Query("goodId")
		goodId, _ := strconv.ParseInt(strGoodId, 10, 64)
		var o service.OrderServiceImpl
		order, flag := o.ProduceOrder(userId, goodId)
		if !flag {
			c.JSON(http.StatusOK, OrderResp{
				Response: Response{StatusCode: 1, StatusMsg: "生成订单失败"},
			})
		} else {
			c.JSON(http.StatusOK, OrderResp{
				Response: Response{StatusCode: 0, StatusMsg: "生成订单成功"},
				order:    order,
			})
		}

	}

}

// GetOrderInfo 返回订单详细
func GetOrderInfo(c *gin.Context) {
	userId, err := strconv.ParseInt(c.GetString("userId"), 10, 64)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusOK, Response{StatusCode: 1})
	} else {
		strOrderId := c.Query("orderId")
		orderId, _ := strconv.ParseInt(strOrderId, 10, 64)
		var o service.OrderServiceImpl
		orders, flag := o.GetOrdersInfo(userId, orderId)
		if !flag {
			c.JSON(http.StatusOK, OrderResp{
				Response: Response{StatusCode: 1, StatusMsg: "获取订单信息失败"},
			})
		} else {
			c.JSON(http.StatusOK, OrderResp{
				Response: Response{StatusCode: 1, StatusMsg: "获取订单信息成功"},
				orders:   orders,
			})
		}

	}

}
