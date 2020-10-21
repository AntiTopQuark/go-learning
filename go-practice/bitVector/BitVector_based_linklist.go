package main

import (
	"fmt"
	"reflect"
)

const bitCounts = 32 << (^uint(0) >> 63) // 32位平台这个值就是32，64位平台这个值就是64

type IntSet_Linklist struct {
	id    int
	num   uint
	pre   *IntSet_Linklist
	next  *IntSet_Linklist
	count int
}

// 判断s是否为空
func (s *IntSet_Linklist) IsEmpty() bool {
	return reflect.DeepEqual(s, IntSet_Linklist{})
}

// 集合中是否存在x
func (s *IntSet_Linklist) Has(x int) bool {
	if s.next.IsEmpty() {
		return false
	}
	word, bit := x/bitCounts, uint(x%bitCounts)
	for {
		if s.next == nil {
			break
		} else if s.next.id > word {
			//从小到大添加
			break
		} else if s.next.id == word {
			// word节点已经存在
			if s.next.num&(1<<bit) != 0 {
				return true
			} else {
				break
			}
		}
		s = s.next
	}
	return false
}

// 添加一个数x到集合中
func (s *IntSet_Linklist) Add(x int) {
	s.count++
	word, bit := x/bitCounts, uint(x%bitCounts)
	tmp := s
	last := false
	for {
		if tmp.next == nil {
			last = true
			break
		} else if tmp.next.id > word {
			//从小到大添加
			break
		} else if tmp.next.id == word {
			// word节点已经存在
			tmp.next.num |= (1 << bit)
		}
		tmp = tmp.next
	}

	if last {
		tmp.next = &IntSet_Linklist{
			id:    word,
			num:   1 << bit,
			pre:   tmp,
			next:  nil,
			count: 1,
		}
	} else {
		new_add := IntSet_Linklist{
			id:    word,
			num:   1 << bit,
			pre:   tmp,
			next:  tmp.next,
			count: 1,
		}
		tmp.next = &new_add
		new_add.next = &new_add
	}
}
func main() {
	fmt.Println("你的机器是32 or 64?", bitCounts)
	set1 := IntSet_Linklist{}
	set1.Add(999999999999)
	fmt.Println(set1.Has(999999999999))
}
