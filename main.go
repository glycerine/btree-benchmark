package main

import (
	"bytes"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"

	"github.com/glycerine/uart"
	gbtree "github.com/google/btree"
	tbtree "github.com/tidwall/btree"
	"github.com/tidwall/lotsa"

	"github.com/dgraph-io/badger/v3/skl"
	"github.com/dgraph-io/badger/v3/y"

	"github.com/zhangyunhao116/skipmap"
)

type keyT string
type valT int64

type itemT struct {
	key keyT
	val valT
}

func int64ToItemT(i int64) itemT {
	return itemT{
		key: keyT(fmt.Sprintf("%016d", i)),
		val: valT(i),
	}
}

func (item itemT) Less(other gbtree.Item) bool {
	return item.key < other.(itemT).key
}

func lessG(a, b itemT) bool {
	return a.key < b.key
}

func less(a, b interface{}) bool {
	return a.(itemT).key < b.(itemT).key
}

func newUART() *uart.Tree {
	tr := uart.NewArtTree()
	//tr.SkipLocking = true
	return tr
}

func newTBTree(degree int) *tbtree.BTree {
	return tbtree.NewOptions(less, tbtree.Options{
		NoLocks: true,
		Degree:  degree,
	})
}
func newTBTreeG(degree int) *tbtree.BTreeG[itemT] {
	return tbtree.NewBTreeGOptions(lessG, tbtree.Options{
		NoLocks: true,
		Degree:  degree,
	})
}
func newTBTreeG_withLocking(degree int) *tbtree.BTreeG[itemT] {
	return tbtree.NewBTreeGOptions(lessG, tbtree.Options{
		NoLocks: false,
		Degree:  degree,
	})
}
func newTBTreeM(degree int) (r *tbtree.Map[keyT, valT]) {
	return tbtree.NewMap[keyT, valT](degree)
}
func newGBTree(degree int) *gbtree.BTree {
	return gbtree.New(degree)
}
func newGBTreeG(degree int) *gbtree.BTreeG[itemT] {
	return gbtree.NewG(degree, lessG)
}

func print_label(label, action string) {
	fmt.Printf("%-11s %-17s ", label+":", action)
}

