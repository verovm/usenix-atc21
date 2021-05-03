// Copyright 2014 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

// evm executes EVM code snippets.
package main

import (
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/cmd/evm/internal/t8ntool"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/internal/flags"
	"gopkg.in/urfave/cli.v1"

	// stage1-substate: import evm/research
	"github.com/ethereum/go-ethereum/cmd/evm/research"
)

var gitCommit = "" // Git SHA1 commit hash of the release (set via linker flags)
var gitDate = ""

var (
	app = flags.NewApp(gitCommit, gitDate, "the evm command line interface")

	DebugFlag = cli.BoolFlag{
		Name:  "debug",
		Usage: "output full trace logs",
	}
	MemProfileFlag = cli.StringFlag{
		Name:  "memprofile",
		Usage: "creates a memory profile at the given path",
	}
	CPUProfileFlag = cli.StringFlag{
		Name:  "cpuprofile",
		Usage: "creates a CPU profile at the given path",
	}
	StatDumpFlag = cli.BoolFlag{
		Name:  "statdump",
		Usage: "displays stack and heap memory information",
	}
	CodeFlag = cli.StringFlag{
		Name:  "code",
		Usage: "EVM code",
	}
	CodeFileFlag = cli.StringFlag{
		Name:  "codefile",
		Usage: "File containing EVM code. If '-' is specified, code is read from stdin ",
	}
	GasFlag = cli.Uint64Flag{
		Name:  "gas",
		Usage: "gas limit for the evm",
		Value: 10000000000,
	}
	PriceFlag = utils.BigFlag{
		Name:  "price",
		Usage: "price set for the evm",
		Value: new(big.Int),
	}
	ValueFlag = utils.BigFlag{
		Name:  "value",
		Usage: "value set for the evm",
		Value: new(big.Int),
	}
	DumpFlag = cli.BoolFlag{
		Name:  "dump",
		Usage: "dumps the state after the run",
	}
	InputFlag = cli.StringFlag{
		Name:  "input",
		Usage: "input for the EVM",
	}
	InputFileFlag = cli.StringFlag{
		Name:  "inputfile",
		Usage: "file containing input for the EVM",
	}
	VerbosityFlag = cli.IntFlag{
		Name:  "verbosity",
		Usage: "sets the verbosity level",
	}
	BenchFlag = cli.BoolFlag{
		Name:  "bench",
		Usage: "benchmark the execution",
	}
	CreateFlag = cli.BoolFlag{
		Name:  "create",
		Usage: "indicates the action should be create rather than call",
	}
	GenesisFlag = cli.StringFlag{
		Name:  "prestate",
		Usage: "JSON file with prestate (genesis) config",
	}
	MachineFlag = cli.BoolFlag{
		Name:  "json",
		Usage: "output trace logs in machine readable format (json)",
	}
	SenderFlag = cli.StringFlag{
		Name:  "sender",
		Usage: "The transaction origin",
	}
	ReceiverFlag = cli.StringFlag{
		Name:  "receiver",
		Usage: "The transaction receiver (execution context)",
	}
	DisableMemoryFlag = cli.BoolFlag{
		Name:  "nomemory",
		Usage: "disable memory output",
	}
	DisableStackFlag = cli.BoolFlag{
		Name:  "nostack",
		Usage: "disable stack output",
	}
	DisableStorageFlag = cli.BoolFlag{
		Name:  "nostorage",
		Usage: "disable storage output",
	}
	DisableReturnDataFlag = cli.BoolFlag{
		Name:  "noreturndata",
		Usage: "disable return data output",
	}
	EVMInterpreterFlag = cli.StringFlag{
		Name:  "vm.evm",
		Usage: "External EVM configuration (default = built-in interpreter)",
		Value: "",
	}
)

// stage1-substate: t8n-substate command
var stateTransitionSubstateCommand = cli.Command{
	Action:    research.TransitionSubstate,
	Name:      "t8n-substate",
	Aliases:   []string{"t8n-substate"},
	Usage:     "executes full state transitions and check output consistency",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	Flags: []cli.Flag{
		research.WorkersFlag,
		research.SkipTransferTxsFlag,
		research.SkipCallTxsFlag,
		research.SkipCreateTxsFlag,
	},
	Description: `
The transition-substate (t8n-substate) command requires
two arguments: <blockNumFirst> <blockNumLast>
<blockNumFirst> and <blockNumLast> are the first and
last block of the inclusive range of blocks to replay transactions.`,
}

// stage1-substate-migration: contract-fuzzer command
var contractFuzzerCommand = cli.Command{
	Action:    research.ContractFuzzer,
	Name:      "contract-fuzzer",
	Aliases:   []string{"cf"},
	Usage:     "execute address and messages given from ContractFuzzer and send output to port 8888",
	ArgsUsage: "<address> <messagesPath> <blockTxsPath>",
	Flags:     []cli.Flag{},
	Description: `
The contract-fuzzer (cf) command requires three arguments:
<address> <callDataPath> <blockTxsPath>
<address> is the account address to test.
<callDataPath> is a path of a file that contains a list of call data (ABI + arguments).
<blockTxsPath> is a path of a file that contains a list of block_tx keys to test.`,
}

var addressToSubstateCommand = cli.Command{
	Action:    research.AddressToSubstate,
	Name:      "address-to-substate",
	Aliases:   []string{"a2s"},
	Usage:     "extract mapping from address to substate for ContractFuzzer",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	Flags:     []cli.Flag{},
	Description: `
The address-to-substate (a2s) command requires two arguments:
<blockNumFirst> <blockNumLast>
<blockNumFirst> and <blockNumLast> are the first and
last block of the inclusive range of blocks to extract.`,
}

func init() {
	app.Flags = []cli.Flag{}
	app.Commands = []cli.Command{
		// stage1-substate: transition-substate (t8n-substate) command
		stateTransitionSubstateCommand,
		// stage1-substate-migration: contract-fuzzer (cf) command
		contractFuzzerCommand,
		// stage1-substate-migration: address-to-substate (a2s) command
		addressToSubstateCommand,
	}
	cli.CommandHelpTemplate = flags.OriginCommandHelpTemplate
}

func main() {
	if err := app.Run(os.Args); err != nil {
		code := 1
		if ec, ok := err.(*t8ntool.NumberedError); ok {
			code = ec.Code()
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(code)
	}
}
