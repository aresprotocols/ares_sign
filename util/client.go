package util

import (
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
)

func DialConn(url string) (*ethclient.Client, string) {
	ip := "165.227.99.131"
	port := 8545

	//url = "https://ethrpc.truescan.network"
	//url = "https://kovan.poa.network/"

	if url == "" {
		url = fmt.Sprintf("http://%s", fmt.Sprintf("%s:%d", ip, port))
	}
	// Create an IPC based RPC connection to a remote node
	// "http://39.100.97.129:8545"
	conn, err := ethclient.Dial(url)
	if err != nil {
		fmt.Printf("Failed to connect to the ethereum client: %v \n", err)
	}
	return conn, url
}
