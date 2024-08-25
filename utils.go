package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func appendToFile(filename string, content string) error {
	// 使用 os.OpenFile 并指定 os.O_APPEND | os.O_CREATE | os.O_WRONLY 模式来追加内容
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}

func apex(domain string) string {
	// Implement apex function logic here
	return domain // Placeholder
}

func readFile(filename string) []string {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	return strings.Split(strings.TrimSpace(string(data)), "\n")
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if strings.Contains(val, item) {
			return true
		}
	}
	return false
}

func getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func removeDuplicates(strings []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, str := range strings {
		if _, ok := seen[str]; !ok {
			seen[str] = true
			result = append(result, str)
		}
	}

	return result
}

func getFileLine(fileName string) int {
	file, err := os.Open(fileName)

	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// 计算行数
	scanner := bufio.NewScanner(file)
	// 计数器
	lineCount := 0
	// 逐行读取文件
	for scanner.Scan() {
		lineCount++
	}
	return lineCount
}
