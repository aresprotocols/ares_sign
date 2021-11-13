package api

import (
	"ares/sign/wallet"
	"fmt"
	"github.com/ethereum/go-ethereum/common"

	"github.com/gin-gonic/gin"
)

func SendCrossTransaction(c *gin.Context) {
	param := make(map[string]string)
	err := c.ShouldBind(&param)

	txHash := common.HexToHash(param["tx_hash"])

	transHash, err := wallet.SendBscTransaction(txHash)
	if err != nil {
		txError := make(wallet.KeyAccount)
		txError[txHash.String()] = uint64(len(wallet.LoadNodesJSON("tx_error")))
		wallet.WriteNodesJSON("tx_error", txError)
	}

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
