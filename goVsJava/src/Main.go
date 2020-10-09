package main

import (
	"fmt"
	"time"
)

func main() {
	// 记录开始时间
	start:= time.Now().UnixNano()
	num := 100000000
	for i := 0; i < num; i++ {
		arr := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
		bubbleSort(&arr)
		//bubbleSort2(arr)
	}
	//打印消耗时间
	fmt.Println(time.Now().UnixNano() -start)
}

func bubbleSort(array *[]int) {
	arr := *array
	for j := 0; j < len(arr)-1; j++ {
		for k := 0; k < len(arr)-1-j; k++ {
			if(arr)[k] < (arr)[k+1] {
				temp := (arr)[k]
				(arr)[k] = (arr)[k+1]
				(arr)[k+1] = temp
			}
		}
	}
}

////排序
//func bubbleSort2(arr []int) {
//	for j := 0; j < len(arr)-1; j++ {
//		for k := 0; k < len(arr)-1-j; k++ {
//			if arr[k] < arr[k+1] {
//				temp := arr[k]
//				arr[k] = arr[k+1]
//				arr[k+1] = temp
//			}
//		}
//	}
//}