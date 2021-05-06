// Copyright 2015 The go-ethereum Authors
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
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"golang.org/x/crypto/sha3"

	// stage1-substate: import state
	"github.com/ethereum/go-ethereum/core/state"
)

func opAdd(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()
	y.Add(&x, y)
	callContext.nstack.createOpComponent(2, ADD, *pc, cost)
	return nil, nil
}

func opSub(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()
	y.Sub(&x, y)
	callContext.nstack.createOpComponent(2, SUB, *pc, cost)
	return nil, nil
}

func opMul(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()
	y.Mul(&x, y)
	callContext.nstack.createOpComponent(2, MUL, *pc, cost)
	return nil, nil
}

func opDiv(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()
	y.Div(&x, y)
	callContext.nstack.createOpComponent(2, DIV, *pc, cost)
	return nil, nil
}

func opSdiv(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()
	y.SDiv(&x, y)
	callContext.nstack.createOpComponent(2, SDIV, *pc, cost)
	return nil, nil
}

func opMod(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()
	y.Mod(&x, y)
	callContext.nstack.createOpComponent(2, MOD, *pc, cost)
	return nil, nil
}

func opSmod(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()
	y.SMod(&x, y)
	callContext.nstack.createOpComponent(2, SMOD, *pc, cost)
	return nil, nil
}

func opExp(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	base, exponent := callContext.stack.pop(), callContext.stack.peek()
	exponent.Exp(&base, exponent)
	callContext.nstack.createOpComponent(2, EXP, *pc, cost)
	return nil, nil
}

func opSignExtend(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	back, num := callContext.stack.pop(), callContext.stack.peek()
	num.ExtendSign(num, &back)
	callContext.nstack.createOpComponent(2, SIGNEXTEND, *pc, cost)
	return nil, nil
}

func opNot(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	x := callContext.stack.peek()
	x.Not(x)
	callContext.nstack.createOpComponent(1, NOT, *pc, cost)
	return nil, nil
}

func opLt(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()
	if x.Lt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	callContext.nstack.createOpComponent(2, LT, *pc, cost)
	return nil, nil
}

func opGt(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()
	if x.Gt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	callContext.nstack.createOpComponent(2, GT, *pc, cost)
	return nil, nil
}

func opSlt(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()
	if x.Slt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	callContext.nstack.createOpComponent(2, SLT, *pc, cost)
	return nil, nil
}

func opSgt(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()
	if x.Sgt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	callContext.nstack.createOpComponent(2, SGT, *pc, cost)
	return nil, nil
}

