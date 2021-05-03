package research

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/research"
	cli "gopkg.in/urfave/cli.v1"
)

// stage1-substate-migration: func ContractFuzzer for contract-fuzzer command
func AddressToSubstate(ctx *cli.Context) error {
	if len(ctx.Args()) != 2 {
		return fmt.Errorf("stage1-substate-migration: address-to-substate (a2s) command requires exactly 2 arguments")
	}

	start := time.Now()
	defer func() {
		duration := time.Since(start) + 1*time.Nanosecond
		fmt.Printf("stage1-substate-migration: AddressToSubstate done in %v\n", duration.Round(1*time.Millisecond))
	}()

	first, ferr := strconv.ParseInt(ctx.Args().Get(0), 10, 64)
	last, lerr := strconv.ParseInt(ctx.Args().Get(1), 10, 64)
	if ferr != nil || lerr != nil {
		return fmt.Errorf("stage1-substate-migration: AddressToSubstate: error in parsing parameters: block number not an integer")
	}
	if first < 0 || last < 0 {
		return fmt.Errorf("stage1-substate-migration: AddressToSubstate: error: block number must be greater than 0")
	}
	if first > last {
		return fmt.Errorf("stage1-substate-migration: AddressToSubstate: error: first block has larger number than last block")
	}

	research.OpenSubstateDBReadOnly()
	defer research.CloseSubstateDB()

	a2s := make(map[common.Address][]string)
	for block := uint64(first); block <= uint64(last); block++ {
		for tx := 0; ; tx++ {
			if !research.HasSubstate(block, tx) {
				break
			}

			substate := research.GetSubstate(block, tx)

			alloc := substate.InputAlloc
			msg := substate.Message

			to := msg.To
			if to != nil {
				// skip regular transactions (ETH transfer)
				if account, exist := alloc[*to]; !exist || len(account.Code) == 0 {
					continue
				}
			}
			if to == nil {
				// skip CREATE transactions
				continue
			}

			blockTx := fmt.Sprintf("%v_%v", block, tx)
			a2s[*to] = append(a2s[*to], blockTx)
		}

		if block%10000 == 0 {
			duration := time.Since(start) + 1*time.Nanosecond
			totalBlockTx := 0
			for _, blockTxs := range a2s {
				totalBlockTx += len(blockTxs)
			}
			fmt.Printf("stage1-substate-migration: elapsed time: %v, number = %v\n", duration.Round(1*time.Millisecond), block)
			fmt.Printf("stage1-substate-migration: len(a2s) = %v, totalBlockTx = %v\n", len(a2s), totalBlockTx)
		}
	}

	outdir := "address-to-substate"
	os.Mkdir(outdir, 0755)

	for addr, blockTxs := range a2s {
		func(addrHex, blockTxsString string) {
			f, err := os.Create(path.Join(outdir, addrHex))
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			f.WriteString(blockTxsString)
		}(strings.ToLower(addr.Hex()), strings.Join(blockTxs, "\n")+"\n")
	}

	return nil
}
