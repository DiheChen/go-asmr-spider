package spider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/DiheChen/go-asmr-spider/config"
	"github.com/DiheChen/go-asmr-spider/utils"
)

var Conf = config.GetConfig()

type ASMRClient struct {
	Authorization string
	WorkerPool    *utils.WorkerPool
	ThreadCount   int
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

func NewASMRClient(maxTask int, maxThread int) *ASMRClient {
	return &ASMRClient{
		WorkerPool:  utils.NewWorkerPool(maxTask),
		ThreadCount: maxThread,
	}
}

func (ac *ASMRClient) Login() error {
	payload, err := json.Marshal(map[string]string{
		"name":     Conf.Account,
		"password": Conf.Password,
	})
	if err != nil {
		fmt.Println("登录失败, 配置文件有误。")
		return err
	}
	client := utils.Client.Get().(*http.Client)
	req, _ := http.NewRequest("POST", "https://api.asmr.one/api/auth/me", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", "https://www.asmr.one/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36")
	resp, err := client.Do(req)
	utils.Client.Put(client)
	if err != nil {
		fmt.Println("登录失败, 网络错误。请尝试通过环境变量的方式设置代理。")
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	all, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print("登录失败, 读取响应失败。")
		return err
	}
	res := make(map[string]string)
	err = json.Unmarshal(all, &res)
	ac.Authorization = "Bearer " + res["token"]
	return nil
}

func (ac *ASMRClient) GetVoiceTracks(id string) ([]track, error) {
	client := utils.Client.Get().(*http.Client)
	req, _ := http.NewRequest("GET", "https://api.asmr.one/api/tracks/"+id, nil)
	req.Header.Set("Authorization", ac.Authorization)
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

func (ac *ASMRClient) Download(id string) {
	id = strings.Replace(id, "RJ", "", 1)
	fmt.Println("作品 RJ 号: " + id)
	tracks, err := ac.GetVoiceTracks(id)
	if err != nil {
		fmt.Println("获取作品失败: " + err.Error())
		return
	}
	ac.EnsureDir(tracks, "RJ"+id)
	fmt.Println("下载完成。")
}

func (ac *ASMRClient) DownloadFile(url string, dirPath string, fileName string) {
	if runtime.GOOS == "windows" {
		for _, str := range []string{"?", "<", ">", ":", "/", "\\", "*", "|"} {
			fileName = strings.Replace(fileName, str, "_", -1)
		}
	}
	savePath := dirPath + "/" + fileName
	if utils.PathExists(savePath) {
		fmt.Println("文件已存在, 跳过。")
		return
	}
	fmt.Println("正在下载 " + savePath)
	downloader := utils.NewDownloader(url, dirPath, fileName, ac.ThreadCount, map[string]string{
		"Referer": "https://www.asmr.one/",
	})
	ac.WorkerPool.TaskQueue <- downloader
}

func (ac *ASMRClient) EnsureDir(tracks []track, basePath string) {
	path := basePath
	_ = os.MkdirAll(path, os.ModePerm)
	for _, t := range tracks {
		if t.Type != "folder" {
			ac.DownloadFile(t.MediaDownloadURL, path, t.Title)
		} else {
			ac.EnsureDir(t.Children, fmt.Sprintf("%s/%s", path, t.Title))
		}
	}
}
