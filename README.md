# Introduction

This repository contains all materials used for experiments in the paper:

**Yeonsoo Kim, Seongho Jeong, Kamil Jezek, Bernd Burgstaller, and Bernhard Scholz**: _An Off-The-Chain Execution Environment for Scalable Testing and Profiling of Smart Contracts_,  ATC Usenix 2021

These materials may be used for relplication studies, follow-up research, experimenting, etc. The following sections contain information about the information set-up, followed by the three use cases from the paper. 

# Getting the Source Code

First, checkout the source code. We expect this is done in your home directory, and all paths in the text bellow refer to a user home directory ```~/```:

```
git clone git@github.com:verovm/usenix-atc21.git
```

# Substate Database Snapshot

The use cases from the paper reguire a pre-exesting sub-state databse snaphost. It can be either generated from the Recorder tool (described below), or we provide a snapshot for download. 

* Substate DB of 9M blocks (stage1-substate-0-9M.tar.zst): [gdrive download](https://drive.google.com/file/d/1jl6vdMea5ROKdrTUJUk8lh5NL48Do9xJ/view?usp=sharing) (139 GB, decompressed size: 285GB)

Download this file and untar on your disk:

```bash
# untar substate DB
cd ~/
tar -xavf stage1-substate-0-9M.tar.zst
mv stage1-substate-0-9M stage1-substate
```

@ Yeonsoo - do we really need these downloads? I guess that a user either uses our sub-state DB snapshot, or he records everything from scratch. Is there any use case where the user would start from the middle - i.e. generate the DB from these snapshots? 

* Exported blockchain files (0-1M.blockchain, 1-2M.blockchain, ...): [gdrive directory](https://drive.google.com/drive/folders/132VLKpxPfulbcg36hiY6C1Sef3yXAirG?usp=sharing) (104 GB)
* Exported blockchain file of 9M blocks (0-9M.blockchain): [gdrive download](https://drive.google.com/file/d/1VoOtMlhcaT_CeVulP8VQ-TpHFZ7eVbqy/view?usp=sharing) (104 GB)


# Build the Recorder and Generate Database Snapshot

It is also possible to generate your own sub-state database snapshot. It is useful either if the download of our snapshot is inpractical, or the recording should happen on alternative blockchains (i.e. our snapshot has been recorded on the mainet).

@Yeonsoo - TODO - do we really need this link to the official documentation? You describe bellow how to build the tool right? I noticed that the official documentation is quite long and if the user just needs to type "build geth" we do not want him to study lenghty manuals. 

The substate recorder/replayer and the use cases are implemented based on go-ethereum (geth). Please follow the official documentation _Building the source_ section of [Geth's README](./go-ethereum/README.md) for more details about building the  tools.

## Sync and Export Blockchain in Files

@Yeonsoo - Can you please describe here what the user must exactly do to get what he needs for the experiment from paper? Bellow you say to export 1M of blocks, but we used the whole blockchain right? This is ok to mention that a user may collect a smaller dataset for experimenting, but we cannot avoid what we have in the paper. 

@Yeonsoo - it does not have to be clear what this path is "/path/to/geth.datadir" - think that a user will likely copy paste the lines, and this path would not exist. Perhaps modify it to somethig relative to the user home. 

```bash
cd ~/usenix-atc21/go-ethereum/
make geth

# press ctrl-c to stop geth sync when it reached the desired block height
./build/bin/geth --datadir /path/to/geth.datadir --syncmode fast --gcmode full

# export from block 2,000,001 to 3,000,000 (total 1M blocks)
./build/bin/geth --datadir /path/to/geth.datadir --syncmode fast --gcmode full 2-3M.blockchain 2000001 3000000
```

## Generate the Substate Database

Substate recorder is implemented by modifying `geth import` command which processes blockchain files exported from Geth full node. To generate substate database, import a blockchain file exported from a Geth full node to our substate recorder. Substate recorder will create substate DB in `./stage1-substate/` directory.

```bash
# build recorder
cd ~/usenix-atc21/record-replay/go-ethereum/
make geth

# record substates in ./stage1-substate/
./build/bin/geth --datadir /path/to/recorder.datadir import /path/to/0-9M.blockchain
```


# Scalability of Substate Replayer

The following experiments provide results from Table 2-3 and Figure 6 in section "5.1 Scalability of Substate Replayer", which compares the time and the space required to replay transactions in 9M blocks using the Geth full node and the substate replayer.

## Geth Full Node - Time and Space

This experiment measures the time and the space to replay transactions with the Geth full node in Table 2-3. To measure the single thread performance in the block processing, the `--cache.noprefetch` option is given. The block import time and the maximum Geth database size of each 1M blocks is saved in `.log` files.

@Yeonsoo - please double check that this path "~/usenix-atc21/go-ethereum/ is correct

```bash
# build geth
cd ~/usenix-atc21/go-ethereum/
make geth

# measure geth block import time and size
./build/bin/geth --datadir geth.ethereum --cache.noprefetch import 0-1M.blockchain 2>&1 | tee -a geth-0-1M.log
du -s geth.ethereum >> geth-0-1M.log
./build/bin/geth --datadir geth.ethereum --cache.noprefetch import 8-9M.blockchain 2>&1 | tee -a geth-1-2M.log
du -s geth.ethereum >> geth-1-2M.log
...
```

@Yeonsoo - please add here a sentence that the log files now contain the results in the Table XY - ideally provide an example log output. 

## Substate Replayer - Time

This experiment measures the execution time of the single- and multi-threaded substate replayer in Table 3 and Figure 6. The substate replayer contains the `evm transition-substate` command (`evm t8n-substate`) that loads substates of a given block range from the sub-state database snapshot in `./stage1-substate/` and replay the transactions. If the substate replayer finds that the replay output is different from the expected output, it returns an error.

@Yeonsoo - probably remove this example - show only what the user must do to run the same experiment as in the paper. 

For example, if you want to replay trasnactions from 46147 to 50000 with 8 replay threads:
```bash
evm t8n-substate 46147 50000 --workers 8
```

The experiment for this section is performed by two scripts: 

[evm-t8n-substate-0-9M.sh](./record-replay/evm-t8n-substate-0-9M.sh) is a bash script that runs substate replayer with different numbers of threads.

[evm-t8n-substate-csv.py](./record-replay/evm-t8n-substate-csv.py) is a python3 script that collects output log files of `evm-t8n-substate-0-9M.sh` and print data in CSV format.

Execute the experiment by typing the following:

```bash
# build substate replayer (evm)
cd ~/usenix-atc21/record-replay/go-ethereum/
make all

# measure replayer performance and print data in CSV
cd ../
./evm-t8n-substate-0-9M.sh
./evm-t8n-substate-csv.py
```

@Yeonsoo - please add here a sentence or two to say what (and where) the scripts produce and how does it match with the paper


## Substate Replayer - Space

This experiment measures the space required to replay transactions with the substate replayer in Table 2. The substate replayer contains the `evm dump-substate` command that reads `./stage1-substate/` and creates a database copy with substates found in a given range of blocks.

@Yeonsoo - it is not clear how this section matches with the paper

For example, to measure space required to replay transactions in 2-3M blocks,
```bash
# build substate replayer (evm)
cd ~/usenix-atc21/record-replay/go-ethereum/
make all

# copy substates of 2-3M blocks and measure database size
cd ../
evm dump-substate ./stage1-substate-2-3M/ 2000001 3000000
du -s ./stage1-substate-2-3M/
```

# Metrics Use Case

The metrics use case analyzes transactions by generating a graph of instruction flow. It counts the number of live instructions and live gases.
To produce a result of metrics and visualize it for 9M blocks execute:

@Seongho - please describe the parameter $threads (or maybe just hardcode a value, such as 4, 8, etc. - we may expect a user will run on a normal computer)

```bash
# build evm for value-graph metrics
cd ~/usenix-atc21/value-graph/go-ethereum
make all

# execute metrics analysis and visualize
cd ..
./metrics-0-9M.sh $threads
```

@Seongho  - please describe what the script produces, where it is, and how does it much with the paper (i.e. refer to particular tables, sections, etc)

@Seongho  - please use in the example exact parameters we used in the paper (unless it is 2000000)

To produce an image of single value graph, the following command generates a PNG image for the first transaction executed in the block 2000000.
```
cd ~/usenix-atc21/value-graph/go-ethereum/build/bin
evm t8n-substate 2000000 2000000 --workers 1 --skip-transfer-txs --skip-create-txs --graph
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

@Yeonsoo - it is not clear what are just examples and what is needed to be run for the experiment in the paper - please streamline it 

For example, to assess the Byzantium hard fork activated at block 4,370,000:
```bash
evm replay-fork 1 4369999 --skip-transfer-txs --skip-create-txs --hard-fork 4370000
```

Note that the Istanbul hard fork is activated at block 9,069,000, while we replayed transactions in 9M blocks:
```bash
evm replay-fork 1 9000000 --skip-transfer-txs --skip-create-txs --hard-fork 9069000
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

[replay-fork-0-9M.sh](./hard-fork/replay-fork-0-9M.sh) is a bash script to assess all hard forks activated before block 9,000,000 with CALL transactions (contract invocations) in initial 9M blocks.

```bash
# build evm for hard fork assessment
cd ~/usenix-atc21/hard-fork/go-ethereum/
make all

# run hard fork assessments with 9M blocks
cd ../
./replay-fork-0-9M.sh
```
