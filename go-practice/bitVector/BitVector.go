package main

import (
	"bytes"
	"fmt"
)

const bitCounts = 32 << (^uint(0) >> 63) // 32位平台这个值就是32，64位平台这个值就是64
// 这是一个包含非负整数的集合
// 零值代表空的集合
type IntSet struct {
	words []uint
	num   int
}

// 集合中是否存在x
func (s *IntSet) Has(x int) bool {
	word, bit := x/bitCounts, uint(x%bitCounts)
	return word < len(s.words) && s.words[word]&(1<<bit) != 0
}

// 添加一个数x到集合中
func (s *IntSet) Add(x int) {
	word, bit := x/bitCounts, uint(x%bitCounts)
	for word >= len(s.words) {
		s.words = append(s.words, 0)
	}
	s.words[word] |= 1 << bit
	s.num += 1
}

// 求并集，并保存到s中
func (s *IntSet) UnionWith(other IntSet) {
	for i, word := range other.words {
		if i >= len(s.words) {
			s.words = append(s.words, word)
		} else {
			s.words[i] |= word
		}
	}
}

// 求差集,并保存在s
func (s *IntSet) DifferenceWith(other IntSet) {
	for i, word := range other.words {
		if i >= len(s.words) {
			s.words[i] &^= word
		}
	}
}

// 求交集,并保存在s
func (s *IntSet) IntersectionWith(other IntSet) {
	for i, word := range other.words {
		if i >= len(s.words) {
			s.words[i] &= word
		}
	}
}

// 打印
func (s *IntSet) String() string {
	var buf bytes.Buffer
	buf.WriteByte('(')
	for i, word := range s.words {
		if word == 0 {
			continue
		}
		for j := 0; j < bitCounts; j++ {
			if word&(1<<uint(j)) != 0 {
				if buf.Len() > len("{") {
					buf.WriteString(", ")
				}
				fmt.Fprintf(&buf, "%d", i*bitCounts+j)

			}
		}
	}
	buf.WriteByte(')')
	return buf.String()
}

//计算置位个数
func (s *IntSet) Len() int {
	return s.num
}

// 一次性添加多个元素
func (s *IntSet) AddAll(arr ...int) {
	for _, v := range arr {
		s.Add(v)
	}
}

// 去除某个元素
func (s *IntSet) Remove(x int) {
	word, bit := x/bitCounts, x%bitCounts
	if word < len(s.words) {
		if s.words[word]&(1<<bit) == 1 {
			s.num--
		}
		s.words[word] &^= 1 << bit
	}
	for i := len(s.words) - 1; i >= 0; i-- {
		if s.words[i] == 0 {
			s.words = s.words[:i]
		} else {
			break
		}
	}
}

// 清空
func (s *IntSet) Clear() {
	s.words = []uint{}
	s.num = 0
}

// 返回包含集合元素的 slice，这适合在 range 循环中使用
func (s *IntSet) Elem() []uint {
	res := []uint{}
	for word, bit := range s.words {
		for j := 0; j < bitCounts; j++ {
			if (bit & (1 << uint(j))) != 0 {
				res = append(res, uint(bitCounts*word+j))
			}
		}
	}
	return res
}

// 复制
func (s *IntSet) Copy() *IntSet {
	res := IntSet{}
	res.words = make([]uint, len(s.words))
	copy(res.words, s.words)
	res.num = s.num
	return &res
}

// 增加多个元素
func main() {
	fmt.Println("你的机器是32 or 64?", bitCounts)
	set1 := IntSet{}
	set1.Add(1)
	set1.Add(3)
	set1.Add(99)
	fmt.Println("测试 Add && Has函数\n", set1.Has(3))
	fmt.Println(set1.Has(1))
	fmt.Println(set1.Has(98))
	fmt.Println("测试String\n", set1.String())

	set2 := IntSet{}
	set2.Add(4)
	set2.Add(3)
	set2.Add(99)
	set2.Add(199)
	fmt.Println("测试并集")
	set1.UnionWith(set2)
	fmt.Println(set1.String())

	set2.Remove(199)
	fmt.Println("测试Remove\n", set2.String())
	fmt.Println("测试Elem\n", set2.Elem())
	for _, v := range set2.Elem() {
		fmt.Println(v)
	}
}
