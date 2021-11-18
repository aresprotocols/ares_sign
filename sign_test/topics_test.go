package main

import (
	"ares/sign/routers/api"
	"ares/sign/wallet"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"gopkg.in/gomail.v2"
	"io/ioutil"
	"math/big"
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
	urlStr := "http://167.179.113.219:9090/api/bridge/crossBsc"
	data := make(url.Values)
	data["tx_hash"] = []string{"0x66262daf297d94d55ac35c83c40dcbc545cbffd8f0ebd4b5fea6dc4eaae78871"}
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

// {"code":0,"data":{"balance":"82112000000000000000000"},"msg":"Get bsc balance success"}
func TestGet(t *testing.T) {
	//urlStr := "http://127.0.0.1:9090/api/bridge/getBscBalance"
	urlStr := "http://167.179.113.219:9090/api/bridge/getBscBalance"
	resp, err := http.Get(urlStr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	fmt.Println(resp.StatusCode)
}

func TestExecErc20(t *testing.T) {

	client, err := ethclient.Dial("wss://mainnet.infura.io/ws/v3/f0001dbfb6c943a09468471b59a01510")
	if err != nil {
		fmt.Println(err)
	}

	// 0x Protocol (ZRX) token address
	contractAddress := common.HexToAddress("0x358AA737e033F34df7c54306960a38d09AaBd523")

	var fromRule []interface{}

	var toRule []interface{}
	toRule = append(toRule, common.HexToAddress("0xbcaf727812a103a7350554b814afa940b9f8b87d"))
	logTransferSig := []byte("Transfer(address,address,uint256)")
	logTransferSigHash := crypto.Keccak256Hash(logTransferSig)

	// Append the event selector to the query parameters and construct the topic set
	query1 := append([][]interface{}{{logTransferSigHash}}, fromRule)
	query1 = append(query1, toRule)

	topics, err := makeTopics(query1...)
	if err != nil {
		fmt.Println("makeTopics", err)
	}
	// Start the background filtering
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(13553806),
		Addresses: []common.Address{
			contractAddress,
		},
		Topics: topics,
	}

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		fmt.Println(err)
	}

	contractAbi, err := abi.JSON(strings.NewReader(string(ERC20ABI)))
	if err != nil {
		fmt.Println(err)
	}

	swapAccount := wallet.LoadSwapJSON("swap")
	if swapAccount == nil {
		swapAccount = make(map[string]*wallet.LogTransfer)
	}

	for _, vLog := range logs {
		fmt.Printf("Log Block Number: %d\n", vLog.BlockNumber)
		fmt.Printf("Log Index: %d\n", vLog.Index)

		switch vLog.Topics[0].Hex() {
		case logTransferSigHash.Hex():

			if _, ok := swapAccount[vLog.TxHash.String()]; ok {
				continue
			}

			var transferEvent wallet.LogTransfer
			fmt.Println("tx ", vLog.TxHash.String())
			err = contractAbi.UnpackIntoInterface(&transferEvent, "Transfer", vLog.Data)
			if err != nil {
				fmt.Println(err)
			}

			transferEvent.From = common.HexToAddress(vLog.Topics[1].Hex())
			transferEvent.To = common.HexToAddress(vLog.Topics[2].Hex())
			transferEvent.ValueS = wallet.ToEth(transferEvent.Value).String()
			transferEvent.Height = vLog.BlockNumber

			fmt.Printf("From: %s\n", transferEvent.From.Hex())
			fmt.Printf("To: %s\n", transferEvent.To.Hex())
			fmt.Printf("Tokens: %s\n", wallet.ToEth(transferEvent.Value).String())

			urlStr := "http://167.179.113.219:9090/api/bridge/crossBsc"
			data := make(url.Values)
			data["tx_hash"] = []string{vLog.TxHash.String()}
			resp, err := http.PostForm(urlStr, data)

			if err != nil {
				fmt.Println(err)
				return
			}
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Println(string(body))
			stu := Data{}
			err = json.Unmarshal(body, &stu)

			if err != nil {
				fmt.Println(err)
			}
			if stu.Data.TxHash != "" {
				fmt.Println(" ", stu.Data.TxHash)
				transferEvent.BscTx = common.HexToHash(stu.Data.TxHash)
				swapAccount[vLog.TxHash.String()] = &transferEvent
			}
		}
	}
	wallet.WriteSwapJSON("swap", swapAccount)
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

func TestFormatSwap(t *testing.T) {

	client, err := ethclient.Dial("wss://mainnet.infura.io/ws/v3/f0001dbfb6c943a09468471b59a01510")

	if err != nil {
		fmt.Println("err", err)
	}

	swapAccount := wallet.LoadSwapJSON("swap")

	for hash, swap := range swapAccount {
		bscHash := hash
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
			swap.Height = receipt.BlockNumber.Uint64()
		}
	}
	wallet.WriteSwapJSON("swap22", swapAccount)

}

func TestUnMarshal(t *testing.T) {
	jsonStr := "{\"code\":0,\"data\":{\"transaction_hash\":\"0x6aad612f2837adf639fd454125d29e2a724cdec69aa36d0f6be74fada3444ade\"},\"msg\":\"Send bsc transaction success\"}\n"
	stu := Data{}
	err := json.Unmarshal([]byte(jsonStr), &stu)

	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(" ", stu.Data.TxHash)
}

func TestUnPack(t *testing.T) {
	str := "000000000000000000000000000000000000000000001168b8b0f627ee100000"
	contractAbi, err := abi.JSON(strings.NewReader(string(ERC20ABI)))

	data, _ := hex.DecodeString(str)

	var number *big.Int
	err = contractAbi.UnpackIntoInterface(&number, "balanceOf", data)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("printBalance erc20", wallet.ToEth(number))
}

func TestSendEmail(t *testing.T) {
	m := gomail.NewMessage()

	//发送人
	m.SetHeader("From", "1032087738@qq.com")
	//接收人
	m.SetHeader("To", "450595468@qq.com")
	//抄送人
	//m.SetAddressHeader("Cc", "xxx@qq.com", "xiaozhujiao")
	//主题
	m.SetHeader("Subject", "账户余额不足")
	//内容
	s1 := "<h1>0xa2990ec3024fd0c8afec70f40a1b51beb9853e641b1b153615a9b01c0d1bc8ae</h1>"
	s2 := "<h1>500000</h1>"
	m.SetBody("text/html", s1+s2)
	//附件
	//m.Attach("./myIpPic.png")

	//拿到token，并进行连接,第4个参数是填授权码
	d := gomail.NewDialer("smtp.qq.com", 587, "1032087738@qq.com", "fvsofxkcmnaqbaja")

	// 发送邮件
	if err := d.DialAndSend(m); err != nil {
		fmt.Printf("DialAndSend err %v:", err)
		panic(err)
	}
	fmt.Printf("send mail success\n")
}

func TestUnpack(t *testing.T) {
	str := "a9059cbb000000000000000000000000bcaf727812a103a7350554b814afa940b9f8b87d00000000000000000000000000000000000000000000d3c21bcecceda1000000"

	contractAbi, err := abi.JSON(strings.NewReader(string(ERC20ABI)))
	data, _ := hex.DecodeString(str)

	transfer := struct {
		From  common.Address
		Value *big.Int
	}{}
	err = contractAbi.UnpackIntoInterface(&transfer, "transfer", data)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("printBalance erc20", wallet.ToEth(transfer.Value), " ", transfer.From)
}
