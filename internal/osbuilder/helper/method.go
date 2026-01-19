package helper

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"strings"
)

func AddNewMethod(layer string, filePath string, kind string, version string, importPath string) error {
	// 加载并解析源文件
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return err
	}

	// 修改 AST 节点
	modifyAST(layer, node, kind, version)
	// 执行添加 import 操作
	if layer == "biz" {
		addImport(fset, node, fmt.Sprintf(
			`%s%s "%s/biz/%s/%s"`,
			strings.ToLower(kind),
			version,
			importPath,
			version,
			strings.ToLower(kind),
		))
	}

	// 将更新后的代码写回文件
	outputFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	err = printer.Fprint(outputFile, fset, node)
	if err != nil {
		return err
	}
	return nil
}

func modifyAST(layer string, node *ast.File, kind string, version string) {
	if layer == "store" {
		methodBody := fmt.Sprintf(`

// %s returns an instance that implements the %sStore.
func (store *store) %s() %sStore {
    return new%sStore(store)
}
`, kind, kind, kind, kind, kind)
		addMethodToInterface(
			node,
			"IStore",
			kind,
			kind+"Store",
			fmt.Sprintf("// %s returns the %sStore interface.", kind, kind),
		)
		addMethodToStruct(node, "store", kind, methodBody)
		return
	}

	methodName := fmt.Sprintf("%s%s", kind, strings.ToUpper(version))
	aliasType := fmt.Sprintf("%s%s", strings.ToLower(kind), version)
	returnType := fmt.Sprintf("%s.%sBiz", aliasType, kind)
	addMethodToInterface(
		node,
		"IBiz",
		methodName,
		returnType,
		fmt.Sprintf("// %sV1 returns an instance that implements the %sBiz interface.", kind, kind),
	)

	methodBody := fmt.Sprintf(`

// %s returns an instance that implements the %sBiz.
func (b *biz) %s() %s {
	return %s.New(b.store)
}
`, methodName, kind, methodName, returnType, aliasType)

	addMethodToStruct(node, "biz", kind, methodBody)
}

func addMethodToInterface(node *ast.File, interfaceName, methodName, returnType, comment string) {
	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != interfaceName {
				continue
			}

			interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
			if !ok {
				continue
			}

			// 检查方法是否已存在
			for _, field := range interfaceType.Methods.List {
				for _, name := range field.Names {
					if name.Name == methodName {
						log.Printf("Method '%s' already exists in interface '%s'\n", methodName, interfaceName)
						return
					}
				}
			}

			// 创建新的方法声明（关键修正点）
			newMethod := &ast.Field{
				Doc: &ast.CommentGroup{ // 使用 Doc 而不是 Comment
					List: []*ast.Comment{{Text: comment}},
				},
				Names: []*ast.Ident{ast.NewIdent(methodName)},
				Type: &ast.FuncType{
					Params: &ast.FieldList{}, // 无参数
					Results: &ast.FieldList{ // 返回值直接关联到方法名
						List: []*ast.Field{
							{Type: ast.NewIdent(returnType)},
						},
					},
				},
			}

			// 将新方法添加到接口中
			interfaceType.Methods.List = append(interfaceType.Methods.List, newMethod)
			log.Printf("Added method '%s() %s' to interface '%s'\n", methodName, returnType, interfaceName)
			return
		}
	}
}

func addMethodToStruct(node *ast.File, structName, methodName, methodBody string) {
	// 检查方法是否已存在
	for _, decl := range node.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Recv == nil {
			continue
		}
		recv := funcDecl.Recv.List[0]
		starExpr, ok := recv.Type.(*ast.StarExpr)
		if !ok {
			continue
		}
		ident, ok := starExpr.X.(*ast.Ident)
		if ok && ident.Name == structName && funcDecl.Name.Name == methodName {
			log.Printf("Method %s already exists in struct %s", methodName, structName)
			return
		}
	}

	// 解析新方法
	methodDecl := parseMethod(methodBody)
	if methodDecl == nil {
		log.Fatalf("Failed to parse method body for method %s", methodName)
	}

	// 将新方法添加到文件声明中
	node.Decls = append(node.Decls, methodDecl)
	log.Printf("Added method %s() to struct %s", methodName, structName)
}

// parseMethod 将方法的代码字符串解析为 AST 节点
func parseMethod(methodBody string) *ast.FuncDecl {
	src := "package main\n" + methodBody
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	if err != nil {
		log.Printf("Failed to parse method body: %v", err)
		return nil
	}

	// 查找函数声明
	for _, decl := range node.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			return funcDecl
		}
	}
	return nil
}

// 添加 import 的核心逻辑
func addImport(fset *token.FileSet, f *ast.File, path string) bool {
	// 如果 import 已存在则跳过
	if hasImport(f, path) {
		return false
	}

	// 构建新 import spec
	newSpec := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: path,
		},
	}

	// 遍历现有 import 声明
	var added bool
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			continue
		}

		// 插入到当前 import 声明中
		genDecl.Specs = append(genDecl.Specs, newSpec)
		added = true
		break
	}

	// 如果不存在 import 声明则创建新的
	if !added {
		newDecl := &ast.GenDecl{
			Tok:   token.IMPORT,
			Specs: []ast.Spec{newSpec},
		}
		f.Decls = append([]ast.Decl{newDecl}, f.Decls...)
	}

	return true
}

// 检查是否存在指定 import
func hasImport(f *ast.File, path string) bool {
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			continue
		}

		for _, spec := range genDecl.Specs {
			importSpec := spec.(*ast.ImportSpec)
			if importSpec.Path.Value == path {
				return true
			}
		}
	}
	return false
}
