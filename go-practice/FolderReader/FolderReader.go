package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var verbose = flag.Bool("v", false, "显示详细的进度消息")

var sema = make(chan struct{}, 50)
var done = make(chan struct{})
var nfile, nbytes, ndirs int64

func main() {
	flag.Parse()
	roots := flag.Args()      //参数中没有能够按照预定义的参数解析的部分，通过flag.Args()即可获取，是一个字符串切片。
	var tick <-chan time.Time // 只写的time chan
	if *verbose {
		tick = time.Tick(500 * time.Millisecond) //time.Tick()返回的是一个channel,每隔指定的时间会有数据从channel中出来
	}
	if len(roots) == 0 {
		roots = []string{"."}
	}
	fileSizes := make(chan int64)
	var n sync.WaitGroup
	for _, root := range roots {
		n.Add(1)
		go walkDir(root, &n, fileSizes)
	}
	go func() {
		n.Wait()
		close(fileSizes)
	}()

	go func() {
		os.Stdin.Read(make([]byte, 1))
		close(done)
	}()
loop:
	for {
		select {
		case <-done:
			for range fileSizes {
				//
			}
		case size, ok := <-fileSizes:
			if !ok {
				break loop
			}
			nfile++
			nbytes += size

		case <-tick:
			printDiskUsage(nfile, nbytes)
		}
	}

	printDiskUsage(nfile, nbytes)
}

func printDiskUsage(nfiles, nbytes int64) {
	fmt.Printf("%d files  %d dirs %.1f KB\n", nfiles, ndirs, float64(nbytes/1024))
}

func walkDir(dir string, n *sync.WaitGroup, fileSizes chan<- int64) {
	defer n.Done()
	if cancelled() {
		return
	}
	for _, entry := range dirents(dir) {
		if entry.IsDir() {
			n.Add(1)
			ndirs++
			subdir := filepath.Join(dir, entry.Name())
			go walkDir(subdir, n, fileSizes)
		} else {
			fileSizes <- entry.Size()
		}
	}
}

func dirents(dir string) []os.FileInfo {
	select {
	case sema <- struct{}{}: //
	case <-done:
		return nil
	}
	defer func() {
		<-sema
	}()
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dul : %v\n", err)
		return nil
	}
	return entries
}

func cancelled() bool {
	select {
	case <-done:
		return true
	default:
		return false
	}
}
