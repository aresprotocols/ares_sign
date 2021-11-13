package wallet

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"io/ioutil"
	"os"
	"sort"
)

const jsonIndent = "    "

func WriteNodesJSON(file string, nodes KeyAccount) {
	for k, v := range LoadNodesJSON(file) {
		nodes[k] = v
	}

	nodesJSON, err := json.MarshalIndent(nodes, "", jsonIndent)
	if err != nil {
		fmt.Println("MarshalIndent error", err)
		return
	}
	if file == "-" {
		os.Stdout.Write(nodesJSON)
		return
	}
	if err := ioutil.WriteFile(file, nodesJSON, 0644); err != nil {
		fmt.Println("writeFile error", err)
	}
}

func LoadNodesJSON(file string) KeyAccount {
	var nodes KeyAccount
	if isExist(file) {
		if err := common.LoadJSON(file, &nodes); err != nil {
			fmt.Println("loadNodesJSON error", err)
		}
	}
	return nodes
}

type KeyAccount map[string]uint64

func isExist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}

type SwapAccount map[string]*LogTransfer

// TxByNonce implements the sort interface to allow sorting a list of transactions
// by their nonces. This is usually only useful for sorting transactions from a
// single account, otherwise a nonce comparison doesn't make much sense.
type TxByNonce Transactions

func (s TxByNonce) Len() int { return len(s) }
func (s TxByNonce) Less(i, j int) bool {
	return s[i].Transfer.Height < s[j].Transfer.Height
}
func (s TxByNonce) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// Transactions implements DerivableList for transactions.
type Transactions []*PrintTransfer

type PrintTransfer struct {
	Id       uint64       `json:"id"`
	TxHash   string       `json:"tx_hash"`
	Transfer *LogTransfer `json:"detail"`
}

func WriteSwapJSON(file string, nodes SwapAccount) {
	for k, v := range LoadSwapJSON(file) {
		nodes[k] = v
	}
	var txs TxByNonce
	for hash, v := range nodes {
		txs = append(txs, &PrintTransfer{
			Id:       uint64(len(txs)),
			TxHash:   hash,
			Transfer: v,
		})
	}
	sort.Sort(txs)

	nodesJSON, err := json.MarshalIndent(txs, "", jsonIndent)
	if err != nil {
		fmt.Println("MarshalIndent error", err)
		return
	}
	if file == "-" {
		os.Stdout.Write(nodesJSON)
		return
	}
	if err := ioutil.WriteFile(file, nodesJSON, 0644); err != nil {
		fmt.Println("writeFile error", err)
	}
}

func LoadSwapJSON(file string) SwapAccount {
	var txs TxByNonce
	if isExist(file) {
		if err := common.LoadJSON(file, &txs); err != nil {
			fmt.Println("loadNodesJSON error", err)
		}
	}
	nodes := make(SwapAccount)
	for _, tx := range txs {
		nodes[tx.TxHash] = tx.Transfer
	}
	return nodes
}