func opEq(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()
	if x.Eq(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	callContext.nstack.createOpComponent(2, EQ, *pc, cost)
	return nil, nil
}

func opIszero(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	x := callContext.stack.peek()
	if x.IsZero() {
		x.SetOne()
	} else {
		x.Clear()
	}
	callContext.nstack.createOpComponent(1, ISZERO, *pc, cost)
	return nil, nil
}

func opAnd(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()
	y.And(&x, y)
	callContext.nstack.createOpComponent(2, AND, *pc, cost)
	return nil, nil
}

func opOr(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()
	y.Or(&x, y)
	callContext.nstack.createOpComponent(2, OR, *pc, cost)
	return nil, nil
}

func opXor(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	x, y := callContext.stack.pop(), callContext.stack.peek()
	y.Xor(&x, y)
	callContext.nstack.createOpComponent(2, XOR, *pc, cost)
	return nil, nil
}

func opByte(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	th, val := callContext.stack.pop(), callContext.stack.peek()
	val.Byte(&th)
	callContext.nstack.createOpComponent(2, BYTE, *pc, cost)
	return nil, nil
}

func opAddmod(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	x, y, z := callContext.stack.pop(), callContext.stack.pop(), callContext.stack.peek()
	if z.IsZero() {
		z.Clear()
	} else {
		z.AddMod(&x, &y, z)
	}
	callContext.nstack.createOpComponent(3, ADDMOD, *pc, cost)
	return nil, nil
}

func opMulmod(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	x, y, z := callContext.stack.pop(), callContext.stack.pop(), callContext.stack.peek()
	z.MulMod(&x, &y, z)
	callContext.nstack.createOpComponent(3, MULMOD, *pc, cost)
	return nil, nil
}

// opSHL implements Shift Left
// The SHL instruction (shift left) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the left by arg1 number of bits.
func opSHL(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	// Note, second operand is left in the stack; accumulate result into it, and no need to push it afterwards
	shift, value := callContext.stack.pop(), callContext.stack.peek()
	if shift.LtUint64(256) {
		value.Lsh(value, uint(shift.Uint64()))
	} else {
		value.Clear()
	}
	callContext.nstack.createOpComponent(2, SHL, *pc, cost)
	return nil, nil
}

// opSHR implements Logical Shift Right
// The SHR instruction (logical shift right) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the right by arg1 number of bits with zero fill.
func opSHR(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	// Note, second operand is left in the stack; accumulate result into it, and no need to push it afterwards
	shift, value := callContext.stack.pop(), callContext.stack.peek()
	if shift.LtUint64(256) {
		value.Rsh(value, uint(shift.Uint64()))
	} else {
		value.Clear()
	}
	callContext.nstack.createOpComponent(2, SHR, *pc, cost)
	return nil, nil
}

// opSAR implements Arithmetic Shift Right
// The SAR instruction (arithmetic shift right) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the right by arg1 number of bits with sign extension.
func opSAR(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	shift, value := callContext.stack.pop(), callContext.stack.peek()
	if shift.GtUint64(256) {
		if value.Sign() >= 0 {
			value.Clear()
		} else {
			// Max negative shift: all bits set
			value.SetAllOne()
		}
		return nil, nil
	}
	n := uint(shift.Uint64())
	value.SRsh(value, n)
	callContext.nstack.createOpComponent(2, SAR, *pc, cost)
	return nil, nil
}

func opSha3(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	offset, size := callContext.stack.pop(), callContext.stack.peek()
	data := callContext.memory.GetPtr(int64(offset.Uint64()), int64(size.Uint64()))
	tmpSize := *size

	if interpreter.hasher == nil {
		interpreter.hasher = sha3.NewLegacyKeccak256().(keccakState)
	} else {
		interpreter.hasher.Reset()
	}
	interpreter.hasher.Write(data)
	interpreter.hasher.Read(interpreter.hasherBuf[:])

	evm := interpreter.evm
	if evm.vmConfig.EnablePreimageRecording {
		evm.StateDB.AddPreimage(interpreter.hasherBuf, data)
	}

	size.SetBytes(interpreter.hasherBuf[:])
	callContext.nstack.createOpComponent(2, SHA3, *pc, cost)
	mNodes := callContext.mtracer.load(int64(offset.Uint64()), int64(tmpSize.Uint64()))
	for _, mNode := range mNodes {
		callContext.nstack.createEdge(mNode, callContext.nstack.peek())
	}
	return nil, nil
}
func opAddress(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	callContext.stack.push(new(uint256.Int).SetBytes(callContext.contract.Address().Bytes()))
	nd := callContext.nstack.createValue(callContext.contract.Address().String())
	callContext.nstack.createValueOp(ADDRESS, *pc, cost, nd)
	return nil, nil
}

func opBalance(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	slot := callContext.stack.peek()
	address := common.Address(slot.Bytes20())
	slot.SetFromBig(interpreter.evm.StateDB.GetBalance(address))
	callContext.nstack.createOpComponent(1, BALANCE, *pc, cost)
	return nil, nil
}

func opOrigin(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	callContext.stack.push(new(uint256.Int).SetBytes(interpreter.evm.Origin.Bytes()))
	nd := callContext.nstack.createValue(interpreter.evm.Origin.String())
	callContext.nstack.createValueOp(ORIGIN, *pc, cost, nd)
	return nil, nil
}
func opCaller(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	callContext.stack.push(new(uint256.Int).SetBytes(callContext.contract.Caller().Bytes()))
	nd := callContext.nstack.createValue(callContext.contract.Caller().String())
	callContext.nstack.createValueOp(CALLER, *pc, cost, nd)
	return nil, nil
}

func opCallValue(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	v, _ := uint256.FromBig(callContext.contract.value)
	callContext.stack.push(v)
	nd := callContext.nstack.createValue(v.String())
	callContext.nstack.createValueOp(CALLVALUE, *pc, cost, nd)
	return nil, nil
}

func opCallDataLoad(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	x := callContext.stack.peek()
	if offset, overflow := x.Uint64WithOverflow(); !overflow {
		data := getData(callContext.contract.Input, offset, 32)
		x.SetBytes(data)
	} else {
		x.Clear()
	}
	callContext.nstack.createOpComponent(1, CALLDATALOAD, *pc, cost)
	return nil, nil
}

func opCallDataSize(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	callContext.stack.push(new(uint256.Int).SetUint64(uint64(len(callContext.contract.Input))))
	nd := callContext.nstack.createValue(new(uint256.Int).SetUint64(uint64(len(callContext.contract.Input))).String())
	callContext.nstack.createValueOp(CALLDATASIZE, *pc, cost, nd)
	return nil, nil
}

func opCallDataCopy(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	var (
		memOffset  = callContext.stack.pop()
		dataOffset = callContext.stack.pop()
		length     = callContext.stack.pop()
	)
	dataOffset64, overflow := dataOffset.Uint64WithOverflow()
	if overflow {
		dataOffset64 = 0xffffffffffffffff
	}
	// These values are checked for overflow during gas cost calculation
	memOffset64 := memOffset.Uint64()
	length64 := length.Uint64()
	callContext.memory.Set(memOffset64, length64, getData(callContext.contract.Input, dataOffset64, length64))
	callContext.nstack.createOpComponent(3, CALLDATACOPY, *pc, cost)
	node := callContext.nstack.pop()
	callContext.mtracer.store(node, int64(memOffset.Uint64()), int64(length.Uint64()))
	return nil, nil
}

func opReturnDataSize(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	callContext.stack.push(new(uint256.Int).SetUint64(uint64(len(interpreter.returnData))))
	nd := callContext.nstack.createValue(new(uint256.Int).SetUint64(uint64(len(interpreter.returnData))).String())
	callContext.nstack.createValueOp(RETURNDATASIZE, *pc, cost, nd)
	return nil, nil
}

func opReturnDataCopy(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	var (
		memOffset  = callContext.stack.pop()
		dataOffset = callContext.stack.pop()
		length     = callContext.stack.pop()
	)

	offset64, overflow := dataOffset.Uint64WithOverflow()
	if overflow {
		return nil, ErrReturnDataOutOfBounds
	}
	// we can reuse dataOffset now (aliasing it for clarity)
	var end = dataOffset
	end.Add(&dataOffset, &length)
	end64, overflow := end.Uint64WithOverflow()
	if overflow || uint64(len(interpreter.returnData)) < end64 {
		return nil, ErrReturnDataOutOfBounds
	}
	callContext.memory.Set(memOffset.Uint64(), length.Uint64(), interpreter.returnData[offset64:end64])
	callContext.nstack.createOpComponent(3, RETURNDATACOPY, *pc, cost)
	node := callContext.nstack.pop()
	callContext.mtracer.store(node, int64(memOffset.Uint64()), int64(length.Uint64()))
	return nil, nil
}

func opExtCodeSize(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	slot := callContext.stack.peek()
	slot.SetUint64(uint64(interpreter.evm.StateDB.GetCodeSize(common.Address(slot.Bytes20()))))
	callContext.nstack.createOpComponent(1, EXTCODESIZE, *pc, cost)
	return nil, nil
}

func opCodeSize(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	l := new(uint256.Int)
	l.SetUint64(uint64(len(callContext.contract.Code)))
	callContext.stack.push(l)
	nd := callContext.nstack.createValue(l.String())
	callContext.nstack.createValueOp(CODESIZE, *pc, cost, nd)
	return nil, nil
}

func opCodeCopy(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	var (
		memOffset  = callContext.stack.pop()
		codeOffset = callContext.stack.pop()
		length     = callContext.stack.pop()
	)
	uint64CodeOffset, overflow := codeOffset.Uint64WithOverflow()
	if overflow {
		uint64CodeOffset = 0xffffffffffffffff
	}
	codeCopy := getData(callContext.contract.Code, uint64CodeOffset, length.Uint64())
	callContext.memory.Set(memOffset.Uint64(), length.Uint64(), codeCopy)

	callContext.nstack.createOpComponent(3, CODECOPY, *pc, cost)
	node := callContext.nstack.pop()
	callContext.mtracer.store(node, int64(memOffset.Uint64()), int64(length.Uint64()))
	return nil, nil
}

func opExtCodeCopy(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	var (
		stack      = callContext.stack
		a          = stack.pop()
		memOffset  = stack.pop()
		codeOffset = stack.pop()
		length     = stack.pop()
	)
	uint64CodeOffset, overflow := codeOffset.Uint64WithOverflow()
	if overflow {
		uint64CodeOffset = 0xffffffffffffffff
	}
	addr := common.Address(a.Bytes20())
	codeCopy := getData(interpreter.evm.StateDB.GetCode(addr), uint64CodeOffset, length.Uint64())
	callContext.memory.Set(memOffset.Uint64(), length.Uint64(), codeCopy)

	callContext.nstack.createOpComponent(4, EXTCODECOPY, *pc, cost)
	node := callContext.nstack.pop()
	callContext.mtracer.store(node, int64(memOffset.Uint64()), int64(length.Uint64()))
	return nil, nil
}

// opExtCodeHash returns the code hash of a specified account.
// There are several cases when the function is called, while we can relay everything
// to `state.GetCodeHash` function to ensure the correctness.
//   (1) Caller tries to get the code hash of a normal contract account, state
// should return the relative code hash and set it as the result.
//
//   (2) Caller tries to get the code hash of a non-existent account, state should
// return common.Hash{} and zero will be set as the result.
//
//   (3) Caller tries to get the code hash for an account without contract code,
// state should return emptyCodeHash(0xc5d246...) as the result.
//
//   (4) Caller tries to get the code hash of a precompiled account, the result
// should be zero or emptyCodeHash.
//
// It is worth noting that in order to avoid unnecessary create and clean,
// all precompile accounts on mainnet have been transferred 1 wei, so the return
// here should be emptyCodeHash.
// If the precompile account is not transferred any amount on a private or
// customized chain, the return value will be zero.
//
//   (5) Caller tries to get the code hash for an account which is marked as suicided
// in the current transaction, the code hash of this account should be returned.
//
//   (6) Caller tries to get the code hash for an account which is marked as deleted,
// this account should be regarded as a non-existent account and zero should be returned.
func opExtCodeHash(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	slot := callContext.stack.peek()
	address := common.Address(slot.Bytes20())
	if interpreter.evm.StateDB.Empty(address) {
		slot.Clear()
	} else {
		slot.SetBytes(interpreter.evm.StateDB.GetCodeHash(address).Bytes())
	}
	callContext.nstack.createOpComponent(1, EXTCODEHASH, *pc, cost)
	return nil, nil
}

func opGasprice(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	v, _ := uint256.FromBig(interpreter.evm.GasPrice)
	callContext.stack.push(v)
	nd := callContext.nstack.createValue(v.String())
	callContext.nstack.createValueOp(GASPRICE, *pc, cost, nd)
	return nil, nil
}

func opBlockhash(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	num := callContext.stack.peek()
	num64, overflow := num.Uint64WithOverflow()

	// stage1-substate: convert vm.StateDB to state.StateDB and save block hash
	defer func() {
		statedb, ok := interpreter.evm.StateDB.(*state.StateDB)
		if ok {
			statedb.ResearchBlockHashes[num64] = common.BytesToHash(num.Bytes())
		}
	}()

	if overflow {
		num.Clear()
		return nil, nil
	}
	var upper, lower uint64
	upper = interpreter.evm.BlockNumber.Uint64()
	if upper < 257 {
		lower = 0
	} else {
		lower = upper - 256
	}
	if num64 >= lower && num64 < upper {
		num.SetBytes(interpreter.evm.GetHash(num64).Bytes())
	} else {
		num.Clear()
	}
	callContext.nstack.createOpComponent(1, BLOCKHASH, *pc, cost)
	return nil, nil
}

func opCoinbase(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	callContext.stack.push(new(uint256.Int).SetBytes(interpreter.evm.Coinbase.Bytes()))
	nd := callContext.nstack.createValue(new(uint256.Int).SetBytes(interpreter.evm.Coinbase.Bytes()).String())
	callContext.nstack.createValueOp(COINBASE, *pc, cost, nd)
	return nil, nil
}

func opTimestamp(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	v, _ := uint256.FromBig(interpreter.evm.Time)
	callContext.stack.push(v)
	nd := callContext.nstack.createValue(v.String())
	callContext.nstack.createValueOp(TIMESTAMP, *pc, cost, nd)
	return nil, nil
}

func opNumber(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	v, _ := uint256.FromBig(interpreter.evm.BlockNumber)
	callContext.stack.push(v)
	nd := callContext.nstack.createValue(v.String())
	callContext.nstack.createValueOp(NUMBER, *pc, cost, nd)
	return nil, nil
}

func opDifficulty(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	v, _ := uint256.FromBig(interpreter.evm.Difficulty)
	callContext.stack.push(v)
	nd := callContext.nstack.createValue(v.String())
	callContext.nstack.createValueOp(DIFFICULTY, *pc, cost, nd)
	return nil, nil
}

func opGasLimit(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	callContext.stack.push(new(uint256.Int).SetUint64(interpreter.evm.GasLimit))
	nd := callContext.nstack.createValue(new(uint256.Int).SetUint64(interpreter.evm.GasLimit).String())
	callContext.nstack.createValueOp(GASLIMIT, *pc, cost, nd)
	return nil, nil
}

func opPop(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	callContext.stack.pop()
	callContext.nstack.createOpComponent(1, POP, *pc, cost)
	callContext.nstack.pop()
	return nil, nil
}

func opMload(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	v := callContext.stack.peek()
	offset := int64(v.Uint64())
	v.SetBytes(callContext.memory.GetPtr(offset, 32))
	callContext.nstack.createOpComponent(1, MLOAD, *pc, cost)
	mNodes := callContext.mtracer.load(offset, 32)
	for _, mNode := range mNodes {
		callContext.nstack.createEdge(mNode, callContext.nstack.peek())
	}
	return nil, nil
}

func opMstore(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	// pop value of the stack
	mStart, val := callContext.stack.pop(), callContext.stack.pop()
	callContext.memory.Set32(mStart.Uint64(), &val)
	callContext.nstack.createOpComponent(2, MSTORE, *pc, cost)
	node := callContext.nstack.pop()
	callContext.mtracer.store(node, int64(mStart.Uint64()), 32)
	return nil, nil
}

func opMstore8(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	off, val := callContext.stack.pop(), callContext.stack.pop()
	callContext.memory.store[off.Uint64()] = byte(val.Uint64())
	callContext.nstack.createOpComponent(2, MSTORE8, *pc, cost)
	node := callContext.nstack.pop()
	callContext.mtracer.store(node, int64(off.Uint64()), 1)
	return nil, nil
}

func opSload(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	loc := callContext.stack.peek()
	hash := common.Hash(loc.Bytes32())
	val := interpreter.evm.StateDB.GetState(callContext.contract.Address(), hash)
	loc.SetBytes(val.Bytes())
	callContext.nstack.createOpComponent(1, SLOAD, *pc, cost)
	return nil, nil
}

func opSstore(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	loc := callContext.stack.pop()
	val := callContext.stack.pop()
	interpreter.evm.StateDB.SetState(callContext.contract.Address(),
		common.Hash(loc.Bytes32()), common.Hash(val.Bytes32()))
	callContext.nstack.createOpComponent(2, SSTORE, *pc, cost)
	callContext.nstack.pop()
	return nil, nil
}

func opJump(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	pos := callContext.stack.pop()
	if !callContext.contract.validJumpdest(&pos) {
		return nil, ErrInvalidJump
	}
	callContext.nstack.createOpComponent(1, JUMP, *pc, cost)
	*pc = pos.Uint64()
	return nil, nil
}

func opJumpi(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	pos, cond := callContext.stack.pop(), callContext.stack.pop()
	callContext.nstack.createOpComponent(2, JUMPI, *pc, cost)
	if !cond.IsZero() {
		if !callContext.contract.validJumpdest(&pos) {
			return nil, ErrInvalidJump
		}
		*pc = pos.Uint64()
	} else {
		*pc++
	}
	return nil, nil
}

func opJumpdest(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	callContext.nstack.createOpComponent(0, JUMPDEST, *pc, cost)
	return nil, nil
}

func opBeginSub(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	return nil, ErrInvalidSubroutineEntry
}

func opJumpSub(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	if len(callContext.rstack.data) >= 1023 {
		return nil, ErrReturnStackExceeded
	}
	pos := callContext.stack.pop()
	if !pos.IsUint64() {
		return nil, ErrInvalidJump
	}
	posU64 := pos.Uint64()
	if !callContext.contract.validJumpSubdest(posU64) {
		return nil, ErrInvalidJump
	}
	callContext.rstack.push(uint32(*pc))
	*pc = posU64 + 1
	return nil, nil
}

func opReturnSub(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	if len(callContext.rstack.data) == 0 {
		return nil, ErrInvalidRetsub
	}
	// Other than the check that the return stack is not empty, there is no
	// need to validate the pc from 'returns', since we only ever push valid
	//values onto it via jumpsub.
	*pc = uint64(callContext.rstack.pop()) + 1
	return nil, nil
}

func opPc(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	callContext.stack.push(new(uint256.Int).SetUint64(*pc))
	nd := callContext.nstack.createValue(new(uint256.Int).SetUint64(*pc).String())
	callContext.nstack.createValueOp(PC, *pc, cost, nd)
	return nil, nil
}

func opMsize(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	callContext.stack.push(new(uint256.Int).SetUint64(uint64(callContext.memory.Len())))
	nd := callContext.nstack.createValue(new(uint256.Int).SetUint64(uint64(callContext.memory.Len())).String())
	callContext.nstack.createValueOp(MSIZE, *pc, cost, nd)
	return nil, nil
}

func opGas(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	callContext.stack.push(new(uint256.Int).SetUint64(callContext.contract.Gas))
	nd := callContext.nstack.createValue(new(uint256.Int).SetUint64(callContext.contract.Gas).String())
	callContext.nstack.createValueOp(GAS, *pc, cost, nd)
	return nil, nil
}

func opCreate(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	var (
		value        = callContext.stack.pop()
		offset, size = callContext.stack.pop(), callContext.stack.pop()
		input        = callContext.memory.GetCopy(int64(offset.Uint64()), int64(size.Uint64()))
		gas          = callContext.contract.Gas
	)
	if interpreter.evm.chainRules.IsEIP150 {
		gas -= gas / 64
	}
	// reuse size int for stackvalue
	stackvalue := size

	callContext.contract.UseGas(gas)
	//TODO: use uint256.Int instead of converting with toBig()
	var bigVal = big0
	if !value.IsZero() {
		bigVal = value.ToBig()
	}

	res, addr, returnGas, suberr := interpreter.evm.Create(callContext.contract, input, gas, bigVal)
	// Push item on the stack based on the returned error. If the ruleset is
	// homestead we must check for CodeStoreOutOfGasError (homestead only
	// rule) and treat as an error, if the ruleset is frontier we must
	// ignore this error and pretend the operation was successful.
	if interpreter.evm.chainRules.IsHomestead && suberr == ErrCodeStoreOutOfGas {
		stackvalue.Clear()
	} else if suberr != nil && suberr != ErrCodeStoreOutOfGas {
		stackvalue.Clear()
	} else {
		stackvalue.SetBytes(addr.Bytes())
	}
	callContext.stack.push(&stackvalue)
	callContext.contract.Gas += returnGas

	if suberr == ErrExecutionReverted {
		return res, nil
	}
	callContext.nstack.createOpComponent(3, CREATE, *pc, cost)
	mNodes := callContext.mtracer.load(int64(offset.Uint64()), int64(size.Uint64()))
	for _, mNode := range mNodes {
		callContext.nstack.createEdge(mNode, callContext.nstack.peek())
	}
	return nil, nil
}

func opCreate2(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	var (
		endowment    = callContext.stack.pop()
		offset, size = callContext.stack.pop(), callContext.stack.pop()
		salt         = callContext.stack.pop()
		input        = callContext.memory.GetCopy(int64(offset.Uint64()), int64(size.Uint64()))
		gas          = callContext.contract.Gas
	)

	// Apply EIP150
	gas -= gas / 64
	callContext.contract.UseGas(gas)
	// reuse size int for stackvalue
	stackvalue := size
	//TODO: use uint256.Int instead of converting with toBig()
	bigEndowment := big0
	if !endowment.IsZero() {
		bigEndowment = endowment.ToBig()
	}
	res, addr, returnGas, suberr := interpreter.evm.Create2(callContext.contract, input, gas,
		bigEndowment, &salt)
	// Push item on the stack based on the returned error.
	if suberr != nil {
		stackvalue.Clear()
	} else {
		stackvalue.SetBytes(addr.Bytes())
	}
	callContext.stack.push(&stackvalue)
	callContext.contract.Gas += returnGas

	if suberr == ErrExecutionReverted {
		return res, nil
	}
	callContext.nstack.createOpComponent(4, CREATE2, *pc, cost)
	mNodes := callContext.mtracer.load(int64(offset.Uint64()), int64(size.Uint64()))
	for _, mNode := range mNodes {
		callContext.nstack.createEdge(mNode, callContext.nstack.peek())
	}
	return nil, nil
}

func opCall(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	stack := callContext.stack
	// Pop gas. The actual gas in interpreter.evm.callGasTemp.
	// We can use this as a temporary value
	temp := stack.pop()
	gas := interpreter.evm.callGasTemp
	// Pop other call parameters.
	addr, value, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := common.Address(addr.Bytes20())
	// Get the arguments from the memory.
	args := callContext.memory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))

	var bigVal = big0
	//TODO: use uint256.Int instead of converting with toBig()
	// By using big0 here, we save an alloc for the most common case (non-ether-transferring contract calls),
	// but it would make more sense to extend the usage of uint256.Int
	if !value.IsZero() {
		gas += params.CallStipend
		bigVal = value.ToBig()
	}

	ret, returnGas, err := interpreter.evm.Call(callContext.contract, toAddr, args, gas, bigVal)

	if err != nil {
		temp.Clear()
	} else {
		temp.SetOne()
	}
	stack.push(&temp)
	if err == nil || err == ErrExecutionReverted {
		callContext.memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	callContext.contract.Gas += returnGas
	callContext.nstack.createOpComponent(7, CALL, *pc, cost)

	thisNode := callContext.nstack.peek()
	mNodes := callContext.mtracer.load(int64(inOffset.Uint64()), int64(inSize.Uint64()))

	for _, mNode := range mNodes {
		callContext.nstack.createEdge(mNode, thisNode)
	}
	callContext.mtracer.store(thisNode, int64(retOffset.Uint64()), int64(retSize.Uint64()))
	return ret, nil
}

