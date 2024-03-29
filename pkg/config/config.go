package config

import (
	"fmt"
	"math/big"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"
)

var (
	rawVersion = "v0.2"
	// Version of the goldchain binaries.
	//
	// Value is defined by a private build flag,
	// or hardcoded to the latest released tag as fallback.
	Version build.ProtocolVersion
)

const (
	// GolchainTokenUnit defines the unit of one Token.
	GolchainTokenUnit = "GFT"
	// GoldchainTokenChainName defines the name of the chain.
	GoldchainTokenChainName = "goldchain"
)

// chain names
const (
	NetworkNameStandard = "standard"
	NetworkNameTest     = "testnet"
	NetworkNameDev      = "devnet"
)

// global network config constants
const (
	BlockFrequency types.BlockHeight = 120 // 1 block per 2 minutes on average
)

// GetBlockchainInfo returns the naming and versioning of tfchain.
func GetBlockchainInfo() types.BlockchainInfo {
	return types.BlockchainInfo{
		Name:            GoldchainTokenChainName,
		NetworkName:     NetworkNameTest,
		CoinUnit:        GolchainTokenUnit,
		ChainVersion:    Version,       // use our own blockChain/build version
		ProtocolVersion: build.Version, // use latest available rivine protocol version
	}
}

// GetStandardnetGenesis explicitly sets all the required constants for the genesis block of the standard (prod) net
func GetStandardnetGenesis() types.ChainConstants {
	cfg := types.StandardnetChainConstants()

	// set transaction versions
	cfg.DefaultTransactionVersion = types.TransactionVersionOne
	cfg.GenesisTransactionVersion = types.TransactionVersionOne

	// 2 minute block time
	cfg.BlockFrequency = BlockFrequency

	// Payouts take roughly 1 day to mature.
	cfg.MaturityDelay = 720

	// The genesis timestamp
	cfg.GenesisTimestamp = types.Timestamp(1522501000) // Human time 03/31/2018 @ 1:03pm (UTC)

	// 1000 block window for difficulty
	cfg.TargetWindow = 1e3

	cfg.MaxAdjustmentUp = big.NewRat(25, 10)
	cfg.MaxAdjustmentDown = big.NewRat(10, 25)

	cfg.FutureThreshold = 1 * 60 * 60        // 1 hour.
	cfg.ExtremeFutureThreshold = 2 * 60 * 60 // 2 hours.

	cfg.StakeModifierDelay = 2000

	// Blockstakes can be used roughly 1 day after receiving
	cfg.BlockStakeAging = 1 << 17 // 2^16s < 1 day < 2^17s

	// Receive 0 coins when you create a block
	cfg.BlockCreatorFee = cfg.CurrencyUnits.OneCoin.Mul64(0)

	// Use 0.001 coins as minimum transaction fee
	cfg.MinimumTransactionFee = cfg.CurrencyUnits.OneCoin.Div64(1000)

	// Foundation receives all transactions fees in a single pool address,
	cfg.TransactionFeeCondition = types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex(
		"")))

	// no  initial coins, except  1 for initial transaction fee payments
	cfg.GenesisCoinDistribution = []types.CoinOutput{}

	// allocate block stakes
	cfg.GenesisBlockStakeAllocation = []types.BlockStakeOutput{
		{
			// 100 BS,
			Value:     types.NewCurrency64(100),
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex(""))),
		},

		{
			// 10 BS, for dev/validation/test purposes
			Value: types.NewCurrency64(10),
			// @foundation @robvanmieghem
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex(""))),
		},
	}

	return cfg
}

// GetTestnetGenesis explicitly sets all the required constants for the genesis block of the testnet
func GetTestnetGenesis() types.ChainConstants {
	cfg := types.TestnetChainConstants()

	// set transaction versions
	cfg.DefaultTransactionVersion = types.TransactionVersionOne
	cfg.GenesisTransactionVersion = types.TransactionVersionOne

	// 2 minute block time
	cfg.BlockFrequency = BlockFrequency

	// Payouts take rougly 1 day to mature.
	cfg.MaturityDelay = 720

	// The genesis timestamp is set to July 26th, 2019
	cfg.GenesisTimestamp = types.Timestamp(1564142400) // July 26th, 2019 @ 12:00pm UTC.

	// 1000 block window for difficulty
	cfg.TargetWindow = 1e3

	cfg.MaxAdjustmentUp = big.NewRat(25, 10)
	cfg.MaxAdjustmentDown = big.NewRat(10, 25)

	cfg.FutureThreshold = 1 * 60 * 60        // 1 hour.
	cfg.ExtremeFutureThreshold = 2 * 60 * 60 // 2 hours.

	cfg.StakeModifierDelay = 2000

	// Blockstake can be used roughly 1 minute after receiving
	cfg.BlockStakeAging = uint64(1 << 6)

	// Receive 0 coins when you create a block
	cfg.BlockCreatorFee = cfg.CurrencyUnits.OneCoin.Mul64(0)

	// Use 0.001 coins as minimum transaction fee
	cfg.MinimumTransactionFee = cfg.CurrencyUnits.OneCoin.Div64(1000)

	// Start with 100M coins
	cfg.GenesisCoinDistribution = []types.CoinOutput{
		{
			// Create 100M coins
			Value: cfg.CurrencyUnits.OneCoin.Mul64(100 * 1000 * 1000),
			// @leesmet
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("01215a03f0098c4fcd801854da4d7bb2e9c78b4d3598fec89f42bc19fb79889bbf7a6aabdbe95f"))),
		},
	}

	// allocate block stakes
	cfg.GenesisBlockStakeAllocation = []types.BlockStakeOutput{
		{
			// Create 3000 blockstakes
			Value: types.NewCurrency64(3000),
			// @leesmet
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("01215a03f0098c4fcd801854da4d7bb2e9c78b4d3598fec89f42bc19fb79889bbf7a6aabdbe95f"))),
		},
	}

	return cfg
}

