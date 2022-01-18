package routers

import (
	"ares/sign/routers/api"
	"ares/sign/routers/middleware"

	"github.com/gin-gonic/gin"
)

func InitRouters() *gin.Engine {
	r := gin.Default()
	r.Use(middleware.Throttle(1000, 20))
	r.Use(middleware.Cors())

	transaction := r.Group("/api/bridge")
	{
		transaction.POST("/crossBsc", api.SendCrossTransaction)
		transaction.GET("/getBscBalance", api.GetBscBalance)
		transaction.GET("/getBscFee", api.GetBscFee)
		transaction.GET("/setBscFee", api.SetBscFee)
		transaction.GET("/getEthFee", api.GetEthFee)
		transaction.GET("/setEthFee", api.SetEthFee)
		transaction.GET("/trxHash", api.GetTrxInfo)
	}

	system := r.Group("/api/system")
	{
		system.GET("/node", api.GetNodeInfo)
		system.GET("/blockNumber", api.GetBlockInfo)
	}

	return r
}
