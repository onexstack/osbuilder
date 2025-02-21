package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/duke-git/lancet/v2/strutil"
	//"github.com/gobuffalo/flect"
	"github.com/fatih/color"
	"github.com/gobuffalo/flect"
	"github.com/rakyll/statik/fs"
)

func UpperCamelCase(kind string) string {
	camelCase := strutil.CamelCase(kind)
	return strutil.UpperFirst(camelCase)
}

func LowerCamelCase(kind string) string {
	camelCase := strutil.CamelCase(kind)
	return strutil.LowerFirst(camelCase)
}

// EnsureDir 根据路径创建目录
func EnsureDir(path string) error {
	// 如果是目录（路径以 / 结尾）
	if strings.HasSuffix(path, string(os.PathSeparator)) {
		return os.MkdirAll(path, os.ModePerm)
	}

	// 如果是文件路径
	dir := filepath.Dir(path) // 提取文件路径中的目录部分
	return os.MkdirAll(dir, os.ModePerm)
}

// 判断文件或目录是否存在
func IsFileExists(path string) bool {
	_, err := os.Stat(path) // 获取文件信息
	if os.IsNotExist(err) { // 判断是否为文件不存在的错误
		return false
	}
	return err == nil // 如果没有错误则文件存在
}

func join(elem ...string) string {
	return filepath.Join(elem...)
}

func ComponentNameFromBinaryName(binaryName string) string {
	splited := strings.Split(binaryName, "-")
	componentName := binaryName
	if len(splited) == 2 {
		componentName = splited[1]
	}
	return componentName
}

func IsOneXStackProject() bool {
	// 定义要检查的文件名
	filename := ".onexstack"

	// 获取文件信息，用于检查文件是否存在
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return false
		}
		panic(err)
	}

	return true
}

func PrintGenerateFile(file string) {
	fmt.Printf("%s: %s\n", color.GreenString("Generated file"), file)
}

func PrintModifiedFile(file string) {
	fmt.Printf("%s: %s\n", color.YellowString("Modified file"), file)
}

// CronJob
func Kind() func(string) string {
	return func(input string) string {
		if len(input) == 0 {
			return input
		}
		return UpperCamelCase(input)
	}
}

// CronJobs
func Kinds() func(string) string {
	return func(input string) string {
		if len(input) == 0 {
			return input
		}
		return flect.Pluralize(UpperCamelCase(input))
	}
}

// V1
func Capitalize() func(string) string {
	return func(input string) string {
		if len(input) == 0 {
			return input
		}
		return strings.ToUpper(input[:1]) + input[1:]
	}
}

func SingularLower() func(string) string {
	return func(input string) string {
		if len(input) == 0 {
			return input
		}
		return strings.ToLower(UpperCamelCase(input))
	}
}

func SingularLowers() func(string) string {
	return func(input string) string {
		if len(input) == 0 {
			return input
		}
		return strings.ToLower(flect.Pluralize(UpperCamelCase(input)))
	}
}
