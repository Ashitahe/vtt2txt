package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func parseVTTFile(filePath string) ([]map[string]string, error) {
	var dialogues []map[string]string
	var currentDialogue map[string]string

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	timePattern := regexp.MustCompile(`(\d{2}:\d{2}:\d{2}\.\d{3}) --> (\d{2}:\d{2}:\d{2}\.\d{3})`)
	musicSymbolPattern := regexp.MustCompile(`â™ª`)

	var buffer string
	defaultPattern := regexp.MustCompile(`<Default><b>((?s).*?)</b></Default>`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和WEBVTT标记
		if line == "" || line == "WEBVTT" || strings.HasPrefix(line, "STYLE") {
			continue
		}

		// 匹配时间戳行
		if timeMatch := timePattern.FindStringSubmatch(line); timeMatch != nil {
			// 处理上一个对话的缓冲区内容
			if buffer != "" && currentDialogue != nil {
				if match := defaultPattern.FindStringSubmatch(buffer); match != nil {
					text := match[1]
					text = strings.ReplaceAll(text, "\n", " ")
					text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
					text = musicSymbolPattern.ReplaceAllString(text, "♪")
					text = strings.TrimSpace(text)
					if text != "" {
						if currentDialogue["text"] != "" {
							currentDialogue["text"] += " "
						}
						currentDialogue["text"] += text
					}
				}
			}
			
			if currentDialogue != nil {
				dialogues = append(dialogues, currentDialogue)
			}
			currentDialogue = map[string]string{
				"start_time": timeMatch[1],
				"end_time":   timeMatch[2],
				"text":       "",
			}
			buffer = ""
			continue
		}

		// 累积所有非时间戳行
		if buffer != "" {
			buffer += " "
		}
		buffer += line
	}
	
	// 处理最后一个对话
	if buffer != "" && currentDialogue != nil {
		if match := defaultPattern.FindStringSubmatch(buffer); match != nil {
			text := match[1]
			text = strings.ReplaceAll(text, "\n", " ")
			text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
			text = musicSymbolPattern.ReplaceAllString(text, "♪")
			text = strings.TrimSpace(text)
			if text != "" {
				if currentDialogue["text"] != "" {
					currentDialogue["text"] += " "
				}
				currentDialogue["text"] += text
			}
		}
	}
	
	if currentDialogue != nil {
		dialogues = append(dialogues, currentDialogue)
	}

	return dialogues, nil
}

func formatDialogue(dialogues []map[string]string) string {
	var formattedText []string
	for _, d := range dialogues {
		if text, ok := d["text"]; ok && text != "" {
			formattedText = append(formattedText, fmt.Sprintf("[%s] %s", d["start_time"], text))
		}
	}
	return strings.Join(formattedText, "\n")
}

func main() {
	// 检查命令行参数
	if len(os.Args) != 2 {
		fmt.Println("使用方法: go run main.go <输入文件路径>")
		return
	}

	inputFile := os.Args[1]
	outputFile := strings.TrimSuffix(inputFile, filepath.Ext(inputFile)) + "_formatted.txt"

	// 处理字幕文件
	dialogues, err := parseVTTFile(inputFile)
	if err != nil {
		fmt.Printf("处理文件出错: %v\n", err)
		return
	}

	// 输出格式化的对话
	formattedOutput := formatDialogue(dialogues)

	// 保存到新文件
	err = os.WriteFile(outputFile, []byte(formattedOutput), 0644)
	if err != nil {
		fmt.Printf("写入文件出错: %v\n", err)
		return
	}

	fmt.Printf("处理完成! 结果已保存到 %s\n", outputFile)
}