package main

import (
	"bufio"
	"fmt"
	"go-ExternalSort/pipeline"
	"os"
)

func main() {
	p := createPipeline("large.in", 800000000, 100)
	writeToFile(p, "large.out")
	printFile("large.out")
}

func printFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	p := pipeline.ReaderSource(file, -1)
	count := 0
	for v := range p {
		fmt.Println(v)
		count++
		if count >= 100 {
			break
		}
	}
}

func writeToFile(p <-chan int, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	defer writer.Flush()

	pipeline.WriteSink(writer, p)

}

func createPipeline(filename string, filesize, chunkCount int) <-chan int {
	chunkSize := filesize / chunkCount
	sortRes := []<-chan int{}
	pipeline.Init()
	for i := 0; i < chunkCount; i++ {
		file, err := os.Open(filename)
		if err != nil {
			panic(err)
		}
		file.Seek(int64(i*chunkSize), 0)

		source := pipeline.ReaderSource(
			bufio.NewReader(file), chunkSize,
		)
		sortRes = append(sortRes, pipeline.InMemSort(source))

	}
	return pipeline.MergeN(sortRes...)

}
