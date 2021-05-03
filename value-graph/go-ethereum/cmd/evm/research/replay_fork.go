package research

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/cmd/evm/internal/t8ntool"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/research"
	"github.com/ethereum/go-ethereum/tests"
	cli "gopkg.in/urfave/cli.v1"
)

var (
	ErrReplayForkOutOfGas     = errors.New("out of gas in replay-fork")
	ErrReplayForkInvalidAlloc = errors.New("invalid alloc in replay-fork")
	ErrReplayForkMoreGas      = errors.New("more gas in replay-fork")
	ErrReplayForkLessGas      = errors.New("less gas in replay-fork")
	ErrReplayForkMisc         = errors.New("misc in replay-fork")
)

type ReplayForkStats map[string]int64

// stage1-substate: func ApplySubstateFork
func ApplySubstateFork(ctx *cli.Context, block uint64, tx int, substate *research.Substate, chainConfig *params.ChainConfig) error {
	inputAlloc := substate.InputAlloc
	inputEnv := substate.Env
	inputMessage := substate.Message

	outputAlloc := substate.OutputAlloc
	outputResult := substate.Result

	var (
		vmConfig    vm.Config
		pre         t8ntool.Prestate
		getTracerFn func(txIndex int) (tracer vm.Tracer, err error)
	)

	vmConfig = vm.Config{}

	// copy inputAlloc to Prestate.Pre
	pre.Pre = make(core.GenesisAlloc)
	for addr, account := range inputAlloc {
		ga := core.GenesisAccount{}
		ga.Nonce = account.Nonce
		ga.Balance = account.Balance
		ga.Storage = account.Storage
		ga.Code = account.Code
		pre.Pre[addr] = ga
	}
	// copy inputEnv to Prestate.Env
	pre.Env.Coinbase = inputEnv.Coinbase
	pre.Env.Difficulty = inputEnv.Difficulty
	pre.Env.GasLimit = inputEnv.GasLimit
	pre.Env.Number = inputEnv.Number
	pre.Env.Timestamp = inputEnv.Timestamp
	pre.Env.BlockHashes = make(map[math.HexOrDecimal64]common.Hash)
	for num64, bhash := range inputEnv.BlockHashes {
		pre.Env.BlockHashes[math.HexOrDecimal64(num64)] = bhash
	}

	getTracerFn = func(txIndex int) (tracer vm.Tracer, err error) {
		return nil, nil
	}

	var hashError error
	getHash := func(num uint64) common.Hash {
		if pre.Env.BlockHashes == nil {
			hashError = fmt.Errorf("getHash(%d) invoked, no blockhashes provided", num)
			return common.Hash{}
		}
		h, ok := pre.Env.BlockHashes[math.HexOrDecimal64(num)]
		if !ok {
			hashError = fmt.Errorf("getHash(%d) invoked, blockhash for that block not provided", num)
		}
		return h
	}

	// Apply Message
	var (
		statedb   = t8ntool.MakePreState(rawdb.NewMemoryDatabase(), pre.Pre)
		gaspool   = new(core.GasPool)
		blockHash = common.Hash{0x13, 0x37}
		txIndex   = tx
	)

	gaspool.AddGas(pre.Env.GasLimit)
	vmContext := vm.Context{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		Coinbase:    pre.Env.Coinbase,
		BlockNumber: new(big.Int).SetUint64(pre.Env.Number),
		Time:        new(big.Int).SetUint64(pre.Env.Timestamp),
		Difficulty:  pre.Env.Difficulty,
		GasLimit:    pre.Env.GasLimit,
		GetHash:     getHash,
		// GasPrice and Origin needs to be set per transaction
	}

	msg := inputMessage.AsMessage()

	tracer, err := getTracerFn(txIndex)
	if err != nil {
		return err
	}
	vmConfig.Tracer = tracer
	vmConfig.Debug = (tracer != nil)
	statedb.Prepare(common.Hash{}, blockHash, txIndex)
	vmContext.GasPrice = msg.GasPrice()
	vmContext.Origin = msg.From()

	evm := vm.NewEVM(vmContext, statedb, chainConfig, vmConfig)
	snapshot := statedb.Snapshot()
	msgResult, err := core.ApplyMessage(evm, msg, gaspool)

	if err != nil {
		statedb.RevertToSnapshot(snapshot)
		return err
	}

	// Ignore hashError
	// if hashError != nil {
	// 	return t8ntool.NewError(t8ntool.ErrorMissingBlockhash, hashError)
	// }

	if chainConfig.IsByzantium(vmContext.BlockNumber) {
		statedb.Finalise(true)
	} else {
		statedb.IntermediateRoot(chainConfig.IsEIP158(vmContext.BlockNumber))
	}

	evmResult := &research.SubstateResult{}
	if msgResult.Failed() {
		evmResult.Status = types.ReceiptStatusFailed
	} else {
		evmResult.Status = types.ReceiptStatusSuccessful
	}
	evmResult.Logs = statedb.GetLogs(common.Hash{})
	evmResult.Bloom = types.BytesToBloom(types.LogsBloom(evmResult.Logs).Bytes())
	if to := msg.To(); to == nil {
		evmResult.ContractAddress = crypto.CreateAddress(evm.Context.Origin, msg.Nonce())
	}
	evmResult.GasUsed = msgResult.UsedGas

	evmAlloc := statedb.ResearchPostAlloc

	if r, a := outputResult.Equal(evmResult), outputAlloc.Equal(evmAlloc); !(r && a) {
		if outputResult.Status == types.ReceiptStatusSuccessful &&
			evmResult.Status == types.ReceiptStatusSuccessful {
			// when both output and evm were successful, check alloc and gas usage

			// check account states
			if len(outputAlloc) != len(evmAlloc) {
				return ErrReplayForkInvalidAlloc
			}
			for addr := range outputAlloc {
				account1 := outputAlloc[addr]
				account2 := evmAlloc[addr]
				if account2 == nil {
					return ErrReplayForkInvalidAlloc
				}

				// check nonce
				if account1.Nonce != account2.Nonce {
					return ErrReplayForkInvalidAlloc
				}

				// check code
				if !bytes.Equal(account1.Code, account2.Code) {
					return ErrReplayForkInvalidAlloc
				}

				// check storage
				storage1 := account1.Storage
				storage2 := account2.Storage
				if len(storage1) != len(storage2) {
					return ErrReplayForkInvalidAlloc
				}
				for k, v1 := range storage1 {
					if v2, exist := storage2[k]; !exist || v1 != v2 {
						return ErrReplayForkInvalidAlloc
					}
				}
			}

			// more gas
			if evmResult.GasUsed > outputResult.GasUsed {
				return ErrReplayForkMoreGas
			}

			// less gas
			if evmResult.GasUsed < outputResult.GasUsed {
				return ErrReplayForkLessGas
			}

			// misc: logs, ...
			return ErrReplayForkMisc

		} else if outputResult.Status == types.ReceiptStatusSuccessful &&
			evmResult.Status == types.ReceiptStatusFailed {
			// if output was successful but evm failed, return runtime error
			return msgResult.Err
		} else {
			// misc (logs, ...)
			return ErrReplayForkMisc
		}
	}

	return nil
}

