package wallet

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Client struct {
	conn *ethclient.Client
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Connect(url string) {
	client, err := ethclient.Dial(url)
	if err != nil {
		log.Printf("dail to %s err:%v\n", url, err)
		return
	}

	c.conn = client
}

func (c *Client) ChainID() (*big.Int, error) {
	return c.conn.ChainID(context.Background())
}

func (c *Client) SendTransaction(tx *types.Transaction) error {
	return c.conn.SendTransaction(context.Background(), tx)
}

func (c *Client) BalanceAt(address common.Address, blockNumber *big.Int) (*big.Int, error) {
	balance, err := c.conn.BalanceAt(context.Background(), address, blockNumber)
	if err != nil {
		log.Println("Get balance err:", err)
		return nil, err
	}
	return balance, nil
}

func (c *Client) PendingNonceAt(address common.Address) (uint64, error) {
	return c.conn.PendingNonceAt(context.Background(), address)
}

func (c *Client) SuggestGasPrice() (*big.Int, error) {
	return c.conn.SuggestGasPrice(context.Background())
}

func (c *Client) EstimateGas(msg ethereum.CallMsg) (uint64, error) {
	return c.conn.EstimateGas(context.Background(), msg)
}

func (c *Client) TransactionReceipt(txHash common.Hash) (*types.Receipt, error) {
	return c.conn.TransactionReceipt(context.Background(), txHash)
}

func (c *Client) BlockByHash(txHash common.Hash) (*types.Block, error) {
	return c.conn.BlockByHash(context.Background(), txHash)
}

func (c *Client) TransactionByHash(txHash common.Hash) (tx *types.Transaction, isPending bool, err error) {
	return c.conn.TransactionByHash(context.Background(), txHash)
}

func (c *Client) CallContract(msg ethereum.CallMsg) ([]byte, error) {
	return c.conn.CallContract(context.Background(), msg, nil)
}
