/**
*  'TestOracle Env'(containing all calls info and status now and then) Init.
*   1.hacker_init
    2 add call
*   3.close call
*   4.hacker_close
*  These operation happens in function Call,CallCode,DelegateCall.
*/

// Copyright 2014 The go-ethereum Authors
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

package vm

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"log"
	"math/big"
	"runtime"
	"strings"
	"sync/atomic"
)

var Println = log.Println
var Printf = log.Printf

type (
	CanTransferFunc func(StateDB, common.Address, *big.Int) bool
	TransferFunc    func(StateDB, common.Address, common.Address, *big.Int)
	// GetHashFunc returns the nth block hash in the blockchain
	// and is used by the BLOCKHASH EVM op code.
	GetHashFunc func(uint64) common.Hash
)

// run runs the given contract and takes care of running precompiles with a fallback to the byte code interpreter.
func run(evm *EVM, snapshot int, contract *Contract, input []byte) ([]byte, error) {
	if contract.CodeAddr != nil {
		precompiledContracts := PrecompiledContracts
		if p := precompiledContracts[*contract.CodeAddr]; p != nil {
			return RunPrecompiledContract(p, input, contract)
		}
	}
	//return evm.interpreter.Run(snapshot, contract, input)
	ret, err := evm.interpreter.Run(snapshot, contract, input)
	return ret, err
}

// Context provides the EVM with auxiliary information. Once provided
// it shouldn't be modified.
type Context struct {
	// CanTransfer returns whether the account contains
	// sufficient ether to transfer the value
	CanTransfer CanTransferFunc
	// Transfer transfers ether from one account to the other
	Transfer TransferFunc
	// GetHash returns the hash corresponding to n
	GetHash GetHashFunc

	// Message information
	Origin   common.Address // Provides information for ORIGIN
	GasPrice *big.Int       // Provides information for GASPRICE

	// Block information
	Coinbase    common.Address // Provides information for COINBASE
	GasLimit    *big.Int       // Provides information for GASLIMIT
	BlockNumber *big.Int       // Provides information for NUMBER
	Time        *big.Int       // Provides information for TIME
	Difficulty  *big.Int       // Provides information for DIFFICULTY
}

// EVM is the Ethereum Virtual Machine base object and provides
// the necessary tools to run a contract on the given state with
// the provided context. It should be noted that any error
// generated through any of the calls should be considered a
// revert-state-and-consume-all-gas operation, no checks on
// specific errors should ever be performed. The interpreter makes
// sure that any errors generated are to be considered faulty code.
//
// The EVM should never be reused and is not thread safe.
type EVM struct {
	// Context provides auxiliary blockchain related information
	Context
	// StateDB gives access to the underlying state
	StateDB StateDB
	// Depth is the current call stack
	depth int

	// chainConfig contains information about the current chain
	chainConfig *params.ChainConfig
	// chain rules contains the chain rules for the current epoch
	chainRules params.Rules
	// virtual machine configuration options used to initialise the
	// evm.
	vmConfig Config
	// global (to this context) ethereum virtual machine
	// used throughout the execution of the tx.
	interpreter *Interpreter
	// abort is used to abort the EVM calling operations
	// NOTE: must be set atomically
	abort int32
}

// NewEVM retutrns a new EVM evmironment. The returned EVM is not thread safe
// and should only ever be used *once*.
func NewEVM(ctx Context, statedb StateDB, chainConfig *params.ChainConfig, vmConfig Config) *EVM {
	evm := &EVM{
		Context:     ctx,
		StateDB:     statedb,
		vmConfig:    vmConfig,
		chainConfig: chainConfig,
		chainRules:  chainConfig.Rules(ctx.BlockNumber),
	}

	evm.interpreter = NewInterpreter(evm, vmConfig)
	return evm
}

// Cancel cancels any running EVM operation. This may be called concurrently and
// it's safe to be called multiple times.
func (evm *EVM) Cancel() {
	atomic.StoreInt32(&evm.abort, 1)
}