func main() {
	N := 1_000_000
	degree := 32
	flag.IntVar(&N, "count", N, "number of items")
	flag.IntVar(&degree, "degree", degree, "B-tree degree")
	flag.Parse()

	items := make([]itemT, N)
	itemsM := make(map[int64]bool)
	itemsBinaryKey := make([][]byte, N)
	leafset := make([]*uart.Leaf, N)
	for i := 0; i < N; i++ {
		for {
			key := rand.Int63n(10000000000000000)
			if !itemsM[key] {
				itemsM[key] = true
				items[i] = int64ToItemT(key)
				if len(items[i].key) != 16 {
					panic("!")
				}
				leafset[i] = &uart.Leaf{}
				break
			}
		}
	}

	lotsa.Output = os.Stdout
	lotsa.MemUsage = true

	sortInts := func() {
		sort.Slice(items, func(i, j int) bool {
			return less(items[i], items[j])
		})
		for i := range items {
			itemsBinaryKey[i] = []byte(items[i].key)
		}
	}

	shuffleInts := func() {
		for i := range items {
			j := rand.Intn(i + 1)
			items[i], items[j] = items[j], items[i]
			itemsBinaryKey[i], itemsBinaryKey[j] = itemsBinaryKey[j], itemsBinaryKey[i]
		}
	}

	gtr := newGBTree(degree)
	gtrG := newGBTreeG(degree)
	ttr := newTBTree(degree)
	ttrG := newTBTreeG(degree)
	ttrGlocking := newTBTreeG_withLocking(degree)
	ttrM := newTBTreeM(degree)
	uART := newUART()
	skiplist := skl.NewSkiplist(int64(N * skl.MaxNodeSize))
	skipm := skipmap.NewFunc[[]byte, int](func(a, b []byte) bool {
		return bytes.Compare(a, b) < 0
	})

	withSeq := true
	withRand := true
	withRandSet := true
	withRandDel := true
	withPivot := true
	withScan := true
	withHints := true
	withDelete := true

	fmt.Printf("\ndegree=%d, key=string (16 bytes), val=int64, count=%d\n",
		degree, N)

	var hint tbtree.PathHint
	var hintG tbtree.PathHint

	if false {
		// titrate degree
		for _, d := range []int{2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 3000} {
			print_label("tidwall(G)", fmt.Sprintf("degree %v", d))
			ttrGlocking = newTBTreeG_withLocking(d)
			//ttrGlocking = newTBTreeG(d) // without locks?
			//ttrGlocking := newGBTree(d) // google b-tree?
			lotsa.Ops(N, 1, func(i, _ int) {
				ttrGlocking.Set(items[i])
				//ttrGlocking.ReplaceOrInsert(items[i])
			})
			ttrGlocking.Get(items[0]) // prevent GC til after lotsa can report.
		}
	}

	if false {
		// fill up, titrate degree for seq read
		for _, d := range []int{2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 3000} {
			//print_label("tidwall(G)", fmt.Sprintf("get-seq degree %v", d))
			print_label("google(G)", fmt.Sprintf("get-seq degree %v", d))

			//ttrGlocking = newTBTreeG_withLocking(d)
			//ttrGlocking = newTBTreeG(d) // without locks?
			ttrGlocking := newGBTree(d) // google b-tree?
			for i := range N {
				//ttrGlocking.Set(items[i])
				ttrGlocking.ReplaceOrInsert(items[i])
			}

			lotsa.Ops(N, 1, func(i, _ int) {
				re := ttrGlocking.Get(items[i])
				if re == nil {
					panic(re)
				}
				//ok := ttrGlocking.Get(items[i])
				//if !ok {
				//	panic(re)
				//}
			})
		}
	}
	/*
	   google(G):  get-seq degree 2  1,000,000 ops in 2569ms, 389,291/sec, 2568 ns/op
	   google(G):  get-seq degree 4  1,000,000 ops in 1960ms, 510,115/sec, 1960 ns/op
	   google(G):  get-seq degree 8  1,000,000 ops in 1885ms, 530,624/sec, 1884 ns/op
	   google(G):  get-seq degree 16 1,000,000 ops in 1692ms, 591,071/sec, 1691 ns/op
	   google(G):  get-seq degree 32 1,000,000 ops in 1582ms, 632,264/sec, 1581 ns/op
	   google(G):  get-seq degree 64 1,000,000 ops in 1532ms, 652,590/sec, 1532 ns/op
	   google(G):  get-seq degree 128 1,000,000 ops in 1524ms, 656,348/sec, 1523 ns/op
	   google(G):  get-seq degree 256 1,000,000 ops in 1536ms, 651,220/sec, 1535 ns/op
	   google(G):  get-seq degree 512 1,000,000 ops in 1628ms, 614,386/sec, 1627 ns/op
	   google(G):  get-seq degree 1024 1,000,000 ops in 1535ms, 651,504/sec, 1534 ns/op
	   google(G):  get-seq degree 2048 1,000,000 ops in 1561ms, 640,725/sec, 1560 ns/op
	   google(G):  get-seq degree 3000 1,000,000 ops in 1551ms, 644,578/sec, 1551 ns/op
	*/

	// 2048 sweet spot
	/*
	   tidwall(G): get-seq degree 2  1,000,000 ops in 1776ms, 563,116/sec, 1775 ns/op
	   tidwall(G): get-seq degree 4  1,000,000 ops in 1425ms, 701,783/sec, 1424 ns/op
	   tidwall(G): get-seq degree 8  1,000,000 ops in 1194ms, 837,763/sec, 1193 ns/op
	   tidwall(G): get-seq degree 16 1,000,000 ops in 1067ms, 937,008/sec, 1067 ns/op
	   tidwall(G): get-seq degree 32 1,000,000 ops in 962ms, 1,039,051/sec, 962 ns/op, 480 bytes, 0.0 bytes/op
	   tidwall(G): get-seq degree 64 1,000,000 ops in 918ms, 1,088,767/sec, 918 ns/op
	   tidwall(G): get-seq degree 128 1,000,000 ops in 903ms, 1,107,745/sec, 902 ns/op
	   tidwall(G): get-seq degree 256 1,000,000 ops in 912ms, 1,096,509/sec, 911 ns/op
	   tidwall(G): get-seq degree 512 1,000,000 ops in 917ms, 1,090,000/sec, 917 ns/op
	   tidwall(G): get-seq degree 1024 1,000,000 ops in 910ms, 1,098,879/sec, 910 ns/op
	   tidwall(G): get-seq degree 2048 1,000,000 ops in 894ms, 1,118,204/sec, 894 ns/op
	   tidwall(G): get-seq degree 3000 1,000,000 ops in 900ms, 1,110,688/sec, 900 ns/op, 480 bytes, 0.0 bytes/op

	*/
	//return
	// degree 64 is the sweet spot.
	/*
	   tidwall(G): degree 2          1,000,000 ops in 1802ms, 555,010/sec, 1801 ns/op, 70.5 MB, 73.9 bytes/op
	   tidwall(G): degree 4          1,000,000 ops in 1317ms, 759,482/sec, 1316 ns/op, 52.9 MB, 55.4 bytes/op
	   tidwall(G): degree 8          1,000,000 ops in 1050ms, 952,317/sec, 1050 ns/op, 52.6 MB, 55.2 bytes/op
	   tidwall(G): degree 16         1,000,000 ops in 901ms, 1,110,231/sec, 900 ns/op, 37.0 MB, 38.8 bytes/op
	   tidwall(G): degree 32         1,000,000 ops in 841ms, 1,189,668/sec, 840 ns/op, 34.9 MB, 36.6 bytes/op
	   tidwall(G): degree 64         1,000,000 ops in 836ms, 1,195,767/sec, 836 ns/op, 34.1 MB, 35.8 bytes/op
	   tidwall(G): degree 128        1,000,000 ops in 939ms, 1,065,436/sec, 938 ns/op, 33.4 MB, 35.0 bytes/op
	   tidwall(G): degree 256        1,000,000 ops in 1040ms, 961,520/sec, 1040 ns/op, 32.3 MB, 33.9 bytes/op
	   tidwall(G): degree 512        1,000,000 ops in 1218ms, 821,043/sec, 1217 ns/op, 37.1 MB, 38.9 bytes/op
	   tidwall(G): degree 1024       1,000,000 ops in 1627ms, 614,668/sec, 1626 ns/op, 30.8 MB, 32.3 bytes/op
	   tidwall(G): degree 2048       1,000,000 ops in 2340ms, 427,399/sec, 2339 ns/op, 25.7 MB, 27.0 bytes/op
	   tidwall(G): degree 3000       1,000,000 ops in 2893ms, 345,677/sec, 2892 ns/op, 33.1 MB, 34.7 bytes/op
	*/
	/*
	   degree=32, key=string (16 bytes), val=int64, count=1000000
	   tidwall(G): degree 32         1,000,000 ops in 874ms, 1,144,162/sec, 874 ns/op, 34.9 MB, 36.6 bytes/op
	   tidwall(G): degree 64         1,000,000 ops in 836ms, 1,195,560/sec, 836 ns/op, 34.0 MB, 35.6 bytes/op
	   tidwall(G): degree 128        1,000,000 ops in 917ms, 1,089,977/sec, 917 ns/op, 33.3 MB, 34.9 bytes/op
	   tidwall(G): degree 256        1,000,000 ops in 1019ms, 981,113/sec, 1019 ns/op, 32.8 MB, 34.4 bytes/op
	   tidwall(G): degree 512        1,000,000 ops in 1234ms, 810,093/sec, 1234 ns/op, 36.9 MB, 38.7 bytes/op
	   tidwall(G): degree 1024       1,000,000 ops in 1601ms, 624,482/sec, 1601 ns/op, 31.4 MB, 32.9 bytes/op
	   tidwall(G): degree 2048       1,000,000 ops in 2338ms, 427,781/sec, 2337 ns/op, 25.1 MB, 26.3 bytes/op
	   tidwall(G): degree 3000       1,000,000 ops in 2857ms, 350,072/sec, 2856 ns/op, 33.2 MB, 34.8 bytes/op
	   tidwall(G): degree 4096       1,000,000 ops in 3723ms, 268,579/sec, 3723 ns/op, 26.7 MB, 28.0 bytes/op
	   tidwall(G): degree 10000      1,000,000 ops in 7272ms, 137,505/sec, 7272 ns/op, 25.4 MB, 26.7 bytes/op
	*/
	// but same thing without locks
	/*
	   degree=32, key=string (16 bytes), val=int64, count=1000000
	   tidwall(G): degree 32         1,000,000 ops in 862ms, 1,159,832/sec, 862 ns/op, 34.9 MB, 36.6 bytes/op
	   tidwall(G): degree 64         1,000,000 ops in 866ms, 1,154,277/sec, 866 ns/op, 34.1 MB, 35.8 bytes/op
	   tidwall(G): degree 128        1,000,000 ops in 956ms, 1,046,441/sec, 955 ns/op, 33.4 MB, 35.1 bytes/op
	   tidwall(G): degree 256        1,000,000 ops in 1039ms, 962,740/sec, 1038 ns/op, 32.2 MB, 33.8 bytes/op
	   tidwall(G): degree 512        1,000,000 ops in 1256ms, 796,354/sec, 1255 ns/op, 36.9 MB, 38.7 bytes/op
	   tidwall(G): degree 1024       1,000,000 ops in 1673ms, 597,784/sec, 1672 ns/op, 31.5 MB, 33.0 bytes/op
	   tidwall(G): degree 2048       1,000,000 ops in 2394ms, 417,780/sec, 2393 ns/op, 25.7 MB, 27.0 bytes/op
	   tidwall(G): degree 3000       1,000,000 ops in 2948ms, 339,214/sec, 2947 ns/op, 33.3 MB, 34.9 bytes/op
	   tidwall(G): degree 4096       1,000,000 ops in 3764ms, 265,694/sec, 3763 ns/op, 27.8 MB, 29.1 bytes/op
	   tidwall(G): degree 10000      1,000,000 ops in 7485ms, 133,592/sec, 7485 ns/op, 25.4 MB, 26.7 bytes/op

	*/
	// same phenomenon for google/btree
	/*
	   degree=32, key=string (16 bytes), val=int64, count=1000000
	   tidwall(G): degree 32         1,000,000 ops in 1411ms, 708,498/sec, 1411 ns/op, 49.2 MB, 51.6 bytes/op
	   tidwall(G): degree 64         1,000,000 ops in 1457ms, 686,513/sec, 1456 ns/op, 45.9 MB, 48.1 bytes/op
	   tidwall(G): degree 128        1,000,000 ops in 1407ms, 710,526/sec, 1407 ns/op, 45.4 MB, 47.6 bytes/op
	   tidwall(G): degree 256        1,000,000 ops in 1516ms, 659,683/sec, 1515 ns/op, 44.3 MB, 46.5 bytes/op
	   tidwall(G): degree 512        1,000,000 ops in 1699ms, 588,566/sec, 1699 ns/op, 46.7 MB, 49.0 bytes/op
	   tidwall(G): degree 1024       1,000,000 ops in 1941ms, 515,245/sec, 1940 ns/op, 45.5 MB, 47.7 bytes/op
	   tidwall(G): degree 2048       1,000,000 ops in 2368ms, 422,316/sec, 2367 ns/op, 40.1 MB, 42.0 bytes/op
	   tidwall(G): degree 3000       1,000,000 ops in 2678ms, 373,465/sec, 2677 ns/op, 46.0 MB, 48.2 bytes/op
	   tidwall(G): degree 4096       1,000,000 ops in 3154ms, 317,075/sec, 3153 ns/op, 42.7 MB, 44.7 bytes/op
	   tidwall(G): degree 10000      1,000,000 ops in 5284ms, 189,254/sec, 5283 ns/op, 41.7 MB, 43.7 bytes/op
	*/

	if withSeq {
		println()
		println("** sequential set **")
		sortInts()

		// google
		print_label("google", "set-seq")
		gtr = newGBTree(degree)
		lotsa.Ops(N, 1, func(i, _ int) {
			gtr.ReplaceOrInsert(items[i])
		})

		print_label("google(G)", "set-seq")
		gtrG = newGBTreeG(degree)
		lotsa.Ops(N, 1, func(i, _ int) {
			gtrG.ReplaceOrInsert(items[i])
		})

		// non-generics tidwall
		print_label("tidwall", "set-seq")
		ttr = newTBTree(degree)
		lotsa.Ops(N, 1, func(i, _ int) {
			ttr.Set(items[i])
		})
		print_label("tidwall(G)", "set-seq")
		ttrG = newTBTreeG(degree)
		lotsa.Ops(N, 1, func(i, _ int) {
			ttrG.Set(items[i])
		})
		print_label("tidwall(M)", "set-seq")
		ttrM = newTBTreeM(degree)
		lotsa.Ops(N, 1, func(i, _ int) {
			ttrM.Set(items[i].key, items[i].val)
		})

		print_label("tidwall(G) with locking", "set-seq")
		ttrGlocking = newTBTreeG_withLocking(degree)
		lotsa.Ops(N, 1, func(i, _ int) {
			ttrGlocking.Set(items[i])
		})
		ttrGlocking.Get(items[0]) // prevent GC til after lotsa can report.

		print_label("badger/skiplist", "set-seq")
		skiplist = skl.NewSkiplist(int64(N * skl.MaxNodeSize))
		lotsa.Ops(N, 1, func(i, _ int) {
			skiplist.Put(itemsBinaryKey[i], y.ValueStruct{Value: itemsBinaryKey[i], Meta: 0, UserMeta: 0})
		})
		skiplist.Get(itemsBinaryKey[0]) // try to prevent too soon GC, does not appear to be working.

		print_label("zhangyunhao116/skipmap", "set-seq")
		skipm = skipmap.NewFunc[[]byte, int](func(a, b []byte) bool {
			return bytes.Compare(a, b) < 0
		})
		lotsa.Ops(N, 1, func(i, _ int) {
			skipm.Store(itemsBinaryKey[i], i)
		})
		skipm.Load(itemsBinaryKey[0]) // prevent GC too soon.

		//fmt.Printf("back from lotsa.Ops for tidwall(M)\n")

		print_label("uART", "set-seq")
		uART = newUART()
		lotsa.Ops(N, 1, func(i, _ int) {
			// remember that Insert copies key, and makes a new leaf.
			if true {
				// uART uses 3x the memory of btrees.
				// tidwall(G): set-seq 1,000,000 ops in 261ms, 3,836,715/sec, 260 ns/op, 49.2 MB, 51.6 bytes/op
				// uART:       set-seq 1,000,000 ops in 330ms, 3,030,409/sec, 329 ns/op, 166.0 MB, 174.0 bytes/op

				uART.Insert(itemsBinaryKey[i], i)
			} else {
				// The key copy could be avoided like this, but the NewLeaf is unavoidable.
				// uART still uses 2x the memory of btrees, not counting the Leaf overhead.
				// tidwall(G): set-seq 1,000,000 ops in 260ms, 3,844,936/sec, 260 ns/op, 49.2 MB, 51.6 bytes/op
				// uART:       set-seq 1,000,000 ops in 255ms, 3,928,167/sec, 254 ns/op, 97.3 MB, 102.0 bytes/op
				lf := leafset[i]
				lf.Key = itemsBinaryKey[i]
				lf.Value = i
				uART.InsertLeaf(lf)
			}
		})

		//_ = uART // does not suffice to keep uART from
		//  being GC-ed before lotsa can measure its memory consumption!
		// We do this to prevent that; otherwise mem measurement will be zero.
		for range uart.Ascend(uART, nil, nil) {
			break
		}

		if withHints {
			print_label("tidwall", "set-seq-hint")
			ttr = newTBTree(degree)
			lotsa.Ops(N, 1, func(i, _ int) {
				ttr.SetHint(items[i], &hint)
			})
			print_label("tidwall(G)", "set-seq-hint")
			ttrG = newTBTreeG(degree)
			lotsa.Ops(N, 1, func(i, _ int) {
				ttrG.SetHint(items[i], &hintG)
			})
		}
		print_label("tidwall", "load-seq")
		ttr = newTBTree(degree)
		lotsa.Ops(N, 1, func(i, _ int) {
			ttr.Load(items[i])
		})
		print_label("tidwall(G)", "load-seq")
		ttrG = newTBTreeG(degree)
		lotsa.Ops(N, 1, func(i, _ int) {
			ttrG.Load(items[i])
		})
		print_label("tidwall(M)", "load-seq")
		ttrM = newTBTreeM(degree)
		lotsa.Ops(N, 1, func(i, _ int) {
			ttrM.Load(items[i].key, items[i].val)
		})

		println()
		println("** sequential get **")
		sortInts()

		print_label("google", "get-seq")
		lotsa.Ops(N, 1, func(i, _ int) {
			re := gtr.Get(items[i])
			if re == nil {
				panic(re)
			}
		})
		print_label("google(G)", "get-seq")
		lotsa.Ops(N, 1, func(i, _ int) {
			re, ok := gtrG.Get(items[i])
			if !ok {
				panic(re)
			}
		})
		print_label("tidwall", "get-seq")
		lotsa.Ops(N, 1, func(i, _ int) {
			re := ttr.Get(items[i])
			if re == nil {
				panic(re)
			}
		})
		print_label("tidwall(G)", "get-seq")
		lotsa.Ops(N, 1, func(i, _ int) {
			re, ok := ttrG.Get(items[i])
			if !ok {
				panic(re)
			}
		})
		print_label("tidwall(M)", "get-seq")
		lotsa.Ops(N, 1, func(i, _ int) {
			re, ok := ttrM.Get(items[i].key)
			if !ok {
				panic(re)
			}
		})
		if withHints {
			print_label("tidwall", "get-seq-hint")
			lotsa.Ops(N, 1, func(i, _ int) {
				re := ttr.GetHint(items[i], &hint)
				if re == nil {
					panic(re)
				}
			})
			print_label("tidwall(G)", "get-seq-hint")
			lotsa.Ops(N, 1, func(i, _ int) {
				re, ok := ttrG.GetHint(items[i], &hintG)
				if !ok {
					panic(re)
				}
			})
		}
	}

	if withDelete {
		println()
		println("** sequential delete **")

		print_label("tidwall(G)", "seq-delete")
		lotsa.Ops(N, 1, func(i, _ int) {
			ttrG.Delete(items[i])
		})

		print_label("uART", "seq-delete")
		lotsa.Ops(N, 1, func(i, _ int) {
			uART.Remove(itemsBinaryKey[i])
		})

		print_label("google(G)", "seq-delete")
		lotsa.Ops(N, 1, func(i, _ int) {
			gtrG.Delete(items[i])
		})

		print_label("google", "seq-delete")
		lotsa.Ops(N, 1, func(i, _ int) {
			gtr.Delete(items[i])
		})

		print_label("tidwall(G) with locking", "seq-delete")
		ttrGlocking = newTBTreeG_withLocking(degree)
		lotsa.Ops(N, 1, func(i, _ int) {
			ttrGlocking.Delete(items[i])
		})
		ttrGlocking.Get(items[0]) // prevent GC til after lotsa can report.

		// does not appear to support delete
		//print_label("badger/skiplist", "seq-delete")
		//skiplist = skl.NewSkiplist(int64(N * skl.MaxNodeSize))
		//lotsa.Ops(N, 1, func(i, _ int) {
		//	skiplist.Remove(itemsBinaryKey[i])
		//})
		//skiplist.Get(itemsBinaryKey[0]) // try to prevent too soon GC, does not appear to be working.

		print_label("zhangyunhao116/skipmap", "seq-delete")
		skipm = skipmap.NewFunc[[]byte, int](func(a, b []byte) bool {
			return bytes.Compare(a, b) < 0
		})
		lotsa.Ops(N, 1, func(i, _ int) {
			skipm.Delete(itemsBinaryKey[i])
		})
		skipm.Load(itemsBinaryKey[0]) // prevent GC too soon.

	}

	if withRand {
		if withRandSet {
			println()
			println("** random set **")
			shuffleInts()
			print_label("google", "set-rand")
			gtr = newGBTree(degree)
			lotsa.Ops(N, 1, func(i, _ int) {
				gtr.ReplaceOrInsert(items[i])
			})
			print_label("google(G)", "set-rand")
			gtrG = newGBTreeG(degree)
			lotsa.Ops(N, 1, func(i, _ int) {
				gtrG.ReplaceOrInsert(items[i])
			})
			print_label("tidwall", "set-rand")
			ttr = newTBTree(degree)
			lotsa.Ops(N, 1, func(i, _ int) {
				ttr.Set(items[i])
			})
			print_label("tidwall(G)", "set-rand")
			ttrG = newTBTreeG(degree)
			lotsa.Ops(N, 1, func(i, _ int) {
				ttrG.Set(items[i])
			})
			print_label("uART", "set-rand")
			uART = newUART()
			lotsa.Ops(N, 1, func(i, _ int) {
				uART.Insert(itemsBinaryKey[i], i)
			})
			print_label("tidwall(M)", "set-rand")
			ttrM = newTBTreeM(degree)
			lotsa.Ops(N, 1, func(i, _ int) {
				ttrM.Set(items[i].key, items[i].val)
			})

			print_label("tidwall(G) with locking", "set-rand")
			ttrGlocking = newTBTreeG_withLocking(degree)
			lotsa.Ops(N, 1, func(i, _ int) {
				ttrGlocking.Set(items[i])
			})
			ttrGlocking.Get(items[0]) // prevent GC til after lotsa can report.

			print_label("badger/skiplist", "set-rand")
			skiplist = skl.NewSkiplist(int64(N * skl.MaxNodeSize))
			lotsa.Ops(N, 1, func(i, _ int) {
				skiplist.Put(itemsBinaryKey[i], y.ValueStruct{Value: itemsBinaryKey[i], Meta: 0, UserMeta: 0})
			})
			skiplist.Get(itemsBinaryKey[0]) // try to prevent too soon GC, does not appear to be working.

			print_label("zhangyunhao116/skipmap", "set-rand")
			skipm = skipmap.NewFunc[[]byte, int](func(a, b []byte) bool {
				return bytes.Compare(a, b) < 0
			})
			lotsa.Ops(N, 1, func(i, _ int) {
				skipm.Store(itemsBinaryKey[i], i)
			})
			skipm.Load(itemsBinaryKey[0]) // prevent GC too soon.

			if withHints {
				print_label("tidwall", "set-rand-hint")
				ttr = newTBTree(degree)
				lotsa.Ops(N, 1, func(i, _ int) {
					ttr.SetHint(items[i], &hint)
				})
				print_label("tidwall(G)", "set-rand-hint")
				ttrG = newTBTreeG(degree)
				lotsa.Ops(N, 1, func(i, _ int) {
					ttrG.SetHint(items[i], &hintG)
				})
			}
			print_label("tidwall", "set-after-copy")
			ttr2 := ttr.Copy()
			lotsa.Ops(N, 1, func(i, _ int) {
				ttr2.Set(items[i])
			})
			print_label("tidwall(G)", "set-after-copy")
			ttrG2 := ttrG.Copy()
			lotsa.Ops(N, 1, func(i, _ int) {
				ttrG2.Set(items[i])
			})
			print_label("tidwall", "load-rand")
			ttr = newTBTree(degree)
			lotsa.Ops(N, 1, func(i, _ int) {
				ttr.Load(items[i])
			})
			print_label("tidwall(G)", "load-rand")
			ttrG = newTBTreeG(degree)
			lotsa.Ops(N, 1, func(i, _ int) {
				ttrG.Load(items[i])
			})
			ttrM = newTBTreeM(degree)
			print_label("tidwall(M)", "load-rand")
			lotsa.Ops(N, 1, func(i, _ int) {
				ttrM.Load(items[i].key, items[i].val)
			})
		}

		if withRandDel {
			println()
			println("** random delete **")
			shuffleInts()

			print_label("tidwall(G)", "rand-delete")
			lotsa.Ops(N, 1, func(i, _ int) {
				ttrG.Delete(items[i])
			})

			print_label("uART", "rand-delete")
			lotsa.Ops(N, 1, func(i, _ int) {
				uART.Remove(itemsBinaryKey[i])
			})

			print_label("google(G)", "rand-delete")
			lotsa.Ops(N, 1, func(i, _ int) {
				gtrG.Delete(items[i])
			})
		}

		println()
		println("** random get **")

		shuffleInts()
		gtr = newGBTree(degree)
		gtrG = newGBTreeG(degree)
		ttr = newTBTree(degree)
		ttrM = newTBTreeM(degree)
		ttrG = newTBTreeG(degree)
		for _, item := range items {
			gtr.ReplaceOrInsert(item)
			gtrG.ReplaceOrInsert(item)
			ttrG.Set(item)
			ttr.Set(item)
			ttrM.Set(item.key, item.val)
		}
		shuffleInts()

		print_label("google", "get-rand")
		lotsa.Ops(N, 1, func(i, _ int) {
			re := gtr.Get(items[i])
			if re == nil {
				panic(re)
			}
		})
		print_label("google(G)", "get-rand")
		lotsa.Ops(N, 1, func(i, _ int) {
			re, ok := gtrG.Get(items[i])
			if !ok {
				panic(re)
			}
		})
		print_label("tidwall", "get-rand")
		lotsa.Ops(N, 1, func(i, _ int) {
			re := ttr.Get(items[i])
			if re == nil {
				panic(re)
			}
		})
		print_label("tidwall(G)", "get-rand")
		lotsa.Ops(N, 1, func(i, _ int) {
			re, ok := ttrG.Get(items[i])
			if !ok {
				panic(re)
			}
		})
		print_label("tidwall(M)", "get-rand")
		lotsa.Ops(N, 1, func(i, _ int) {
			re, ok := ttrM.Get(items[i].key)
			if !ok {
				panic(re)
			}
		})
		if withHints {
			print_label("tidwall", "get-rand-hint")
			lotsa.Ops(N, 1, func(i, _ int) {
				re := ttr.GetHint(items[i], &hint)
				if re == nil {
					panic(re)
				}
			})
			print_label("tidwall(G)", "get-rand-hint")
			lotsa.Ops(N, 1, func(i, _ int) {
				re, ok := ttrG.GetHint(items[i], &hintG)
				if !ok {
					panic(re)
				}
			})
		}
	}

	if !withRand {
		shuffleInts()
		gtr = newGBTree(degree)
		gtrG = newGBTreeG(degree)
		ttr = newTBTree(degree)
		ttrM = newTBTreeM(degree)
		ttrG = newTBTreeG(degree)
		for _, item := range items {
			gtr.ReplaceOrInsert(item)
			gtrG.ReplaceOrInsert(item)
			ttrG.Set(item)
			ttr.Set(item)
			ttrM.Set(item.key, item.val)
		}
	}

	if withPivot {
		sortInts()
		const M = 10
		var hint tbtree.PathHint
		println()
		fmt.Printf("** sequential pivot **\n")
		fmt.Printf("Test getting %d consecutive items starting at a pivot.\n", M)
		print_label("google", "ascend-seq")
		lotsa.Ops(N, 1, func(i, _ int) {
			var count int
			gtr.AscendGreaterOrEqual(items[i], func(item gbtree.Item) bool {
				count++
				return count < M
			})
		})
		print_label("google", "descend-seq")
		lotsa.Ops(N, 1, func(i, _ int) {
			var count int
			gtr.DescendLessOrEqual(items[i], func(item gbtree.Item) bool {
				count++
				return count < M
			})
		})
		print_label("google(G)", "ascend-seq")
		lotsa.Ops(N, 1, func(i, _ int) {
			var count int
			gtrG.AscendGreaterOrEqual(items[i], func(item itemT) bool {
				count++
				return count < M
			})
		})
		print_label("google(G)", "descend-seq")
		lotsa.Ops(N, 1, func(i, _ int) {
			var count int
			gtrG.DescendLessOrEqual(items[i], func(item itemT) bool {
				count++
				return count < M
			})
		})
		print_label("tidwall", "ascend-seq")
		lotsa.Ops(N, 1, func(i, _ int) {
			var count int
			ttr.Ascend(items[i], func(item any) bool {
				count++
				return count < M
			})
		})
		print_label("tidwall", "descend-seq")
		lotsa.Ops(N, 1, func(i, _ int) {
			var count int
			ttr.Ascend(items[i], func(item any) bool {
				count++
				return count < M
			})
		})
		if withHints {
			print_label("tidwall", "ascend-seq-hint")
			lotsa.Ops(N, 1, func(i, _ int) {
				var count int
				ttr.AscendHint(items[i], func(item any) bool {
					count++
					return count < M
				}, &hint)
			})
			print_label("tidwall", "descend-seq-hint")
			lotsa.Ops(N, 1, func(i, _ int) {
				var count int
				ttr.DescendHint(items[i], func(item any) bool {
					count++
					return count < M
				}, &hint)
			})
		}
		print_label("tidwall(G)", "ascend-seq")
		lotsa.Ops(N, 1, func(i, _ int) {
			var count int
			ttrG.Ascend(items[i], func(item itemT) bool {
				count++
				return count < M
			})
		})
		print_label("tidwall(G)", "descend-seq")
		lotsa.Ops(N, 1, func(i, _ int) {
			var count int
			ttrG.Descend(items[i], func(item itemT) bool {
				count++
				return count < M
			})
		})
		if withHints {
			print_label("tidwall(G)", "ascend-seq-hint")
			lotsa.Ops(N, 1, func(i, _ int) {
				var count int
				ttrG.AscendHint(items[i], func(item itemT) bool {
					count++
					return count < M
				}, &hint)
			})
			print_label("tidwall(G)", "descend-seq-hint")
			lotsa.Ops(N, 1, func(i, _ int) {
				var count int
				ttrG.DescendHint(items[i], func(item itemT) bool {
					count++
					return count < M
				}, &hint)
			})
		}
		print_label("tidwall(G)", "iter-seq")
		lotsa.Ops(N, 1, func(i, _ int) {
			iter := ttrG.Iter()
			var count int
			for ok := iter.Seek(items[i]); ok; ok = iter.Next() {
				count++
				if count == M {
					break
				}
			}
			iter.Release()
		})
		print_label("tidwall(G)", "iter-seq-hint")
		lotsa.Ops(N, 1, func(i, _ int) {
			iter := ttrG.Iter()
			var count int
			for ok := iter.SeekHint(items[i], &hint); ok; ok = iter.Next() {
				count++
				if count == M {
					break
				}
			}
			iter.Release()
		})
	}

	if withPivot {
		shuffleInts()
		const M = 10
		var hint tbtree.PathHint
		println()
		fmt.Printf("** random pivot **\n")
		fmt.Printf("Test getting %d consecutive items starting at a pivot.\n", M)
		print_label("google", "ascend-rand")
		lotsa.Ops(N, 1, func(i, _ int) {
			var count int
			gtr.AscendGreaterOrEqual(items[i], func(item gbtree.Item) bool {
				count++
				return count < M
			})
		})
		print_label("google", "descend-rand")
		lotsa.Ops(N, 1, func(i, _ int) {
			var count int
			gtr.DescendLessOrEqual(items[i], func(item gbtree.Item) bool {
				count++
				return count < M
			})
		})
		print_label("google(G)", "ascend-rand")
		lotsa.Ops(N, 1, func(i, _ int) {
			var count int
			gtrG.AscendGreaterOrEqual(items[i], func(item itemT) bool {
				count++
				return count < M
			})
		})
		print_label("google(G)", "descend-rand")
		lotsa.Ops(N, 1, func(i, _ int) {
			var count int
			gtrG.DescendLessOrEqual(items[i], func(item itemT) bool {
				count++
				return count < M
			})
		})
		print_label("tidwall", "ascend-rand")
		lotsa.Ops(N, 1, func(i, _ int) {
			var count int
			ttr.Ascend(items[i], func(item any) bool {
				count++
				return count < M
			})
		})
		print_label("tidwall", "descend-rand")
		lotsa.Ops(N, 1, func(i, _ int) {
			var count int
			ttr.Ascend(items[i], func(item any) bool {
				count++
				return count < M
			})
		})
		if withHints {
			print_label("tidwall", "ascend-rand-hint")
			lotsa.Ops(N, 1, func(i, _ int) {
				var count int
				ttr.AscendHint(items[i], func(item any) bool {
					count++
					return count < M
				}, &hint)
			})
			print_label("tidwall", "descend-rand-hint")
			lotsa.Ops(N, 1, func(i, _ int) {
				var count int
				ttr.DescendHint(items[i], func(item any) bool {
					count++
					return count < M
				}, &hint)
			})
		}
		print_label("tidwall(G)", "ascend-rand")
		lotsa.Ops(N, 1, func(i, _ int) {
			var count int
			ttrG.Ascend(items[i], func(item itemT) bool {
				count++
				return count < M
			})
		})
		print_label("tidwall(G)", "descend-rand")
		lotsa.Ops(N, 1, func(i, _ int) {
			var count int
			ttrG.Descend(items[i], func(item itemT) bool {
				count++
				return count < M
			})
		})
		if withHints {
			print_label("tidwall(G)", "ascend-rand-hint")
			lotsa.Ops(N, 1, func(i, _ int) {
				var count int
				ttrG.AscendHint(items[i], func(item itemT) bool {
					count++
					return count < M
				}, &hint)
			})
			print_label("tidwall(G)", "descend-rand-hint")
			lotsa.Ops(N, 1, func(i, _ int) {
				var count int
				ttrG.DescendHint(items[i], func(item itemT) bool {
					count++
					return count < M
				}, &hint)
			})
		}
		print_label("tidwall(G)", "iter-rand")
		lotsa.Ops(N, 1, func(i, _ int) {
			iter := ttrG.Iter()
			var count int
			for ok := iter.Seek(items[i]); ok; ok = iter.Next() {
				count++
				if count == M {
					break
				}
			}
			iter.Release()
		})
		print_label("tidwall(G)", "iter-rand-hint")
		lotsa.Ops(N, 1, func(i, _ int) {
			iter := ttrG.Iter()
			var count int
			for ok := iter.SeekHint(items[i], &hint); ok; ok = iter.Next() {
				count++
				if count == M {
					break
				}
			}
			iter.Release()
		})
	}

	if withScan {
		println()
		println("** scan **")
		println("Test scanning over every item in the tree")
		print_label("google", "ascend")
		lotsa.Ops(N, 1, func(i, _ int) {
			if i == 0 {
				gtr.Ascend(func(item gbtree.Item) bool {
					return true
				})
			}
		})
		print_label("google(G)", "ascend")
		lotsa.Ops(N, 1, func(i, _ int) {
			if i == 0 {
				gtrG.Ascend(func(item itemT) bool {
					return true
				})
			}
		})
		print_label("tidwall", "ascend")
		lotsa.Ops(N, 1, func(i, _ int) {
			if i == 0 {
				ttr.Ascend(nil, func(item interface{}) bool {
					return true
				})
			}
		})
		print_label("tidwall(G)", "scan")
		lotsa.Ops(N, 1, func(i, _ int) {
			if i == 0 {
				ttrG.Scan(func(item itemT) bool {
					return true
				})
			}
		})
		print_label("tidwall(G)", "walk")
		lotsa.Ops(N, 1, func(i, _ int) {
			if i == 0 {
				ttrG.Walk(func(items []itemT) bool {
					for j := 0; j < len(items); j++ {

					}
					return true
				})
			}
		})
		print_label("tidwall(G)", "iter")
		lotsa.Ops(N, 1, func(i, _ int) {
			if i == 0 {
				iter := ttrG.Iter()
				for ok := iter.First(); ok; ok = iter.Next() {
				}
				iter.Release()
			}
		})
	}
}

