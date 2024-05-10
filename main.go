package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const reposLocalDir = "./repos"

const mdOutputDir = "./outputs"

var markdownExtension = ".md"

// create a dir to clone the repositorys named by the last part of the URL

func main() {

	// 检查是否提供了仓库 URL 参数
	if len(os.Args) < 2 {
		fmt.Println("Usage: gh2md <repository-url>")
		os.Exit(1)
	}

	repoURL := os.Args[1]

	// 确定本地存储仓库的路径
	repoName := strings.TrimSuffix(filepath.Base(repoURL), ".git")
	localRepoPath := filepath.Join(reposLocalDir, repoName)

	// clone
	err := cloneRepository(repoURL, localRepoPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Traverse the cloned repository files and generate the markdown
	err = traverseFiles(localRepoPath, mdOutputDir)
	if err != nil {
		fmt.Println("Error traversing files:", err)
		os.Exit(1)
	}

	fmt.Printf("Markdown file created successfully at %s\n", mdOutputDir)

}

func cloneRepository(repoURL string, localRepoPath string) error {
	// 检查目录是否存在，如果存在则先删除(todo: 优化)
	if _, err := os.Stat(localRepoPath); os.IsNotExist(err) {
		err := os.MkdirAll(localRepoPath, 0755)
		if err != nil {
			return err
		}
	} else {
		// 如果目录存在，先删除
		err := os.RemoveAll(localRepoPath)
		if err != nil {
			return err
		}
		err = os.MkdirAll(localRepoPath, 0755)
		if err != nil {
			return err
		}
	}

	// print a log message
	fmt.Println("Cloning repository: ", repoURL)

	cmd := exec.Command("git", "clone", "--depth", "1", repoURL, localRepoPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to clone repository: %s", output)
	}
	fmt.Println("Repository cloned successfully.")
	return nil
}

func traverseFiles(repoPath, outputsDir string) error {

	// 通过repoPath获取repo名称，然后在outputsDir下创建一个repo名称的md文件
	repoName := filepath.Base(repoPath)
	outputFile := filepath.Join(outputsDir, repoName+".md")

	return filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// 跳过根目录
			if path == repoPath {
				return nil
			}
			return nil
		}

		relPath, _ := filepath.Rel(repoPath, path)
		fileExt := strings.ToLower(filepath.Ext(relPath))

		// 打开文件以读取其内容
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// 根据文件类型处理
		if fileExt == markdownExtension {
			// 直接读取 Markdown 文件内容
			content, err := ioutil.ReadAll(file)
			if err != nil {
				return err
			}
			return appendToFile(outputFile, string(content))
		} else {
			// 对于其他文件，获取其内容并确定语言
			language := fileExt
			if language == "" {
				language = "text"
			}

			// 读取内容并写入 Markdown 文件
			scanner := bufio.NewScanner(file)
			var codeBlock bytes.Buffer
			codeBlock.WriteString("```" + language + "\n")
			for scanner.Scan() {
				codeBlock.WriteString(scanner.Text() + "\n")
			}
			codeBlock.WriteString("```\n")

			// 添加文件路径
			codeBlock.WriteString("\n<!-- " + relPath + " -->\n")

			return appendToFile(outputFile, codeBlock.String())
		}
	})
}

func appendToFile(outputFile, content string) error {

	// 确保输出文件的父目录存在
	dir := filepath.Dir(outputFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.WriteString(file, content)
	return err
}
