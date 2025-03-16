package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"

	"github.com/glycerine/uart"
	gbtree "github.com/google/btree"
	tbtree "github.com/tidwall/btree"
	"github.com/tidwall/lotsa"
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
	tr.SkipLocking = true
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
func newTBTreeM(degree int) *tbtree.Map[keyT, valT] {
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
	ttrM := newTBTreeM(degree)
	uART := newUART()

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
			ttr = ttr.Copy()
			lotsa.Ops(N, 1, func(i, _ int) {
				ttr.Set(items[i])
			})
			print_label("tidwall(G)", "set-after-copy")
			ttrG = ttrG.Copy()
			lotsa.Ops(N, 1, func(i, _ int) {
				ttrG.Set(items[i])
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
