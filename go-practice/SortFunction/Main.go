package main

import (
	"fmt"
)

// 冒泡排序
// 时间复杂度: 平均 n**2 最好 n 最坏 n**2
// 空间复杂度 1
// 内排序 稳定
func bubbleSort(arr []int) []int {
	length := len(arr)
	for i := 0; i < length; i++ {
		for j := 0; j < length-1; j++ {
			if arr[j] > arr[j+1] {
				arr[j], arr[j+1] = arr[j+1], arr[j]
			}
		}
	}
	return arr
}

// 选择排序
// 时间复杂度: 平均 n**2 最好 n**2 最坏 n**2
// 空间复杂度 1
// 内排序 不稳定
func selectSort(arr []int) []int {
	length := len(arr)
	for i := 0; i < length; i++ {
		min_index := i
		for j := i; j < length; j++ {
			if arr[j] < arr[min_index] {
				min_index = j
			}
		}
		arr[i], arr[min_index] = arr[min_index], arr[i]
	}
	return arr

}

// 插入排序
// 时间复杂度: 平均 n**2 最好 n 最坏 n**2
// 空间复杂度 1
// 内排序 稳定
func insertSort(arr []int) []int {
	for i := range arr {
		preIndex := i - 1
		cur := arr[i]
		for preIndex >= 0 && arr[preIndex] > cur {
			arr[preIndex+1] = arr[preIndex]
			preIndex--
		}
		arr[preIndex+1] = cur
	}
	return arr

}

// 希尔排序
// 时间复杂度: 平均 nlogn 最好 nlog2n 最坏 nlog2n
// 空间复杂度 1
// 内排序 不稳定
func shellSort(arr []int) []int {
	length := len(arr)
	gap := length / 2
	for gap > 0 {
		for i := gap; i < length; i++ {
			preIndex := i - gap
			cur := arr[i]
			for preIndex >= 0 && arr[preIndex] > cur {
				arr[preIndex+gap] = arr[preIndex]
				preIndex -= gap
			}
			arr[preIndex+gap] = cur
		}
		gap /= 2
	}
	return arr
}

// 归并排序
// 时间复杂度: 平均 nlogn 最好 nlogn 最坏 nlogn
// 空间复杂度 n
// 外排序 不稳定
func mergeSort(arr []int) []int {
	length := len(arr)
	if length < 2 {
		return arr
	}
	mid := length / 2
	left := arr[:mid]
	right := arr[mid:]
	return merge(mergeSort(left), mergeSort(right))
}
func merge(left, right []int) []int {
	var res []int
	left_i := 0
	right_i := 0
	for left_i < len(left) && right_i < len(right) {
		if left[left_i] < right[right_i] {
			res = append(res, left[left_i])
			left_i++
		} else {
			res = append(res, right[right_i])
			right_i++
		}
	}
	for left_i < len(left) {
		res = append(res, left[left_i])
		left_i++
	}
	for right_i < len(right) {
		res = append(res, right[right_i])
		right_i++
	}
	return res
}

// 快速排序
// 时间复杂度: 平均 nlogn 最好 nlogn 最坏 n**2
// 空间复杂度 logn
// 内排序 不稳定
func quickSort(arr []int) []int {
	return _quickSort(arr, 0, len(arr)-1)
}
func _quickSort(arr []int, left, right int) []int {

	if left < right {
		p := partition(arr, left, right)
		_quickSort(arr, left, p)
		_quickSort(arr, p+1, right)
	}
	return arr
}
func partition(arr []int, left, right int) int {
	p := arr[left]
	index := left
	for i := left + 1; i <= right; i++ {
		if arr[i] < p {
			arr[index], arr[i] = arr[i], arr[index]
			index++
		}
	}
	arr[index] = p
	return index
}

// 堆排序
// 时间复杂度: 平均 nlogn 最好 nlogn 最坏 nlogn
// 空间复杂度 1
// 内排序 不稳定
func heapSort(arr []int) []int {
	length := len(arr)
	buildMaxHeap(arr, length)
	for i := length - 1; i >= 0; i-- {
		arr[i], arr[0] = arr[0], arr[i]
		length -= 1
		heapify(arr, 0, length)
	}
	return arr
}
func buildMaxHeap(arr []int, length int) {
	for i := length / 2; i >= 0; i-- {
		heapify(arr, i, length)
	}
}
func heapify(arr []int, i, length int) {
	left := 2*i + 1
	right := 2*i + 2
	largest := i
	if left < length && arr[left] > arr[largest] {
		largest = left
	}
	if right < length && arr[right] > arr[largest] {
		largest = right
	}
	if largest != i {
		arr[i], arr[largest] = arr[largest], arr[i]
		heapify(arr, largest, length)
	}

}

func main() {
	arr := []int{2, 7, 3, 9, 5, 9, 1, 5, 8, 4, 2, 7, 9, 3, 0, 11}
	fmt.Println(heapSort(arr))
}
