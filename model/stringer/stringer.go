package main

import (
	"bytes"
	"embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/fatih/structtag"
	"github.com/iancoleman/strcase"
)

//go:embed stringer.go.tmpl
var tmpFS embed.FS

func tagParse(tag, key string) (*structtag.Tag, error) {
	t, err := structtag.Parse(tag[1 : len(tag)-1])
	if err != nil {
		return nil, fmt.Errorf("%v %s", err, tag)
	}
	return t.Get(key)
}

type modifier int

const (
	None = iota
	Pointer
	Array
)

func (m modifier) String() string {
	switch m {
	case Array:
		return "[]"
	case Pointer:
		return "*"
	default:
		return ""
	}
}

type modifiers []modifier

func (mm modifiers) String() string {
	var s []string
	for _, m := range mm {
		s = append(s, m.String())
	}
	return strings.Join(s, "")
}

func structFields(file *ast.File) (string, error) {
	//log.Printf("%#v", file.Name.Name)
	for _, dl := range file.Decls {
		if gd, ok := dl.(*ast.GenDecl); ok && gd != nil && gd.Tok == token.TYPE {
			//log.Printf("%#v", gd)
			for _, spec := range gd.Specs {
				switch ts := spec.(type) {
				case *ast.TypeSpec:
					//log.Printf("%#v", ts)
					switch st := ts.Type.(type) {
					case *ast.StructType:
						//log.Printf("%#v", st.Fields)
						for _, field := range st.Fields.List {
							//log.Printf("%#v", field)
							tag, err := tagParse(field.Tag.Value, "json")
							if err != nil {
								return "", err
							}
							if len(field.Names) != 1 {
								return "", fmt.Errorf("FIXME5 %#v ", field)
							}
							name := field.Names[0]
							var mod modifiers
							for {
								switch ft := field.Type.(type) {
								case *ast.SelectorExpr:
									switch pt := ft.X.(type) {
									case *ast.Ident:
										field.Type = &ast.Ident{
											NamePos: pt.NamePos,
											Name:    strings.Join([]string{pt.Name, ft.Sel.Name}, "."),
											Obj:     pt.Obj,
										}
										continue
									default:
										return "", fmt.Errorf("FIXME3 %#v ", field)
									}
								case *ast.Ident:
									log.Printf("\t%s %s%s :%s", name.Name, mod, ft.Name, tag.Name)
								case *ast.StarExpr:
									field.Type = ft.X
									mod = append(mod, Pointer)
									continue
								case *ast.ArrayType:
									field.Type = ft.Elt
									mod = append(mod, Array)
									continue
								default:
									return "", fmt.Errorf("FIXME1 %#v ", field)
								}
								break
							}
						}
					default:
						return "", fmt.Errorf("FIXME2 %#v", st)
					}
				default:
					return "", fmt.Errorf("FIXME4 %#v", ts)
				}
			}
		}
	}

	return file.Name.Name, nil
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal(os.ErrInvalid)
	}
	file, err := parser.ParseFile(token.NewFileSet(), strcase.ToSnake(os.Args[1])+".go", nil, 0)
	if err != nil {
		log.Fatal(err)
	}

	log.Print(os.Args[1])

	pkg, err := structFields(file)
	if err != nil {
		log.Fatal(err)
	}

	tmp, err := template.ParseFS(tmpFS)
	if err != nil {
		log.Fatal(err)
	}

	out, err := os.Create(strcase.ToSnake(os.Args[1]) + "_stringer.go1")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	type data struct {
		Package string
		Type    string
	}
	var buf bytes.Buffer

	err = tmp.Execute(&buf, map[string]interface{}{
		"package": pkg,
		"type":    os.Args[1],
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = buf.WriteTo(out)
	if err != nil {
		return
	}
}
