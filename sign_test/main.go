package main

import (
	"ares/sign/wallet"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const ERC20ABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

func main() {
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

	contractAbi, err := abi.JSON(strings.NewReader(string(ERC20ABI)))
	if err != nil {
		fmt.Println(err)
	}

	swapAccount := wallet.LoadSwapJSON("swap")
	if swapAccount == nil {
		swapAccount = make(map[string]*wallet.LogTransfer)
	}

	for {
		logs, err := client.FilterLogs(context.Background(), query)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("FilterLogs", len(logs))
		LoopQueryCross(logs, contractAbi, logTransferSigHash, swapAccount)
		wallet.WriteSwapJSON("swap", swapAccount)

		time.Sleep(time.Minute * 10)
	}
}

func LoopQueryCross(logs []types.Log, contractAbi abi.ABI, signHash common.Hash, swapAccount wallet.SwapAccount) {
	for _, vLog := range logs {

		switch vLog.Topics[0].Hex() {
		case signHash.Hex():

			if _, ok := swapAccount[vLog.TxHash.String()]; ok {
				continue
			}

			fmt.Printf("Log Block Number: %d\n", vLog.BlockNumber)
			fmt.Printf("Log Index: %d\n", vLog.Index)

			var transferEvent wallet.LogTransfer
			fmt.Println("tx ", vLog.TxHash.String())
			err := contractAbi.UnpackIntoInterface(&transferEvent, "Transfer", vLog.Data)
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
}

type Resp struct {
	TxHash string `json:"transaction_hash"`
	Msg    string `json:"msg"`
}

type Data struct {
	Data Resp `json:"data"`
}
