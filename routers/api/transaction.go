package api

import (
	"ares/sign/wallet"
	"fmt"

	"github.com/gin-gonic/gin"
)

func SendCrossTransaction(c *gin.Context) {
	param := make(map[string]string)
	err := c.ShouldBind(&param)

	transHash, err := wallet.SendBscTransaction(param)

	data := make(map[string]string)
	if err != nil {
		fmt.Println("SendCrossTransaction", err)
		data["error"] = err.Error()
		SuccessResponse(c, 0, "Send bsc transaction error", data)
	} else {
		data["transaction_hash"] = transHash
		SuccessResponse(c, 0, "Send bsc transaction success", data)
	}

}
