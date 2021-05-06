#!/bin/bash
set -e

EVM="./hard-fork/go-ethereum/build/bin/evm"
EVM_LOG="evm-replay-fork.log"

forkns="1150000 2463000 2675000 4370000 7280000 9069000"
for forkn in ${forkns}
do
    lastn=$(expr ${forkn} - 1)
    lastn=$([ ${lastn} -le 9000000 ] && echo ${lastn} || echo 9000000)
    log="evm-replay-fork-${forkn}.log"
    echo "$log"
    # if log file exists
    if [[ -f "$log" ]]; then
        continue
    fi
    ${EVM} replay-fork 1 ${lastn} --workers $(nproc) --skip-transfer-txs --skip-create-txs --hard-fork ${forkn} | tee $EVM_LOG
    mv ${EVM_LOG} ${log}
done
