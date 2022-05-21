package utils

import (
	"fmt"
	"sync"
)

type WorkerChan chan *MultiThreadDownloader

type Worker struct {
	TaskChan WorkerChan
}

type WorkerPool struct {
	sync.WaitGroup
	WorkerCount int
	TaskQueue   WorkerChan
}

func NewWorker() *Worker {
	return &Worker{TaskChan: make(chan *MultiThreadDownloader)}
}

func NewWorkerPool(WorkerCount int) *WorkerPool {
	return &WorkerPool{
		WorkerCount: WorkerCount,
		TaskQueue:   make(WorkerChan, WorkerCount),
	}
}

func (w *Worker) Run(owner *WorkerPool) {
	for t := range owner.TaskQueue {
		owner.Add(1)
		go func(t *MultiThreadDownloader) {
			err := t.Download()
			if err != nil {
				fmt.Printf("下载 %s 时出现错误 %s\n", t.FullPath, err)
				return
			}
			fmt.Println("下载完成", t.FullPath)
			owner.Done()
		}(t)
	}
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.WorkerCount; i++ {
		w := NewWorker()
		go w.Run(wp)
	}
}
