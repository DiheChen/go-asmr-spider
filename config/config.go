package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Config struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

func generateDefaultConfig() {
	config, err := json.Marshal(map[string]string{
		"account":  "guest",
		"password": "guest",
	})
	if err != nil {
		fmt.Print("生成默认配置文件失败", err)
		os.Exit(0)
	}
	_ = os.WriteFile("config.json", config, 0644)
	fmt.Print("已生成默认配置文件config.json, 请修改配置文件后重新运行程序, 若不修改则会使用游客账号登录。")
	os.Exit(0)
}

func GetConfig() *Config {
	if _, err := os.Stat("config.json"); os.IsNotExist(err) {
		generateDefaultConfig()
	}
	file, err := os.Open("config.json")
	if err != nil {
		fmt.Print("打开配置文件失败", err)
		os.Exit(0)
	}
	defer func() { _ = file.Close() }()
	all, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Print("读取配置文件失败", err)
		os.Exit(0)
	}
	var config Config
	err = json.Unmarshal(all, &config)
	if err != nil {
		fmt.Print("解析配置文件失败", err)
		os.Exit(0)
	}
	return &config
}
