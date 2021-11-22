package wallet

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"log"
	"math/big"
	"time"
)

type LogTransfer struct {
	From   common.Address `json:"from"`
	To     common.Address `json:"to"`
	Value  *big.Int       `json:"value"`
	ValueS string         `json:"value_f"`
	Height uint64         `json:"height"`
	BscTx  common.Hash    `json:"bsc_tx"`
}

func GetBscBalance() (*big.Int, error) {
	return mywallet.getAresBalance()
}

func SendBscTransaction(txHash common.Hash) (string, error) {
	mywallet.lock.Lock()
	defer mywallet.lock.Unlock()

	tx, pending, err := mywallet.client.TransactionByHash(txHash)

	if _, ok := mywallet.swapAccount[txHash.String()]; ok {
		return "", errors.New("cross bsc already exists")
	}

	swapAccount := LoadSwapJSON("tx_success")
	if swapAccount == nil {
		swapAccount = make(map[string]*LogTransfer)
	}
	if _, ok := swapAccount[txHash.String()]; ok {
		return "", fmt.Errorf("cross bsc local already exists: %v", txHash)
	}

	if pending {
		fmt.Println("Please waiting ", " txHash ", txHash.String())
	}

	fmt.Println("txHash ", txHash)

	if tx.To() == nil {
		return "", errors.New("to address is nil")
	}

	if *tx.To() != mywallet.contractAddress {
		return "", errors.New("contractAddress address is correct")
	}

	receipt, err := mywallet.client.TransactionReceipt(txHash)
	if err != nil {
		return "", err
	}

	if receipt.Status == types.ReceiptStatusSuccessful {
		block, err := mywallet.client.BlockByHash(receipt.BlockHash)
		if err != nil {
			return "", err
		}

		fmt.Println("Transaction Success", " block Number", receipt.BlockNumber.Uint64(), " block txs", len(block.Transactions()), "blockhash", block.Hash().Hex())
	}

	var transferEvent LogTransfer
	for _, log := range receipt.Logs {
		fmt.Println("log ", hex.EncodeToString(log.Data))
		err = mywallet.contractAbi.UnpackIntoInterface(&transferEvent, "Transfer", log.Data)
		if err != nil {
			return "", errors.New("UnpackIntoInterface address is correct")
		}
		transferEvent.From = common.HexToAddress(log.Topics[1].Hex())
		transferEvent.To = common.HexToAddress(log.Topics[2].Hex())
		transferEvent.ValueS = ToEth(transferEvent.Value).String()
		transferEvent.Height = log.BlockNumber
		fmt.Println("", transferEvent.From, " ", transferEvent.To, "  ", transferEvent.Value, " ", ToEth(transferEvent.Value))
	}

	if transferEvent.To != mywallet.adminAddress {
		return "", errors.New("adminAddress address is correct")
	}

	input := mywallet.packInput("transfer", transferEvent.From, transferEvent.Value)
	mywallet.update = true
	number, _ := mywallet.getAresBalance()
	if number.Cmp(transferEvent.Value) < 0 {
		mywallet.sendDepositEmail(number, txHash.String()+"  "+ToEth(transferEvent.Value).String(), true)
		return "", errors.New("account balance is low")
	}

	bscHash, err := mywallet.sendBscTransaction(mywallet.bscContractAddress, nil, input)
	if err != nil {
		return "", err
	}

	transferEvent.BscTx = common.HexToHash(bscHash)
	txSwap := make(SwapAccount)

	fmt.Println("Please waiting ", " bscHash ", bscHash)

	mywallet.sendDepositEmail(transferEvent.Value, txHash.String()+"  "+ToEth(transferEvent.Value).String()+"  "+transferEvent.From.String(), true)

	count := 0
	for {
		_, isPending, err := mywallet.bscClient.TransactionByHash(common.HexToHash(bscHash))
		if err != nil {
			fmt.Println("Please use TransactionByHash sub command query later.", err)
			return "", err
		}
		count++
		if !isPending {
			break
		}
		time.Sleep(time.Millisecond * 200)
		if count >= 40 {
			fmt.Println("Please use querytx sub command query later.")
			return "", errors.New("bsc tx error")
		}
	}

	txSwap[txHash.String()] = &transferEvent
	WriteSwapJSON("tx_success", txSwap)
	mywallet.swapAccount[txHash.String()] = &transferEvent
	mywallet.update = true

	//for {
	//	receipt, err = mywallet.bscClient.TransactionReceipt(common.HexToHash(bscHash))
	//	if err != nil {
	//		fmt.Println("Please use TransactionReceipt sub command query later.", err)
	//	}
	//	count++
	//	if err == nil {
	//		if receipt.Status == types.ReceiptStatusSuccessful {
	//			break
	//		}
	//	}
	//	fmt.Println("count ", count)
	//	time.Sleep(time.Millisecond * 200)
	//	if count >= 10 {
	//		fmt.Println("Please use TransactionReceipt sub command query later.", err)
	//		return "", errors.New("bsc tx error")
	//	}
	//}

	return bscHash, nil
}

