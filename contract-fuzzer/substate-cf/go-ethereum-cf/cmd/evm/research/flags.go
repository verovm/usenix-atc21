package research

import (
	"fmt"

	cli "gopkg.in/urfave/cli.v1"
)

// stage1-substate: flags for t8n-substate command
var (
	WorkersFlag = cli.IntFlag{
		Name:  "workers",
		Usage: "Number of worker threads that execute in parallel",
		Value: 4,
	}
	SkipTransferTxsFlag = cli.BoolFlag{
		Name:  "skip-transfer-txs",
		Usage: "Skip executing transactions that only transfer ETH",
	}
	SkipCallTxsFlag = cli.BoolFlag{
		Name:  "skip-call-txs",
		Usage: "Skip executing CALL transactions to accounts with contract bytecode",
	}
	SkipCreateTxsFlag = cli.BoolFlag{
		Name:  "skip-create-txs",
		Usage: "Skip executing CREATE transactions",
	}
	HardForkNums = []int64{
		1,
		1150000,
		2463000,
		2675000,
		4370000,
		7280000,
		9069000,
		0,
	}
	HardForkName = map[int64]string{
		1:       "Frontier",
		1150000: "Homestead",
		2463000: "Tangerine Whistle",
		2675000: "Spurious Dragon",
		4370000: "Byzantium",
		7280000: "Constantinople + Petersburg",
		9069000: "Istanbul",
		0:       "Mainnet",
	}
	HardForkFlag = cli.Int64Flag{
		Name: "hard-fork",
		Usage: func() string {
			s := ""
			s += "Hard-fork block number, it will not change block number in substate env"
			for _, num64 := range HardForkNums {
				s += fmt.Sprintf("\n\t  %v: %s", num64, HardForkName[num64])
			}
			return s
		}(),
		Value: 0,
	}
)