// Call executes the contract associated with the addr with the given input as parameters. It also handles any
// necessary value transfer required and takes the necessary steps to create accounts and reverses the state in
// case of an execution error or failed value transfer.
func (evm *EVM) Call(caller ContractRef, addr common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, leftOverGas uint64, err error) {
	if caller == nil {
		fmt.Println("caller is nil")
	}
	defer func() { // ??????????????????defer????????????????????????panic??????
		if err := recover(); err != nil {
			Println("error happened in Evm.Call")
			Printf("%v", err) // ?????????err????????????panic???????????????
			for i := 0; i < 10; i++ {
				funcName, file, line, ok := runtime.Caller(i)
				if ok {
					Printf("frame %v:[func:%v,file:%v,line:%v]\n", i, runtime.FuncForPC(funcName).Name(), file, line)
				}
			}
		}
	}()

	if evm.vmConfig.NoRecursion && evm.depth > 0 {
		return nil, gas, nil
	}

	// Depth check execution. Fail if we're trying to execute above the
	// limit.
	if evm.depth > int(params.CallCreateDepth) {
		return nil, gas, ErrDepth
	}

	if !evm.Context.CanTransfer(evm.StateDB, caller.Address(), value) {
		return nil, gas, ErrInsufficientBalance
	}

	var (
		to       = AccountRef(addr)
		snapshot = evm.StateDB.Snapshot()
	)
	if !evm.StateDB.Exist(addr) {
		if PrecompiledContracts[addr] == nil && evm.ChainConfig().IsEIP158(evm.BlockNumber) && value.Sign() == 0 {
			return nil, gas, nil
		}
		evm.StateDB.CreateAccount(addr)
	}
	evm.Transfer(evm.StateDB, caller.Address(), to.Address(), value)

	// initialise a new contract and set the code that is to be used by the
	// E The contract is a scoped evmironment for this execution context
	// only.
	contract := NewContract(caller, to, value, gas)
	contract.SetCallCode(&addr, evm.StateDB.GetCodeHash(addr), evm.StateDB.GetCode(addr))

	if caller != nil && !isRelOracle(contract.Address()) {
		/***
		*record Call action.And create HackerContractCall object
		*then push the object to the stack.
		**/
		hacker_init(evm, contract, input)
		if hacker_call_stack == nil {
			Println("call stack is nil")
			return
		}
		call := hacker_call_stack.peek()
		if call == nil {
			Println("call is nil")
			return
		}
		nextCall := call.OnCall(caller, contract.Address(), *value, *new(big.Int).SetUint64(gas), input)
		nextCall.snapshotId = snapshot
		if nextCall == nil {
			Println("nextcall is nil")
			return
		}
		hacker_call_stack.push(nextCall)
		Printf("\npush call@%p into stack", nextCall)
	}

	ret, err = run(evm, snapshot, contract, input)
	// When an error was returned by the EVM or when setting the creation code
	// above we revert to the snapshot and consume any gas remaining. Additionally
	// when we're in homestead this also counts for code storage gas errors.
	nextRevisionId := evm.StateDB.GetNextRevisionId()
	if err != nil {
		contract.UseGas(contract.Gas)
		evm.StateDB.RevertToSnapshot(snapshot)
		nextRevisionId = snapshot
	}

	if caller != nil && !isRelOracle(contract.Address()) {
		Println("\nclose call...")
		/**
		*Call action finish. So pop the object on top of the stack.
		*and set object to "close" state, also record final related state.
		**/
		if hacker_call_stack != nil {
			call := hacker_call_stack.pop()
			call.nextRevisionId = nextRevisionId
			if err != nil {
				call.throwException = true
				if strings.EqualFold(ErrOutOfGas.Error(), err.Error()) {
					call.errOutGas = true
				}
			}
			if call == nil {
				Println("call is nil")
				return
			}
			call.OnCloseCall(*new(big.Int).SetUint64(contract.Gas))
			if hacker_call_stack.len() == 1 {
				hacker_close()
			}
		}
	}
	return ret, contract.Gas, err
}

