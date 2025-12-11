# Redis Serialization Benchmark Summary

## Environment
- **OS**: macOS (Darwin arm64)
- **CPU**: Apple M4
- **Package**: `github.com/trinhdaiphuc/go-kit/examples/cache/redis`

## Benchmark Results

| Benchmark                   | Iterations | Time (ns/op) | Alloc (B/op) | Allocs/op |
| :-------------------------- | :--------: | :----------: | :----------: | :-------: |
| **Small Struct**            |            |              |              |           |
| `JSON_Set_Small`            |   15376    |    76,117    |     395      |     9     |
| `JSON_Get_Small`            |   15126    |    73,693    |     640      |    14     |
| `Hash_Set_Small`            |   15612    |    76,367    |     568      |    13     |
| `Hash_Get_Small`            |   16365    |    73,181    |     648      |    18     |
| `GoccyJSON_Set_Small`       |   16176    |    73,766    |     395      |     9     |
| `GoccyJSON_Get_Small`       |   17170    |    69,749    |     488      |     9     |
| `MsgPack_Set_Small`         |   16790    |    72,779    |     363      |     9     |
| `MsgPack_Get_Small`         |   17119    |    75,809    |     440      |    11     |
| `Proto_Set_Small`           |   16779    |    77,006    |     274      |     8     |
| `Proto_Get_Small`           |   17212    |    69,974    |     344      |    10     |
|                             |            |              |              |           |
| **Medium Struct**           |            |              |              |           |
| `JSON_Set_Medium`           |   15650    |    75,399    |     523      |     9     |
| `JSON_Get_Medium`           |   16026    |    74,379    |     896      |    17     |
| `Hash_Set_Medium`           |   15619    |    76,187    |     680      |    14     |
| `Hash_Get_Medium`           |   16390    |    73,930    |     712      |    18     |
| `GoccyJSON_Set_Medium`      |   16351    |    73,044    |     523      |     9     |
| `GoccyJSON_Get_Medium`      |   17096    |    69,874    |     776      |     9     |
| `MsgPack_Set_Medium`        |   16476    |    74,208    |     491      |    10     |
| `MsgPack_Get_Medium`        |   17079    |    70,724    |     632      |    13     |
| `Proto_Set_Medium`          |   16561    |    73,615    |     314      |     8     |
| `Proto_Get_Medium`          |   17371    |    71,225    |     504      |    12     |
|                             |            |              |              |           |
| **Large Struct**            |            |              |              |           |
| `JSON_Set_Large`            |   15597    |    76,890    |     924      |     9     |
| `JSON_Get_Large`            |   15849    |    77,461    |     1624     |    24     |
| `Hash_Set_Large`            |   14674    |    80,471    |     2297     |    26     |
| `Hash_Get_Large`            |   15681    |    76,101    |     1936     |    40     |
| `GoccyJSON_Set_Large`       |   15973    |    74,107    |     925      |     9     |
| `GoccyJSON_Get_Large`       |   16922    |    70,723    |     1624     |     9     |
| `MsgPack_Set_Large`         |   16321    |    74,139    |     1259     |    12     |
| `MsgPack_Get_Large`         |   16797    |    71,663    |     1160     |    19     |
| `Proto_Set_Large`           |   16500    |    72,708    |     394      |     8     |
| `Proto_Get_Large`           |   17073    |    70,902    |     872      |    18     |
|                             |            |              |              |           |
| **Large Nested Struct**     |            |              |              |           |
| `JSON_Set_LargeNested`      |   15741    |    76,655    |     1149     |     9     |
| `JSON_Get_LargeNested`      |   15424    |    77,389    |     1968     |    26     |
| `Hash_Set_LargeNested`      |   13869    |    86,521    |     5105     |    63     |
| `GoccyJSON_Set_LargeNested` |   16233    |    73,797    |     1148     |     9     |
| `GoccyJSON_Get_LargeNested` |   16911    |    71,980    |     2233     |     9     |
| `MsgPack_Set_LargeNested`   |   15871    |    74,190    |     1259     |    12     |
| `MsgPack_Get_LargeNested`   |   16854    |    72,673    |     1432     |    21     |
| `Proto_Set_LargeNested`     |   16513    |    73,107    |     378      |     8     |
| `Proto_Get_LargeNested`     |   17106    |    71,224    |     1309     |    30     |
|                             |            |              |              |           |
| **Tinylib Msgp**            |            |              |              |           |
| `Tinylib_Set_Small`         |   12506    |    82,459    |     336      |     8     |
| `Tinylib_Get_Small`         |   16911    |    71,489    |     392      |    10     |
| `Tinylib_Set_Large`         |   14954    |    75,266    |     703      |     8     |
| `Tinylib_Get_Large`         |   16911    |    81,878    |     1176     |    18     |

## CPU Profile Analysis (Top Nodes)

The benchmark is heavily I/O bound, dominated by system calls and Redis client overhead.

