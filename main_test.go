package main

import (
	"ares/sign/routers/api"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestPost(t *testing.T) {
	urlStr := "http://127.0.0.1:9090/api/bridge/crossBsc"
	teamworkinfo := api.Transaction{
		From:  "111",
		To:    "333",
		Value: 100,
	}
	jsons, _ := json.Marshal(teamworkinfo)
	result := string(jsons)
	jsoninfo := strings.NewReader(result)
	req, _ := http.NewRequest("POST", urlStr, jsoninfo)
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	//stu:=Result{}
	//err =json.Unmarshal(body,&stu)
	//
	//if err!=nil{
	//	fmt.Println(err)
	//}
}

func TestPostForm(t *testing.T) {
	urlStr := "http://127.0.0.1:9090/api/bridge/crossBsc"
	data := make(url.Values)
	data["tx_hash"] = []string{"0x6aad612f2837adf639fd454125d29e2a724cdec69aa36d0f6be74fada3444ade"}
	resp, err := http.PostForm(urlStr, data)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	//stu:=Result{}
	//err =json.Unmarshal(body,&stu)
	//
	//if err!=nil{
	//	fmt.Println(err)
	//}
}

func TestReadErc20(t *testing.T) {

	client, err := ethclient.Dial("https://bsc-dataseed.binance.org")
	if err != nil {
		fmt.Println("err", err)
	}

	bscHash := "0xa2990ec3024fd0c8afec70f40a1b51beb9853e641b1b153615a9b01c0d1bc8ae"
	// 0x Protocol (ZRX) token address
	count := 0
	for {
		time.Sleep(time.Millisecond * 200)
		_, isPending, err := client.TransactionByHash(context.Background(), common.HexToHash(bscHash))
		if err != nil {
			fmt.Println("err", err)
		}
		count++
		if !isPending {
			break
		}
		if count >= 40 {
			fmt.Println("Please use querytx sub command query later.")
		}
	}
	receipt, err := client.TransactionReceipt(context.Background(), common.HexToHash(bscHash))
	if err != nil {
		fmt.Println("err", err)
	}

	if receipt.Status == types.ReceiptStatusSuccessful {
		block, err := client.BlockByHash(context.Background(), receipt.BlockHash)
		if err != nil {
			fmt.Println("err", err)
		}

		fmt.Println("Bsc transaction Success", " block Number", receipt.BlockNumber.Uint64(), " block txs", len(block.Transactions()), "blockhash", block.Hash().Hex())
	}
}
