package api

import (
	"ares/sign/routers/jsonrpc"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

/*
Go-Ethereum JSON-RPC API：https://github.com/ethereum/go-ethereum/wiki/Management-APIs#admin_nodeinfo
*/
func GetNodeInfo(context *gin.Context) {
	// 送出查詢請求
	response := jsonrpc.ClientRequest("admin_nodeInfo", nil)
	fmt.Println("response", response)
	// 回傳執行結果
	context.JSON(http.StatusOK, gin.H{
		"enode": response.Get("result").Get("enode").MustString(),
		"name":  response.Get("result").Get("name").MustString(),
	})
}

/*
Go-Ethereum JSON-RPC API：https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_getblockbynumber
*/
func GetBlockInfo(context *gin.Context) {
	blockNumberS := context.DefaultQuery("blockNumber", "10")
	if blockNumberS == "" {
		InternalError(context, -1, "Param parse error")
		return
	}

	blockNumber, _ := strconv.Atoi(blockNumberS)
	params := []interface{}{fmt.Sprintf("0x%x", blockNumber), true}

	// 送出查詢請求
	response := jsonrpc.ClientRequest("eth_getBlockByNumber", params)
	fmt.Println("response", response)

	// 若查無該區塊, 則回傳空陣列
	if len(response.Get("result").Get("hash").MustString()) == 0 {
		context.JSON(http.StatusOK, gin.H{
			"message": "block not found",
			"data":    []string{},
		})

		return
	}

	// 回傳執行結果
	context.JSON(http.StatusOK, gin.H{
		"difficulty":      response.Get("result").Get("difficulty").MustString(),
		"gasLimit":        response.Get("result").Get("gasLimit").MustString(),
		"gasUsed":         response.Get("result").Get("gasUsed").MustString(),
		"hash":            response.Get("result").Get("hash").MustString(),
		"miner":           response.Get("result").Get("miner").MustString(),
		"parentHash":      response.Get("result").Get("parentHash").MustString(),
		"totalDifficulty": response.Get("result").Get("totalDifficulty").MustString(),
	})
}

// Transaction struct - 用以儲存要交易的設定
type Transaction struct {
	From  string `json:"from"`  // 轉出帳號
	To    string `json:"to"`    // 轉入帳號
	Value int    `json:"value"` // 交易金額
}

/*
Go-Ethereum JSON-RPC API：https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_gettransactionbyhash
*/
func GetTrxInfo(context *gin.Context) {
	// 準備要查詢的交易編碼
	trxHash := []interface{}{context.Param("trxHash")}

	// 送出交易查詢
	response := jsonrpc.ClientRequest("eth_getTransactionByHash", trxHash)

	// 回傳執行結果
	context.JSON(http.StatusOK, gin.H{
		"blockHash":   response.Get("result").Get("blockHash").MustString(),
		"blockNumber": response.Get("result").Get("blockNumber").MustString(),
		"from":        response.Get("result").Get("from").MustString(),
		"gas":         response.Get("result").Get("gas").MustString(),
		"gasPrice":    response.Get("result").Get("gasPrice").MustString(),
		"hash":        response.Get("result").Get("hash").MustString(),
		"nonce":       response.Get("result").Get("nonce").MustString(),
		"to":          response.Get("result").Get("to").MustString(),
		"value":       response.Get("result").Get("value").MustString(),
	})
}

/*
Go-Ethereum JSON-RPC API：https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_sendtransaction
*/
func SendTrxWithLocal(context *gin.Context) {
	// 讀入要執行的交易設定, 透過 FORM POST 傳送上來的 RAW ( JSON )
	var trx Transaction
	context.ShouldBind(&trx)

	// 準備要查詢的交易編碼
	//trxDetail := map[string]interface{}{"from": trx.From, "to": trx.To, "value": trx.Value}
	//params := []map[string]interface{}{trxDetail}

	// 送出交易查詢
	//response := jsonrpc.ClientRequest("eth_sendTransaction", params)

	//url := "https://mainnet.infura.io/v3/9aa3d95b3bc440fa88ea12eaa4456161"

	fmt.Println("trx.From", trx.From, " trx.From ", trx.To, " value ", trx.Value)

	// 回傳執行結果
	//context.JSON(http.StatusCreated, gin.H{
	//	"trx": response.Get("result").MustString(),
	//}
	context.JSON(http.StatusOK, gin.H{
		"message": "block not found",
		"data":    []string{},
	})
}
