package wallet

import (
	"ares/sign/console"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"io/ioutil"
	"log"
)

func (w *Wallet) initPrivateKey() (string, error) {
	var (
		key *keystore.Key
		err error
	)

	files, err := ioutil.ReadDir(w.keydir)
	if err != nil {
		log.Printf("Read %s directory err:%v\n", w.keydir, err)
		return "", err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		} else {
			filename := w.keydir + "/" + file.Name()

			for trials := 0; trials < 2; trials++ {
				password := getPassPhrase("Please enter the password to decrypt the'"+file.Name()+"':", false)

				//password := "secret"
				key, err = Decrypt(filename, password)

				if err != nil {
					if trials == 1 {
						fmt.Println("Input error password limit ", file.Name())
						break
					}
					fmt.Println("Please input correct password")
				} else {
					log.Println("Decrypt keystore success", "keyfile", file.Name())
					//decrypt the private key successful
					w.privateKey = key.PrivateKey
					w.account = key.Address.Hex()
					w.keyFile = file.Name()
					return w.account, nil
				}
			}
		}
	}

	return "", err
}

func (w *Wallet) Create(password string) (string, error) {
	ks := keystore.NewKeyStore(w.keydir, keystore.StandardScryptN, keystore.StandardScryptP)

	account, err := ks.NewAccount(password)
	if err != nil {
		log.Println(err)
		return "", err
	}

	w.account = account.Address.Hex()

	return w.account, nil
}

func (w *Wallet) getPrivateKey(password string) (string, error) {
	filename := w.keydir + "/" + w.keyFile
	key, err := Decrypt(filename, password)
	if err != nil {
		return "", err
	}

	privateKeyBytes := crypto.FromECDSA(key.PrivateKey)

	return hexutil.Encode(privateKeyBytes), nil
}

// getPassPhrase retrieves the password associated with truekey, either fetched
// from a list of preloaded passphrases, or requested interactively from the user.
func getPassPhrase(prompt string, confirmation bool) string {
	fmt.Println(prompt)
	password, err := console.Stdin.PromptPassword("Password: ")
	if err != nil {
		utils.Fatalf("Failed to read password: %v", err)
	}
	if confirmation {
		confirm, err := console.Stdin.PromptPassword("Repeat password: ")
		if err != nil {
			utils.Fatalf("Failed to read password confirmation: %v", err)
		}
		if password != confirm {
			utils.Fatalf("Passwords do not match")
		}
	}
	return password
}