func opCallCode(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	// Pop gas. The actual gas is in interpreter.evm.callGasTemp.
	stack := callContext.stack
	// We use it as a temporary value
	temp := stack.pop()
	gas := interpreter.evm.callGasTemp
	// Pop other call parameters.
	addr, value, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := common.Address(addr.Bytes20())
	// Get arguments from the memory.
	args := callContext.memory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))

	//TODO: use uint256.Int instead of converting with toBig()
	var bigVal = big0
	if !value.IsZero() {
		gas += params.CallStipend
		bigVal = value.ToBig()
	}

	ret, returnGas, err := interpreter.evm.CallCode(callContext.contract, toAddr, args, gas, bigVal)
	if err != nil {
		temp.Clear()
	} else {
		temp.SetOne()
	}
	stack.push(&temp)
	if err == nil || err == ErrExecutionReverted {
		callContext.memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	callContext.contract.Gas += returnGas
	callContext.nstack.createOpComponent(7, CALLCODE, *pc, cost)

	thisNode := callContext.nstack.peek()
	mNodes := callContext.mtracer.load(int64(inOffset.Uint64()), int64(inSize.Uint64()))
	for _, mNode := range mNodes {
		callContext.nstack.createEdge(mNode, thisNode)
	}
	callContext.mtracer.store(thisNode, int64(retOffset.Uint64()), int64(retSize.Uint64()))
	return ret, nil
}

