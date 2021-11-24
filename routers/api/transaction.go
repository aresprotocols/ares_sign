package api

import (
	"ares/sign/wallet"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"math/big"
)

func SendCrossTransaction(c *gin.Context) {
	param := make(map[string]string)
	err := c.ShouldBind(&param)

	txHash := common.HexToHash(param["tx_hash"])
	transHash, err := wallet.SendBscTransaction(txHash, c)
	if err != nil {
		txError := make(wallet.KeyAccount)
		txError[txHash.String()] = uint64(len(wallet.LoadNodesJSON("tx_error")))
		wallet.WriteNodesJSON("tx_error", txError)
	}

	data := make(map[string]string)
	if err != nil {
		fmt.Println("SendCrossTransaction", err)
		data["error"] = err.Error()
		SuccessResponse(c, 0, "Cross bsc tx error", data)
	} else {
		data["transaction_hash"] = transHash
		SuccessResponse(c, 0, "Cross bsc tx success", data)
	}

}

func GetBscBalance(c *gin.Context) {

	// 送出交易查詢
	response, err := wallet.GetBscBalance()

	data := make(map[string]string)
	if err != nil {
		data["error"] = err.Error()
		SuccessResponse(c, 0, "Get bsc balance error", data)
	} else {
		value := new(big.Int).Add(wallet.EthToWei(200), response)
		data["balance"] = value.String()
		SuccessResponse(c, 0, "Get bsc balance success", data)
	}
}

func GetBscFee(c *gin.Context) {
	// 送出交易查詢
	data := make(map[string]string)
	data["fee"] = "10"
	SuccessResponse(c, 0, "Get bsc fee success", data)
}
