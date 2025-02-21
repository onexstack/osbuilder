package main

import (
	"os"
	"text/template"
	//"fmt"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/duke-git/lancet/v2/strutil"
	//"github.com/gobuffalo/flect"
	"github.com/fatih/color"
	"github.com/gobuffalo/flect"
	"github.com/rakyll/statik/fs"

	_ "github.com/onexstack/osbuilder/examples/template/statik"
)

type Data struct {
	RootDir       string // The root directory for the project
	APIVersion    string // The version of the API
	ModuleName    string // Go module name for the project
	ComponentName string // Component name, for example: apiserver
	WebFramework  string // The mode in which the server operates
	APIAlias      string // Alias name for the API
	Boilerplate   string // Boilerplate text for license/header files
	ProjectName   string
	// Registry-related settings.
	RegistryPrefix string // Prefix for the image registry (e.g., docker.io/project)

	// Project properties.
	BinaryName       string // Binary name that must adhere to the xx-xxxx format
	StorageType      string // Defines the storage backend (e.g., local or remote)
	DeploymentMethod string // Defines the type of deployment: kubernetes or systemd

	// Resource-specific details.
	UseStructuredMakefile bool // Determines if the Makefile follows a structured format
	GRPCServiceName       string
	WithHealthz           bool

	// Additional configurable settings.
	WithUser          bool   // Indicates whether user-related features are included
	EnvironmentPrefix string // Prefix for generated environment variables (e.g., APP_)

	SingularName       string // Singular form of the kind (e.g., "CronJob")
	PluralName         string // Plural form of the kind (e.g., "CronJobs")
	SingularLower      string // Singular name in lower format(e.g., "cronjob")
	PluralLower        string // Plural name in lower format(e.g., "cronjobs")
	SingularLowerFirst string // Singular name with the first letter lowercase (e.g., "cronJob")
	PluralLowerFirst   string // Plural name with the first letter lowercase (e.g., "cronJobs")

	GORMModel           string // Name of the associated GORM model (e.g., "CronJobModel")
	MapModelToAPIFunc   string // Function name to map the model to the API
	MapAPIToModelFunc   string // Function name to map the API to the model
	BusinessFactoryName string // Name of the business layer factory

	FileName      string // Name of the generated Go file
	APIImportPath string // Import path for the API package
}

func main() {
	statikFS, _ := fs.New()
	// Access individual files by their paths.
	r, _ := statikFS.Open("/tpl.go")
	contentBytes, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}

	data := &Data{
		WithUser:      true,
		WebFramework:  "grpc",
		ComponentName: "gateway",
		APIVersion:    "v1",
		ModuleName:    "github.com/onexstack/onex",
		StorageType:   "memory",
		APIAlias:      "v1",
		WithHealthz:   false,

		SingularName:       "CronJob",
		PluralName:         "CronJobs",
		SingularLower:      "cronjob",
		PluralLower:        "cronjobs",
		SingularLowerFirst: "cronJob",
		PluralLowerFirst:   "cronJobs",
	}
	upperAPIVersion := "V1"
	data.GORMModel = data.SingularName + "M"
	data.MapModelToAPIFunc = fmt.Sprintf("%sMTo%s%s", data.SingularName, data.SingularName, upperAPIVersion)
	data.MapAPIToModelFunc = fmt.Sprintf("%s%sTo%sM", data.SingularName, upperAPIVersion, data.SingularName)
	data.BusinessFactoryName = fmt.Sprintf("%s%s", data.SingularName, upperAPIVersion)

	data.FileName = data.SingularLower + ".go"
	data.APIImportPath = fmt.Sprintf(`%s "%s/pkg/api/%s/%s"`, data.APIAlias, data.ModuleName, data.ComponentName, data.APIVersion)

	// 解析模板
	t, err := template.New("example").Funcs(template.FuncMap{
		"kind":       Kind(),
		"kinds":      Kinds(),
		"capitalize": Capitalize(),
		"lowerkind":  SingularLower(),
		"lowerkinds": SingularLowers(),
	}).Parse(string(contentBytes))
	if err != nil {
		log.Fatal(err)
	}

	// 执行模板并将输出打印到标准输出
	if err := t.Execute(os.Stdout, data); err != nil {
		log.Fatal(err)
	}
}

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
