package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/buger/jsonparser"
	"github.com/projectdiscovery/goflags"
)

type Options struct {
	maxConcurrent  int
	tabs           int
	crawlergoPath  string
	chromiumPath   string
	targetFile     string
	rootdomainFile string
	paramFile      string
}

var (
	paramKeys   []string
	options     Options
	semaphore   chan struct{}
	rootdomains []string
	wg          sync.WaitGroup
)

func readFlags() {
	flagSet := goflags.NewFlagSet()
	flagSet.CreateGroup("input", "Input",
		// mac下: /Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome
		flagSet.StringVarP(&options.crawlergoPath, "crawlergo-path", "cgo", "/usr/lib/golang/bin/crawlergo", "crawlergo path"),
		flagSet.StringVarP(&options.chromiumPath, "chromium", "c", "/usr/bin/chromium-browser", "chromium-browser path"),
		flagSet.StringVarP(&options.targetFile, "target-file", "tf", "uro.txt", "input file which contains web url to craw"),
		flagSet.StringVarP(&options.rootdomainFile, "root-file", "rf", "rootdomains.txt", "to filter url"),
		flagSet.IntVarP(&options.maxConcurrent, "thread", "t", 10, "how may crawlergo at same time"),
		flagSet.IntVar(&options.tabs, "tabs", 10, "crawlergo concurrent tabs"),
	)
	flagSet.CreateGroup("Output", "Output",
		flagSet.StringVarP(&options.paramFile, "param-file", "pf", "params.txt", "output filename.txt which contains params of post data"),
	)

	_ = flagSet.Parse(os.Args[1:]...)
}

func runCrawler(targets <-chan string, wg *sync.WaitGroup, bar *ProcessBar) {

	for target := range targets {

		// 执行crawlergo的命令
		cmd := exec.Command(
			options.crawlergoPath,
			"-c", options.chromiumPath,
			"-t", strconv.Itoa(options.tabs),
			"-o", "json",
			"--robots-path",
			//"--fuzz-path",
			target)

		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Println("Error running crawler:", err)
			bar.PrintAccumulator(1, 0, 1, "crawlergo Program")
			return
		}

		// @TODO 这里post的data也应该取出来生成params.txt文件, 但是需要不能all_req_list, 只能req_list
		resultJson := strings.Split(string(out), "--[Mission Complete]--")[1]
		result := make(map[string]interface{})
		json.Unmarshal([]byte(resultJson), &result)
		reqList := result["all_req_list"].([]interface{})

		urls := make(map[string]map[string]bool)
		for _, res := range reqList {
			resBytes, _ := json.Marshal(res)

			m, _ := jsonparser.GetString(resBytes, "method")
			if m == "POST" {
				// 拿参数, 添加入param.txt
				d, _ := jsonparser.GetString(resBytes, "data")
				// json类型
				if isValidJSON(d) {
					keys := getAllKeys([]byte(d))
					for _, key := range keys {
						paramKeys = append(paramKeys, key)
					}
				} else if strings.Contains(d, "&") {
					// raw类型
					params := strings.Split(d, "&")
					for _, param := range params {
						paramKey := strings.Split(param, "=")[0]
						paramKeys = append(paramKeys, paramKey)
					}
				}
			}

			// url
			u, _ := jsonparser.GetString(resBytes, "url")
			parsedUrl, _ := url.Parse(u)
			domain := parsedUrl.Hostname() // 这里不带port
			domain = apex(domain)          // 这里暂时返回domain/ip

			if contains(rootdomains, domain) { // 相应的策略为根域名在domain中即可string.Contains
				if urls[domain] == nil {
					urls[domain] = make(map[string]bool)
				}
				urls[domain][u] = true
			}
		}

		directoryPath := "./crawler_output"
		for domain, urlSet := range urls {
			filename := filepath.Join(directoryPath, strings.ReplaceAll(domain, ".", "_"))
			content := strings.Join(getKeys(urlSet), "\n") + "\n"
			if err := appendToFile(filename, content); err != nil {
				log.Println("Error appending to file:", err)
			}
		}
		bar.PrintAccumulator(1, 1, 0, "crawlergo Program")

		wg.Done()
	}
}

func isValidJSON(s string) bool {
	var result interface{}
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return false
	}
	return true
}

// 解析 JSON 并获取所有键名
func getAllKeys(data []byte) []string {
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil
	}
	return getKeysFromMap(jsonData)
}

// 递归获取 map[string]interface{} 中的所有键名
func getKeysFromMap(m map[string]interface{}) []string {
	var keys []string

	for key, value := range m {
		keys = append(keys, key)

		// 如果值是 map[string]interface{} 类型，则递归获取键名
		if subMap, ok := value.(map[string]interface{}); ok {
			subKeys := getKeysFromMap(subMap)
			keys = append(keys, subKeys...)
		}
	}

	return keys
}

func main() {

	readFlags()

	rootdomains = readFile(options.rootdomainFile)

	if err := os.MkdirAll("./crawler_output", 0755); err != nil {
		log.Fatal(err)
	}

	bar := NewProcessBar(int64(getFileLine(options.targetFile)))

	// 开启协程
	targets := make(chan string)
	for w := 1; w <= options.maxConcurrent; w++ {
		go runCrawler(targets, &wg, bar)
	}

	// 传递数据
	file, err := os.Open(options.targetFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		wg.Add(1)
		targets <- scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	wg.Wait()

	// 最后写文件param
	file, _ = os.OpenFile(options.paramFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	_, _ = file.WriteString(strings.Join(removeDuplicates(paramKeys), "\n"))

	println()

}