// CallCode executes the contract associated with the addr with the given input as parameters. It also handles any
// necessary value transfer required and takes the necessary steps to create accounts and reverses the state in
// case of an execution error or failed value transfer.
//
// CallCode differs from Call in the sense that it executes the given address' code with the caller as context.
func (evm *EVM) CallCode(caller ContractRef, addr common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, leftOverGas uint64, err error) {
	defer func() { // ??????????????????defer????????????????????????panic??????
		if err := recover(); err != nil {
			Println("CallCode")
			Println(err) // ?????????err????????????panic??????????????????55
		}
	}()

	if evm.vmConfig.NoRecursion && evm.depth > 0 {
		return nil, gas, nil
	}

	// Depth check execution. Fail if we're trying to execute above the
	// limit.
	if evm.depth > int(params.CallCreateDepth) {
		return nil, gas, ErrDepth
	}
	if !evm.CanTransfer(evm.StateDB, caller.Address(), value) {
		return nil, gas, ErrInsufficientBalance
	}

	var (
		snapshot       = evm.StateDB.Snapshot()
		to             = AccountRef(caller.Address())
		nextRevisionId = snapshot
	)
	// initialise a new contract and set the code that is to be used by the
	// E The contract is a scoped evmironment for this execution context
	// only.
	contract := NewContract(caller, to, value, gas)
	contract.SetCallCode(&addr, evm.StateDB.GetCodeHash(addr), evm.StateDB.GetCode(addr))

	if caller != nil && !isRelOracle(contract.Address()) {
		/***
		*record CallCode action.And create HackerContractCall object
		*then push the object to the stack.
		**/
		hacker_init(evm, contract, input)
		if hacker_call_stack == nil {
			Println("call stack is nil")
			return
		}
		call := hacker_call_stack.peek()
		if call == nil {
			Println("call is nil")
			return
		}
		nextCall := call.OnCallCode(caller, contract.Address(), *value, *new(big.Int).SetUint64(gas), input)
		nextCall.snapshotId = snapshot
		//nextCall.nextRevisionId = evm.StateDB.GetNextRevisonId()
		if nextCall == nil {
			Println("nextcall is nil")
			return
		}
		hacker_call_stack.push(nextCall)
		Printf("\npush call@%p into stack", nextCall)
		/***/
	}

	ret, err = run(evm, snapshot, contract, input)
	if err != nil {
		contract.UseGas(contract.Gas)
		evm.StateDB.RevertToSnapshot(snapshot)
	}
	if caller != nil && !isRelOracle(contract.Address()) {
		Println("\nclose call...")
		// subcriber.Close()
		// subcriber.Write()
		/**
		*CallCode action finish. So pop the object on top of the stack.
		*and set object to "close" state, also record final related state.
		**/
		if hacker_call_stack != nil {
			call := hacker_call_stack.pop()
			call.nextRevisionId = nextRevisionId
			if err != nil {
				call.throwException = true
			}
			if call == nil {
				Println("call is nil")
				return
			}
			call.OnCloseCall(*new(big.Int).SetUint64(contract.Gas))
			if hacker_call_stack.len() == 1 {
				hacker_close()
			}
		}
	}
	return ret, contract.Gas, err
}

// DelegateCall executes the contract associated with the addr with the given input as parameters.
// It reverses the state in case of an execution error.
//
// DelegateCall differs from CallCode in the sense that it executes the given address' code with the caller as context
// and the caller is set to the caller of the caller.
func (evm *EVM) DelegateCall(caller ContractRef, addr common.Address, input []byte, gas uint64) (ret []byte, leftOverGas uint64, err error) {
	defer func() { // ??????????????????defer????????????????????????panic??????
		if err := recover(); err != nil {
			Println("DelegateCall")
			Println(err) // ?????????err????????????panic??????????????????55
		}
	}()

	if evm.vmConfig.NoRecursion && evm.depth > 0 {
		return nil, gas, nil
	}

	// Depth check execution. Fail if we're trying to execute above the
	// limit.
	if evm.depth > int(params.CallCreateDepth) {
		return nil, gas, ErrDepth
	}

	var (
		snapshot       = evm.StateDB.Snapshot()
		to             = AccountRef(caller.Address())
		nextRevisionId = snapshot
	)

	// Iinitialise a new contract and make initialise the delegate values
	contract := NewContract(caller, to, nil, gas).AsDelegate()
	contract.SetCallCode(&addr, evm.StateDB.GetCodeHash(addr), evm.StateDB.GetCode(addr))

	if caller != nil && !isRelOracle(contract.Address()) {
		/***
		*record DelegateCall action.And create HackerContractCall object
		*then push the object to the stack.
		**/
		hacker_init(evm, contract, input)
		if hacker_call_stack == nil {
			Println("call stack is nil")
			return
		}
		call := hacker_call_stack.peek()
		if call == nil {
			Println("call is nil")
			return
		}
		nextCall := call.OnDelegateCall(caller, contract.Address(), *new(big.Int).SetUint64(gas), input)
		nextCall.snapshotId = snapshot
		//nextCall.nextRevisionId = evm.StateDB.GetNextRevisonId()
		if nextCall == nil {
			Println("nextcall is nil")
			return
		}
		hacker_call_stack.push(nextCall)
		Printf("\npush call@%p into stack", nextCall)
	}
	ret, err = run(evm, snapshot, contract, input)
	if err != nil {
		contract.UseGas(contract.Gas)
		evm.StateDB.RevertToSnapshot(snapshot)
	}
	if caller != nil && !isRelOracle(contract.Address()) {
		Println("\nclose call...")
		// subcriber.Close()
		// subcriber.Write()
		/**
		*DeletegateCall action finish. So pop the object on top of the stack.
		*and set object to "close" state, also record final related state.
		**/
		if hacker_call_stack != nil {
			call := hacker_call_stack.pop()
			call.nextRevisionId = nextRevisionId
			if err != nil {
				call.throwException = true
			}
			if call == nil {
				Println("call is nil")
				return
			}
			call.OnCloseCall(*new(big.Int).SetUint64(contract.Gas))
			if hacker_call_stack.len() == 1 {
				hacker_close()
			}
		}
	}

	return ret, contract.Gas, err
}

