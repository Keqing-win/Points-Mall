package main

import (
	"Points_Mall/api"
	"Points_Mall/middlerware/jwt"
	"github.com/gin-gonic/gin"
)

func initRouter(r *gin.Engine) {
	apiRoute := r.Group("/pointsmall")
	//后台api
	apiRoute.PUT("/score/add", api.AddScore)
	apiRoute.POST("/good/add", api.CreateGood)
	apiRoute.PUT("/good/update", api.UpdateGood)
	apiRoute.DELETE("/good/delete", api.DeleteGood)
	apiRoute.POST("/task/add", api.AddTask)
	apiRoute.DELETE("/task/delete", api.DeleteTask)
	apiRoute.POST("/order/produce", api.ProduceOrder)

	//客户端接口
	apiRoute.POST("/user/register", api.Register)
	apiRoute.POST("/user/login", api.Login)
	apiRoute.GET("/user/", jwt.Auth(), api.GetUserInfo)
	apiRoute.GET("/good/", jwt.Auth(), api.SelectGood)
	apiRoute.GET("/order/", jwt.Auth(), api.GetOrderInfo)
	apiRoute.POST("/signin/", jwt.Auth(), api.SingIn)
	apiRoute.POST("/good/buy", jwt.Auth(), api.ExchangeGood)
	apiRoute.GET("/score/", jwt.Auth(), api.GetScoreRecord)
	apiRoute.POST("/task/finish", jwt.Auth(), api.FinishTask)
	apiRoute.GET("/signin/rank", jwt.Auth(), api.Rank)

}
