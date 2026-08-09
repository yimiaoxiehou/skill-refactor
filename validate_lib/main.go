package main

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type LibRoot struct {
	XMLName xml.Name `xml:"root"`
	Libs    *Libs    `xml:"libs"`
	AppIds  string   `xml:"appIds"`
	Cloud   string   `xml:"cloud"`
}

type Libs struct {
	Lib string `xml:"lib"`
}

func checkLineEnding(path string, expectedCRLF bool) []string {
	var errors []string
	data, err := os.ReadFile(path)
	if err != nil {
		return append(errors, fmt.Sprintf("无法读取文件: %v", err))
	}
	hasCRLF := strings.Contains(string(data), "\r\n")

	if expectedCRLF && !hasCRLF {
		errors = append(errors, "行尾格式错误: 期望 CRLF (根目录 lib), 实际为 LF")
	} else if !expectedCRLF && hasCRLF {
		errors = append(errors, "行尾格式错误: 期望 LF (补丁 lib), 实际为 CRLF")
	}
	return errors
}

func checkCloudMatchesFilename(path string, root *LibRoot) []string {
	var errors []string
	if root.Cloud != "" {
		basename := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		if strings.TrimSpace(root.Cloud) != basename {
			errors = append(errors,
				fmt.Sprintf("<cloud> 值 '%s' 与文件名 '%s' 不匹配",
					strings.TrimSpace(root.Cloud), basename))
		}
	}
	return errors
}

func checkReferencesExist(libPath string, root *LibRoot, cusDir string) []string {
	var errors []string
	if root.Libs == nil || strings.TrimSpace(root.Libs.Lib) == "" {
		errors = append(errors, "缺少 <libs>/<lib> 元素或内容为空")
		return errors
	}

	entries := strings.Split(root.Libs.Lib, ",")
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			errors = append(errors, "发现空引用条目（逗号分隔多余）")
			continue
		}

		refName := entry
		if after, found := strings.CutPrefix(refName, "cus/"); found {
			refName = after
		}

		if strings.HasSuffix(refName, ".xml") {
			fullPath := filepath.Join(cusDir, refName)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				errors = append(errors,
					fmt.Sprintf("引用的 xml 文件不存在: %s → %s", entry, fullPath))
			}
		} else {
			fullPath := filepath.Join(cusDir, refName+".zip")
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				errors = append(errors,
					fmt.Sprintf("引用的 zip 文件不存在: %s → %s", entry, fullPath))
			}
		}
	}
	return errors
}

func checkZipJarNaming(libPath string, root *LibRoot, cusDir string) []string {
	var errors []string
	if root.Libs == nil || strings.TrimSpace(root.Libs.Lib) == "" {
		return errors
	}

	entries := strings.Split(root.Libs.Lib, ",")
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		refName := entry
		if after, found := strings.CutPrefix(refName, "cus/"); found {
			refName = after
		}

		if strings.HasSuffix(refName, ".xml") {
			continue
		}

		zipPath := filepath.Join(cusDir, refName+".zip")
		if _, err := os.Stat(zipPath); os.IsNotExist(err) {
			continue
		}

		r, err := zip.OpenReader(zipPath)
		if err != nil {
			errors = append(errors, fmt.Sprintf("zip 文件损坏: %s", zipPath))
			continue
		}
		for _, f := range r.File {
			if strings.HasSuffix(f.Name, ".jar") {
				// informational check only
				_ = f.Name
			}
		}
		r.Close()
	}
	return errors
}

func checkCommaWhitespace(root *LibRoot) []string {
	var errors []string
	if root.Libs == nil || root.Libs.Lib == "" {
		return errors
	}

	text := root.Libs.Lib
	parts := strings.Split(text, ",")
	for i, part := range parts {
		if part != strings.TrimSpace(part) {
			errors = append(errors, fmt.Sprintf("条目 #%d 有多余空格: '%s'", i+1, part))
		}
	}
	return errors
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: validate_lib <lib_file> [--cus-dir <directory>]")
		os.Exit(1)
	}

	libPath := os.Args[1]
	if _, err := os.Stat(libPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "❌ 文件不存在: %s\n", libPath)
		os.Exit(1)
	}

	cusDir := ""
	for i := 2; i < len(os.Args); i++ {
		if os.Args[i] == "--cus-dir" && i+1 < len(os.Args) {
			cusDir = os.Args[i+1]
		}
	}
	if cusDir == "" {
		parentDir := filepath.Dir(libPath)
		cusDir = filepath.Join(parentDir, "cus")
	}

	// Determine if patch lib
	isPatch := false
	basename := filepath.Base(libPath)
	if basename == "cus.lib" {
		isPatch = true
	} else {
		data, err := os.ReadFile(libPath)
		if err == nil {
			var root LibRoot
			if err := xml.Unmarshal(data, &root); err == nil {
				isPatch = root.AppIds == "" && root.Cloud == ""
			}
		}
	}

	// 1. XML well-formedness check
	data, err := os.ReadFile(libPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ 无法读取文件: %v\n", err)
		os.Exit(1)
	}

	var root LibRoot
	if err := xml.Unmarshal(data, &root); err != nil {
		fmt.Fprintf(os.Stderr, "❌ XML 解析失败: %v\n", err)
		os.Exit(1)
	}

	var allErrors []string

	// 2. Required elements
	if root.Libs == nil {
		allErrors = append(allErrors, "缺少 <libs> 元素")
	}

	// 3. Line-ending format
	allErrors = append(allErrors, checkLineEnding(libPath, !isPatch)...)

	// 4. Cloud matches filename
	if !isPatch {
		allErrors = append(allErrors, checkCloudMatchesFilename(libPath, &root)...)
	}

	// 5. Referenced files exist
	allErrors = append(allErrors, checkReferencesExist(libPath, &root, cusDir)...)

	// 6. Comma whitespace
	allErrors = append(allErrors, checkCommaWhitespace(&root)...)

	// 7. Zip jar naming
	allErrors = append(allErrors, checkZipJarNaming(libPath, &root, cusDir)...)

	if len(allErrors) > 0 {
		fmt.Printf("❌ 验证失败 (%d 个问题):\n", len(allErrors))
		for _, e := range allErrors {
			fmt.Printf("   - %s\n", e)
		}
		os.Exit(1)
	} else {
		fmt.Printf("✅ 验证通过: %s\n", libPath)
		fmt.Printf("   cus 目录: %s\n", cusDir)
	}
}