// Create creates a new contract using code as deployment code.
func (evm *EVM) Create(caller ContractRef, code []byte, gas uint64, value *big.Int) (ret []byte, contractAddr common.Address, leftOverGas uint64, err error) {
	if evm.vmConfig.NoRecursion && evm.depth > 0 {
		return nil, common.Address{}, gas, nil
	}

	// Depth check execution. Fail if we're trying to execute above the
	// limit.
	if evm.depth > int(params.CallCreateDepth) {
		return nil, common.Address{}, gas, ErrDepth
	}
	if !evm.CanTransfer(evm.StateDB, caller.Address(), value) {
		return nil, common.Address{}, gas, ErrInsufficientBalance
	}

	// Create a new account on the state
	nonce := evm.StateDB.GetNonce(caller.Address())
	evm.StateDB.SetNonce(caller.Address(), nonce+1)

	snapshot := evm.StateDB.Snapshot()
	contractAddr = crypto.CreateAddress(caller.Address(), nonce)
	evm.StateDB.CreateAccount(contractAddr)
	if evm.ChainConfig().IsEIP158(evm.BlockNumber) {
		evm.StateDB.SetNonce(contractAddr, 1)
	}
	evm.Transfer(evm.StateDB, caller.Address(), contractAddr, value)

	// initialise a new contract and set the code that is to be used by the
	// E The contract is a scoped evmironment for this execution context
	// only.
	contract := NewContract(caller, AccountRef(contractAddr), value, gas)
	contract.SetCallCode(&contractAddr, crypto.Keccak256Hash(code), code)

	ret, err = run(evm, snapshot, contract, nil)
	// check whether the max code size has been exceeded
	maxCodeSizeExceeded := len(ret) > params.MaxCodeSize
	// if the contract creation ran successfully and no errors were returned
	// calculate the gas required to store the code. If the code could not
	// be stored due to not enough gas set an error and let it be handled
	// by the error checking condition below.
	if err == nil && !maxCodeSizeExceeded {
		createDataGas := uint64(len(ret)) * params.CreateDataGas
		if contract.UseGas(createDataGas) {
			evm.StateDB.SetCode(contractAddr, ret)
		} else {
			err = ErrCodeStoreOutOfGas
		}
	}

	// When an error was returned by the EVM or when setting the creation code
	// above we revert to the snapshot and consume any gas remaining. Additionally
	// when we're in homestead this also counts for code storage gas errors.
	if maxCodeSizeExceeded ||
		(err != nil && (evm.ChainConfig().IsHomestead(evm.BlockNumber) || err != ErrCodeStoreOutOfGas)) {
		contract.UseGas(contract.Gas)
		evm.StateDB.RevertToSnapshot(snapshot)
	}
	// If the vm returned with an error the return value should be set to nil.
	// This isn't consensus critical but merely to for behaviour reasons such as
	// tests, RPC calls, etc.
	if err != nil {
		ret = nil
	}

	return ret, contractAddr, contract.Gas, err
}

// ChainConfig returns the evmironment's chain configuration
func (evm *EVM) ChainConfig() *params.ChainConfig { return evm.chainConfig }

// Interpreter returns the EVM interpreter
func (evm *EVM) Interpreter() *Interpreter { return evm.interpreter }