func opDelegateCall(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	stack := callContext.stack
	// Pop gas. The actual gas is in interpreter.evm.callGasTemp.
	// We use it as a temporary value
	temp := stack.pop()
	gas := interpreter.evm.callGasTemp
	// Pop other call parameters.
	addr, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := common.Address(addr.Bytes20())
	// Get arguments from the memory.
	args := callContext.memory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))

	ret, returnGas, err := interpreter.evm.DelegateCall(callContext.contract, toAddr, args, gas)
	if err != nil {
		temp.Clear()
	} else {
		temp.SetOne()
	}
	stack.push(&temp)
	if err == nil || err == ErrExecutionReverted {
		callContext.memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	callContext.contract.Gas += returnGas
	callContext.nstack.createOpComponent(6, DELEGATECALL, *pc, cost)

	thisNode := callContext.nstack.peek()
	mNodes := callContext.mtracer.load(int64(inOffset.Uint64()), int64(inSize.Uint64()))
	for _, mNode := range mNodes {
		callContext.nstack.createEdge(mNode, thisNode)
	}
	callContext.mtracer.store(thisNode, int64(retOffset.Uint64()), int64(retSize.Uint64()))
	return ret, nil
}

func opStaticCall(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	// Pop gas. The actual gas is in interpreter.evm.callGasTemp.
	stack := callContext.stack
	// We use it as a temporary value
	temp := stack.pop()
	gas := interpreter.evm.callGasTemp
	// Pop other call parameters.
	addr, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := common.Address(addr.Bytes20())
	// Get arguments from the memory.
	args := callContext.memory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))

	ret, returnGas, err := interpreter.evm.StaticCall(callContext.contract, toAddr, args, gas)
	if err != nil {
		temp.Clear()
	} else {
		temp.SetOne()
	}
	stack.push(&temp)
	if err == nil || err == ErrExecutionReverted {
		callContext.memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	callContext.contract.Gas += returnGas
	callContext.nstack.createOpComponent(6, STATICCALL, *pc, cost)

	thisNode := callContext.nstack.peek()
	mNodes := callContext.mtracer.load(int64(inOffset.Uint64()), int64(inSize.Uint64()))
	for _, mNode := range mNodes {
		callContext.nstack.createEdge(mNode, thisNode)
	}
	callContext.mtracer.store(thisNode, int64(retOffset.Uint64()), int64(retSize.Uint64()))
	return ret, nil
}

