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

var conf = config.GetConfig()

type OAuth2 struct {
	Authorization string
}

func Login() (*OAuth2, error) {
	payload, err := json.Marshal(map[string]string{
		"name":     conf.Account,
		"password": conf.Password,
	})
	if err != nil {
		fmt.Print(err)
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
		fmt.Print(err)
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	all, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err)
		return nil, err
	}
	res := make(map[string]string)
	err = json.Unmarshal(all, &res)
	return &OAuth2{Authorization: "Bearer " + res["token"]}, nil
}

func (OAuth2 *OAuth2) GetVoiceTracks(id string) ([]interface{}, error) {
	client := utils.Client.Get().(*http.Client)
	req, _ := http.NewRequest("GET", "https://api.asmr.one/api/tracks/"+id, nil)
	req.Header.Set("Authorization", OAuth2.Authorization)
	req.Header.Set("Referer", "https://www.asmr.one/")
	req.Header.Set("User-Agent", "PostmanRuntime/7.29.0")
	resp, err := client.Do(req)
	utils.Client.Put(client)
	if err != nil {
		fmt.Print(err)
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	all, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err)
		return nil, err
	}
	var res []interface{}
	err = json.Unmarshal(all, &res)
	return res, nil
}

func DownloadFile(url string, dirPath string, fileName string) {
	if runtime.GOOS == "windows" {
		for _, str := range []string{"?", "<", ">", ":", "/", "\\", "*", "|"} {
			fileName = strings.Replace(fileName, str, "_", -1)
		}
	}
	client := utils.Client.Get().(*http.Client)
	fmt.Println("正在下载 " + dirPath + "/" + fileName)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Referer", "https://www.asmr.one/")
	req.Header.Set("User-Agent", "PostmanRuntime/7.29.0")
	resp, err := client.Do(req)
	utils.Client.Put(client)
	if err != nil {
		fmt.Print(err)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	all, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err)
		return
	}
	err = os.WriteFile(dirPath+"/"+fileName, all, 0644)
	if err != nil {
		return
	}
}

func EnsureDir(tracks []interface{}, basePath string) {
	path := basePath
	_ = os.MkdirAll(path, os.ModePerm)
	for _, track := range tracks {
		if track.(map[string]interface{})["type"].(string) != "folder" {
			DownloadFile(track.(map[string]interface{})["mediaDownloadUrl"].(string), path, track.(map[string]interface{})["title"].(string))
		} else {
			_ = os.MkdirAll(path+"/"+track.(map[string]interface{})["title"].(string), os.ModePerm)
			EnsureDir(track.(map[string]interface{})["children"].([]interface{}), path+"/"+track.(map[string]interface{})["title"].(string))
		}
	}
}

func (OAuth2 *OAuth2) Download(id string) {
	id = strings.Replace(id, "RJ", "", 1)
	fmt.Println("作品 RJ 号: " + id)
	tracks, err := OAuth2.GetVoiceTracks(id)
	if err != nil {
		fmt.Print(err)
		return
	}
	EnsureDir(tracks, "RJ"+id)
}