// GetDevnetGenesis explicitly sets all the required constants for the genesis block of the devnet
func GetDevnetGenesis() types.ChainConstants {
	cfg := types.DevnetChainConstants()

	// set transaction versions
	cfg.DefaultTransactionVersion = types.TransactionVersionOne
	cfg.GenesisTransactionVersion = types.TransactionVersionOne

	// 12 seconds, slow enough for developers to see
	// ~each block, fast enough that blocks don't waste time
	cfg.BlockFrequency = 12

	// 120 seconds before a delayed output matters
	// as it's expressed in units of blocks
	cfg.MaturityDelay = 10
	cfg.MedianTimestampWindow = 11

	// The genesis timestamp is set to February 21st, 2018
	cfg.GenesisTimestamp = types.Timestamp(1519200000) // February 21st, 2018 @ 8:00am UTC.

	// difficulity is adjusted based on prior 20 blocks
	cfg.TargetWindow = 20

	// Difficulty adjusts quickly.
	cfg.MaxAdjustmentUp = big.NewRat(120, 100)
	cfg.MaxAdjustmentDown = big.NewRat(100, 120)

	cfg.FutureThreshold = 2 * 60        // 2 minutes
	cfg.ExtremeFutureThreshold = 4 * 60 // 4 minutes

	cfg.StakeModifierDelay = 2000

	// Blockstake can be used roughly 1 minute after receiving
	cfg.BlockStakeAging = uint64(1 << 6)

	// Receive 10 coins when you create a block
	cfg.BlockCreatorFee = cfg.CurrencyUnits.OneCoin.Mul64(10)

	// Use 0.1 coins as minimum transaction fee
	cfg.MinimumTransactionFee = cfg.CurrencyUnits.OneCoin.Mul64(1)

	// distribute initial coins
	cfg.GenesisCoinDistribution = []types.CoinOutput{
		{
			// Create 100M coins
			Value: cfg.CurrencyUnits.OneCoin.Mul64(100 * 1000 * 1000),
			// belong to wallet with mnemonic:
			// carbon boss inject cover mountain fetch fiber fit tornado cloth wing dinosaur proof joy intact fabric thumb rebel borrow poet chair network expire else
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f"))),
		},
	}

	// allocate block stakes
	cfg.GenesisBlockStakeAllocation = []types.BlockStakeOutput{
		{
			// Create 3K blockstakes
			Value: types.NewCurrency64(3000),
			// belongs to wallet with mnemonic:
			// carbon boss inject cover mountain fetch fiber fit tornado cloth wing dinosaur proof joy intact fabric thumb rebel borrow poet chair network expire else
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f"))),
		},
	}

	return cfg
}

// GetStandardnetGenesisAuthCoinCondition returns the genesis auth condition used for the standard (prod) net
func GetStandardnetGenesisAuthCoinCondition() types.UnlockConditionProxy {
	// TODO: adapt to real condition, also being multi-sig
	return types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f")))
}

// GetStandardnetBootstrapPeers sets the standard bootstrap node addresses
func GetStandardnetBootstrapPeers() []modules.NetAddress {
	return []modules.NetAddress{
		"bootstrap1.goldtoken.bnh.com:22112",
		"bootstrap2.goldtoken.bnh.com:22112",
		"bootstrap3.goldtoken.bnh.com:22112",
		"bootstrap4.goldtoken.bnh.com:22112",
		"bootstrap5.goldtoken.bnh.com:22112",
	}
}

// GetTestnetGenesisAuthCoinCondition returns the genesis auth condition used for the testnet
func GetTestnetGenesisAuthCoinCondition() types.UnlockConditionProxy {
	// @leesmet, to be changed to 1 out of 2 multisig once we have an automated
	// faucet for this
	return types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("01215a03f0098c4fcd801854da4d7bb2e9c78b4d3598fec89f42bc19fb79889bbf7a6aabdbe95f")))
}

// GetTestnetBootstrapPeers sets the testnet bootstrap node addresses
func GetTestnetBootstrapPeers() []modules.NetAddress {
	return []modules.NetAddress{
		"bootstrap1.testnet.nbh-digital.com:22112",
		"bootstrap2.testnet.nbh-digital.com:22112",
		"bootstrap3.testnet.nbh-digital.com:22112",
		"bootstrap4.testnet.nbh-digital.com:22112",
		"bootstrap5.testnet.nbh-digital.com:22112",
	}
}

// GetDevnetGenesisAuthCoinCondition returns the genesis auth condition used for the devnet
func GetDevnetGenesisAuthCoinCondition() types.UnlockConditionProxy {
	return types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f")))
}

// GetDevnetBootstrapPeers sets the default devnet bootstrap node addresses
func GetDevnetBootstrapPeers() []modules.NetAddress {
	return []modules.NetAddress{
		"localhost:22112",
	}
}

func unlockHashFromHex(hstr string) (uh types.UnlockHash) {
	err := uh.LoadString(hstr)
	if err != nil {
		panic(fmt.Sprintf("func unlockHashFromHex(%s) failed: %v", hstr, err))
	}
	return
}

func init() {
	Version = build.MustParse(rawVersion)
}
