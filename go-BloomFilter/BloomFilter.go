package main

import (
	"fmt"
	"github.com/willf/bitset"
)

const SIZE = 2 << 24

var seeds = []uint{7, 11, 13, 31, 37, 61}

type Hasher struct {
	cap  uint
	seed uint
}
type BloomFilter struct {
	set   *bitset.BitSet
	funcs [6]Hasher
}

func InitBloomFilter() *BloomFilter {
	bf := new(BloomFilter)
	for i := 0; i < len(seeds); i++ {
		bf.funcs[i] = Hasher{SIZE, seeds[i]}
	}
	bf.set = bitset.New(SIZE)
	return bf
}
func (b *BloomFilter) Add(i string) {
	for _, v := range b.funcs {
		b.set.Set(v.hash(i))
	}
}
func (b *BloomFilter) contain(value string) bool {
	res := true
	for _, v := range b.funcs {
		res = res && b.set.Test(v.hash(value))
	}
	return res
}
func (s Hasher) hash(value string) uint {
	var res uint = 0
	for i := 0; i < len(value); i++ {
		res = res*s.seed + uint(value[i])
	}
	return (s.cap - 1) & res
}

func main() {
	filter := InitBloomFilter()
	fmt.Println(filter.funcs[1].seed)
	str1 := "hello,bloom filter!"
	filter.Add(str1)
	str2 := "Hello ,go"
	filter.Add(str2)
	str3 := "PingCap"
	filter.Add(str3)
	fmt.Println(filter.contain(str1))
	fmt.Println(filter.contain(str2))
	fmt.Println(filter.contain(str3))
	fmt.Println(filter.contain("The Great Wall"))
}
