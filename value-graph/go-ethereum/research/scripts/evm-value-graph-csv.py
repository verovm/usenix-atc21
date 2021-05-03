#!/usr/bin/env python3

import os

header = "first,last,total,live"
print(header)
for name in sorted(os.listdir()):
    if name.startswith("evm-value-graph-") and name.endswith("k.log"):
        with open(name) as f:
            lines = f.readlines()
        for i in range(len(lines)):
            if lines[i].startswith(header):
                print(lines[i+1], end="")

