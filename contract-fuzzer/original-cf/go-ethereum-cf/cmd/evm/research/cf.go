package research

import (
	"bufio"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/cmd/evm/internal/t8ntool"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/research"
	cli "gopkg.in/urfave/cli.v1"
)

// stage1-substate-migration: func ApplyFuzzerMessage
func ApplyFuzzerMessage(address common.Address, callData []byte, block uint64, tx int) {
	var err error
	substate := research.GetSubstate(block, tx)

	inputAlloc := substate.InputAlloc
	inputEnv := substate.Env
	inputMessage := substate.Message

	if inputMessage.To == nil || *inputMessage.To != address {
		panic(fmt.Errorf("stage1-substate-transition: ApplyFuzzerMessage: %v_%v's inputMessage.To is not address %s", block, tx, address.Hex()))
	}
	// if len(inputMessage.Data) < 4 || !bytes.Equal(inputMessage.Data[:4], callData[:4]) {
	// 	panic(fmt.Errorf("stage1-substate-transition: ApplyFuzzerMessage: %v_%v's ABI signature is not %s", block, tx, common.ToHex(callData[:4])))
	// }

	inputMessage.Data = callData

	var (
		vmConfig    vm.Config
		chainConfig *params.ChainConfig
		pre         t8ntool.Prestate
		getTracerFn func(txIndex int) (tracer vm.Tracer, err error)
	)

	vmConfig = vm.Config{}

	chainConfig = &params.ChainConfig{}
	*chainConfig = *params.MainnetChainConfig
	// disable DAOForkSupport, otherwise account states will be overwritten
	chainConfig.DAOForkSupport = false

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
	memdb, _ := ethdb.NewMemDatabase()
	var (
		statedb   = t8ntool.MakePreState(memdb, pre.Pre)
		gaspool   = new(core.GasPool)
		blockHash = common.Hash{0x13, 0x37}
		txIndex   = tx
	)

	gaspool.AddGas(new(big.Int).SetUint64(pre.Env.GasLimit))
	vmContext := vm.Context{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		Coinbase:    pre.Env.Coinbase,
		BlockNumber: new(big.Int).SetUint64(pre.Env.Number),
		Time:        new(big.Int).SetUint64(pre.Env.Timestamp),
		Difficulty:  pre.Env.Difficulty,
		GasLimit:    new(big.Int).SetUint64(pre.Env.GasLimit),
		GetHash:     getHash,
		// GasPrice and Origin needs to be set per transaction
	}

	msg := inputMessage.AsMessage()

	tracer, err := getTracerFn(txIndex)

	vmConfig.Tracer = tracer
	vmConfig.Debug = (tracer != nil)
	statedb.Prepare(common.Hash{}, blockHash, txIndex)
	vmContext.GasPrice = msg.GasPrice()
	vmContext.Origin = msg.From()

	evm := vm.NewEVM(vmContext, statedb, chainConfig, vmConfig)
	snapshot := statedb.Snapshot()
	_, _, err = core.ApplyMessage(evm, msg, gaspool)

	if err != nil {
		statedb.RevertToSnapshot(snapshot)
	}
}

// stage1-substate-migration: func ContractFuzzer for contract-fuzzer command
func ContractFuzzer(ctx *cli.Context) error {
	if len(ctx.Args()) != 3 {
		return fmt.Errorf("stage1-substate-migration: contract-fuzzer (cf) command requires exactly 3 arguments")
	}

	start := time.Now()
	defer func() {
		duration := time.Since(start) + 1*time.Nanosecond
		fmt.Printf("stage1-substate: ContractFuzzer done in %v\n", duration.Round(1*time.Millisecond))
	}()

	address := common.HexToAddress(ctx.Args().Get(0))

	// read callDataPath into callDataList
	callDataPath := ctx.Args().Get(1)
	cdfile, cderr := os.Open(callDataPath)
	defer cdfile.Close()
	if cderr != nil {
		return fmt.Errorf("stage1-substate-migration: failed to open callDataPath: %v", cderr)
	}
	cdscanner := bufio.NewScanner(cdfile)
	cdscanner.Split(bufio.ScanWords)

	callDataList := [][]byte{}
	for cdscanner.Scan() {
		callData := common.FromHex(cdscanner.Text())
		callDataList = append(callDataList, callData)
	}

	// read blockTxsPath into blockTxs
	blockTxsPath := ctx.Args().Get(2)
	btfile, bterr := os.Open(blockTxsPath)
	defer btfile.Close()
	if bterr != nil {
		return fmt.Errorf("stage1-substate-migration: failed to open blockTxsPath: %v", cderr)
	}
	btscanner := bufio.NewScanner(btfile)
	btscanner.Split(bufio.ScanWords)

	blockTxs := []string{}
	for btscanner.Scan() {
		blockTx := btscanner.Text()
		blockTxs = append(blockTxs, blockTx)
	}

	research.OpenSubstateDBReadOnly()
	defer research.CloseSubstateDB()

	for _, callData := range callDataList {
		for _, blockTx := range blockTxs {
			blockTxSplit := strings.Split(blockTx, "_")
			block, err := strconv.ParseUint(blockTxSplit[0], 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse blockTx data: %s", blockTx)
			}
			tx, err := strconv.ParseInt(blockTxSplit[1], 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse blockTx data: %s", blockTx)
			}
			ApplyFuzzerMessage(address, callData, block, int(tx))
		}
	}

	return nil
}
