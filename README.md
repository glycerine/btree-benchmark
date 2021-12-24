# btree-benchmark

Benchmark utility for the [tidwall/btree](https://github.com/tidwall/btree) Go package

- `google`: The [google/btree](https://github.com/google/btree) package (without generics)
- `tidwall`: The [tidwall/btree](https://github.com/tidwall/btree) package (without generics)
- `tidwall(G)`: The [tidwall/btree](https://github.com/tidwall/btree) package (generics using the `btree.Generic` type)
- `tidwall(M)`: The [tidwall/btree](https://github.com/tidwall/btree) package (generics using the `btree.Map` type)
- `go-arr`: A simple Go array

The following benchmarks were run on my 2019 Macbook Pro (2.4 GHz 8-Core Intel Core i9) 
using Go Development version 1.18 (beta1).
The items are simple 8-byte ints. 

```
** sequential set **
google:     set-seq        1,000,000 ops in 156ms, 6,426,724/sec, 155 ns/op, 39.0 MB, 40 bytes/op
tidwall:    set-seq        1,000,000 ops in 135ms, 7,380,627/sec, 135 ns/op, 23.5 MB, 24 bytes/op
tidwall(G): set-seq        1,000,000 ops in 78ms, 12,881,995/sec, 77 ns/op, 8.2 MB, 8 bytes/op
tidwall(M): set-seq        1,000,000 ops in 46ms, 21,892,141/sec, 45 ns/op, 8.2 MB, 8 bytes/op
tidwall:    set-seq-hint   1,000,000 ops in 73ms, 13,789,017/sec, 72 ns/op, 23.5 MB, 24 bytes/op
tidwall(G): set-seq-hint   1,000,000 ops in 48ms, 20,969,431/sec, 47 ns/op, 8.2 MB, 8 bytes/op
tidwall:    load-seq       1,000,000 ops in 45ms, 22,452,523/sec, 44 ns/op, 23.5 MB, 24 bytes/op
tidwall(G): load-seq       1,000,000 ops in 22ms, 46,242,274/sec, 21 ns/op, 8.2 MB, 8 bytes/op
tidwall(M): load-seq       1,000,000 ops in 13ms, 74,371,903/sec, 13 ns/op, 8.2 MB, 8 bytes/op
go-arr:     append         1,000,000 ops in 21ms, 47,141,875/sec, 21 ns/op, 8.1 MB, 8 bytes/op

** sequential get **
google:     get-seq        1,000,000 ops in 119ms, 8,389,459/sec, 119 ns/op
tidwall:    get-seq        1,000,000 ops in 110ms, 9,068,759/sec, 110 ns/op
tidwall(G): get-seq        1,000,000 ops in 78ms, 12,813,135/sec, 78 ns/op
tidwall(M): get-seq        1,000,000 ops in 62ms, 16,053,728/sec, 62 ns/op
tidwall:    get-seq-hint   1,000,000 ops in 64ms, 15,509,696/sec, 64 ns/op
tidwall(G): get-seq-hint   1,000,000 ops in 41ms, 24,144,951/sec, 41 ns/op

** random set **
google:     set-rand       1,000,000 ops in 563ms, 1,777,592/sec, 562 ns/op, 29.7 MB, 31 bytes/op
tidwall:    set-rand       1,000,000 ops in 542ms, 1,844,397/sec, 542 ns/op, 29.6 MB, 31 bytes/op
tidwall(G): set-rand       1,000,000 ops in 234ms, 4,271,764/sec, 234 ns/op, 11.2 MB, 11 bytes/op
tidwall(M): set-rand       1,000,000 ops in 189ms, 5,292,236/sec, 188 ns/op, 11.2 MB, 11 bytes/op
tidwall:    set-rand-hint  1,000,000 ops in 602ms, 1,659,852/sec, 602 ns/op, 29.6 MB, 31 bytes/op
tidwall(G): set-rand-hint  1,000,000 ops in 278ms, 3,595,435/sec, 278 ns/op, 11.2 MB, 11 bytes/op
tidwall:    set-after-copy 1,000,000 ops in 679ms, 1,471,954/sec, 679 ns/op
tidwall(G): set-after-copy 1,000,000 ops in 238ms, 4,196,854/sec, 238 ns/op
tidwall:    load-rand      1,000,000 ops in 532ms, 1,880,877/sec, 531 ns/op, 29.6 MB, 31 bytes/op
tidwall(G): load-rand      1,000,000 ops in 232ms, 4,316,475/sec, 231 ns/op, 11.2 MB, 11 bytes/op
tidwall(M): load-rand      1,000,000 ops in 209ms, 4,790,169/sec, 208 ns/op, 11.2 MB, 11 bytes/op

** random get **
google:     get-rand       1,000,000 ops in 807ms, 1,238,703/sec, 807 ns/op
tidwall:    get-rand       1,000,000 ops in 812ms, 1,231,551/sec, 811 ns/op
tidwall(G): get-rand       1,000,000 ops in 255ms, 3,914,819/sec, 255 ns/op
tidwall(M): get-rand       1,000,000 ops in 190ms, 5,249,966/sec, 190 ns/op
tidwall:    get-rand-hint  1,000,000 ops in 876ms, 1,141,313/sec, 876 ns/op
tidwall(G): get-rand-hint  1,000,000 ops in 258ms, 3,877,775/sec, 257 ns/op

** range **
google:     ascend        1,000,000 ops in 26ms, 39,101,882/sec, 25 ns/op
tidwall:    ascend        1,000,000 ops in 20ms, 50,223,988/sec, 19 ns/op
tidwall(G): iter          1,000,000 ops in 8ms, 119,155,937/sec, 8 ns/op
tidwall(G): scan          1,000,000 ops in 6ms, 168,275,407/sec, 5 ns/op
tidwall(G): walk          1,000,000 ops in 5ms, 186,941,046/sec, 5 ns/op
go-arr:     for-loop      1,000,000 ops in 4ms, 272,234,997/sec, 3 ns/op
```
