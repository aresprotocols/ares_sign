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
	From  common.Address `json:"from"`
	To    common.Address `json:"to"`
	Value *big.Int       `json:"value"`
}

func SendBscTransaction(param map[string]string) (string, error) {
	txHash := common.HexToHash(param["tx_hash"])

	tx, pending, err := mywallet.client.TransactionByHash(txHash)

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
		fmt.Println("", transferEvent.From, " ", transferEvent.To, "  ", transferEvent.Value, " ", ToEth(transferEvent.Value))
	}

	if transferEvent.To != mywallet.adminAddress {
		return "", errors.New("adminAddress address is correct")
	}

	input := mywallet.packInput("transfer", transferEvent.From, transferEvent.Value)
	bscHash, err := mywallet.sendBscTransaction(mywallet.bscContractAddress, nil, input)
	if err != nil {
		return "", err
	}

	fmt.Println("Please waiting ", " bscHash ", bscHash)

	count := 0
	for {
		time.Sleep(time.Millisecond * 200)
		_, isPending, err := mywallet.bscClient.TransactionByHash(common.HexToHash(bscHash))
		if err != nil {
			return "", err
		}
		count++
		if !isPending {
			break
		}
		if count >= 40 {
			fmt.Println("Please use querytx sub command query later.")
			return "", errors.New("bsc tx error")
		}
	}
	receipt, err = mywallet.bscClient.TransactionReceipt(common.HexToHash(bscHash))
	if err != nil {
		return "", err
	}

	if receipt.Status == types.ReceiptStatusSuccessful {
		block, err := mywallet.client.BlockByHash(receipt.BlockHash)
		if err != nil {
			return "", err
		}

		fmt.Println("Bsc transaction Success", " block Number", receipt.BlockNumber.Uint64(), " block txs", len(block.Transactions()), "blockhash", block.Hash().Hex())
	}

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

	err = w.bscClient.SendTransaction(signedTx)
	if err != nil {
		log.Println("Send transaction err:", err)
		return "", err
	}

	return signedTx.Hash().Hex(), nil
}

func GetGasPrice() (string, error) {
	return mywallet.getGasPrice()
}