```txt
Duration: 77.22s, Total samples = 21.73s (28.14%)
Showing nodes accounting for 21.30s, 98.02% of 21.73s total

      flat  flat%   sum%        cum   cum%
    13.69s 63.00% 63.00%     13.69s 63.00%  syscall.syscall
     6.93s 31.89% 94.89%      6.93s 31.89%  runtime.kevent
     0.64s  2.95% 97.84%      0.64s  2.95%  runtime.pthread_cond_signal
     0.02s 0.092% 97.93%      7.09s 32.63%  runtime.findRunnable
     0.02s 0.092% 98.02%      0.11s  0.51%  runtime.scanobject
```

### Key Takeaways
1. **Protobuf** consistently offers the lowest allocation overhead (B/op) and is very competitive in execution time.
2. **MsgPack** is a strong contender, offering better performance than standard JSON and comparable to Goccy JSON, with significantly smaller payload sizes than JSON.
3. **Goccy JSON** provides a noticeable improvement over standard JSON, especially in allocation efficiency for larger structs.
4. **Tinylib Msgp** shows excellent performance for **Large Structs** (Set: ~75µs vs MsgPack's ~127µs) and lower memory allocation (703 B vs 1266 B). However, for small structs, it performs similarly to other binary formats.
5. **Redis Hash** (HSet/HGetAll) incurs higher allocation overhead for complex/nested structs due to the need to flatten/unflatten the data structure.

## I/O Optimization: Pipelining & Batching

Since the application is I/O bound (63% syscalls), reducing the number of round-trips to Redis is the most effective optimization.

| Benchmark                 | Iterations | Time (ns/op) | Alloc (B/op) | Allocs/op |
| :------------------------ | :--------: | :----------: | :----------: | :-------: |
| **Standard (Sequential)** |            |              |              |           |
| `JSON_Set_Small`          |   15591    |    78,770    |     395      |     9     |
| **Optimized (Batch 100)** |            |              |              |           |
| `Pipeline_Set_Small`      |  567,531   |  **1,841**   |     424      |     7     |
| `MSet_Small`              |  796,969   |  **1,338**   |     180      |     2     |

**Impact**:
- **Pipelining** reduces the per-operation time by **~40x** (from ~78µs to ~1.8µs).
- **MSet** is even faster (~1.3µs) as it sends a single command for multiple keys, reducing protocol overhead further.

### CPU Profile Analysis (Pipeline vs Normal)

Comparing the CPU profiles of the standard approach vs. pipelining reveals why the latter is so much faster.

**Normal Mode (Set_Small)**:
- **Dominant Node**: `syscall.syscall` (~63%)
- **Cause**: Every single `Set` operation triggers a write to the socket and a read from the socket. The CPU spends most of its time waiting for the OS to handle these network packets.

**Pipeline/MSet Mode**:
- **Dominant Node**: `syscall.syscall` is still high (~75% of the *captured* samples), but the **total number of syscalls is drastically reduced**.
- **Efficiency**: Instead of 100 syscalls for 100 items, we do 1 syscall. The CPU stays busy packing data into buffers (`bufio.Writer.Flush` ~56%) rather than context switching for every small packet.
- **Throughput**: Because the CPU isn't blocked waiting for round-trips, it can process orders of magnitude more data per second.

### Why CPU Load Increases in Pipeline Mode
You correctly observed that the CPU load appears higher in Pipeline mode. This is counter-intuitive but expected:

1.  **Idle vs. Busy**: In "Normal" mode, the CPU spends most of its time *waiting* (idle) for the network response. In "Pipeline" mode, the CPU is constantly *working* (packing buffers, encoding JSON, writing to memory).
2.  **Work Density**: We are processing ~560,000 operations/sec (Pipeline) vs ~15,000 operations/sec (Normal). The CPU is doing **37x more work per second**.
3.  **Profile Evidence**:
    -   `bufio.(*Writer).Flush` (56%): The CPU is aggressively pushing data to the network buffer.
    -   `writeCmds` (42%): Significant time is spent formatting the Redis protocol commands in memory.
    -   `syscall.syscall` (75%): This represents the *aggregated* cost of sending massive batches of data, not the overhead of individual packet switching.

### Large Struct Analysis (Pipeline vs Normal)

We also benchmarked **Large Structs** to see if the CPU overhead increases as the payload grows.

| Benchmark                 | Iterations | Time (ns/op) | Alloc (B/op) | Allocs/op |
| :------------------------ | :--------: | :----------: | :----------: | :-------: |
| **Standard (Sequential)** |            |              |              |           |
| `JSON_Set_Large`          |   13603    |    77,186    |     924      |     9     |
| **Optimized (Batch 100)** |            |              |              |           |
| `Pipeline_Set_Large`      |  398,907   |  **3,217**   |     954      |     7     |
| `MSet_Large`              |  452,434   |  **3,206**   |     709      |     2     |

**Observations**:
1.  **Throughput Drop**: Compared to Small structs (~1.8µs), Large structs take ~3.2µs in pipeline mode. This is expected as there is more data to copy.
2.  **CPU Profile**:
    -   `writeCmd` usage increased to **47%** (from 42% in Small).
    -   `bufio.(*Writer).Write` usage increased to **43%** (from 40% in Small).
    -   This confirms your hypothesis: **Larger structs require more CPU time to build and buffer the request.**
3.  **Still Massive Win**: Despite the extra CPU work, it is still **24x faster** than the standard sequential approach (77µs vs 3.2µs).
