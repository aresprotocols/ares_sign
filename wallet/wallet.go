package wallet

import (
	"ares/sign/config"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"log"
	"math/big"
	"strings"
	"sync"
)

type Wallet struct {
	client             *Client
	bscClient          *Client
	keydir             string //key store directory
	keyFile            string //the account key store file
	account            string
	privateKey         *ecdsa.PrivateKey
	adminAddress       common.Address
	contractAddress    common.Address
	bscContractAddress common.Address
	contractAbi        abi.ABI
	balance            *big.Int
	update             bool
	swapAccount        map[string]*LogTransfer
	lock               sync.RWMutex
}

var (
	mywallet *Wallet
)

const ERC20ABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

func InitWallet() {
	mywallet = NewWallet(config.GetString("app.key_store_dir"))

	client := NewClient()
	client.Connect(config.GetString("app.ether_net_url"))

	bscClient := NewClient()
	bscClient.Connect(config.GetString("app.bsc_net_url"))
	mywallet.adminAddress = common.HexToAddress(config.GetString("app.admin_address"))
	mywallet.contractAddress = common.HexToAddress(config.GetString("app.contract_address"))
	mywallet.bscContractAddress = common.HexToAddress(config.GetString("app.bsc_contract_address"))

	if client.conn != nil {
		mywallet.client = client
	}
	if bscClient.conn != nil {
		mywallet.bscClient = bscClient
	}

	mywallet.initPrivateKey()

	contractAbi, err := abi.JSON(strings.NewReader(string(ERC20ABI)))
	if err != nil {
		log.Fatal(err)
	}
	mywallet.contractAbi = contractAbi

	mywallet.update = true
	_, err = mywallet.printBalance()
	if err != nil {
		log.Fatal(err)
	}
	swapAccount := LoadSwapJSON("tx_success")
	if swapAccount == nil {
		swapAccount = make(map[string]*LogTransfer)
	}
	mywallet.swapAccount = swapAccount
}

func NewWallet(keydir string) *Wallet {
	return &Wallet{
		keydir: keydir,
	}
}

func (w *Wallet) printBalance() (string, error) {
	if w.bscClient == nil {
		return "", errors.New("Please check network connection")
	}

	address := common.HexToAddress(w.account)
	balance, err := w.bscClient.BalanceAt(address, nil)
	if err != nil {
		log.Println("Get balance err:", err)
		return "", err
	}
	fmt.Println("printBalance", ToEth(balance))

	_, err = w.getAresBalance()
	if err != nil {
		log.Println("Get erc20 balance err:", err)
		return "", err
	}

	return balance.String(), err
}

func (w *Wallet) getAresBalance() (*big.Int, error) {
	if w.bscClient == nil {
		return nil, errors.New("Please check network connection")
	}

	if !w.update {
		return w.balance, nil
	}

	address := common.HexToAddress(w.account)

	// Pack the input, call and unpack the results
	input, err := w.contractAbi.Pack("balanceOf", address)
	if err != nil {
		return nil, err
	}

	msg := ethereum.CallMsg{From: address, To: &w.bscContractAddress, Data: input}

	output, err := w.bscClient.CallContract(msg)

	var number *big.Int
	err = w.contractAbi.UnpackIntoInterface(&number, "balanceOf", output)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("printBalance erc20", ToEth(number), " address ", address)
	w.balance = number
	w.update = false
	return number, err
}