// stage1-substate: func replayFork for replay-fork task
func replayForkBlock(ctx *cli.Context, block int64, chainConfig *params.ChainConfig) ReplayForkStats {
	var err error
	var blockStats = make(ReplayForkStats)

	for block, tx := uint64(block), 0; ; tx++ {

		if !research.HasSubstate(block, tx) {
			break
		}

		substate := research.GetSubstate(block, tx)

		alloc := substate.InputAlloc
		msg := substate.Message

		to := msg.To
		if ctx.Bool(SkipTransferTxsFlag.Name) && to != nil {
			// skip regular transactions (ETH transfer)
			if account, exist := alloc[*to]; !exist || len(account.Code) == 0 {
				continue
			}
		}
		if ctx.Bool(SkipCallTxsFlag.Name) && to != nil {
			// skip CALL trasnactions with contract bytecode
			if account, exist := alloc[*to]; exist && len(account.Code) > 0 {
				continue
			}
		}
		if ctx.Bool(SkipCreateTxsFlag.Name) && to == nil {
			// skip CREATE transactions
			continue
		}

		err = ApplySubstateFork(ctx, block, tx, substate, chainConfig)
		if err != nil {
			errstr := fmt.Sprintf("%v", err)
			blockStats[errstr]++
		}
	}
	return blockStats
}

