from datetime import timedelta
import re


workers = [1, 2, 4, 8, 12, 16, 24, 32, 48, 64]
segments = ["{}-{}M".format(i-1, i) for i in range(1, 9+1)]


# 00h00m00.000s
delta_re = re.compile(r'((?P<hours>\d+?)h)?((?P<minutes>\d+?)m)?((?P<seconds>\d+(\.\d+)??)s)?')
def parse_delta(delta_str):
    parts = delta_re.match(delta_str)
    if not parts:
        return

    delta_params = parts.groupdict()
    for (name, param) in dict(delta_params).items():
        if param is None:
            del delta_params[name]
        else:
            delta_params[name] = round(float(param))

    return timedelta(**delta_params)

print("block", end="\t");
for w in workers:
    print(w, end="\t")
print()

for s in segments:
    print(s.replace("-", "--"), end="\t")
    for w in workers:
        logname = "evm-t8n-substate-w{}-{}.log".format(w, s)
        logfile = open(logname)
        loglines = logfile.readlines()
        prefix = "stage1-substate: TransitionSubstate done in "
        for line in loglines:
            if line[:len(prefix)] == prefix:
                delta_str = line[len(prefix):].strip()
                delta = parse_delta(delta_str)
                print(delta.total_seconds(), end="\t")
                break
    print()
