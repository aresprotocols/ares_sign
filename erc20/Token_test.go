package erc20

import (
	"ares/sign/util"
	"ares/sign/wallet"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
)

func init() {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlWarn, log.StreamHandler(os.Stderr, log.TerminalFormat(false))))
}

var (
	key, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	addr   = crypto.PubkeyToAddress(key.PublicKey)

	key2, _  = crypto.HexToECDSA("8a1f9a8f95be41cd7ccb6168179afb4504aefe388d1e14474d32c45c72ce7b7a")
	key3, _  = crypto.HexToECDSA("49a7b37aa6f6645917e7b807e9d1c00d4fa71f18343b0d4122a4d2df64dd6fee")
	testAddr = crypto.PubkeyToAddress(key2.PublicKey)
	add3     = crypto.PubkeyToAddress(key3.PublicKey)
)

func TestErc20(t *testing.T) {
	contractBackend := backends.NewSimulatedBackend(core.GenesisAlloc{
		addr:     {Balance: big.NewInt(1000000000000000000)},
		testAddr: {Balance: big.NewInt(1000000000000000000)}},
		10000000000)
	transactOpts, _ := bind.NewKeyedTransactorWithChainID(key, big.NewInt(1337))
	keyOpts, _ := bind.NewKeyedTransactorWithChainID(key2, big.NewInt(1337))
	// Deploy the ENS registry

	ensAddr, _, _, err := DeployToken(transactOpts, contractBackend)
	if err != nil {
		t.Fatalf("can't DeployContract: %v", err)
	}
	ens, err := NewToken(ensAddr, contractBackend)
	if err != nil {
		t.Fatalf("can't NewContract: %v", err)
	}

	contractBackend.Commit()

	// Set ourself as the owner of the name.
	name, err := ens.Name(nil)
	if err != nil {
		log.Error("Failed to retrieve token ", "name: %v", err)
	}
	fmt.Println("Token name:", name)

	// Set ourself as the owner of the name.
	symbol, err := ens.Symbol(nil)
	if err != nil {
		log.Error("Failed to retrieve token ", "name: %v", err)
	}
	fmt.Println("Token symbol:", symbol)

	totalSupply, err := ens.TotalSupply(nil)
	if err != nil {
		log.Error("Failed to retrieve token ", "name: %v", err)
	}
	fmt.Println("totalSupply ", totalSupply)

	balance, err := ens.BalanceOf(nil, addr)
	if err != nil {
		log.Error("Failed to retrieve token ", "name: %v", err)
	}
	fmt.Println("addr balance BalanceOf", balance)

	tx, err := ens.Transfer(transactOpts, testAddr, big.NewInt(50000))
	if err != nil {
		log.Error("Failed to request token transfer: %v", err)
	}
	fmt.Printf("Transfer pending: 0x%x\n", tx.Hash())
	contractBackend.Commit()

	balance, err = ens.BalanceOf(nil, testAddr)
	if err != nil {
		log.Error("Failed to retrieve token ", "name: %v", err)
	}
	fmt.Println("testAddr balance BalanceOf", balance)

	tx, err = ens.Approve(keyOpts, addr, big.NewInt(10000))
	if err != nil {
		log.Error("Failed to retrieve Approve ", "name: %v", err)
	}
	contractBackend.Commit()

	balance, err = ens.Allowance(nil, testAddr, addr)
	if err != nil {
		log.Error("Failed to retrieve token ", "name: %v", err)
	}
	fmt.Println("balance Allowance", balance)

	tx, err = ens.TransferFrom(transactOpts, testAddr, add3, big.NewInt(5000))
	if err != nil {
		log.Error("Failed to request token transfer: %v", err)
	}
	fmt.Printf("Transfer pending: 0x%x\n", tx.Hash())
	contractBackend.Commit()

	balance, err = ens.Allowance(nil, testAddr, addr)
	if err != nil {
		log.Error("Failed to retrieve token ", "name: %v", err)
	}
	fmt.Println("Allowance balance ", balance)

	balance, err = ens.BalanceOf(nil, testAddr)
	if err != nil {
		log.Error("Failed to retrieve token ", "name: %v", err)
	}
	fmt.Println("balance BalanceOf", balance)

	balance, err = ens.BalanceOf(nil, add3)
	if err != nil {
		log.Error("Failed to retrieve token ", "name: %v", err)
	}
	fmt.Println("balance BalanceOf", balance)
	contractBackend.Commit()

	balance, _ = contractBackend.BalanceAt(context.Background(), testAddr, nil)
	fmt.Println("balance", balance)
	keyOpts.Value = new(big.Int).SetUint64(10000000)
	tx, err = ens.Approve(keyOpts, addr, big.NewInt(10000))
	if err != nil {
		log.Error("Failed to retrieve ApproveOne ", "name: %v", err)
	}
	contractBackend.Commit()

	balance, _ = contractBackend.BalanceAt(context.Background(), testAddr, nil)
	fmt.Println("11balance", balance)
}

