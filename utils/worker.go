package utils

import (
	"fmt"
	"sync"
)

type TaskQueue chan *MultiThreadDownloader

type Worker struct {
	TaskChan TaskQueue
}

type WorkerPool struct {
	sync.WaitGroup

	WorkerCount int
	TaskQueue   TaskQueue
	WorkerQueue chan TaskQueue
}

func NewWorker() *Worker {
	return &Worker{TaskChan: make(chan *MultiThreadDownloader)}
}

func NewWorkerPool(WorkerCount int) *WorkerPool {
	return &WorkerPool{
		WorkerCount: WorkerCount,
		TaskQueue:   make(TaskQueue),
		WorkerQueue: make(chan TaskQueue, WorkerCount),
	}
}

func (w *Worker) Run(wq chan TaskQueue, owner *WorkerPool) {
	go func() {
		for {
			wq <- w.TaskChan
			select {
			case t := <-w.TaskChan:
				err := t.Download()
				if err != nil {
					fmt.Printf("下载 %s 时出现错误 %s\n", t.FullPath, err)
					return
				}
				fmt.Println("下载完成", t.FullPath)
				owner.Done()
			}
		}
	}()
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.WorkerCount; i++ {
		w := NewWorker()
		w.Run(wp.WorkerQueue, wp)
	}
	go func() {
		for {
			select {
			case t := <-wp.TaskQueue:
				wp.Add(1)
				w := <-wp.WorkerQueue
				w <- t
			}
		}
	}()
}
