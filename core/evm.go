// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"math/big"

	"github.com/PlatONnetwork/PlatON-Go/x/xutil"

	"github.com/PlatONnetwork/PlatON-Go/common"
	"github.com/PlatONnetwork/PlatON-Go/consensus"
	"github.com/PlatONnetwork/PlatON-Go/core/types"
	"github.com/PlatONnetwork/PlatON-Go/core/vm"
)

// ChainContext supports retrieving headers and consensus parameters from the
// current blockchain to be used during transaction processing.
type ChainContext interface {
	// Engine retrieves the chain's consensus engine.
	Engine() consensus.Engine

	// GetHeader returns the hash corresponding to their hash.
	GetHeader(common.Hash, uint64) *types.Header
}

// NewEVMContext creates a new context for use in the EVM.
func NewEVMContext(msg Message, header *types.Header, chain ChainContext) vm.Context {

	beneficiary := header.Coinbase // we're must using header validation

	blockHash := common.ZeroHash
	// store the sign in  header.Extra[32:97]
	if !xutil.IsWorker(header.Extra) {
		blockHash = header.CacheHash()
	}

	return vm.Context{
		CanTransfer: CanTransfer,
		Transfer:    Transfer,
		GetHash:     GetHashFn(header, chain),
		GetNonce:    GetNonceFn(header, chain),
		Origin:      msg.From(),
		Coinbase:    beneficiary,
		BlockNumber: new(big.Int).Set(header.Number),
		Time:        new(big.Int).SetUint64(header.Time),
		GasLimit:    header.GasLimit,
		GasPrice:    new(big.Int).Set(msg.GasPrice()),
		BlockHash:   blockHash,
		Difficulty:  new(big.Int).SetUint64(0), // This one must not be deleted, otherwise the solidity contract will be failed
		Nonce:       header.Nonce,
		ParentHash:  header.ParentHash,
	}
}

// GetHashFn returns a GetHashFunc which retrieves header hashes by number
func GetHashFn(ref *types.Header, chain ChainContext) func(n uint64) common.Hash {
	var cache map[uint64]common.Hash

	return func(n uint64) common.Hash {
		// If there's no hash cache yet, make one
		if cache == nil {
			cache = map[uint64]common.Hash{
				ref.Number.Uint64() - 1: ref.ParentHash,
			}
		}
		// Try to fulfill the request from the cache
		if hash, ok := cache[n]; ok {
			return hash
		}
		// Not cached, iterate the blocks and cache the hashes
		/*for header := chain.GetHeader(ref.ParentHash, ref.Number.Uint64()-1); header != nil; header = chain.GetHeader(header.ParentHash, header.Number.Uint64()-1) {
			cache[header.Number.Uint64()-1] = header.ParentHash
			if n == header.Number.Uint64()-1 {
				return header.ParentHash
			}
		}*/
		for block := chain.Engine().GetBlockByHashAndNum(ref.ParentHash, ref.Number.Uint64()-1); block != nil; block = chain.Engine().GetBlockByHashAndNum(block.Header().ParentHash, block.Header().Number.Uint64()-1) {
			cache[block.Header().Number.Uint64()-1] = block.Header().ParentHash
			if n == block.Header().Number.Uint64()-1 {
				return block.Header().ParentHash
			}
		}
		return common.Hash{}
	}
}

func GetNonceFn(ref *types.Header, chain ChainContext) func(n uint64) []byte {
	var cache map[uint64][]byte

	return func(n uint64) []byte {
		// If there's no hash cache yet, make one
		if cache == nil {
			cache = map[uint64][]byte{
				ref.Number.Uint64() - 1: ref.Nonce.Bytes(),
			}
		}
		// Try to fulfill the request from the cache
		if nonce, ok := cache[n]; ok {
			return nonce
		}
		for block := chain.Engine().GetBlockByHashAndNum(ref.ParentHash, ref.Number.Uint64()-1); block != nil; block = chain.Engine().GetBlockByHashAndNum(block.Header().ParentHash, block.Header().Number.Uint64()-1) {
			cache[block.Header().Number.Uint64()-1] = block.Header().Nonce.Bytes()
			if n == block.Header().Number.Uint64()-1 {
				return block.Header().Nonce.Bytes()
			}
		}
		return nil
	}
}

// CanTransfer checks whether there are enough funds in the address' account to make a transfer.
// This does not take the necessary gas in to account to make the transfer valid.
func CanTransfer(db vm.StateDB, addr common.Address, amount *big.Int) bool {
	return db.GetBalance(addr).Cmp(amount) >= 0
}

// Transfer subtracts amount from sender and adds amount to recipient using the given Db
func Transfer(db vm.StateDB, sender, recipient common.Address, amount *big.Int) {
	db.SubBalance(sender, amount)
	db.AddBalance(recipient, amount)
}

type transferItem struct {
	From   common.Address
	To     common.Address
	Amount *big.Int
}

func buildTransferItem(from, to common.Address, amount *big.Int) *transferItem {
	return &transferItem{
		From:   from,
		To:     to,
		Amount: amount,
	}
}
