package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/DiheChen/go-asmr-spider/spider"
)

func main() {
	fmt.Println("请输入要下载的音声的 RJ 号, 如: RJ373001, 如果要下载多个, 请用空格分开。")
	readString, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	tasks := strings.Split(strings.TrimSpace(readString), " ")
	s, err := spider.Login()
	if err != nil {
		fmt.Println(err)
	}
	for _, task := range tasks {
		s.Download(task)
	}
}
