package helper

import (
	"fmt"
	"io"
	iofs "io/fs"
	"io/ioutil"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/strutil"
	"github.com/fatih/color"
	"github.com/gobuffalo/flect"
	"github.com/rakyll/statik/fs"
	"k8s.io/apimachinery/pkg/util/sets"
	"resty.dev/v3"

	_ "github.com/onexstack/osbuilder/internal/osbuilder/statik"
)

var underscoreReplacer = strings.NewReplacer(".", "_", "-", "_")

// // FileSystem wraps a base path for managing file-related operations.
type FileSystem struct {
	BasePath string
}

// NewFileSystem creates a new instance of FileSystem.
func NewFileSystem(basePath string) FileSystem {
	return FileSystem{BasePath: basePath}
}

// GetFile reads and retrieves the content of a file relative to the base path.
func (f *FileSystem) Content(relPath string) string {
	return ReadFile(filepath.Join(f.BasePath, relPath))
}

// GetTemplate retrieves the content of the keep template file.
func (f *FileSystem) GetKeep() string {
	return ReadFile("/keep.tpl")
}

func (f *FileSystem) CopyFiles(src, dst string) error {
	src = filepath.Join(f.BasePath, src)
	// 初始化 statik 文件系统
	statikFS, err := fs.New()
	if err != nil {
		return fmt.Errorf("failed to create statik fs: %v", err)
	}

	// 确保源路径以 / 开头
	if !strings.HasPrefix(src, "/") {
		src = "/" + src
	}

	// 创建目标根目录
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	// 遍历源目录中的所有文件
	return fs.Walk(statikFS, src, func(path string, info iofs.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking path %s: %v", path, err)
		}

		// 计算相对路径
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}
		if relPath == "." {
			return nil
		}

		// 构建目标路径
		dstPath := filepath.Join(dst, relPath)

		// 如果是目录，创建对应的目标目录
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// 复制文件
		return copyFileFromStatik(statikFS, path, dstPath, info.Mode())
	})
}

// ReadFile reads the content of a file from the statik file system.
func ReadFile(path string) string {
	red := color.New(color.FgRed)
	defer func() {
		if r := recover(); r != nil {
			red.Printf("Panic occurred. File `%s`, Error: %v\n", path, r)
		}
	}()
	statikFS, err := fs.New()
	if err != nil {
		red.Printf("Failed to initialize the file system. File `%s`, Error: %v\n", path, err)
		return ""
	}
	file, err := statikFS.Open(path)
	if err != nil {
		red.Printf("Failed to open file `%s`, Error: %v\n", path, err)
		return ""
	}
	contentBytes, err := ioutil.ReadAll(file)
	if err != nil {
		red.Printf("Failed to read file `%s`, Error: %v\n", path, err)
		return ""
	}
	return string(contentBytes)
}

// ToUpperCamelCase converts a string to upper CamelCase.
func ToUpperCamelCase(input string) string {
	return strutil.UpperFirst(strutil.CamelCase(input))
}

// ToLowerCamelCase converts a string to lower CamelCase.
func ToLowerCamelCase(input string) string {
	return strutil.LowerFirst(strutil.CamelCase(input))
}

// EnsureDirectory ensures that the directory for the given path exists.
func EnsureDirectory(path string) error {
	if strings.HasSuffix(path, string(os.PathSeparator)) {
		return os.MkdirAll(path, os.ModePerm)
	}
	dir := filepath.Dir(path) // Get the directory part of the path
	return os.MkdirAll(dir, os.ModePerm)
}

// FileExists checks whether a file or directory exists at the specified path.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// JoinPaths joins multiple elements into a single valid file system path.
func JoinPaths(elements ...string) string {
	return filepath.Join(elements...)
}

// GetComponentName extracts the component name from a binary name.
func GetComponentName(binaryName string) string {
	parts := strings.Split(binaryName, "-")
	if len(parts) == 2 {
		return parts[1]
	}
	return binaryName
}

// IsOneXStackProject checks if the current directory is a OneXStack project.
func IsOneXStackProject(file string) bool {
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			return false
		}
		panic(err)
	}
	return true
}

