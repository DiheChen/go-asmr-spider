package utils

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
)

type WorkerChan chan *MultiThreadDownloader

type WorkerPool struct {
	sync.WaitGroup
	Limit     int32
	TaskQueue WorkerChan
	Count     int32
}

func NewWorkerPool(WorkerCount int) *WorkerPool {
	return &WorkerPool{
		Limit:     int32(WorkerCount),
		TaskQueue: make(WorkerChan, WorkerCount),
	}
}

func (wp *WorkerPool) Start() {
	go func() {
		for t := range wp.TaskQueue {
			wp.Add(1)
			for {
				if atomic.LoadInt32(&wp.Count) < wp.Limit {
					break
				} else {
					runtime.Gosched()
				}
			}
			go func(t *MultiThreadDownloader) {
				atomic.AddInt32(&wp.Count, 1)
				defer atomic.AddInt32(&wp.Count, -1)
				err := t.Download()
				if err != nil {
					fmt.Printf("下载 %s 时出现错误 %s\n", t.FullPath, err)
					wp.Done()
					return
				}
				fmt.Println("下载完成", t.FullPath)
				wp.Done()
			}(t)
		}
	}()
}
