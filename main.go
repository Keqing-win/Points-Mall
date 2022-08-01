package main

import (
	"Points_Mall/dao"
	"Points_Mall/middlerware/rabbitmq"
	"Points_Mall/middlerware/redis"
	"github.com/gin-gonic/gin"
)

func main() {

	initDeps()
	//gin
	r := gin.Default()
	initRouter(r)

	r.Run(":8080") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

// 加载项目依赖
func initDeps() {
	dao.Init()
	redis.InitRedis()
	rabbitmq.InitRabbitMQ()
	rabbitmq.InitOrderRabbitMQ()

}
