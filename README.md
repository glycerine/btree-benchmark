# btree-benchmark

Benchmark utility for the [tidwall/btree](https://github.com/tidwall/btree) Go package

- `google`: The [google/btree](https://github.com/google/btree) package (without generics)
- `google(G)`: The [google/btree](https://github.com/google/btree) package (generics)
- `tidwall`: The [tidwall/btree](https://github.com/tidwall/btree) package (without generics)
- `tidwall(G)`: The [tidwall/btree](https://github.com/tidwall/btree) package (generics)
- `tidwall(M)`: The [tidwall/btree](https://github.com/tidwall/btree) package (generics using the `btree.Map` type)

The following benchmarks were run on my 2021 Macbook Pro M1 Max 
using Go version 1.20.4.  
All items are key/value pairs where the key is a string filled with 16 random digits such as `5204379379828236`, and the value is the int64 representation of the key.
The degree is 32.  

```
degree=32, key=string (16 bytes), val=int64, count=1000000

** sequential set **
google:     set-seq           1,000,000 ops in 219ms, 4,568,211/sec, 218 ns/op, 56.9 MB, 59.6 bytes/op
google(G):  set-seq           1,000,000 ops in 167ms, 5,991,695/sec, 166 ns/op, 49.7 MB, 52.1 bytes/op
tidwall:    set-seq           1,000,000 ops in 139ms, 7,186,487/sec, 139 ns/op, 56.4 MB, 59.1 bytes/op
tidwall(G): set-seq           1,000,000 ops in 119ms, 8,419,681/sec, 118 ns/op, 49.2 MB, 51.6 bytes/op
tidwall(M): set-seq           1,000,000 ops in 102ms, 9,761,406/sec, 102 ns/op, 49.2 MB, 51.6 bytes/op
tidwall:    set-seq-hint      1,000,000 ops in 87ms, 11,538,644/sec, 86 ns/op, 56.4 MB, 59.1 bytes/op
tidwall(G): set-seq-hint      1,000,000 ops in 58ms, 17,338,334/sec, 57 ns/op, 49.2 MB, 51.6 bytes/op
tidwall:    load-seq          1,000,000 ops in 53ms, 18,909,653/sec, 52 ns/op, 56.4 MB, 59.1 bytes/op
tidwall(G): load-seq          1,000,000 ops in 32ms, 31,300,333/sec, 31 ns/op, 49.2 MB, 51.6 bytes/op
tidwall(M): load-seq          1,000,000 ops in 28ms, 36,346,564/sec, 27 ns/op, 49.2 MB, 51.6 bytes/op

** sequential get **
google:     get-seq           1,000,000 ops in 207ms, 4,835,779/sec, 206 ns/op
google(G):  get-seq           1,000,000 ops in 169ms, 5,927,806/sec, 168 ns/op
tidwall:    get-seq           1,000,000 ops in 156ms, 6,405,236/sec, 156 ns/op
tidwall(G): get-seq           1,000,000 ops in 125ms, 8,023,836/sec, 124 ns/op
tidwall(M): get-seq           1,000,000 ops in 102ms, 9,822,836/sec, 101 ns/op
tidwall:    get-seq-hint      1,000,000 ops in 84ms, 11,866,425/sec, 84 ns/op
tidwall(G): get-seq-hint      1,000,000 ops in 53ms, 18,696,903/sec, 53 ns/op

** random set **
google:     set-rand          1,000,000 ops in 1134ms, 881,661/sec, 1134 ns/op, 46.5 MB, 48.8 bytes/op
google(G):  set-rand          1,000,000 ops in 743ms, 1,345,894/sec, 743 ns/op, 34.6 MB, 36.3 bytes/op
tidwall:    set-rand          1,000,000 ops in 838ms, 1,193,269/sec, 838 ns/op, 46.5 MB, 48.8 bytes/op
tidwall(G): set-rand          1,000,000 ops in 571ms, 1,751,498/sec, 570 ns/op, 34.8 MB, 36.5 bytes/op
tidwall(M): set-rand          1,000,000 ops in 560ms, 1,784,123/sec, 560 ns/op, 34.8 MB, 36.5 bytes/op
tidwall:    set-rand-hint     1,000,000 ops in 888ms, 1,125,569/sec, 888 ns/op, 46.5 MB, 48.8 bytes/op
tidwall(G): set-rand-hint     1,000,000 ops in 606ms, 1,648,850/sec, 606 ns/op, 34.8 MB, 36.5 bytes/op
tidwall:    set-after-copy    1,000,000 ops in 974ms, 1,026,808/sec, 973 ns/op
tidwall(G): set-after-copy    1,000,000 ops in 612ms, 1,634,365/sec, 611 ns/op
tidwall:    load-rand         1,000,000 ops in 868ms, 1,151,739/sec, 868 ns/op, 46.5 MB, 48.8 bytes/op
tidwall(G): load-rand         1,000,000 ops in 594ms, 1,682,555/sec, 594 ns/op, 34.8 MB, 36.5 bytes/op
tidwall(M): load-rand         1,000,000 ops in 551ms, 1,816,466/sec, 550 ns/op, 34.8 MB, 36.5 bytes/op

** random get **
google:     get-rand          1,000,000 ops in 1523ms, 656,452/sec, 1523 ns/op
google(G):  get-rand          1,000,000 ops in 858ms, 1,165,715/sec, 857 ns/op
tidwall:    get-rand          1,000,000 ops in 1110ms, 901,008/sec, 1109 ns/op
tidwall(G): get-rand          1,000,000 ops in 655ms, 1,526,501/sec, 655 ns/op
tidwall(M): get-rand          1,000,000 ops in 625ms, 1,599,495/sec, 625 ns/op
tidwall:    get-rand-hint     1,000,000 ops in 1159ms, 863,083/sec, 1158 ns/op
tidwall(G): get-rand-hint     1,000,000 ops in 716ms, 1,395,710/sec, 716 ns/op

** sequential pivot **
Test getting 10 consecutive items starting at a pivot.
google:     ascend-seq        1,000,000 ops in 341ms, 2,929,622/sec, 341 ns/op
google:     descend-seq       1,000,000 ops in 431ms, 2,322,505/sec, 430 ns/op
google(G):  ascend-seq        1,000,000 ops in 213ms, 4,702,900/sec, 212 ns/op
google(G):  descend-seq       1,000,000 ops in 275ms, 3,638,184/sec, 274 ns/op
tidwall:    ascend-seq        1,000,000 ops in 197ms, 5,075,743/sec, 197 ns/op
tidwall:    descend-seq       1,000,000 ops in 205ms, 4,883,863/sec, 204 ns/op
tidwall:    ascend-seq-hint   1,000,000 ops in 121ms, 8,252,368/sec, 121 ns/op
tidwall:    descend-seq-hint  1,000,000 ops in 126ms, 7,950,834/sec, 125 ns/op
tidwall(G): ascend-seq        1,000,000 ops in 151ms, 6,631,924/sec, 150 ns/op
tidwall(G): descend-seq       1,000,000 ops in 153ms, 6,521,033/sec, 153 ns/op
tidwall(G): ascend-seq-hint   1,000,000 ops in 80ms, 12,487,434/sec, 80 ns/op
tidwall(G): descend-seq-hint  1,000,000 ops in 81ms, 12,329,924/sec, 81 ns/op
tidwall(G): iter-seq          1,000,000 ops in 213ms, 4,691,868/sec, 213 ns/op
tidwall(G): iter-seq-hint     1,000,000 ops in 138ms, 7,248,690/sec, 137 ns/op

** random pivot **
Test getting 10 consecutive items starting at a pivot.
google:     ascend-rand       1,000,000 ops in 1916ms, 521,904/sec, 1916 ns/op
google:     descend-rand      1,000,000 ops in 2351ms, 425,348/sec, 2351 ns/op
google(G):  ascend-rand       1,000,000 ops in 1043ms, 959,150/sec, 1042 ns/op
google(G):  descend-rand      1,000,000 ops in 1173ms, 852,528/sec, 1172 ns/op
tidwall:    ascend-rand       1,000,000 ops in 1166ms, 857,545/sec, 1166 ns/op
tidwall:    descend-rand      1,000,000 ops in 1164ms, 858,910/sec, 1164 ns/op
tidwall:    ascend-rand-hint  1,000,000 ops in 1211ms, 825,455/sec, 1211 ns/op
tidwall:    descend-rand-hint 1,000,000 ops in 1224ms, 817,205/sec, 1223 ns/op
tidwall(G): ascend-rand       1,000,000 ops in 726ms, 1,376,815/sec, 726 ns/op
tidwall(G): descend-rand      1,000,000 ops in 712ms, 1,403,801/sec, 712 ns/op
tidwall(G): ascend-rand-hint  1,000,000 ops in 771ms, 1,296,249/sec, 771 ns/op
tidwall(G): descend-rand-hint 1,000,000 ops in 768ms, 1,302,896/sec, 767 ns/op
tidwall(G): iter-rand         1,000,000 ops in 811ms, 1,232,549/sec, 811 ns/op
tidwall(G): iter-rand-hint    1,000,000 ops in 848ms, 1,179,822/sec, 847 ns/op

** scan **
Test scanning over every item in the tree
google:     ascend            1,000,000 ops in 10ms, 97,801,108/sec, 10 ns/op
google(G):  ascend            1,000,000 ops in 11ms, 89,880,910/sec, 11 ns/op
tidwall:    ascend            1,000,000 ops in 8ms, 124,326,554/sec, 8 ns/op
tidwall(G): scan              1,000,000 ops in 9ms, 108,123,701/sec, 9 ns/op
tidwall(G): walk              1,000,000 ops in 3ms, 331,945,689/sec, 3 ns/op
tidwall(G): iter              1,000,000 ops in 11ms, 94,063,785/sec, 10 ns/op
```

Adding skipmap, locked tidwall(G), and a skiplist:

```
Compilation started at Mon Mar 17 22:03:58

go run main.go

degree=32, key=string (16 bytes), val=int64, count=1000000

** sequential set **
tidwall(G) with locking: 270ms, 3,705,761/sec, 269 ns/op, 49.2 MB, 51.6 bytes/op
google(G): (no locking)  281ms, 3,557,008/sec, 281 ns/op, 49.7 MB, 52.1 bytes/op
tidwall:    set-seq      285ms, 3,508,612/sec, 285 ns/op, 56.4 MB, 59.1 bytes/op
uART:       set-seq      288ms, 3,466,263/sec, 288 ns/op, 166.0 MB,174.1 bytes/op
zhangyunhao116/skipmap:  304ms, 3,294,504/sec, 303 ns/op, 99.5 MB, 104.4 bytes/op
badger/skiplist:         353ms, 2,832,188/sec, 353 ns/op

** sequential delete **
zhangyunhao116/skipmap:  1,000,000 ops in 16ms, 63,214,721/sec, 15 ns/op
tidwall(G) with locking: 1,000,000 ops in 33ms, 30,352,844/sec, 32 ns/op
uART:       seq-delete   1,000,000 ops in 122ms, 8,218,107/sec, 121 ns/op
tidwall(G): seq-delete   1,000,000 ops in 248ms, 4,039,629/sec, 247 ns/op
google(G):  seq-delete   1,000,000 ops in 254ms, 3,939,167/sec, 253 ns/op
google:     seq-delete   1,000,000 ops in 320ms, 3,128,727/sec, 319 ns/op


** random set **
tidwall(M): (no locking)      1,000,000 ops in 989ms, 1,011,151/sec, 988 ns/op, 34.9 MB, 36.6 bytes/op
google(G): (no locking)       1,000,000 ops in 997ms, 1,003,214/sec, 996 ns/op, 35.0 MB, 36.7 bytes/op
tidwall(G) with locking:      1,000,000 ops in 1054ms, 948,954/sec, 1053 ns/op, 34.9 MB, 36.6 bytes/op
badger/skiplist: set-rand     1,000,000 ops in 1554ms, 643,388/sec, 1554 ns/op
zhangyunhao116/skipmap:       1,000,000 ops in 2065ms, 484,192/sec, 2065 ns/op, 99.5 MB, 104.4 bytes/op

full run, darwin (Sonoma 14.0) go1.24.1

Compilation started at Mon Mar 17 23:54:39

go run main.go

degree=32, key=string (16 bytes), val=int64, count=1000000

** sequential set **
google:     set-seq           1,000,000 ops in 309ms, 3,235,436/sec, 309 ns/op, 60.8 MB, 63.7 bytes/op
google(G):  set-seq           1,000,000 ops in 280ms, 3,569,261/sec, 280 ns/op, 49.7 MB, 52.1 bytes/op
tidwall:    set-seq           1,000,000 ops in 304ms, 3,292,711/sec, 303 ns/op, 56.4 MB, 59.1 bytes/op
tidwall(G): set-seq           1,000,000 ops in 258ms, 3,877,245/sec, 257 ns/op, 49.2 MB, 51.6 bytes/op
tidwall(M): set-seq           1,000,000 ops in 213ms, 4,693,959/sec, 213 ns/op, 49.2 MB, 51.6 bytes/op
tidwall(G) with locking: set-seq           1,000,000 ops in 280ms, 3,576,637/sec, 279 ns/op, 49.2 MB, 51.6 bytes/op
badger/skiplist: set-seq           1,000,000 ops in 350ms, 2,859,587/sec, 349 ns/op
zhangyunhao116/skipmap: set-seq           1,000,000 ops in 340ms, 2,943,933/sec, 339 ns/op, 99.5 MB, 104.4 bytes/op
uART:       set-seq           1,000,000 ops in 333ms, 3,001,450/sec, 333 ns/op, 166.1 MB, 174.1 bytes/op
tidwall:    set-seq-hint      1,000,000 ops in 174ms, 5,761,038/sec, 173 ns/op, 56.4 MB, 59.1 bytes/op
tidwall(G): set-seq-hint      1,000,000 ops in 147ms, 6,787,871/sec, 147 ns/op, 49.2 MB, 51.6 bytes/op
tidwall:    load-seq          1,000,000 ops in 139ms, 7,181,074/sec, 139 ns/op, 56.4 MB, 59.1 bytes/op
tidwall(G): load-seq          1,000,000 ops in 74ms, 13,431,216/sec, 74 ns/op, 49.2 MB, 51.6 bytes/op
tidwall(M): load-seq          1,000,000 ops in 69ms, 14,536,901/sec, 68 ns/op, 49.2 MB, 51.6 bytes/op

** sequential get **
google:     get-seq           1,000,000 ops in 276ms, 3,622,003/sec, 276 ns/op
google(G):  get-seq           1,000,000 ops in 209ms, 4,792,806/sec, 208 ns/op
tidwall:    get-seq           1,000,000 ops in 275ms, 3,635,404/sec, 275 ns/op
tidwall(G): get-seq           1,000,000 ops in 220ms, 4,548,412/sec, 219 ns/op
tidwall(M): get-seq           1,000,000 ops in 188ms, 5,318,605/sec, 188 ns/op
tidwall:    get-seq-hint      1,000,000 ops in 176ms, 5,672,012/sec, 176 ns/op
tidwall(G): get-seq-hint      1,000,000 ops in 143ms, 7,005,698/sec, 142 ns/op

** sequential delete **
tidwall(G): seq-delete        1,000,000 ops in 238ms, 4,202,762/sec, 237 ns/op
uART:       seq-delete        1,000,000 ops in 111ms, 9,005,916/sec, 111 ns/op
google(G):  seq-delete        1,000,000 ops in 243ms, 4,108,229/sec, 243 ns/op
google:     seq-delete        1,000,000 ops in 310ms, 3,225,549/sec, 310 ns/op
tidwall(G) with locking: seq-delete        1,000,000 ops in 33ms, 30,370,789/sec, 32 ns/op
zhangyunhao116/skipmap: seq-delete        1,000,000 ops in 17ms, 58,613,475/sec, 17 ns/op

** random set **
google:     set-rand          1,000,000 ops in 1552ms, 644,381/sec, 1551 ns/op, 49.2 MB, 51.6 bytes/op
google(G):  set-rand          1,000,000 ops in 972ms, 1,028,429/sec, 972 ns/op, 35.1 MB, 36.8 bytes/op
tidwall:    set-rand          1,000,000 ops in 1579ms, 633,384/sec, 1578 ns/op, 46.7 MB, 48.9 bytes/op
tidwall(G): set-rand          1,000,000 ops in 1025ms, 975,594/sec, 1025 ns/op, 35.0 MB, 36.7 bytes/op
uART:       set-rand          1,000,000 ops in 1040ms, 961,299/sec, 1040 ns/op, 165.7 MB, 173.8 bytes/op
tidwall(M): set-rand          1,000,000 ops in 976ms, 1,024,343/sec, 976 ns/op, 35.0 MB, 36.7 bytes/op
tidwall(G) with locking: set-rand          1,000,000 ops in 1044ms, 957,457/sec, 1044 ns/op, 35.0 MB, 36.7 bytes/op
badger/skiplist: set-rand          1,000,000 ops in 1511ms, 661,621/sec, 1511 ns/op
zhangyunhao116/skipmap: set-rand          1,000,000 ops in 2039ms, 490,437/sec, 2038 ns/op, 99.5 MB, 104.4 bytes/op
tidwall:    set-rand-hint     1,000,000 ops in 1618ms, 618,176/sec, 1617 ns/op, 46.7 MB, 48.9 bytes/op
tidwall(G): set-rand-hint     1,000,000 ops in 1060ms, 943,603/sec, 1059 ns/op, 35.0 MB, 36.7 bytes/op
tidwall:    set-after-copy    1,000,000 ops in 1800ms, 555,614/sec, 1799 ns/op
tidwall(G): set-after-copy    1,000,000 ops in 1107ms, 903,194/sec, 1107 ns/op
tidwall:    load-rand         1,000,000 ops in 1568ms, 637,947/sec, 1567 ns/op, 46.7 MB, 48.9 bytes/op
tidwall(G): load-rand         1,000,000 ops in 1014ms, 985,720/sec, 1014 ns/op, 35.0 MB, 36.7 bytes/op
tidwall(M): load-rand         1,000,000 ops in 987ms, 1,012,837/sec, 987 ns/op, 35.0 MB, 36.7 bytes/op

** random delete **
tidwall(G): rand-delete       1,000,000 ops in 970ms, 1,030,559/sec, 970 ns/op
uART:       rand-delete       1,000,000 ops in 1260ms, 793,463/sec, 1260 ns/op
google(G):  rand-delete       1,000,000 ops in 992ms, 1,008,536/sec, 991 ns/op

** random get **
google:     get-rand          1,000,000 ops in 1910ms, 523,570/sec, 1909 ns/op
google(G):  get-rand          1,000,000 ops in 1072ms, 932,965/sec, 1071 ns/op
tidwall:    get-rand          1,000,000 ops in 1950ms, 512,710/sec, 1950 ns/op
tidwall(G): get-rand          1,000,000 ops in 1136ms, 880,522/sec, 1135 ns/op
tidwall(M): get-rand          1,000,000 ops in 1052ms, 950,524/sec, 1052 ns/op
tidwall:    get-rand-hint     1,000,000 ops in 1995ms, 501,295/sec, 1994 ns/op
tidwall(G): get-rand-hint     1,000,000 ops in 1272ms, 785,958/sec, 1272 ns/op

** sequential pivot **
Test getting 10 consecutive items starting at a pivot.
google:     ascend-seq        1,000,000 ops in 425ms, 2,355,115/sec, 424 ns/op
google:     descend-seq       1,000,000 ops in 542ms, 1,845,632/sec, 541 ns/op
google(G):  ascend-seq        1,000,000 ops in 252ms, 3,974,388/sec, 251 ns/op
google(G):  descend-seq       1,000,000 ops in 314ms, 3,186,957/sec, 313 ns/op
tidwall:    ascend-seq        1,000,000 ops in 340ms, 2,945,268/sec, 339 ns/op
tidwall:    descend-seq       1,000,000 ops in 354ms, 2,822,731/sec, 354 ns/op
tidwall:    ascend-seq-hint   1,000,000 ops in 320ms, 3,123,429/sec, 320 ns/op
tidwall:    descend-seq-hint  1,000,000 ops in 314ms, 3,184,809/sec, 313 ns/op
tidwall(G): ascend-seq        1,000,000 ops in 246ms, 4,072,619/sec, 245 ns/op
tidwall(G): descend-seq       1,000,000 ops in 242ms, 4,125,182/sec, 242 ns/op
tidwall(G): ascend-seq-hint   1,000,000 ops in 170ms, 5,889,374/sec, 169 ns/op
tidwall(G): descend-seq-hint  1,000,000 ops in 173ms, 5,769,961/sec, 173 ns/op
tidwall(G): iter-seq          1,000,000 ops in 315ms, 3,176,052/sec, 314 ns/op
tidwall(G): iter-seq-hint     1,000,000 ops in 251ms, 3,977,071/sec, 251 ns/op

** random pivot **
Test getting 10 consecutive items starting at a pivot.
google:     ascend-rand       1,000,000 ops in 2195ms, 455,636/sec, 2194 ns/op
google:     descend-rand      1,000,000 ops in 3024ms, 330,664/sec, 3024 ns/op
google(G):  ascend-rand       1,000,000 ops in 1274ms, 784,967/sec, 1273 ns/op
google(G):  descend-rand      1,000,000 ops in 1596ms, 626,556/sec, 1596 ns/op
tidwall:    ascend-rand       1,000,000 ops in 2042ms, 489,796/sec, 2041 ns/op
tidwall:    descend-rand      1,000,000 ops in 2057ms, 486,206/sec, 2056 ns/op
tidwall:    ascend-rand-hint  1,000,000 ops in 2051ms, 487,605/sec, 2050 ns/op
tidwall:    descend-rand-hint 1,000,000 ops in 2118ms, 472,227/sec, 2117 ns/op
tidwall(G): ascend-rand       1,000,000 ops in 1295ms, 772,485/sec, 1294 ns/op
tidwall(G): descend-rand      1,000,000 ops in 1227ms, 815,123/sec, 1226 ns/op
tidwall(G): ascend-rand-hint  1,000,000 ops in 1277ms, 782,783/sec, 1277 ns/op
tidwall(G): descend-rand-hint 1,000,000 ops in 1300ms, 769,248/sec, 1299 ns/op
tidwall(G): iter-rand         1,000,000 ops in 1337ms, 748,019/sec, 1336 ns/op
tidwall(G): iter-rand-hint    1,000,000 ops in 1410ms, 709,255/sec, 1409 ns/op

** scan **
Test scanning over every item in the tree
google:     ascend            1,000,000 ops in 11ms, 92,144,166/sec, 10 ns/op
google(G):  ascend            1,000,000 ops in 12ms, 86,702,468/sec, 11 ns/op
tidwall:    ascend            1,000,000 ops in 10ms, 103,109,568/sec, 9 ns/op
tidwall(G): scan              1,000,000 ops in 10ms, 101,119,105/sec, 9 ns/op
tidwall(G): walk              1,000,000 ops in 5ms, 209,795,784/sec, 4 ns/op
tidwall(G): iter              1,000,000 ops in 12ms, 80,602,565/sec, 12 ns/op

Compilation finished at Mon Mar 17 23:56:19


```
