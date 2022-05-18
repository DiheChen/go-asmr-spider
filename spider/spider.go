package spider

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/DiheChen/go-asmr-spider/config"
	"github.com/DiheChen/go-asmr-spider/utils"
)

var conf = config.GetConfig()

type OAuth2 struct {
	Authorization string
}

type track struct {
	Type             string  `json:"type"`
	Title            string  `json:"title"`
	Children         []track `json:"children,omitempty"`
	Hash             string  `json:"hash,omitempty"`
	WorkTitle        string  `json:"workTitle,omitempty"`
	MediaStreamURL   string  `json:"mediaStreamUrl,omitempty"`
	MediaDownloadURL string  `json:"mediaDownloadUrl,omitempty"`
}

func Login() (*OAuth2, error) {
	payload, err := json.Marshal(map[string]string{
		"name":     conf.Account,
		"password": conf.Password,
	})
	if err != nil {
		fmt.Println("登录失败, 配置文件有误。")
		return nil, err
	}
	client := utils.Client.Get().(*http.Client)
	req, _ := http.NewRequest("POST", "https://api.asmr.one/api/auth/me", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", "https://www.asmr.one/")
	req.Header.Set("User-Agent", "PostmanRuntime/7.29.0")
	resp, err := client.Do(req)
	utils.Client.Put(client)
	if err != nil {
		fmt.Println("登录失败, 网络错误。请尝试通过环境变量的方式设置代理。")
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	all, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print("登录失败, 读取响应失败。")
		return nil, err
	}
	res := make(map[string]string)
	err = json.Unmarshal(all, &res)
	return &OAuth2{Authorization: "Bearer " + res["token"]}, nil
}

func (OAuth2 *OAuth2) GetVoiceTracks(id string) ([]track, error) {
	client := utils.Client.Get().(*http.Client)
	req, _ := http.NewRequest("GET", "https://api.asmr.one/api/tracks/"+id, nil)
	req.Header.Set("Authorization", OAuth2.Authorization)
	req.Header.Set("Referer", "https://www.asmr.one/")
	req.Header.Set("User-Agent", "PostmanRuntime/7.29.0")
	resp, err := client.Do(req)
	utils.Client.Put(client)
	if err != nil {
		fmt.Println("获取音声信息失败:", err)
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	all, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("获取音声信息失败: ", err)
		return nil, err
	}
	res := make([]track, 0)
	err = json.Unmarshal(all, &res)
	return res, nil
}

func DownloadFile(url string, dirPath string, fileName string) {
	if runtime.GOOS == "windows" {
		for _, str := range []string{"?", "<", ">", ":", "/", "\\", "*", "|"} {
			fileName = strings.Replace(fileName, str, "_", -1)
		}
	}
	savePath := dirPath + "/" + fileName
	if pathExists(savePath) {
		fmt.Println("文件已存在, 跳过。")
		return
	}
	client := utils.Client.Get().(*http.Client)
	fmt.Println("正在下载 " + savePath)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Referer", "https://www.asmr.one/")
	req.Header.Set("User-Agent", "PostmanRuntime/7.29.0")
	resp, err := client.Do(req)
	utils.Client.Put(client)
	if err != nil {
		fmt.Println("下载"+savePath+"失败: ", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	file, err := os.OpenFile(dirPath+"/"+fileName+".temp", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("下载"+savePath+"失败: ", err)
		return
	}
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Println("写入"+savePath+"失败: ", err)
		return
	}
	_ = file.Close()
	err = os.Rename(dirPath+"/"+fileName+".temp", dirPath+"/"+fileName)
	if err != nil {
		fmt.Println("重命名"+savePath+"失败: ", err)
		return
	}
	return
}

func EnsureDir(tracks []track, basePath string) {
	path := basePath
	_ = os.MkdirAll(path, os.ModePerm)
	for _, t := range tracks {
		if t.Type != "folder" {
			DownloadFile(t.MediaDownloadURL, path, t.Title)
		} else {
			EnsureDir(t.Children, fmt.Sprintf("%s/%s", path, t.Title))
		}
	}
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || errors.Is(err, os.ErrExist)
}

func (OAuth2 *OAuth2) Download(id string) {
	id = strings.Replace(id, "RJ", "", 1)
	fmt.Println("作品 RJ 号: " + id)
	tracks, err := OAuth2.GetVoiceTracks(id)
	if err != nil {
		fmt.Println("获取作品失败: " + err.Error())
		return
	}
	EnsureDir(tracks, "RJ"+id)
	fmt.Println("下载完成。")
}
