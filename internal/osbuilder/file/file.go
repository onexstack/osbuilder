// filemanager/filemanager.go
package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/fatih/color"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/afero"
	iofs "io/fs"
	"net/http"
	"strings"
)

// FileOperation 表示文件操作类型
type FileOperation string

const (
	Created FileOperation = "CREATED" // 创建操作
	Updated FileOperation = "UPDATED" // 更新操作
)

// FileInfo 存储文件信息和操作类型
type FileInfo struct {
	Path string
	//Operation FileOperation
}

// FileManager 文件管理器
type FileManager struct {
	mu      sync.RWMutex
	FS      afero.Fs
	workDir string
	force   bool

	cache map[string]FileInfo
}

// NewFileManager 创建新的文件管理器实例
func NewFileManager(workDir string, force bool) *FileManager {
	return &FileManager{
		FS:      afero.NewOsFs(),
		workDir: workDir,
		force:   force,
		cache:   make(map[string]FileInfo),
	}
}

// CreateOrUpdateFile 创建或更新文件
func (fm *FileManager) WriteFile(path string, content []byte) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if exists, _ := afero.Exists(fm.FS, path); exists && !fm.force {
		return nil
	}

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := fm.FS.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 检查文件是否存在
	_, err := fm.FS.Stat(path)
	/*
		//operation := Created
		if err == nil {
			operation = Updated
		}
	*/

	// 写入文件
	err = afero.WriteFile(fm.FS, path, content, 0644)
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	/*
		fm.cache[path] = FileInfo{
			Path:      path,
			Operation: operation,
		}
	*/

	fm.Print(Created, path)
	return nil
}

/*
func (fm *FileManager) PrintAll() {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	fmt.Println("文件操作记录:")
	for _, info := range fm.cache {
		fm.Print(info.Path)
	}
}
*/

func (fm *FileManager) Print(operation FileOperation, path string) {
	//info, _ := os.Stat(path)
	//fmt.Printf("%s %s (%v bytes)\n", color.GreenString("CREATED"), strings.Replace(path, filepath.Dir(fm.workDir)+"/", "", -1), info.Size())
	fmt.Printf("%s %s\n", color.GreenString(string(operation)), strings.Replace(path, filepath.Dir(fm.workDir)+"/", "", -1))
}

// GetFileCount 获取缓存中的文件数量
func (fm *FileManager) GetFileCount() int {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return len(fm.cache)
}

// ClearCache 清空缓存
func (fm *FileManager) ClearCache() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.cache = make(map[string]FileInfo)
}

func (fm *FileManager) CopyFiles(src, dst string) error {
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
	if err := fm.FS.MkdirAll(dst, 0755); err != nil {
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
		return fm.copyFileFromStatik(statikFS, path, dstPath, info.Mode())
	})
}

// copyFileFromStatik 从 statik 文件系统复制单个文件到目标位置
func (fm *FileManager) copyFileFromStatik(statikFS http.FileSystem, src, dst string, mode os.FileMode) error {
	if exists, _ := afero.Exists(fm.FS, dst); exists && !fm.force {
		return nil
	}

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

	fm.Print(Created, dst)
	return nil
}
