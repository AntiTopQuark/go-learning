package main

import (
	"bytes"
	"fmt"
)

// 1.布尔值无法隐式转换成数值，反之也不行。如果需要把布尔值转成0或1，需要显示的使用if：
// 如果转换操作使用频繁，值得专门写成一个函数：
func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

func itob(i int) bool {
	return i != 0
}

//2. 函数 intsToString 与 fmt.Sprintf(values) 类似，但插入了逗号\
func intsToString(arr []int) string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	for i, v := range arr {
		if i > 0 {
			buf.WriteString(", ")
		}
		fmt.Fprintf(&buf, "%d", v)
	}
	buf.WriteByte(']')
	return buf.String()
}

// 3. 切片的平移\反转\比较
func reverse(s []int) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
func moveLeft(n int, s []int) {
	reverse(s[n:])
	reverse(s[:n])
	reverse(s)

}

func moveRight(n int, s []int) {
	moveLeft(len(s)-n, s)
}

func equal(s1, s2 []int) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i, v := range s1 {
		if s2[i] != v {
			return false
		}
	}
	return true
}

func main() {
	// 2.
	fmt.Println(intsToString([]int{1, 2, 3})) // "[1, 2, 3]"
	fmt.Println([]int{1, 2, 3})               // "[1 2 3]"

	//3.
	arr := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	reverse(arr)
	fmt.Println(arr)
	moveLeft(3, arr)
	fmt.Println(arr)
	moveRight(3, arr)
	fmt.Println(arr)

}
