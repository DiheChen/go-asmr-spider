package utils

import (
	"bufio"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
)

var (
	defaultUA                    = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36"
	ErrUnsupportedMultiThreading = errors.New("unsupported multi-threading")
)

type BlockMetaData struct {
	BeginOffset    int64
	EndOffset      int64
	DownloadedSize int64
}

type MultiThreadDownloader struct {
	Url         string
	SavePath    string
	FileName    string
	FullPath    string
	Client      *http.Client
	Headers     map[string]string
	Blocks      []*BlockMetaData
	ThreadCount int
}

func NewDownloader(url string, path string, name string, threadCount int, headers map[string]string) *MultiThreadDownloader {
	return &MultiThreadDownloader{
		Url:         url,
		SavePath:    path,
		FileName:    name,
		FullPath:    path + "/" + name,
		Client:      Client.Get().(*http.Client),
		Headers:     headers,
		Blocks:      nil,
		ThreadCount: threadCount,
	}
}

func (m *MultiThreadDownloader) Download() error {
	if m.ThreadCount < 2 {
		err := SingleThreadDownload(m.Url, m.FullPath, m.Headers)
		return err
	}
	if err := m.initDownload(); err != nil {
		if err == ErrUnsupportedMultiThreading {
			return nil
		}
		return err
	}
	wg := sync.WaitGroup{}
	wg.Add(len(m.Blocks))
	var lastErr error
	for i := range m.Blocks {
		go func(b *BlockMetaData) {
			defer wg.Done()
			if err := m.downloadBlocks(b); err != nil {
				lastErr = err
			}
		}(m.Blocks[i])
	}
	wg.Wait()
	return lastErr
}

func (m *MultiThreadDownloader) initDownload() error {
	var contentLength int64
	copyStream := func(s io.ReadCloser) error {
		file, err := os.OpenFile(m.FullPath, os.O_WRONLY|os.O_CREATE, 0o666)
		if err != nil {
			return err
		}
		defer func() { _ = file.Close() }()
		if _, err = file.ReadFrom(s); err != nil {
			return err
		}
		return ErrUnsupportedMultiThreading
	}
	req, err := http.NewRequest("GET", m.Url, nil)
	if err != nil {
		return err
	}

	for k, v := range m.Headers {
		req.Header.Set(k, v)
	}
	if _, ok := m.Headers["User-Agent"]; !ok {
		req.Header["User-Agent"] = []string{defaultUA}
	}
	req.Header.Set("range", "bytes=0-")
	resp, err := m.Client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("response status unsuccessful: " + strconv.FormatInt(int64(resp.StatusCode), 10))
	}
	if resp.StatusCode == 200 {
		return copyStream(resp.Body)
	}
	if resp.StatusCode == 206 {
		contentLength = resp.ContentLength
		blockSize := func() int64 {
			if contentLength > 1024*1024 {
				return (contentLength / int64(m.ThreadCount)) - 10
			}
			return contentLength
		}()
		if blockSize == contentLength {
			return copyStream(resp.Body)
		}
		var tmp int64
		for tmp+blockSize < contentLength {
			m.Blocks = append(m.Blocks, &BlockMetaData{
				BeginOffset: tmp,
				EndOffset:   tmp + blockSize - 1,
			})
			tmp += blockSize
		}
		m.Blocks = append(m.Blocks, &BlockMetaData{
			BeginOffset: tmp,
			EndOffset:   contentLength - 1,
		})
		return nil
	}
	return errors.New("unknown status code")
}

func (m *MultiThreadDownloader) downloadBlocks(block *BlockMetaData) error {
	req, _ := http.NewRequest("GET", m.Url, nil)
	file, err := os.OpenFile(m.FullPath, os.O_WRONLY|os.O_CREATE, 0o666)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	_, _ = file.Seek(block.BeginOffset, io.SeekStart)
	writer := bufio.NewWriter(file)
	defer func() { _ = writer.Flush() }()

	for k, v := range m.Headers {
		req.Header.Set(k, v)
	}
	if _, ok := m.Headers["User-Agent"]; !ok {
		req.Header["User-Agent"] = []string{defaultUA}
	}
	req.Header.Set("range", "bytes="+strconv.FormatInt(block.BeginOffset, 10)+"-"+strconv.FormatInt(block.EndOffset, 10))
	resp, err := m.Client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("response status unsuccessful: " + strconv.FormatInt(int64(resp.StatusCode), 10))
	}
	buffer := make([]byte, 1024)
	i, err := resp.Body.Read(buffer)
	for {
		if err != nil && err != io.EOF {
			return err
		}
		i64 := int64(len(buffer[:i]))
		needSize := block.EndOffset + 1 - block.BeginOffset
		if i64 > needSize {
			i64 = needSize
			err = io.EOF
		}
		_, e := writer.Write(buffer[:i64])
		if e != nil {
			return e
		}
		block.BeginOffset += i64
		block.DownloadedSize += i64
		if err == io.EOF || block.BeginOffset > block.EndOffset {
			break
		}
		i, err = resp.Body.Read(buffer)
	}
	return nil
}

func SingleThreadDownload(url, path string, headers map[string]string) error {
	client := Client.Get().(*http.Client)
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0o666)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if _, ok := headers["User-Agent"]; !ok {
		req.Header["User-Agent"] = []string{defaultUA}
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	_, err = file.ReadFrom(resp.Body)
	if err != nil {
		return err
	}
	return nil
}
