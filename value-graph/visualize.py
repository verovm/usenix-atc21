import pandas as pd
import matplotlib.pyplot as plt
import numpy as np
#import tikzplotlib
#from tikzplotlib import save as tikz_save

for n in range(0, 9):
    filename = str(n) + 'M.csv'
    raw = pd.read_csv(filename, usecols=[0,1,2,3,4,5])
    data = raw.sort_values(['block', 'txIndex'], ascending = (True, True))
    data['countInst'] = data['totalInst']-data['liveInst']
    data['ratioInst'] = ((data['totalInst']-data['liveInst']) / data['totalInst']).fillna(0)
    data['countGas'] = data['totalGas']-data['liveGas']
    data['ratioGas'] = ((data['totalGas']-data['liveGas']) / data['totalGas']).fillna(0)
    data.to_csv('refined_' + filename, index = False)


plt.style.use("ggplot")
plt.rcParams["figure.dpi"] = 300
label = ['0-1M', '1-2M', '2-3M', '3-4M', '4-5M', '5-6M', '6-7M', '7-8M', '8-9M'] 

deadInstRatio = []
wasteGasPerTx = []
gasPerTx = []

print("Reading data...")
for bn in range(0,9):
    filename = 'refined_' + str(bn) + "M.csv"
    data = pd.read_csv(filename,usecols=[0,1,2,3,4])
    deadInstRatio.append((data['totalInst'].sum()-data['liveInst'].sum())/data['totalInst'].sum()*100)
    wasteGasPerTx.append((data['totalGas'].sum()-data['liveGas'].sum())/len(data))
    gasPerTx.append(data['totalGas'].sum()/len(data))

print("Generating avg_ratio.png...")
df = pd.DataFrame({'Block':label, 'Unnecessary instruction':deadInstRatio})
df.plot.bar(x='Block', y='Unnecessary instruction', rot=20, legend=None)
plt.ylabel('Percentage (%)')
plt.savefig('avg_ratio.png')
#tikz_save('avg_ratio.tex', axis_height='6cm', axis_width='9cm')
plt.close()

print("Generating wasted_gas.png...")
df = pd.DataFrame({'Block':label, 'Wasted gas per TX':wasteGasPerTx})
df.plot.bar(x='Block', y='Wasted gas per TX', rot=20, legend=None)
plt.legend('', frameon=False)
plt.ylabel('Gas (k)')
plt.savefig('wasted_gas.png')
#tikz_save('wasted_gas.tex', axis_height='6cm', axis_width='9cm')
plt.close()


print("Generating boxplot.png...")
ratioInst = pd.DataFrame()
for bn in range(0,9):
    filename = 'refined_' + str(bn) + "M.csv"
    data = pd.read_csv(filename,usecols=[0,7])
    tmp = pd.DataFrame()
    tmp[str(bn)+'M'] = data['ratioInst']*100
    ratioInst = pd.concat([ratioInst, tmp], ignore_index=True, axis=1)

ratioInst.boxplot(showfliers=False, rot=20)
plt.xticks(np.arange(len(label)), label)
plt.xlabel('Block')
plt.ylabel('Percentage (%)')
plt.ylim(-0.2, 100.2)
plt.savefig('boxplot.png')
#tikz_save('boxplot.tex', axis_height='6cm', axis_width='9cm')
plt.close()
