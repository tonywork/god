package radix

import (
	"../murmur"
	"bytes"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var benchmarkTestTree *Tree
var benchmarkTestKeys [][]byte
var benchmarkTestValues []Hasher

func init() {
	rand.Seed(time.Now().UnixNano())
	benchmarkTestTree = new(Tree)
}

func assertExistance(t *testing.T, tree *Tree, k, v string) {
	if value, existed := tree.Get([]byte(k)); !existed || value != StringHasher(v) {
		t.Errorf("%v should contain %v => %v, got %v, %v", tree.Describe(), rip([]byte(k)), v, value, existed)
	}
}

func assertNewPut(t *testing.T, tree *Tree, k, v string) {
	if value, existed := tree.Put([]byte(k), StringHasher(v)); existed || value != nil {
		t.Errorf("%v should not contain %v, got %v, %v", tree.Describe(), rip([]byte(k)), value, existed)
	}
}

func assertOldPut(t *testing.T, tree *Tree, k, v, old string) {
	if value, existed := tree.Put([]byte(k), StringHasher(v)); !existed || value != StringHasher(old) {
		t.Errorf("%v should contain %v => %v, got %v, %v", tree.Describe(), rip([]byte(k)), v, value, existed)
	}
}

func assertNonExistance(t *testing.T, tree *Tree, k string) {
	if value, existed := tree.Get([]byte(k)); existed || value != nil {
		t.Errorf("%v should not contain %v, got %v, %v", tree, k, value, existed)
	}
}

func TestRadixHash(t *testing.T) {
	tree1 := new(Tree)
	var keys [][]byte
	var vals []StringHasher
	n := 10000
	for i := 0; i < n; i++ {
		k := []byte(fmt.Sprint(rand.Int63()))
		v := StringHasher(fmt.Sprint(rand.Int63()))
		keys = append(keys, k)
		vals = append(vals, v)
		tree1.Put(k, v)
	}
	tree2 := new(Tree)
	for i := 0; i < n; i++ {
		index := rand.Int() % len(keys)
		k := keys[index]
		v := vals[index]
		tree2.Put(k, v)
		keys = append(keys[:index], keys[index+1:]...)
		vals = append(vals[:index], vals[index+1:]...)
	}
	if bytes.Compare(tree1.Hash(), tree2.Hash()) != 0 {
		t.Errorf("%v and %v have hashes\n%v\n%v\nand they should be equal!", tree1.Describe(), tree2.Describe(), tree1.Hash(), tree2.Hash())
	}
}

func TestRadixBasicOps(t *testing.T) {
	tree := new(Tree)
	assertNonExistance(t, tree, "apple")
	assertNewPut(t, tree, "apple", "stonefruit")
	assertOldPut(t, tree, "apple", "fruit", "stonefruit")
	assertNonExistance(t, tree, "crab")
	assertNewPut(t, tree, "crab", "critter")
	assertOldPut(t, tree, "crab", "animal", "critter")
	assertNonExistance(t, tree, "crabapple")
	assertNewPut(t, tree, "crabapple", "poop")
	assertOldPut(t, tree, "crabapple", "fruit", "poop")
	assertNonExistance(t, tree, "banana")
	assertNewPut(t, tree, "banana", "yellow")
	assertOldPut(t, tree, "banana", "fruit", "yellow")
	assertNonExistance(t, tree, "guava")
	assertNewPut(t, tree, "guava", "fart")
	assertOldPut(t, tree, "guava", "fruit", "fart")
	assertNonExistance(t, tree, "guanabana")
	assertNewPut(t, tree, "guanabana", "man")
	assertOldPut(t, tree, "guanabana", "city", "man")
	if old, existed := tree.Put(nil, StringHasher("nil")); old != nil || existed {
		t.Error("should not exist yet")
	}
	if old, existed := tree.Put([]byte("nil"), nil); old != nil || existed {
		t.Error("should not exist yet")
	}
	assertExistance(t, tree, "apple", "fruit")
	assertExistance(t, tree, "crab", "animal")
	assertExistance(t, tree, "crabapple", "fruit")
	assertExistance(t, tree, "banana", "fruit")
	assertExistance(t, tree, "guava", "fruit")
	assertExistance(t, tree, "guanabana", "city")
	if value, existed := tree.Get(nil); !existed || value != StringHasher("nil") {
		t.Errorf("%v should contain %v => %v, got %v, %v", tree, nil, "nil", value, existed)
	}
	if value, existed := tree.Get([]byte("nil")); !existed || value != nil {
		t.Errorf("%v should contain %v => %v, got %v, %v", tree, "nil", nil, value, existed)
	}
	tree.Put([]byte("crab"), StringHasher("crustacean"))
	assertExistance(t, tree, "crab", "crustacean")
	assertNonExistance(t, tree, "gua")

}

func benchTree(b *testing.B, n int) {
	b.StopTimer()
	for len(benchmarkTestKeys) < n {
		benchmarkTestKeys = append(benchmarkTestKeys, murmur.HashString(fmt.Sprint(len(benchmarkTestKeys))))
		benchmarkTestValues = append(benchmarkTestValues, StringHasher(fmt.Sprint(len(benchmarkTestValues))))
	}
	for benchmarkTestTree.Size() < n {
		benchmarkTestTree.Put(benchmarkTestKeys[benchmarkTestTree.Size()], benchmarkTestValues[benchmarkTestTree.Size()])
	}
	var keys [][]byte
	var vals []Hasher
	for i := 0; i < b.N; i++ {
		keys = append(keys, murmur.HashString(fmt.Sprint(rand.Int63())))
		vals = append(vals, StringHasher(fmt.Sprint(rand.Int63())))
	}
	var k []byte
	var v Hasher
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		k = benchmarkTestKeys[i % len(benchmarkTestKeys)]
		v = benchmarkTestValues[i % len(benchmarkTestValues)]
		benchmarkTestTree.Put(k, v)
		j, existed := benchmarkTestTree.Get(k)
		if j != v {
			b.Fatalf("%v should contain %v, but got %v, %v", benchmarkTestTree.Describe(), v, j, existed)
		}
	}
}

func BenchmarkTree10(b *testing.B) {
	benchTree(b, 10)
}

func BenchmarkTree100(b *testing.B) {
	benchTree(b, 100)
}

func BenchmarkTree1000(b *testing.B) {
	benchTree(b, 1000)
}

func BenchmarkTree10000(b *testing.B) {
	benchTree(b, 10000)
}

func BenchmarkTree100000(b *testing.B) {
	benchTree(b, 100000)
}

func BenchmarkTree1000000(b *testing.B) {
	benchTree(b, 1000000)
}