func opReturn(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	offset, size := callContext.stack.pop(), callContext.stack.pop()
	ret := callContext.memory.GetPtr(int64(offset.Uint64()), int64(size.Uint64()))
	callContext.nstack.createOpComponent(2, RETURN, *pc, cost)
	thisNode := callContext.nstack.pop()
	mNodes := callContext.mtracer.load(int64(offset.Uint64()), int64(size.Uint64()))
	for _, mNode := range mNodes {
		callContext.nstack.createEdge(mNode, thisNode)
		callContext.nstack.stopFlag = true
	}
	return ret, nil
}

func opRevert(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	offset, size := callContext.stack.pop(), callContext.stack.pop()
	ret := callContext.memory.GetPtr(int64(offset.Uint64()), int64(size.Uint64()))
	callContext.nstack.createOpComponent(2, REVERT, *pc, cost)
	thisNode := callContext.nstack.pop()
	mNodes := callContext.mtracer.load(int64(offset.Uint64()), int64(size.Uint64()))
	for _, mNode := range mNodes {
		callContext.nstack.createEdge(mNode, thisNode)
	}
	return ret, nil
}

func opStop(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	return nil, nil
}

func opSuicide(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	beneficiary := callContext.stack.pop()
	balance := interpreter.evm.StateDB.GetBalance(callContext.contract.Address())
	interpreter.evm.StateDB.AddBalance(common.Address(beneficiary.Bytes20()), balance)
	interpreter.evm.StateDB.Suicide(callContext.contract.Address())
	callContext.nstack.createOpComponent(1, SELFDESTRUCT, *pc, cost)
	callContext.nstack.pop()
	return nil, nil
}

