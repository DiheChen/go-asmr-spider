package utils

import (
	"fmt"
)

type WorkerChan chan *MultiThreadDownloader

type WorkerPool struct {
	WorkerCount int
	TaskQueue   WorkerChan
	ResQueue    chan bool
}

func NewWorkerPool(WorkerCount int) *WorkerPool {
	return &WorkerPool{
		WorkerCount: WorkerCount,
		TaskQueue:   make(WorkerChan, WorkerCount),
	}
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.WorkerCount; i++ {
		go func() {
			for t := range wp.TaskQueue {
				err := t.Download()
				if err != nil {
					fmt.Printf("下载 %s 时出现错误 %s\n", t.FullPath, err)
					return
				}
				fmt.Println("下载完成", t.FullPath)
			}
		}()
	}
}
