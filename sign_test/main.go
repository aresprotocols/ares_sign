package main

import (
	"ares/sign/wallet"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"gopkg.in/gomail.v2"
	"math/big"
	"os"
	"os/signal"
	"strings"
	"time"
)

const ERC20ABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

var height = uint64(0)

const Debug = true

func main() {

	// send

	blacklist := []common.Address{
		common.HexToAddress("0x65d19dbbcbf1d9126b1bfff07610ab21ec725ece"),
		common.HexToAddress("0x506332957899155ade8dd01f789bd14ef2ebdbb6"),
	}
	go LoopQueryCrossChainTx("wss://mainnet.infura.io/ws/v3/f0001dbfb6c943a09468471b59a01510",
		"0x358AA737e033F34df7c54306960a38d09AaBd523", "0xbcaf727812a103a7350554b814afa940b9f8b87d",
		"swap", blacklist, 5000)

	go LoopQueryCrossChainTx("wss://bsc-ws-node.nariox.org",
		"0xf9752a6e8a5e5f5e6eb3ab4e7d8492460fb319f0", "0xbcaf727812a103a7350554b814afa940b9f8b87d",
		"swapEth", blacklist, 5000)

	abortChan := make(chan os.Signal, 1)
	signal.Notify(abortChan, os.Interrupt)

	sig := <-abortChan

	fmt.Printf("Exiting... signal %v\n", sig)
}

func LoopQueryCrossChainTx(ws, contract, to, file string, blacklist []common.Address, step uint64) {
	client, err := ethclient.Dial(ws)
	if err != nil {
		fmt.Println(err)
	}

	// 0x Protocol (ZRX) token address
	contractAddress := common.HexToAddress(contract)

	var fromRule []interface{}

	var toRule []interface{}
	toRule = append(toRule, common.HexToAddress(to))
	logTransferSig := []byte("Transfer(address,address,uint256)")
	logTransferSigHash := crypto.Keccak256Hash(logTransferSig)

	// Append the event selector to the query parameters and construct the topic set
	query1 := append([][]interface{}{{logTransferSigHash}}, fromRule)
	query1 = append(query1, toRule)

	topics, err := makeTopics(query1...)
	if err != nil {
		fmt.Println("makeTopics", err)
	}
	height, err = client.BlockNumber(context.Background())
	if err != nil {
		fmt.Println("BlockNumber", err)
		return
	}
	// Start the background filtering
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(height),
		Addresses: []common.Address{
			contractAddress,
		},
		ToBlock: new(big.Int).SetUint64(height + step),
		Topics:  topics,
	}

	contractAbi, err := abi.JSON(strings.NewReader(string(ERC20ABI)))
	if err != nil {
		fmt.Println(err)
	}

	swapAccount := wallet.LoadSwapJSON(file)
	if swapAccount == nil {
		swapAccount = make(map[string]*wallet.LogTransfer)
	}

	for {
		logs, err := client.FilterLogs(context.Background(), query)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("FilterLogs", len(logs))
		LoopQueryCross(logs, contractAbi, logTransferSigHash, blacklist, file)

		if query.ToBlock.Uint64()+step < height {
			query.FromBlock = new(big.Int).Set(query.ToBlock)
		} else {
			query.FromBlock = new(big.Int).SetUint64(height)
		}
		query.ToBlock = new(big.Int).Add(query.FromBlock, new(big.Int).SetUint64(step))
		time.Sleep(time.Minute * 15)
	}
}

func LoopQueryCross(logs []types.Log, contractAbi abi.ABI, signHash common.Hash, blacklist []common.Address, file string) {
	for _, vLog := range logs {

		switch vLog.Topics[0].Hex() {
		case signHash.Hex():

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
			find := false
			for _, black := range blacklist {
				if black == transferEvent.From {
					find = true
				}
			}

			var body string
			if file == "swapEth" {
				explorer := fmt.Sprintf("%s/token/%s?a=%s", "https://bscscan.com/", "0xf9752a6e8a5e5f5e6eb3ab4e7d8492460fb319f0", transferEvent.From.String())
				bodyTemplate := `<h1>who: %s</h1>
							<h1>balance: %s</h1>
							<h1>txHash: %s</h1>
							<div>explorer: %s</div`
				body = fmt.Sprintf(bodyTemplate, transferEvent.From.String(), wallet.ToEth(transferEvent.Value).String(), vLog.TxHash.String(), explorer)
			} else {
				explorer := fmt.Sprintf("%s/token/%s?a=%s", "cn.etherscan.com", "0x358AA737e033F34df7c54306960a38d09AaBd523", transferEvent.From.String())
				bodyTemplate := `<h1>who: %s</h1>
							<h1>balance: %s</h1>
							<h1>txHash: %s</h1>
							<div>explorer: %s</div`
				body = fmt.Sprintf(bodyTemplate, transferEvent.From.String(), wallet.ToEth(transferEvent.Value).String(), vLog.TxHash.String(), explorer)
			}
			if Debug {
				sendDepositEmail(body, find)
			} else {
				fmt.Println("body", body)
			}

			fmt.Printf("From: %s\n", transferEvent.From.Hex())
			fmt.Printf("To: %s\n", transferEvent.To.Hex())
			fmt.Printf("Tokens: %s\n", wallet.ToEth(transferEvent.Value).String())
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

func sendDepositEmail(body string, black bool) {
	m := gomail.NewMessage()

	//Sender
	m.SetHeader("From", "1032087738@qq.com")
	//Receiver
	m.SetHeader("To", "450595468@qq.com")
	//CC
	//m.SetAddressHeader("Cc", "xxx@qq.com", "xiaozhujiao")
	//Subject
	if black {
		m.SetHeader("Subject", "黑名单账户")
	} else {
		m.SetHeader("Subject", "账户余额不足")
	}

	m.SetBody("text/html", body)
	//attach
	//m.Attach("./myIpPic.png")

	//拿到token，并进行连接,第4个参数是填授权码
	d := gomail.NewDialer("smtp.qq.com", 587, "1032087738@qq.com", "fvsofxkcmnaqbaja")

	// 发送邮件
	if err := d.DialAndSend(m); err != nil {
		fmt.Printf("DialAndSend err %v:", err)
	}
	fmt.Printf("send mail success\n")

}
