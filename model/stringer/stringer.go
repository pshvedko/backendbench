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

//go:embed stringer.go.tmp
var tmpFS embed.FS

func tagParse(tag, key string) (*structtag.Tag, error) {
	t, err := structtag.Parse(strings.Trim(tag, "` "))
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

type field struct {
	First bool
	Name  string
	Type  string
	Alias string
	Mod   modifiers
}

func structFields(file *ast.File) (string, []field, error) {
	var fields []field
	for _, dl := range file.Decls {
		if gd, ok := dl.(*ast.GenDecl); ok && gd != nil && gd.Tok == token.TYPE {
			for _, sp := range gd.Specs {
				switch ts := sp.(type) {
				case *ast.TypeSpec:
					switch st := ts.Type.(type) {
					case *ast.StructType:
						for _, fd := range st.Fields.List {
							if fd.Tag == nil {
								continue
							}
							jt, err := tagParse(fd.Tag.Value, "json")
							if err != nil {
								return "", nil, err
							}
							if len(fd.Names) != 1 {
								return "", nil, fmt.Errorf("FIXME5 %#v ", fd)
							}
							fn := fd.Names[0]
							var mod modifiers
							for {
								switch ft := fd.Type.(type) {
								case *ast.SelectorExpr:
									switch pt := ft.X.(type) {
									case *ast.Ident:
										fd.Type = &ast.Ident{
											NamePos: pt.NamePos,
											Name:    strings.Join([]string{pt.Name, ft.Sel.Name}, "."),
											Obj:     pt.Obj,
										}
										continue
									default:
										return "", nil, fmt.Errorf("FIXME3 %#v ", fd)
									}
								case *ast.Ident:
									fields = append(fields, field{
										First: 0 == len(fields),
										Name:  fn.Name,
										Type:  ft.Name,
										Alias: jt.Name,
										Mod:   mod,
									})
								case *ast.StarExpr:
									fd.Type = ft.X
									mod = append(mod, Pointer)
									continue
								case *ast.ArrayType:
									fd.Type = ft.Elt
									mod = append(mod, Array)
									continue
								default:
									return "", nil, fmt.Errorf("FIXME1 %#v ", fd)
								}
								break
							}
						}
					default:
						return "", nil, fmt.Errorf("FIXME2 %#v", st)
					}
				default:
					return "", nil, fmt.Errorf("FIXME4 %#v", ts)
				}
			}
		}
	}

	return file.Name.Name, fields, nil
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal(os.ErrInvalid)
	}

	file, err := parser.ParseFile(token.NewFileSet(), strcase.ToSnake(os.Args[1])+".go", nil, 0)
	if err != nil {
		log.Fatal(err)
	}

	name, fields, err := structFields(file)
	if err != nil {
		log.Fatal(err)
	}

	tmp, err := template.ParseFS(tmpFS, "*")
	if err != nil {
		log.Fatal(err)
	}

	out, err := os.Create(strcase.ToSnake(os.Args[1]) + "_stringer.go")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	defer out.Sync()

	var buf bytes.Buffer
	err = tmp.Execute(&buf, map[string]interface{}{
		"Type":    os.Args[1],
		"Fields":  fields,
		"Package": name,
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = buf.WriteTo(out)
	if err != nil {
		return
	}
}