// following functions are used by the instruction jump  table

// make log instruction function
func makeLog(size int) executionFunc {
	return func(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
		topics := make([]common.Hash, size)
		stack := callContext.stack
		mStart, mSize := stack.pop(), stack.pop()
		for i := 0; i < size; i++ {
			addr := stack.pop()
			topics[i] = common.Hash(addr.Bytes32())
		}

		d := callContext.memory.GetCopy(int64(mStart.Uint64()), int64(mSize.Uint64()))
		interpreter.evm.StateDB.AddLog(&types.Log{
			Address: callContext.contract.Address(),
			Topics:  topics,
			Data:    d,
			// This is a non-consensus field, but assigned here because
			// core/state doesn't know the current block number.
			BlockNumber: interpreter.evm.BlockNumber.Uint64(),
		})

		callContext.nstack.createOpComponent(2+size, OpCode(byte(LOG0)+byte(size)), *pc, cost)
		node := callContext.nstack.pop()
		callContext.mtracer.store(node, int64(mStart.Uint64()), int64(mSize.Uint64()))
		return nil, nil
	}
}

// opPush1 is a specialized version of pushN
func opPush1(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
	var nd *ValueNode
	var (
		codeLen = uint64(len(callContext.contract.Code))
		integer = new(uint256.Int)
	)
	*pc += 1
	if *pc < codeLen {
		callContext.stack.push(integer.SetUint64(uint64(callContext.contract.Code[*pc])))
		nd = callContext.nstack.createValue(integer.SetUint64(uint64(callContext.contract.Code[*pc])).String())
	} else {
		callContext.stack.push(integer.Clear())
		nd = callContext.nstack.createValue(integer.Clear().String())
	}
	callContext.nstack.createValueOp(PUSH1, *pc-1, cost, nd)
	return nil, nil
}

