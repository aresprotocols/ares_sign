package api

import (
	"ares/sign/wallet"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"math/big"
	"net/http"
	"strconv"
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
		value := new(big.Int).Add(wallet.EthToWei(4000000), response)
		data["balance"] = value.String()
		SuccessResponse(c, 0, "Get bsc balance success", data)
	}
}

func GetBscFee(c *gin.Context) {
	// 送出交易查詢
	data := make(map[string]string)
	data["fee"] = strconv.Itoa(int(wallet.GetBridgeFee()))
	SuccessResponse(c, 0, "Get bsc fee success", data)
}

func SetBscFee(c *gin.Context) {
	fee, exist := c.GetQuery("fee")
	if !exist {
		c.JSON(http.StatusBadRequest, "bad request")
		return
	}
	value, _ := strconv.Atoi(fee)
	wallet.SetBridgeFee(uint32(value))
	SuccessResponse(c, 0, "Set bsc fee success", value)
}
