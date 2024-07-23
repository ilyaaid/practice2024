import math
import random 
import tqdm
import pandas as pd

def genBigGraph(p, q):
    # 2*10^9 узлов
    # 1.7*10^9 ребер
    # E?N = 1.7/2 = 0.85

    dencity = 0.85
    nodesNum = p*10**q
    l = int(math.log2(nodesNum))
    edgesNum = int(nodesNum*dencity)
    targetStrLen = 34
    print('-> generating', nodesNum, 'nodes and', edgesNum, 'edges')
    
    s = f'{0:0{l}b}'
    tobinst = lambda x: s.format(x)

    l1, l2 = [], []
    nodeMax = int(nodesNum*1.42)
    nodesSet = set()
    for i in tqdm.tqdm(range(edgesNum)):
        v1, v2 = random.randint(0, nodeMax), random.randint(0, nodeMax)
        nodesSet.add(v1)
        nodesSet.add(v2)
        s1, s2 = bin(v1), bin(v2)
        s1 += 'x'*(targetStrLen - len(s1))
        s2 += 'x'*(targetStrLen - len(s2))
        l1.append(s1)
        l2.append(s2)
    
    print("-> generated", len(nodesSet), 'nodes and', edgesNum, 'edges')
    # return 

    df = pd.DataFrame({
        "node1": l1, 
        "node2": l2
    })
    filename = f'test_graphs/big-{p}e{q}.csv'
    print('saving to', filename)
    df.to_csv(
        filename, 
        index=False, header=False
    )

genBigGraph(1, 1)