func (w *Wallet) packInput(abiMethod string, params ...interface{}) []byte {
	input, err := w.contractAbi.Pack(abiMethod, params...)
	if err != nil {
		fmt.Println(abiMethod, " error ", err)
	}
	return input
}

func (w *Wallet) sendTransaction(toAccount string, amount, gasPrice *big.Int, gasLimit uint64) (string, error) {
	if w.client == nil {
		return "", errors.New("Please check network connection")
	}

	fromAddress := common.HexToAddress(w.account)

	nonce, err := w.client.PendingNonceAt(fromAddress)
	if err != nil {
		log.Println("Get nonce err:", err)
		return "", err
	}

	networkId, err := w.client.ChainID()
	if err != nil {
		log.Println("Get network id err:", err)
		return "", err
	}

	var data []byte

	toAddress := common.HexToAddress(toAccount)
	tx := types.NewTransaction(nonce, toAddress, amount, gasLimit, gasPrice, data)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(networkId), w.privateKey)
	if err != nil {
		log.Println("Signed transaction err:", err)
		return "", err
	}

	err = w.client.SendTransaction(signedTx)
	if err != nil {
		log.Println("Send transaction err:", err)
		return "", err
	}

	return signedTx.Hash().Hex(), nil
}

func (w *Wallet) sendBscTransaction(toAccount common.Address, amount *big.Int, input []byte) (string, error) {
	if w.bscClient == nil {
		return "", errors.New("Please check network connection")
	}

	fromAddress := common.HexToAddress(w.account)

	// Ensure a valid value field and resolve the account nonce
	nonce, err := w.bscClient.PendingNonceAt(fromAddress)
	if err != nil {
		log.Println("Get nonce err:", err)
		return "", err
	}

	gasPrice, err := w.bscClient.SuggestGasPrice()
	if err != nil {
		log.Fatal(err)
	}

	gasLimit := uint64(2100000) // in units
	// If the contract surely has code (or code is not needed), estimate the transaction
	msg := ethereum.CallMsg{From: fromAddress, To: &toAccount, GasPrice: gasPrice, Value: amount, Data: input}
	gasLimit, err = w.bscClient.EstimateGas(msg)
	if err != nil {
		fmt.Println("Contract exec failed", err)
	}
	if gasLimit < 1 {
		gasLimit = 866328
	}

	networkId, err := w.bscClient.ChainID()
	if err != nil {
		log.Println("Get network id err:", err)
		return "", err
	}

	tx := types.NewTransaction(nonce, toAccount, amount, gasLimit, gasPrice, input)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(networkId), w.privateKey)
	if err != nil {
		log.Println("Signed transaction err:", err)
		return "", err
	}

	//err = w.bscClient.SendTransaction(signedTx)
	//if err != nil {
	//	log.Println("Send transaction err:", err)
	//	return "", err
	//}

	return signedTx.Hash().Hex(), nil
}
