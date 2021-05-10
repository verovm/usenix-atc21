# Introduction

This repository contains all materials used for experiments in the paper:

**Yeonsoo Kim, Seongho Jeong, Kamil Jezek, Bernd Burgstaller, and Bernhard Scholz**: _An Off-The-Chain Execution Environment for Scalable Testing and Profiling of Smart Contracts_,  ATC Usenix 2021

These materials may be used for relplication studies, follow-up research, experimenting, etc. The following sections contain information about the environment set-up, followed by the three use cases from the paper. 

# Getting the Source Code

First, checkout the source code. We expect this is done in your home directory, and all paths in the text bellow refer to the user home directory ```~/```:

```
git clone git@github.com:verovm/usenix-atc21.git
```

# Substate Database Snapshot

The use cases from the paper reguire a pre-exesting sub-state database snapshots. It can be either generated from the Recorder tool (described below), or we provide a snapshot for download. 

* Substate DB of 9M blocks (stage1-substate-0-9M.tar.zst): [gdrive download](https://drive.google.com/file/d/1jl6vdMea5ROKdrTUJUk8lh5NL48Do9xJ/view?usp=sharing) (139 GB, decompressed size: 285GB)

Download this file and untar on your disk:

```bash
# untar substate DB
cd ~/
tar -xavf stage1-substate-0-9M.tar.zst
mv stage1-substate-0-9M stage1-substate
```


# Build the Recorder and Generate Database Snapshot

It is also possible to generate your own sub-state database snapshot. It is useful either if the download of our snapshot is inpractical, or the recording should happen on alternative blockchains (i.e. our snapshot has been recorded on the mainet).

## Sync and Export Blockchain in Files

Exported blockchain files are used to generate the substate database and to measure the time and the space of the Geth full node. You can export blockchain in a file with the following commands. To sync in a reasonable amount of time, the `--datadir` parameter must be the path to an empty directory on an SSD with at least 1TB of space. The syncing step can take more than a day to finish.

```bash
cd ~/usenix-atc21/go-ethereum/
make geth

# press ctrl-c to stop geth sync when it reached the desired block height
./build/bin/geth --datadir geth.datadir/ --syncmode fast --gcmode full

# export from block 2,000,001 to 3,000,000 (total 1M blocks)
./build/bin/geth --datadir geth.datadir/ --syncmode fast --gcmode full 2-3M.blockchain 2000001 3000000
```

We provide exported blockchain files for download. To generate the substate database, use `0-9M.blockchain` which contains initial 9M blocks. To measure time and space of Geth full node, use blockchain files segmented by 1M blocks.

* Initial 9M blocks (`0-9M.blockchain`): [gdrive download](https://drive.google.com/file/d/1VoOtMlhcaT_CeVulP8VQ-TpHFZ7eVbqy/view?usp=sharing) (total 104 GB)
* 1M-block segments (`0-1M.blockchain`, `1-2M.blockchain`, ...): [gdrive directory](https://drive.google.com/drive/folders/132VLKpxPfulbcg36hiY6C1Sef3yXAirG?usp=sharing) (104 GB)


## Generate the Substate Database

The substate recorder is implemented by modifying the `geth import` command, which processes blockchain files exported from the Geth full node. To generate substate database, import a blockchain file exported from a Geth full node to our substate recorder. The substate recorder will create the substate DB in `./stage1-substate/` directory.

```bash
# build recorder
cd ~/usenix-atc21/record-replay/go-ethereum/
make geth

# record substates in ./stage1-substate/
./build/bin/geth --datadir /path/to/recorder.datadir import /path/to/0-9M.blockchain
```


# Scalability of Substate Replayer

The following experiments provide results from Table 2-3 and Figure 6 in section "5.1 Scalability of Substate Replayer", which compares the time and the space required to replay the transactions in 9M blocks using the Geth full node and the substate replayer.

## Geth Full Node - Time and Space

This experiment measures the time and the space to replay the transactions with the Geth full node in Table 2-3. To measure the single thread performance in the block processing, the `--cache.noprefetch` option is given. The block import time and the maximum Geth database size of each 1M blocks is saved in `.log` files.

```bash
# build geth
cd ~/usenix-atc21/go-ethereum/
make geth

# replay transactions in 0-1M.blockchain
./build/bin/geth --datadir geth.datadir/ --cache.noprefetch import 0-1M.blockchain 2>&1 | tee -a geth-0-1M.log

# extract block import time
grep 'Import done' geth-0-1M.log > geth-time-0-1M.log

# measure Geth database
du -sb geth.datadir/ > geth-size-0-1M.log


# continue the measurement with next 1M blocks
./build/bin/geth --datadir geth.datadir/ --cache.noprefetch import 8-9M.blockchain 2>&1 | tee -a geth-1-2M.log
grep 'Import done' geth-1-2M.log > geth-time-1-2M.log
du -sb geth.datadir/ > geth-size-1-2M.log

...

```

The values in `geth-time-0-1M.log`, `geth-time-1-2M.log`, ... are used in the paper, Section 5.1, Table 3, _Geth full node - Time (s)_ column.
These logs contain time spent to import and replay transactions in `0-1M.blockchain`, `1-2M.blockchain`, ...:
```
Import done in 1m20.480482469s.
```

The values in `geth-size-0-1M.log`, `geth-size-1-2M.log`, ... are used in the paper, Section 5.1, Table 3, _Geth full node (GB)_ column in Table 2.
These logs contain space (bytes) required to import and replay transactions in `0-1M.blockchain`, `1-2M.blockchain`, ...:
```
103707444	geth.datadir/
```

## Substate Replayer - Time

This experiment measures the execution time of the single- and multi-threaded substate replayer in Table 3 and Figure 6. The substate replayer contains the `evm transition-substate` command (`evm t8n-substate`) that loads substates of a given block range from the sub-state database snapshot in `./stage1-substate/` and replay the transactions. If the substate replayer finds that the replay output is different from the expected output, it returns an error.

The experiment for this section is performed by two scripts: 

[evm-t8n-substate-0-9M.sh](./record-replay/evm-t8n-substate-0-9M.sh) is a bash script that runs substate replayer with different numbers of threads.

[evm-t8n-substate-csv.py](./record-replay/evm-t8n-substate-csv.py) is a python3 script that collects output log files of `evm-t8n-substate-0-9M.sh` and print data in CSV format.

Execute the experiment by typing the following:

```bash
# build substate replayer (evm)
cd ~/usenix-atc21/record-replay/go-ethereum/
make all

# measure replayer performance and print data in tab-separated values (TSV)
cd ../
./evm-t8n-substate-0-9M.sh
python3 ./evm-t8n-substate-tsv.py > evm-t8n-substate.tsv
```

`evm-t8n-substate.csv` should look like:
```
block	1	2	4	8	12	16	24	32	48	64	
0--1M	526.0	419.0	163.0	94.0	70.0	58.0	47.0	44.0	43.0	46.0	
1--2M	1517.0	836.0	470.0	271.0	190.0	163.0	128.0	127.0	122.0	130.0	
2--3M	24125.0	14563.0	9468.0	7179.0	6818.0	6699.0	6557.0	6456.0	6257.0	6212.0	
3--4M	5222.0	2728.0	1586.0	913.0	678.0	562.0	465.0	439.0	429.0	431.0	
4--5M	28873.0	15270.0	8583.0	4763.0	3454.0	2744.0	2126.0	1972.0	1916.0	1975.0	
5--6M	35390.0	19037.0	10617.0	5891.0	4249.0	3452.0	2649.0	2411.0	2418.0	2619.0	
6--7M	33672.0	18476.0	10171.0	5606.0	3989.0	3192.0	2495.0	2309.0	2276.0	2484.0	
7--8M	38060.0	20898.0	11312.0	6242.0	4432.0	3579.0	2803.0	2503.0	2448.0	2586.0	
8--9M	41999.0	22746.0	12222.0	6753.0	4800.0	3851.0	3032.0	2741.0	2767.0	2880.0	
```


## Substate Replayer - Space

This experiment measures the space required to store transaction substates of every 1M blocks in the substate DB. The results of this experiment are contained in _Substate replayer_  column in Table 2. The substate replayer contains the `evm dump-substate` command that reads `./stage1-substate/` and creates a database copy with substates found in a given range of blocks.

For example, to measure space required to replay transactions in 2-3M blocks:
```bash
# build substate replayer (evm)
cd ~/usenix-atc21/record-replay/go-ethereum/
make all

# copy substates of 2-3M blocks and measure database size (bytes)
cd ../
evm dump-substate ./stage1-substate-2-3M/ 2000001 3000000
du -sb ./stage1-substate-2-3M/
```
Repeat these steps for other block heights, 0-1M, 1-2M, 2-3M, etc. 

# Metrics Use Case

The metrics use case analyzes transactions by generating a graph of instruction flow. It counts the number of live instructions and live gases.
To produce a result of metrics and visualize it for 9M blocks execute:

```bash
# build evm for value-graph metrics
cd ~/usenix-atc21/value-graph/go-ethereum
make all

# execute metrics analysis and visualize. $numThreads sets the number of workers for replayer.
cd ..
./metrics-0-9M.sh $numThreads
```

The script will replay 9M blocks and produce metrics from the value graph analysis. The outputs contain raw data for 9M blocks (csv files) and visualization of the data (Figure 7, 8, and 9). 

To produce an image of single value graph, the following command generates a PNG image for the first transaction executed in the block 2000000.
```
cd ~/usenix-atc21/value-graph/go-ethereum/build/bin
evm t8n-substate 6011051 6011051 --workers 1 --skip-transfer-txs --skip-create-txs --graph
```

# Contract Fuzzer Use Case

This experiment provides results for "Section 5.3 Fuzzer Use Case". This repository contais two variants of ContractFuzzer - an original version, and our fork that enables transaction replay. 

The experiment requires:
* the sub-state database, 
* contracs' ABIs, 
* addresses mapping (address-to-substate/): [gdrive download](https://drive.google.com/file/d/13eTEpu7Bt1XRpKDFFHYNhy_phwuLujGV/view?usp=sharing) (108 MB, decompressed size: 805 MB)
* [NodeJS Installation](https://nodejs.org/en/download/), 
* [Docker installation](https://docs.docker.com/get-docker/).

## Contracts' ABIs

The contract's ABIs can be obtained by the script:
```
cd ~/usenix-atc21/contract-fuzzer/substate-cf/contract_downloader/
./download_contracts.sh ~/address-to-substate/ ~/contracts 10
```
The parameters of the script tell (1) the directory with the addresses mappings (2) the output dir (3) the size of the batch, the ABIs will be grouped into - in the paper we have used 10. 

Notice that the script will try to download all available ABIs for the whole blockchain. It is possible to interrupt the script anytime earlier and continue the experiment on a smaller dataset. In the paper, we have dowloaed several hundreds of ABIs. 

## Build Docker Images

This repository contains docker images to simplify the run of the experiments. Build the images by following commands:

```
cd ~/usenix-atc21/contract-fuzzer/original-cf/
docker build -t contractfuzzer-original-experiment .

cd ~/usenix-atc21/contract-fuzzer/substate-cf/
docker build -t contractfuzzer-experiment .

cd ~/usenix-atc21/contract-fuzzer/substate-cf/contract_experiments/
docker build -t cf-experiment-master  .
```

## Run the Experiment - Original Contract Fuzzer

Now the experiment may be triggered for the original contract fuzzer:
```
cd ~/usenix-atc21/contract-fuzzer/original-cf/contract_experiments/
```
Edit the ```docker-compose.yaml``` file with a text editor and update the following lines to contain correct paths on your system - absolute paths must be used (modify only the part before colon):
```
   - /opt/cf-experiments/contracts-original/:/contracts     # Directory with contracs's ABIs, must point to /absolute/path//usenix-atc21/contract-fuzzer/contracts-original
   - /opt/cf-experiments/address-to-substate/:/addresses    # Addresses mappings
   - /opt/cf-experiments/stage1-substate:/ContractFuzzer/stage1-substate/     # Substate database
```

The experiment may be now invoked via docker:
```
 docker swarm init
 docker stack deploy -c  docker-compose.yaml CF
 ```
 These commands run the experiment as docker services. 
 Now, periodically monitor the service logs, which will contain speed of the execution, by using the following command: 
 ```
 docker service logs CF_master
 ```
 This will show for instance an output:
 ```
 CF_master.1  | Next task is 10 Index: 2/1165
 CF_master.1  | Speed:  diffTime: 4.0002, finishedTasks: 10, speed: 2.4998750062496873
 ```
 After some time the experiment will process batches and the value of the  ```speed ``` stabilises. We have executed several hundrets of batches (```tasks```). The value is used in the paper in Table 4: ContractFuzzer — performance improvements, the first row.
 
 The running experiment may be interrupted by typing:
 ```
 docker stack rm CF
 ```
 ## Run the Experiment - Contract Fuzzer with Substate Reply
 
 To run the experiment, go to the directory 
 ```
 cd ~/usenix-atc21/contract-fuzzer/substate-cf/contract_experiments
 ```
 and repeat the same steps as in the previous experiment. Now the ContractFuzzer uses contracts data from the substate database via the Replay tool. 
 
 Edit the ```docker-compose.yaml``` file with a text editor and update again the paths (notice that the paths to contracts will be different than in the first experiment):
 
 ```
   - /opt/cf-experiments/contracts/:/contracts     # Directory with contracs's ABIs downloaded by the script above, must point to ABIs downloaded by the script /download_contracts.sh
   - /opt/cf-experiments/address-to-substate/:/addresses    # Addresses mappings
   - /opt/cf-experiments/stage1-substate:/ContractFuzzer/stage1-substate/     # Substate database 
```
 Furthermore, change the number of parallel executions. After each edit, run the experiment again. 
 
 ```
   deploy:
      replicas: 1    # Numner of parallel executions
 ```

The same as in the previous experiment, the speed is monitored and the results are used for Table 4: ContractFuzzer — performance improvements, second and next rows. The number of parallel tasks in the first column matches to the number of selected ```replicas```.

Last two columns of the table were calculated. 

Notice that all replicas use the same sub-state database mounted via a file mount, and no blockchain is needed. 

# Hard Fork Assesment Use Case

This experiment provides results of Table 5 in section "5.4 Hard Fork Assessment". This experiment assesses hard forks by replaying the transactions in the same context they were executed except the protocols is changed by the new hard fork.

[replay-fork-0-9M.sh](./hard-fork/replay-fork-0-9M.sh) is a bash script to assess all hard forks activated before block 9,000,000 with CALL transactions (contract invocations) in initial 9M blocks.

```bash
# build evm for hard fork assessment
cd ~/usenix-atc21/hard-fork/go-ethereum/
make all

# run hard fork assessments with 9M blocks
cd ../
./replay-fork-0-9M.sh
```

After replayed all transactions, `evm replay-fork` prints numbers and types of errors like:
```
stage1-substate: ReplayFork: =    303300443 total #tx
stage1-substate: ReplayFork: =       193524 invalid jump destination
stage1-substate: ReplayFork: =       599698 invalid opcode: opcode 0xfe not defined
stage1-substate: ReplayFork: =      3061728 execution reverted
stage1-substate: ReplayFork: =       430290 invalid alloc in replay-fork
stage1-substate: ReplayFork: =    207742570 more gas in replay-fork
stage1-substate: ReplayFork: =     56251999 less gas in replay-fork
stage1-substate: ReplayFork: =     11108391 misc in replay-fork
stage1-substate: ReplayFork: =     23912243 out of gas
```

| `evm replay-fork` error                 | Table 6 column |
|-----------------------------------------|-----------------------------------------|
| invalid jump destination                | EVM runtime exception - Invalid JUMP    |
| invalid opcode: opcode 0xfe not defined | EVM runtime exception - Invalid opcode  |
| execution reverted                      | EVM runtime exception - Reverted        |
| invalid alloc in replay-fork            | Output Changed                          |
| misc in replay-fork                     | Output Changed                          |
| out of gas                              | Gas usage changed - Out-of-gas          |
| more gas in replay-fork                 | Gas usage changed - Increased           |
| less gas in replay-fork                 | Gas usage changed - Decreased           |
