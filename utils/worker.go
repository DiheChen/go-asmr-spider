package utils

import (
	"fmt"
	"sync"
)

type WorkerChan chan *MultiThreadDownloader

type WorkerPool struct {
	sync.WaitGroup
	cond      *sync.Cond
	TaskQueue WorkerChan
	Limit     int
	Count     int
}

func NewWorkerPool(WorkerCount int) *WorkerPool {
	return &WorkerPool{
		cond:      sync.NewCond(&sync.Mutex{}),
		Limit:     WorkerCount,
		TaskQueue: make(WorkerChan, WorkerCount),
	}
}

func (wp *WorkerPool) Start() {
	go func() {
		for t := range wp.TaskQueue {
			wp.cond.L.Lock()
			for wp.Count >= wp.Limit {
				wp.cond.Wait()
			}
			wp.Add(1)
			wp.cond.L.Unlock()
			go func(t *MultiThreadDownloader) {
				wp.cond.L.Lock()
				wp.Count++
				wp.cond.L.Unlock()
				defer func() {
					wp.cond.L.Lock()
					wp.Count--
					wp.Done()
					wp.cond.Broadcast()
					wp.cond.L.Unlock()
				}()
				err := t.Download()
				if err != nil {
					fmt.Printf("下载 %s 时出现错误 %s\n", t.FullPath, err)
					return
				}
				fmt.Println("下载完成", t.FullPath)
			}(t)
		}
	}()
}