func TestReadErc20(t *testing.T) {
	url := "https://mainnet.infura.io/v3/9aa3d95b3bc440fa88ea12eaa4456161"

	client, url := util.DialConn(url)

	ens, err := NewToken(common.HexToAddress("0x358AA737e033F34df7c54306960a38d09AaBd523"), client)
	if err != nil {
		t.Fatalf("can't NewContract: %v", err)
	}

	// Set ourself as the owner of the name.
	name, err := ens.Name(nil)
	if err != nil {
		log.Error("Failed to retrieve token ", "name: %v", err)
	}
	fmt.Println("Token name:", name)

	// Set ourself as the owner of the name.
	symbol, err := ens.Symbol(nil)
	if err != nil {
		log.Error("Failed to retrieve token ", "name: %v", err)
	}
	fmt.Println("Token symbol:", symbol)

	totalSupply, err := ens.TotalSupply(nil)
	if err != nil {
		log.Error("Failed to retrieve token ", "name: %v", err)
	}
	fmt.Println("totalSupply ", totalSupply)

	balance, err := ens.BalanceOf(nil, common.HexToAddress("0x7a646ee13eb104853c651e1d90d143acc9e72cdb"))
	if err != nil {
		log.Error("Failed to retrieve token ", "name: %v", err)
	}
	fmt.Println("addr balance BalanceOf", balance)
	//tx, err := ens.Transfer(nil, common.HexToAddress("0x7a646ee13eb104853c651e1d90d143acc9e72cdb"), wallet.EthToWei(200))
	//msg := ethereum.CallMsg{
	//	From:  common.HexToAddress("0x7a646ee13eb104853c651e1d90d143acc9e72cdb"),
	//	To:    tx.To(),
	//	Value: tx.Value(),
	//	Data:  tx.Data(),
	//}
	//limit, err := client.EstimateGas(context.Background(), msg)
}

func TestStakingErc20(t *testing.T) {
	url := "https://mainnet.infura.io/v3/9aa3d95b3bc440fa88ea12eaa4456161"

	client, url := util.DialConn(url)

	ens, err := NewToken(common.HexToAddress("0x358AA737e033F34df7c54306960a38d09AaBd523"), client)
	if err != nil {
		t.Fatalf("can't NewContract: %v", err)
	}
	arrs := []common.Address{
		common.HexToAddress("0xdf1afbc5d532a607329b095e39a013eb672a4eb3"),
		common.HexToAddress("0xa99d9fa06dd1827fd39ab2d6e0d8eb1dae9c4b93"),
		common.HexToAddress("0x4c4f6d9fae70236888c4d613199ea4419ada23e8"),
		common.HexToAddress("0xb31d8eba3f5e2d758b54544e4446b39f9cb769ea"),
	}
	total := new(big.Int)
	for _, arr := range arrs {
		balance, err := ens.BalanceOf(nil, arr)
		if err != nil {
			log.Error("Failed to retrieve token ", "name: %v", err)
		}
		total.Add(total, balance)
	}

	fmt.Println("addr balance BalanceOf", total, " ", wallet.ToEth(total))
}

func ConvertTime(utime uint64) string {
	format := time.Unix(int64(utime), 0).Format("2006-01-02 15:04:05")
	return format
}

func TestStakingBSC(t *testing.T) {
	//url := "https://mainnet.infura.io/v3/9aa3d95b3bc440fa88ea12eaa4456161"
	url := "https://bsc-dataseed.binance.org"

	client, url := util.DialConn(url)

	ens, err := NewToken(common.HexToAddress("0xf9752a6e8a5e5f5e6eb3ab4e7d8492460fb319f0"), client)
	if err != nil {
		t.Fatalf("can't NewContract: %v", err)
	}
	arrs := []common.Address{
		common.HexToAddress("0xbcaf727812a103a7350554b814afa940b9f8b87d"),
		common.HexToAddress("0x9408953119fea0612e6a7bd8b9af03cd66baeb56"),
		common.HexToAddress("0x065c4b7de1c25aeb1ab021461a0f6f56cc38b7cf"),
		common.HexToAddress("0x1bb37bdAf2cBcD65E6f185e02f7541EB40706d30"),

		common.HexToAddress("0x21F2ccfD76897C58e0083A3Ab1bbD40A066d1516"),
		common.HexToAddress("0xd5713b34E240713417b8e1341aE4FF64A9fD2828"),
	}
	total := new(big.Int)
	for _, arr := range arrs {
		balance, err := ens.BalanceOf(nil, arr)
		if err != nil {
			log.Error("Failed to retrieve token ", "name: %v", err)
		}
		total.Add(total, balance)
	}
	//12254487.516730427419216494
	val := wallet.EthToWei(1500000 + 12254487 - 510000 + 1287210)
	total.Add(total, val)
	sub, _ := wallet.ToEth(total).Int64()
	rel := 1000000000 - sub
	fmt.Println("addr balance BalanceOf", total, " ", wallet.ToEth(total))
	fmt.Println("rel ", rel)
	fmt.Println("ver ", 24246/rel)
}

func TestBnb(t *testing.T) {
	cal1 := 44.36 + 49.76 + 41.318 - 84 + 0.7 + 5
	cal2 := cal1*444.15 + 1400
	fmt.Println("cal1 ", cal1, " cal2 ", cal2)
	//
	air := 6863 * 50.0
	div := 3838143.0 - 600000 - 300000 - 300000 - air

	fmt.Println("div ", cal2/div, " air ", air)
	fmt.Println(" div ", div)
}
