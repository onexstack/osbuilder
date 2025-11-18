package file

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"strings"

	"mvdan.cc/gofumpt/format"
)

func (fm *FileManager) AddNewGRPCMethod(filePath string, kind string, grpcServiceName string, importPath string) error {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	updated, changed, err := applyUpdates(string(b), kind, grpcServiceName, importPath)
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}

	if err := backupFile(filePath, b); err != nil {
		return err
	}
	if err := os.WriteFile(filePath, []byte(updated), 0o644); err != nil {
		return err
	}

	return nil
}

func (fm *FileManager) AddNewMethod(layer string, filePath string, kind string, version string, importPath string) error {
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
			`%s%s "%s"`,
			strings.ToLower(kind),
			version,
			importPath,
		), importPath)
	}

	oldSRC, err := os.ReadFile(filePath)

	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, node); err != nil {
		return err
	}

	if bytes.Equal(oldSRC, buf.Bytes()) {
		return nil
	}

	formatted, err := format.Source(buf.Bytes(), format.Options{})
	if err != nil {
		return err
	}

	if err := os.WriteFile(filePath, formatted, 0o644); err != nil {
		return err
	}

	fm.Print(Updated, filePath)
	return nil
}

func modifyAST(layer string, node *ast.File, kind string, version string) {
	if layer == "store" {
		addMethodToInterface(node, "IStore", kind, kind+"Store", "// aaa")
		addMethodToStruct(node, layer, kind, version)
		return
	}

	methodName := fmt.Sprintf("%s%s", kind, strings.ToUpper(version))
	aliasType := fmt.Sprintf("%s%s", strings.ToLower(kind), version)
	returnType := fmt.Sprintf("%s.%sBiz", aliasType, kind)
	addMethodToInterface(node, "IBiz", methodName, returnType, "// bbb")
	addMethodToStruct(node, layer, kind, version)
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
						return
					}
				}
			}

			// 创建新的方法声明（关键修正点）
			newMethod := &ast.Field{
				Doc: &ast.CommentGroup{ // 使用 Doc 而不是 Comment
					List: []*ast.Comment{{Text: comment}},
				},
				Names: []*ast.Ident{
					{
						Name:    methodName,
						NamePos: token.Pos(1),
						Obj: &ast.Object{
							Kind: ast.Fun,
							Name: methodName,
						},
					},
				},
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
			return
		}
	}
}

func addMethodToStruct(file *ast.File, layer string, kind string, version string) {
	recv := "b"
	retType := fmt.Sprintf("%s%s.%sBiz", strings.ToLower(kind), version, kind)
	body := fmt.Sprintf("%s%s.New(b.store)", strings.ToLower(kind), version)
	methodName := fmt.Sprintf("%s%s", kind, strings.ToUpper(version))
	structName := "biz"
	if layer == "store" {
		recv = "store"
		retType = "TenantStore"
		retType = fmt.Sprintf("%sStore", kind)
		body = fmt.Sprintf("new%sStore(store)", kind)
		structName = "datastore"
		methodName = kind
	}

	// 检查结构体是否存在
	var targetStructExists bool
	var methodExists bool

	// 首先检查结构体是否存在
	ast.Inspect(file, func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if _, ok := typeSpec.Type.(*ast.StructType); ok && typeSpec.Name.Name == structName {
				targetStructExists = true
				return false
			}
		}
		return true
	})

	if !targetStructExists {
		return
	}

	// 检查是否存在同名方法
	ast.Inspect(file, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			// 检查是否是方法（有接收者）
			if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
				// 获取接收者类型
				recv := funcDecl.Recv.List[0].Type
				var recvName string

				// 处理指针接收者
				if starExpr, ok := recv.(*ast.StarExpr); ok {
					if ident, ok := starExpr.X.(*ast.Ident); ok {
						recvName = ident.Name
					}
				} else if ident, ok := recv.(*ast.Ident); ok {
					recvName = ident.Name
				}

				// 检查接收者类型和方法名
				if recvName == structName && funcDecl.Name.Name == methodName {
					methodExists = true
					return false
				}
			}
		}
		return true
	})

	if methodExists {
		return
	}

	// 创建新的方法声明
	newMethod := &ast.FuncDecl{
		Doc: &ast.CommentGroup{
			List: []*ast.Comment{
				{
					Text:  "// ValidatePassword 验证用户密码是否正确",
					Slash: token.Pos(1),
				},
				{
					Text:  "// 参数：password - 待验证的密码",
					Slash: token.Pos(1),
				},
				{
					Text:  "// 返回：bool - 密码是否正确，error - 错误信息",
					Slash: token.Pos(1),
				},
			},
		},
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{
						{
							Name: recv,
						},
					},
					Type: &ast.StarExpr{
						X: ast.NewIdent(structName),
					},
				},
			},
		},
		Name: ast.NewIdent(methodName),
		Type: &ast.FuncType{
			/*
				Params: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{
								{
									Name: "password",
								},
							},
							Type: ast.NewIdent("string"),
						},
					},
				},
			*/
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: ast.NewIdent(retType),
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.Ident{Name: body},
					},
				},
			},
		},
	}

	// 将新方法添加到文件声明列表中
	file.Decls = append(file.Decls, newMethod)
	/*
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
	*/
}

// parseMethod 将方法的代码字符串解析为 AST 节点
func parseMethod(methodBody string) *ast.FuncDecl {
	src := "package main\n" + methodBody
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", src, parser.AllErrors|parser.ParseComments)
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
func addImport(fset *token.FileSet, f *ast.File, path string, packageName string) bool {
	// 如果 import 已存在则跳过
	if hasImport(f, packageName) {
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
func hasImport(f *ast.File, packageName string) bool {
	hasImport := false
	for _, imp := range f.Imports {
		if strings.Contains(imp.Path.Value, packageName) {
			hasImport = true
			break
		}
	}

	return hasImport
}