// PrintGeneratedFile logs that a file has been generated.
func PrintGeneratedFile(file string) {
	fmt.Printf("%s: %s\n", color.GreenString("Generated"), RelativePath(file))
}

// PrintModifiedFile logs that a file has been modified.
func PrintModifiedFile(file string) {
	fmt.Printf("%s: %s\n", color.YellowString("Modified"), RelativePath(file))
}

func RelativePath(path string) string {
	// 获取当前工作目录
	currentDir, err := filepath.Abs(".")
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return ""
	}

	// 使用 filepath.Rel 计算目标路径相对于当前目录的相对路径
	relativePath, err := filepath.Rel(currentDir, path)
	if err != nil {
		fmt.Println("Error calculating relative path:", err)
		return ""
	}
	return relativePath
}

// ToUnderscore converts dots (.) and hyphens (-) to underscores (_)
func ToUnderscore() func(string) string {
	return func(s string) string {
		return underscoreReplacer.Replace(s)
	}
}

// Kind returns a function to convert strings to upper CamelCase.
func Kind() func(string) string {
	return func(input string) string {
		if len(input) == 0 {
			return input
		}
		return ToUpperCamelCase(input)
	}
}

// Kinds returns a function to convert strings to plural upper CamelCase.
func Kinds() func(string) string {
	return func(input string) string {
		if len(input) == 0 {
			return input
		}
		return flect.Pluralize(ToUpperCamelCase(input))
	}
}

// Capitalize returns a function to capitalize the first character of a string.
func Capitalize() func(string) string {
	return func(input string) string {
		if len(input) == 0 {
			return input
		}
		return strings.ToUpper(input[:1]) + input[1:]
	}
}

// SingularLower returns a function to convert strings to lower CamelCase.
func SingularLower() func(string) string {
	return func(input string) string {
		if len(input) == 0 {
			return input
		}
		return strings.ToLower(ToUpperCamelCase(input))
	}
}

// SingularLowers returns a function to convert strings to plural lower CamelCase.
func SingularLowers() func(string) string {
	return func(input string) string {
		if len(input) == 0 {
			return input
		}
		return strings.ToLower(flect.Pluralize(ToUpperCamelCase(input)))
	}
}

func CurrentYear() func() int {
	return func() int {
		return time.Now().Year()
	}
}

func Tree(path string, dir string) {
	_ = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err == nil && info != nil && !info.IsDir() {
			fmt.Printf("%s %s (%v bytes)\n", color.GreenString("CREATED"), strings.Replace(path, dir+"/", "", -1), info.Size())
		}
		return nil
	})
}

// copyFileFromStatik 从 statik 文件系统复制单个文件到目标位置
func copyFileFromStatik(statikFS http.FileSystem, src, dst string, mode os.FileMode) error {
	// 打开源文件
	srcFile, err := statikFS.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %v", src, err)
	}
	defer srcFile.Close()

	// 创建目标文件
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %v", dst, err)
	}
	defer dstFile.Close()

	// 使用缓冲区复制文件内容
	buf := make([]byte, 32*1024) // 32KB 缓冲区
	if _, err = io.CopyBuffer(dstFile, srcFile, buf); err != nil {
		return fmt.Errorf("failed to copy file contents from %s to %s: %v", src, dst, err)
	}

	return nil
}

func Available[T string](s sets.Set[string]) string {
	return strings.Join(sets.List(s), ",")
}

func Merge[K comparable, V any](m1, m2 map[K]V) map[K]V {
	out := maps.Clone(m1) // 复制 m1
	maps.Copy(out, m2)    // 将 m2 合入，冲突时以 m2 为准
	return out
}

// CountRequest 请求结构体
type CountRequest struct {
	Type   string `json:"type"`
	Status string `json:"status"`
}

// RecordOSBuilderUsage 记录 OSBuilder 工具使用统计.
func RecordOSBuilderUsage(apiType string, err error) {
	status := "success"
	if err != nil {
		status = "fail"
	}

	// 构建请求体
	request := CountRequest{Type: apiType, Status: status}

	// 发送请求
	client := resty.New().SetTimeout(2 * time.Second)
	_, _ = client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(request).
		Post("http://43.139.4.14:33331/count")
}
