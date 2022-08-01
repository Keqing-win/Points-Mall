package api

import (
	"Points_Mall/config"
	"Points_Mall/dao"
	"Points_Mall/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type GoodResp struct {
	Response
	ScoreNeed int64  //需要的积分
	GoodName  string //商品的名字
	Number    int32  //商品的数目
	CoverUrl  string //图片的url
}

type Good struct {
	ScoreNeed int64  `form:"score_need" json:"score_need" binding:"required" ` //需要的积分
	GoodName  string `form:"good_name" json:"good_name" binding:"required" `   //商品的名字
	Number    int32  `form:"number" json:"number" binding:"required"`          //商品的数目
	CoverUrl  string `form:"cover_url" json:"cover_url" binding:"required"`    //图片的url
}

// ExchangeGood 交换商品时如果成功
func ExchangeGood(c *gin.Context) {
	userId, err := strconv.ParseInt(c.GetString("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1})
	}
	actionType, _ := strconv.ParseInt(c.Query("actionType"), 10, 64)
	goodId, _ := strconv.ParseInt(c.Query("goodId"), 10, 64)
	g := service.GoodServiceImpl{}
	if actionType == config.ValidGood {
		//先判断用户积分是否够
		flag, err := g.CheckUserBalance(userId, goodId)
		if flag && err != nil {
			c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "用户积分不足"})
		}
		if !flag && err != nil {
			c.JSON(http.StatusOK, Response{StatusCode: 1})
		}
		//客户端生产订单后，查询订单id
		orderId, _ := strconv.ParseInt(c.Query("orderId"), 10, 64)

		//然后下订单
		flag = g.OrderAction(userId, goodId, orderId, config.ValidOrder)
		if !flag {
			c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "完成订单失败"})
		}
		c.JSON(http.StatusOK, Response{StatusCode: 0, StatusMsg: "已经下单"})
	} else {
		//取消订单
		orderId, _ := strconv.ParseInt(c.Query("orderId"), 10, 64)
		flag := g.OrderAction(userId, goodId, orderId, config.InvalidOrder)
		if !flag {
			c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "取消订单失败"})
		}
		c.JSON(http.StatusOK, Response{StatusCode: 0, StatusMsg: "已经取消订单"})
	}

}

func CreateGood(c *gin.Context) {
	var good Good
	if err := c.ShouldBind(&good); err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "数据处理错误"})
	}
	g := service.GoodServiceImpl{}
	var goodTable dao.GoodTable
	goodTable.GoodName = good.GoodName
	goodTable.CoverUrl = good.CoverUrl
	goodTable.Number = good.Number
	goodTable.ScoreNeed = good.ScoreNeed
	flag := g.CreateGood(&goodTable)
	if !flag {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "创建商品失败"})
	} else {
		c.JSON(http.StatusOK, Response{StatusCode: 0, StatusMsg: "创建商品成功"})
	}

}

func UpdateGood(c *gin.Context) {
	var good Good
	goodId, _ := strconv.ParseInt(c.Query("goodId"), 10, 64)
	if err := c.ShouldBind(&good); err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "数据处理错误"})
	} else {
		g := service.GoodServiceImpl{}
		var goodTable dao.GoodTable
		goodTable.GoodName = good.GoodName
		goodTable.CoverUrl = good.CoverUrl
		goodTable.Number = good.Number
		goodTable.ScoreNeed = good.ScoreNeed
		flag := g.ChangeGood(goodId, &goodTable)
		if !flag {
			c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "修改商品失败"})
		} else {
			c.JSON(http.StatusOK, Response{StatusCode: 0, StatusMsg: "修改商品成功"})
		}

	}

}

func DeleteGood(c *gin.Context) {
	goodIdStr := c.Query("goodId")
	goodId, err := strconv.ParseInt(goodIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1})
	} else {
		g := service.GoodServiceImpl{}
		flag := g.DeleteGood(goodId)
		if !flag {
			c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "删除商品失败"})
		} else {
			c.JSON(http.StatusOK, Response{StatusCode: 0, StatusMsg: "删除商品成功"})
		}
	}

}

func SelectGood(c *gin.Context) {
	goodIdStr := c.Query("good_id")
	goodId, err := strconv.ParseInt(goodIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1})
	} else {
		g := service.GoodServiceImpl{}
		goodTable, flag := g.SearchGood(goodId)
		if !flag {
			c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "查询商品失败"})
		} else {
			c.JSON(http.StatusOK, GoodResp{
				Response:  Response{StatusCode: 0, StatusMsg: "查询商品成功"},
				ScoreNeed: goodTable.ScoreNeed,
				GoodName:  goodTable.GoodName,
				Number:    goodTable.Number,
				CoverUrl:  goodTable.CoverUrl,
			})
		}

	}

}
