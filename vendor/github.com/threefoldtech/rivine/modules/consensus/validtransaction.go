package consensus

import (
	"errors"

	bolt "github.com/rivine/bbolt"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"
)

type consensusStateGetter struct {
	tx *bolt.Tx
}

func newConsensusStateGetter(tx *bolt.Tx) *consensusStateGetter {
	return &consensusStateGetter{tx: tx}
}

var _ modules.ConsensusStateGetter = (*consensusStateGetter)(nil)

func (csg *consensusStateGetter) BlockAtID(id types.BlockID) (modules.ConsensusBlock, error) {
	pb, err := getBlockMap(csg.tx, id)
	if err != nil {
		return modules.ConsensusBlock{}, err
	}
	return modules.ConsensusBlock{
		Block:       pb.Block,
		Height:      pb.Height,
		Depth:       pb.Depth,
		ChildTarget: pb.ChildTarget,
	}, nil
}

func (csg *consensusStateGetter) BlockAtHeight(height types.BlockHeight) (modules.ConsensusBlock, error) {
	id, err := getPath(csg.tx, height)
	if err != nil {
		return modules.ConsensusBlock{}, err
	}
	return csg.BlockAtID(id)
}

func (csg *consensusStateGetter) UnspentCoinOutputGet(id types.CoinOutputID) (types.CoinOutput, error) {
	return getCoinOutput(csg.tx, id)
}

func (csg *consensusStateGetter) UnspentBlockStakeOutputGet(id types.BlockStakeOutputID) (types.BlockStakeOutput, error) {
	return getBlockStakeOutput(csg.tx, id)
}

// validTransaction checks that all fields are valid within the current
// consensus state. If not an error is returned.
func (cs *ConsensusSet) validTransaction(tx *bolt.Tx, t types.Transaction, constants types.TransactionValidationConstants, blockHeight types.BlockHeight, blockTimestamp types.Timestamp, isBlockCreatingTx bool) error {
	csg := newConsensusStateGetter(tx)
	ctx := types.TransactionValidationContext{
		ValidationContext: types.ValidationContext{
			Confirmed:         true,
			BlockHeight:       blockHeight,
			BlockTime:         blockTimestamp,
			IsBlockCreatingTx: isBlockCreatingTx,
		},
		BlockSizeLimit:         constants.BlockSizeLimit,
		ArbitraryDataSizeLimit: constants.ArbitraryDataSizeLimit,
		MinimumMinerFee:        constants.MinimumMinerFee,
	}

	// return the first error reported by a validator
	var err error

	// check if we have stand alone validators specific for this tx version, if so apply them
	if validators, ok := cs.txVersionMappedValidators[t.Version]; ok {
		for _, validator := range validators {
			err = validator(t, ctx, csg)
			if err != nil {
				return err
			}
		}
	}

	// validate all transactions using the stand alone validators
	for _, validator := range cs.txValidators {
		err = validator(t, ctx, csg)
		if err != nil {
			return err
		}
	}

	// validate using the plugins, both version-specific as well as global
	return cs.validateTransactionUsingPlugins(t, ctx, csg, tx)
}

// TryTransactionSet applies the input transactions to the consensus set to
// determine if they are valid. An error is returned IFF they are not a valid
// set in the current consensus set. The size of the transactions and the set
// is not checked. After the transactions have been validated, a consensus
// change is returned detailing the diffs that the transaction set would have.
func (cs *ConsensusSet) TryTransactionSet(txns []types.Transaction) (modules.ConsensusChange, error) {
	err := cs.tg.Add()
	if err != nil {
		return modules.ConsensusChange{}, err
	}
	defer cs.tg.Done()
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	// applyTransaction will apply the diffs from a transaction and store them
	// in a block node. diffHolder is the blockNode that tracks the temporary
	// changes. At the end of the function, all changes that were made to the
	// consensus set get reverted.
	diffHolder := new(processedBlock)

	// Boltdb will only roll back a tx if an error is returned. In the case of
	// TryTransactionSet, we want to roll back the tx even if there is no
	// error. So errSuccess is returned. An alternate method would be to
	// manually manage the tx instead of using 'Update', but that has safety
	// concerns and is more difficult to implement correctly.
	errSuccess := errors.New("success")
	err = cs.db.Update(func(tx *bolt.Tx) error {
		diffHolder.Height = blockHeight(tx)
		blockTime, err := blockTimeStamp(tx, diffHolder.Height)
		if err != nil {
			return err
		}
		for _, txn := range txns {
			// a transaction can only be "block creating" in the context of a block,
			// which we don't have here, so just pass in false for the "isBlockCreatingTx"
			// argument. In other words, a block creating transaction can never be part
			// of a transaction pool and must be inserted when the block is actually created
			err := cs.validTransaction(tx, txn, types.TransactionValidationConstants{
				BlockSizeLimit:         cs.chainCts.BlockSizeLimit,
				ArbitraryDataSizeLimit: cs.chainCts.ArbitraryDataSizeLimit,
				MinimumMinerFee:        cs.chainCts.MinimumTransactionFee,
			}, diffHolder.Height, blockTime, false)
			if err != nil {
				cs.log.Printf("WARN: try-out tx %v is invalid: %v", txn.ID(), err)
				return err
			}
			applyTransaction(tx, diffHolder, txn)

			// apply transaction for all plugins
			for name, plugin := range cs.plugins {
				bucket := cs.bucketForPlugin(tx, name)
				err := plugin.ApplyTransaction(txn, diffHolder.Block, diffHolder.Height, bucket)
				if err != nil {
					return err
				}
			}
		}
		return errSuccess
	})
	if err != errSuccess {
		return modules.ConsensusChange{}, err
	}
	cc := modules.ConsensusChange{
		CoinOutputDiffs:       diffHolder.CoinOutputDiffs,
		BlockStakeOutputDiffs: diffHolder.BlockStakeOutputDiffs,
	}
	return cc, nil
}
