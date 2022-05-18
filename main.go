package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/DiheChen/go-asmr-spider/spider"
)

var (
	maxTask   int
	maxThread int
)

func init() {
	flag.IntVar(&maxTask, "w", spider.Conf.MaxTask, "最多同时进行的下载任务")
	flag.IntVar(&maxThread, "t", spider.Conf.MaxThread, "单文件最大线程")
}

func main() {
	flag.Parse()
	fmt.Printf("正在使用 %d 线程, %d 任务同时下载\n", maxThread, maxTask)
	fmt.Println("请输入要下载的音声的 RJ 号, 如: RJ373001, 如果要下载多个, 请用空格分开。")
	readString, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	tasks := strings.Split(strings.TrimSpace(readString), " ")
	c := spider.NewASMRClient(maxTask, maxThread)
	c.WorkerPool.Start()
	err := c.Login()
	if err != nil {
		fmt.Println("登录失败:", err)
		return
	}
	for _, task := range tasks {
		c.Download(task)
	}
}
