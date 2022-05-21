package utils

import (
	"fmt"
	"sync"
)

type WorkerChan chan *MultiThreadDownloader

type WorkerPool struct {
	sync.WaitGroup
	WorkerCount int
	TaskQueue   WorkerChan
}

func NewWorkerPool(WorkerCount int) *WorkerPool {
	return &WorkerPool{
		WorkerCount: WorkerCount,
		TaskQueue:   make(WorkerChan, WorkerCount),
	}
}

func (wp *WorkerPool) Start() {
	go func() {
		for t := range wp.TaskQueue {
			wp.Add(1)
			go func(t *MultiThreadDownloader) {
				err := t.Download()
				if err != nil {
					fmt.Printf("下载 %s 时出现错误 %s\n", t.FullPath, err)
					return
				}
				fmt.Println("下载完成", t.FullPath)
				wp.Done()
			}(t)
		}
	}()
}