// make push instruction function
func makePush(size uint64, pushByteSize int) executionFunc {
	return func(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
		codeLen := len(callContext.contract.Code)

		startMin := codeLen
		if int(*pc+1) < startMin {
			startMin = int(*pc + 1)
		}

		endMin := codeLen
		if startMin+pushByteSize < endMin {
			endMin = startMin + pushByteSize
		}

		integer := new(uint256.Int)
		callContext.stack.push(integer.SetBytes(common.RightPadBytes(
			callContext.contract.Code[startMin:endMin], pushByteSize)))
		nd := callContext.nstack.createValue(integer.SetBytes(common.RightPadBytes(
			callContext.contract.Code[startMin:endMin], pushByteSize)).String())
		callContext.nstack.createValueOp(PUSH, *pc, cost, nd)
		*pc += size
		return nil, nil
	}
}

// make dup instruction function
func makeDup(size int64) executionFunc {
	return func(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
		callContext.stack.dup(int(size))
		callContext.nstack.dup(int(size), *pc, cost)
		return nil, nil
	}
}

// make swap instruction function
func makeSwap(size int64) executionFunc {
	// switch n + 1 otherwise n would be swapped with n
	size++
	return func(pc *uint64, interpreter *EVMInterpreter, callContext *callCtx, cost uint64) ([]byte, error) {
		callContext.stack.swap(int(size))
		callContext.nstack.swap(int(size), *pc, cost)
		return nil, nil
	}
}
