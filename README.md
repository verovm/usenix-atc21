# usenix-atc21

1. go-ethereum/

Original go-ethereum v1.9.18 source code

2. record-replay/

Transaction substate recorder and replayer

Substate DB of 9M blocks: [link](https://drive.google.com/file/d/1jl6vdMea5ROKdrTUJUk8lh5NL48Do9xJ/view?usp=sharing)
```
# untar substate DB
tar -xavf stage1-substate-0-9M.tar.zst
mv stage1-substate-0-9M stage1-substate
```

3. value-graph/

Use case 1: Metric

4. contract-fuzzer/

Use case 2: Fuzzer

5. hard-fork/

Use case 3: Hard-fork assessment