// full recent run
/*
Compilation started at Wed Mar 19 01:57:33

go run main.go

degree=32, key=string (16 bytes), val=int64, count=1000000

** sequential set **
google:     set-seq           1,000,000 ops in 304ms, 3,294,776/sec, 303 ns/op, 60.8 MB, 63.7 bytes/op
google(G):  set-seq           1,000,000 ops in 269ms, 3,719,898/sec, 268 ns/op, 49.7 MB, 52.1 bytes/op
tidwall:    set-seq           1,000,000 ops in 287ms, 3,489,826/sec, 286 ns/op, 56.4 MB, 59.1 bytes/op
tidwall(G): set-seq           1,000,000 ops in 266ms, 3,761,495/sec, 265 ns/op, 49.2 MB, 51.6 bytes/op
tidwall(M): set-seq           1,000,000 ops in 229ms, 4,375,311/sec, 228 ns/op, 49.2 MB, 51.6 bytes/op
tidwall(G) with locking: set-seq           1,000,000 ops in 278ms, 3,600,127/sec, 277 ns/op, 49.2 MB, 51.6 bytes/op
badger/skiplist: set-seq           1,000,000 ops in 365ms, 2,743,355/sec, 364 ns/op
zhangyunhao116/skipmap: set-seq           1,000,000 ops in 318ms, 3,146,359/sec, 317 ns/op, 99.5 MB, 104.4 bytes/op
uART:       set-seq           1,000,000 ops in 352ms, 2,842,015/sec, 351 ns/op, 166.0 MB, 174.0 bytes/op
tidwall:    set-seq-hint      1,000,000 ops in 177ms, 5,642,442/sec, 177 ns/op, 56.4 MB, 59.1 bytes/op
tidwall(G): set-seq-hint      1,000,000 ops in 153ms, 6,533,930/sec, 153 ns/op, 49.2 MB, 51.6 bytes/op
tidwall:    load-seq          1,000,000 ops in 141ms, 7,072,614/sec, 141 ns/op, 56.4 MB, 59.1 bytes/op
tidwall(G): load-seq          1,000,000 ops in 77ms, 13,041,139/sec, 76 ns/op, 49.2 MB, 51.6 bytes/op
tidwall(M): load-seq          1,000,000 ops in 69ms, 14,519,117/sec, 68 ns/op, 49.2 MB, 51.6 bytes/op

** sequential get **
google:     get-seq           1,000,000 ops in 274ms, 3,646,738/sec, 274 ns/op
google(G):  get-seq           1,000,000 ops in 198ms, 5,044,442/sec, 198 ns/op
tidwall:    get-seq           1,000,000 ops in 256ms, 3,909,873/sec, 255 ns/op
tidwall(G): get-seq           1,000,000 ops in 221ms, 4,533,030/sec, 220 ns/op
tidwall(M): get-seq           1,000,000 ops in 178ms, 5,620,924/sec, 177 ns/op
tidwall:    get-seq-hint      1,000,000 ops in 182ms, 5,503,299/sec, 181 ns/op
tidwall(G): get-seq-hint      1,000,000 ops in 141ms, 7,072,239/sec, 141 ns/op

** sequential delete **
tidwall(G): seq-delete        1,000,000 ops in 247ms, 4,046,482/sec, 247 ns/op
uART:       seq-delete        1,000,000 ops in 114ms, 8,799,501/sec, 113 ns/op
google(G):  seq-delete        1,000,000 ops in 252ms, 3,970,541/sec, 251 ns/op
google:     seq-delete        1,000,000 ops in 363ms, 2,755,507/sec, 362 ns/op
tidwall(G) with locking: seq-delete        1,000,000 ops in 33ms, 30,403,797/sec, 32 ns/op
zhangyunhao116/skipmap: seq-delete        1,000,000 ops in 16ms, 63,796,162/sec, 15 ns/op

** random set **
google:     set-rand          1,000,000 ops in 1592ms, 628,044/sec, 1592 ns/op, 49.2 MB, 51.6 bytes/op
google(G):  set-rand          1,000,000 ops in 1006ms, 994,072/sec, 1005 ns/op, 35.2 MB, 36.9 bytes/op
tidwall:    set-rand          1,000,000 ops in 1569ms, 637,502/sec, 1568 ns/op, 46.7 MB, 49.0 bytes/op
tidwall(G): set-rand          1,000,000 ops in 1006ms, 994,329/sec, 1005 ns/op, 35.1 MB, 36.8 bytes/op
uART:       set-rand          1,000,000 ops in 1048ms, 954,583/sec, 1047 ns/op, 165.6 MB, 173.7 bytes/op
tidwall(M): set-rand          1,000,000 ops in 1027ms, 973,714/sec, 1026 ns/op, 35.1 MB, 36.8 bytes/op
tidwall(G) with locking: set-rand          1,000,000 ops in 1154ms, 866,366/sec, 1154 ns/op, 35.1 MB, 36.8 bytes/op
badger/skiplist: set-rand          1,000,000 ops in 1527ms, 655,049/sec, 1526 ns/op
zhangyunhao116/skipmap: set-rand          1,000,000 ops in 2097ms, 476,962/sec, 2096 ns/op, 99.5 MB, 104.4 bytes/op
tidwall:    set-rand-hint     1,000,000 ops in 1745ms, 573,074/sec, 1744 ns/op, 46.7 MB, 49.0 bytes/op
tidwall(G): set-rand-hint     1,000,000 ops in 1120ms, 893,001/sec, 1119 ns/op, 35.1 MB, 36.8 bytes/op
tidwall:    set-after-copy    1,000,000 ops in 1904ms, 525,127/sec, 1904 ns/op
tidwall(G): set-after-copy    1,000,000 ops in 1137ms, 879,819/sec, 1136 ns/op
tidwall:    load-rand         1,000,000 ops in 1659ms, 602,799/sec, 1658 ns/op, 46.7 MB, 49.0 bytes/op
tidwall(G): load-rand         1,000,000 ops in 1061ms, 942,847/sec, 1060 ns/op, 35.1 MB, 36.8 bytes/op
tidwall(M): load-rand         1,000,000 ops in 1066ms, 938,305/sec, 1065 ns/op, 35.1 MB, 36.8 bytes/op

** random delete **
tidwall(G): rand-delete       1,000,000 ops in 1004ms, 996,169/sec, 1003 ns/op
uART:       rand-delete       1,000,000 ops in 1308ms, 764,757/sec, 1307 ns/op
google(G):  rand-delete       1,000,000 ops in 1094ms, 914,072/sec, 1094 ns/op

** random get **
google:     get-rand          1,000,000 ops in 2217ms, 451,145/sec, 2216 ns/op
google(G):  get-rand          1,000,000 ops in 1319ms, 758,131/sec, 1319 ns/op
tidwall:    get-rand          1,000,000 ops in 2496ms, 400,690/sec, 2495 ns/op
tidwall(G): get-rand          1,000,000 ops in 1348ms, 741,721/sec, 1348 ns/op
tidwall(M): get-rand          1,000,000 ops in 1209ms, 826,999/sec, 1209 ns/op
tidwall:    get-rand-hint     1,000,000 ops in 2285ms, 437,715/sec, 2284 ns/op
tidwall(G): get-rand-hint     1,000,000 ops in 1378ms, 725,837/sec, 1377 ns/op

** sequential pivot **
Test getting 10 consecutive items starting at a pivot.
google:     ascend-seq        1,000,000 ops in 495ms, 2,018,986/sec, 495 ns/op
google:     descend-seq       1,000,000 ops in 584ms, 1,713,734/sec, 583 ns/op
google(G):  ascend-seq        1,000,000 ops in 274ms, 3,646,012/sec, 274 ns/op
google(G):  descend-seq       1,000,000 ops in 342ms, 2,924,057/sec, 341 ns/op
tidwall:    ascend-seq        1,000,000 ops in 381ms, 2,623,779/sec, 381 ns/op
tidwall:    descend-seq       1,000,000 ops in 413ms, 2,424,197/sec, 412 ns/op
tidwall:    ascend-seq-hint   1,000,000 ops in 334ms, 2,994,857/sec, 333 ns/op
tidwall:    descend-seq-hint  1,000,000 ops in 336ms, 2,975,374/sec, 336 ns/op
tidwall(G): ascend-seq        1,000,000 ops in 285ms, 3,511,029/sec, 284 ns/op
tidwall(G): descend-seq       1,000,000 ops in 299ms, 3,339,176/sec, 299 ns/op
tidwall(G): ascend-seq-hint   1,000,000 ops in 185ms, 5,405,460/sec, 184 ns/op
tidwall(G): descend-seq-hint  1,000,000 ops in 184ms, 5,421,958/sec, 184 ns/op
tidwall(G): iter-seq          1,000,000 ops in 363ms, 2,758,031/sec, 362 ns/op
tidwall(G): iter-seq-hint     1,000,000 ops in 307ms, 3,252,669/sec, 307 ns/op

** random pivot **
Test getting 10 consecutive items starting at a pivot.
google:     ascend-rand       1,000,000 ops in 2659ms, 376,142/sec, 2658 ns/op
google:     descend-rand      1,000,000 ops in 3536ms, 282,808/sec, 3535 ns/op
google(G):  ascend-rand       1,000,000 ops in 1528ms, 654,511/sec, 1527 ns/op
google(G):  descend-rand      1,000,000 ops in 1827ms, 547,302/sec, 1827 ns/op
tidwall:    ascend-rand       1,000,000 ops in 2410ms, 414,952/sec, 2409 ns/op
tidwall:    descend-rand      1,000,000 ops in 2262ms, 442,014/sec, 2262 ns/op
tidwall:    ascend-rand-hint  1,000,000 ops in 2182ms, 458,293/sec, 2182 ns/op
tidwall:    descend-rand-hint 1,000,000 ops in 2183ms, 458,026/sec, 2183 ns/op
tidwall(G): ascend-rand       1,000,000 ops in 1296ms, 771,686/sec, 1295 ns/op
tidwall(G): descend-rand      1,000,000 ops in 1299ms, 769,945/sec, 1298 ns/op
tidwall(G): ascend-rand-hint  1,000,000 ops in 1348ms, 741,877/sec, 1347 ns/op
tidwall(G): descend-rand-hint 1,000,000 ops in 1361ms, 734,523/sec, 1361 ns/op
tidwall(G): iter-rand         1,000,000 ops in 1412ms, 708,393/sec, 1411 ns/op
tidwall(G): iter-rand-hint    1,000,000 ops in 1499ms, 667,054/sec, 1499 ns/op

** scan **
Test scanning over every item in the tree
google:     ascend            1,000,000 ops in 11ms, 93,682,343/sec, 10 ns/op
google(G):  ascend            1,000,000 ops in 11ms, 86,964,227/sec, 11 ns/op
tidwall:    ascend            1,000,000 ops in 10ms, 99,877,839/sec, 10 ns/op
tidwall(G): scan              1,000,000 ops in 11ms, 93,493,462/sec, 10 ns/op
tidwall(G): walk              1,000,000 ops in 5ms, 198,087,112/sec, 5 ns/op
tidwall(G): iter              1,000,000 ops in 13ms, 75,106,341/sec, 13 ns/op

Compilation finished at Wed Mar 19 01:59:20
*/
