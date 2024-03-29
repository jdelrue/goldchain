package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/threefoldtech/rivine/extensions/authcointx"
	"github.com/threefoldtech/rivine/pkg/api"
	"github.com/threefoldtech/rivine/types"

	gtypes "github.com/nbh-digital/goldchain/pkg/types"
	authapi "github.com/threefoldtech/rivine/extensions/authcointx/api"
)

var (
	// errUnauthorized is returned when an address wants to receive coins, but
	// is currently unauthorized
	errUnauthorized = errors.New("can't send coins to a currently unauthorized address")
)

func updateAddressAuthorization(address types.UnlockHash, authorize bool) (types.TransactionID, error) {
	// Create transaction
	tx := authcointx.AuthAddressUpdateTransaction{Nonce: types.RandomTransactionNonce()}
	if authorize {
		log.Println("[DEBUG] Updating address", address.String(), "to be authorized")
		tx.AuthAddresses = []types.UnlockHash{address}
	} else {
		log.Println("[DEBUG] Updating address", address.String(), "to be deauthorized")
		tx.DeauthAddresses = []types.UnlockHash{address}
	}

	// Sign transaction
	log.Println("[DEBUG] Signing authorization transaction")
	var signedTx interface{}
	data, err := json.Marshal(tx.Transaction(types.TransactionVersion(gtypes.TransactionVersionAuthAddressUpdateTx)))
	if err != nil {
		return types.TransactionID{}, err
	}
	err = httpClient.PostResp("/wallet/sign", string(data), &signedTx)
	if err != nil {
		return types.TransactionID{}, err
	}

	// Post transaction
	log.Println("[DEBUG] Pushing authorization transaction")
	data, err = json.Marshal(signedTx)
	if err != nil {
		return types.TransactionID{}, err
	}

	var resp api.TransactionPoolPOST
	err = httpClient.PostResp("/transactionpool/transactions", string(data), &resp)
	return resp.TransactionID, err
}

func dripCoins(address types.UnlockHash, amount types.Currency) (types.TransactionID, error) {
	// Check if address is authorized first
	var result authapi.GetAddressesAuthStateResponse
	err := httpClient.GetAPI(fmt.Sprintf("/consensus/authcoin/status?addr=%s", address.String()), &result)
	if err != nil {
		return types.TransactionID{}, err
	}
	if len(result.AuthStates) == 0 {
		return types.TransactionID{}, fmt.Errorf(
			"failed to check authorization state for address %s: no auth states or error returned",
			address.String())
	}
	if len(result.AuthStates) > 1 {
		return types.TransactionID{}, fmt.Errorf(
			"failed to check authorization state for address %s: ambiguity issue: more than one auth state returned, while one was expected",
			address.String())
	}

	if !result.AuthStates[0] {
		return types.TransactionID{}, errUnauthorized
	}

	data, err := json.Marshal(api.WalletCoinsPOST{
		CoinOutputs: []types.CoinOutput{
			{
				Value:     amount,
				Condition: types.NewCondition(types.NewUnlockHashCondition(address)),
			},
		},
	})
	if err != nil {
		return types.TransactionID{}, err
	}

	log.Println("[DEBUG] Dripping", amount.String(), "coins to address", address.String())

	var resp api.WalletCoinsPOSTResp
	err = httpClient.PostResp("/wallet/coins", string(data), &resp)
	if err != nil {
		log.Println("[ERROR] /wallet/coins - request body:", string(data))
	}
	return resp.TransactionID, err
}