// stage1-substate: func ReplayFork for replay-fork command
func ReplayFork(ctx *cli.Context) error {
	if len(ctx.Args()) != 2 {
		return fmt.Errorf("stage1-substate: transition-substate (replay-fork) command requires exactly 2 arguments")
	}

	start := time.Now()

	first, ferr := strconv.ParseInt(ctx.Args().Get(0), 10, 64)
	last, lerr := strconv.ParseInt(ctx.Args().Get(1), 10, 64)
	if ferr != nil || lerr != nil {
		return fmt.Errorf("stage1-substate: ReplayFork: error in parsing parameters: block number not an integer")
	}
	if first < 0 || last < 0 {
		return fmt.Errorf("stage1-substate: ReplayFork: error: block number must be greater than 0")
	}
	if first > last {
		return fmt.Errorf("stage1-substate: ReplayFork: error: first block has larger number than last block")
	}

	hardFork := ctx.Int64(HardForkFlag.Name)
	chainConfig := &params.ChainConfig{}
	if hardForkName, exist := HardForkName[hardFork]; !exist {
		return fmt.Errorf("stage1-substate: ReplayFork: invalid hard-fork block number %v", hardFork)
	} else {
		fmt.Printf("stage1-substate: ReplayFork: hard-fork: block %v (%s)\n", hardFork, hardForkName)
	}
	switch hardFork {
	case 0:
		*chainConfig = *params.MainnetChainConfig
	case 1:
		*chainConfig = *tests.Forks["Frontier"]
	case 1150000:
		*chainConfig = *tests.Forks["Homestead"]
	case 2463000:
		*chainConfig = *tests.Forks["EIP150"] // Tangerine Whistle
	case 2675000:
		*chainConfig = *tests.Forks["EIP158"] // Spurious Dragon
	case 4370000:
		*chainConfig = *tests.Forks["Byzantium"]
	case 7280000:
		*chainConfig = *tests.Forks["ConstantinopleFix"]
	case 9069000:
		*chainConfig = *tests.Forks["Istanbul"]
	}

	research.OpenSubstateDBReadOnly()
	defer research.CloseSubstateDB()

	var totalNumBlock, totalNumTx int64
	var stats = make(ReplayForkStats)
	var statsLock = &sync.Mutex{}
	defer func() {
		duration := time.Since(start) + 1*time.Nanosecond
		sec := duration.Seconds()

		nb, nt := atomic.LoadInt64(&totalNumBlock), atomic.LoadInt64(&totalNumTx)
		blkPerSec := float64(nb) / sec
		txPerSec := float64(nt) / sec
		fmt.Printf("stage1-substate: ReplayFork: total #block = %v\n", nb)
		fmt.Printf("stage1-substate: ReplayFork: total #tx    = %v\n", nt)
		fmt.Printf("stage1-substate: ReplayFork: %.2f blk/s, %.2f tx/s\n", blkPerSec, txPerSec)
		fmt.Printf("stage1-substate: ReplayFork done in %v\n", duration.Round(1*time.Millisecond))

		statsLock.Lock()
		fmt.Printf("stage1-substate: ReplayFork: = list of differences = \n")
		fmt.Printf("stage1-substate: ReplayFork: = %12v total #tx\n", nt)
		for errstr, n := range stats {
			fmt.Printf("stage1-substate: ReplayFork: = %12v %s\n", n, errstr)
		}
		statsLock.Unlock()
	}()

	numWorker := ctx.Int(WorkersFlag.Name)
	// numProcs + work producer (1) + main thread (1)
	numProcs := numWorker + 2
	if goMaxProcs := runtime.GOMAXPROCS(0); goMaxProcs < numProcs {
		runtime.GOMAXPROCS(numProcs)
	}

	fmt.Printf("stage1-substate: ReplayFork: #CPU = %v, #worker = %v, #thread = %v\n", runtime.NumCPU(), numWorker, runtime.GOMAXPROCS(0))

	workChan := make(chan int64, numWorker*10)
	doneChan := make(chan interface{}, numWorker*10)
	stopChan := make(chan struct{}, numWorker)
	wg := sync.WaitGroup{}
	defer func() {
		// stop all workers
		for i := 0; i < numWorker; i++ {
			stopChan <- struct{}{}
		}
		// stop work producer (1)
		stopChan <- struct{}{}

		wg.Wait()
		close(workChan)
		close(doneChan)
	}()
	for i := 0; i < numWorker; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {

				case block := <-workChan:
					blockStats := replayForkBlock(ctx, block, chainConfig)
					statsLock.Lock()
					for err, n := range blockStats {
						stats[err] += n
						atomic.AddInt64(&totalNumTx, n)
					}
					atomic.AddInt64(&totalNumBlock, 1)
					statsLock.Unlock()
					doneChan <- block

				case <-stopChan:
					return

				}
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		for block := first; block <= last; block++ {
			select {

			case workChan <- block:
				continue

			case <-stopChan:
				return

			}
		}
	}()

	var lastSec float64
	var lastNumBlock, lastNumTx int64
	for block := first; block <= last; block++ {
		data := <-doneChan
		switch t := data.(type) {

		case int64:
			block := data.(int64)
			if block%10000 == 0 {
				duration := time.Since(start) + 1*time.Nanosecond
				sec := duration.Seconds()

				nb, nt := atomic.LoadInt64(&totalNumBlock), atomic.LoadInt64(&totalNumTx)
				blkPerSec := float64(nb-lastNumBlock) / (sec - lastSec)
				txPerSec := float64(nt-lastNumTx) / (sec - lastSec)
				fmt.Printf("stage1-substate: elapsed time: %v, number = %v\n", duration.Round(1*time.Millisecond), block)
				fmt.Printf("stage1-substate: %.2f blk/s, %.2f tx/s\n", blkPerSec, txPerSec)

				lastSec, lastNumBlock, lastNumTx = sec, nb, nt
			}

		case error:
			err := data.(error)
			return err

		default:
			return fmt.Errorf("stage1-substate: ReplayFork: unknown type %T value from doneChan", t)

		}
	}

	return nil
}
