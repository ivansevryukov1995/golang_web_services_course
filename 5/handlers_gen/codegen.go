package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"reflect"
	"strings"
	"text/template"
)

var (
	packageTpl = template.Must(template.New("packageTpl").Parse(
		`package {{.NamePackage}}
	
import (
	"errors"
	"net/http"
	"slices"
	"strconv"
)

`))
	serveTpl = template.Must(template.New("serveTpl").Parse(
		`{{range .Receivers}}func (h *{{.Receiver}}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path { {{range .Routes}}
	case "{{.Meta.URL}}":
		h.handler{{.Method.Name.Name}}(w, r){{end}}
	default:
		http.Error(w, "unknown method", http.StatusNotFound)
	}
}

{{end}}`))
	respTpl = template.Must(template.New("packageTpl").Parse(
		`type resp struct {
	Error    string      ` + "`" + `json:"error"` + "`" + `
	Response interface{} ` + "`" + `json:"response,omitempty"` + "`" + `
}
	`))
	requiredTpl = `if params.{{.Param}} == "" {
		http.Error(w, "{{.Param}} must me not empty", http.StatusBadRequest)
		return
	}`
	authTpl = `if r.Header.Get("X-Auth") != "100500" {
		w.Header().Set("WWW-Authenticate", "Basic realm='api'")
		http.Error(w, "unauthorized", http.StatusForbidden)
		return
	}`
	apiErrorCheckTpl = `if err != nil {
		var ae *ApiError
		if errors.As(err, ae) {
			http.Error(w, err.Error(), ae.HTTPStatus)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}`
	handlerTpl = template.Must(template.New("handlerTpl").
			Funcs(template.FuncMap{
			"authTpl":             func() string { return authTpl },
			"apiErrorCheckTpl":    func() string { return apiErrorCheckTpl },
			"getNameParamsStruct": getNameParamsStruct,
			"parseTag":            parseTag,
			"hasKey":              hasKey,
			"toLowercase":         func(s string) string { return strings.ToLower(s) },
			"getTypeField":        getTypeField,
			"paramKey":            paramKey,
			"split":               func(s, sep string) []string { return strings.Split(s, sep) },
		}).
		Parse(
			`{{$root := .}}
	{{range $rec := .Receivers}}
		{{range $rts := $rec.Routes}}
			func (h *{{$rec.Receiver}}) handler{{$rts.Method.Name.Name}}(w http.ResponseWriter, r *http.Request) {
			{{if $rts.Meta.Auth}}
				{{authTpl}}
			{{end}}
			{{if eq $rts.Meta.Method "POST"}}
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			{{else}}
			values := r.URL.Query()
			{{end}}
			var params {{(index $rts.Method.Type.Params.List 1).Type}}
			// валидирование параметров и заполнение структуры params
			{{range $st := $root.Structs}}
				{{$paramsType := $rts.Method.Type.Params.List | getNameParamsStruct}}
				{{if eq $st.TypeSpec.Name.Name $paramsType}}
					{{range $field := $st.StructType.Fields.List}}	
						{{if $field.Tag}}
							{{$tagMap := $field.Tag.Value | parseTag}}
							{{$paramname := paramKey $tagMap (index $field.Names 0).Name}}
							{{if eq $rts.Meta.Method "POST"}}
								{{if eq "int" ($field | getTypeField)}}
									{{(index $field.Names 0).Name | toLowercase}}, err := strconv.Atoi(r.FormValue("{{(index $field.Names 0).Name | toLowercase}}")) 
									if err != nil {
										http.Error(w, err.Error(), http.StatusInternalServerError)
										return
									}
								{{else}}
									{{(index $field.Names 0).Name | toLowercase}} := r.FormValue("{{$paramname}}")
								{{end}}
							{{else}}
								{{if eq "int" ($field | getTypeField)}}
									{{(index $field.Names 0).Name | toLowercase}}, err := strconv.Atoi(values.Get("{{(index $field.Names 0).Name | toLowercase}}")) 
									if err != nil {
										http.Error(w, err.Error(), http.StatusInternalServerError)
										return
									}
								{{else}}
									{{(index $field.Names 0).Name | toLowercase}} :=  values.Get("{{$paramname}}")
								{{end}}
							{{end}}
							{{range $key, $val := $tagMap}}
								{{if eq $key "default"}}
									if {{(index $field.Names 0).Name | toLowercase}} == "" {
										{{(index $field.Names 0).Name | toLowercase}} = "{{$tagMap.default}}"
									}
								{{else if eq $key "required"}}
									if {{(index $field.Names 0).Name | toLowercase}} == "" {
										http.Error(w, "{{$paramname}} must me not empty", http.StatusBadRequest)
										return
									}
								{{else if eq $key "min"}}
									{{if eq "string" ($field | getTypeField)}}
										if len({{(index $field.Names 0).Name | toLowercase}}) < {{$tagMap.min}} {
											http.Error(w, "{{$paramname}} len must be >= {{$tagMap.min}}", http.StatusBadRequest)
											return
										}
									{{else if eq "int" ($field | getTypeField)}}
										if {{(index $field.Names 0).Name | toLowercase}} < {{$tagMap.min}} {
											http.Error(w, "{{$paramname}} must be >= {{$tagMap.min}}", http.StatusBadRequest)
											return
										}
									{{end}}
								{{else if eq $key "max"}}
									if {{(index $field.Names 0).Name | toLowercase}} > {{$tagMap.max}} {
										http.Error(w, "{{$paramname}} must be <= {{$tagMap.max}}", http.StatusBadRequest)
										return
									}
								{{else if eq $key "enum"}}
									{{$allowed := split $tagMap.enum "|"}}
									if !slices.Contains([]string{ {{range $allowed}}"{{.}}",{{end}} }, {{$paramname}}) {
										http.Error(w, "{{$paramname}} must be one of {{$tagMap.enum}}", http.StatusBadRequest)
										return
									}
										
								{{end}}	
							{{end}}
							params.{{(index $field.Names 0).Name}} = {{(index $field.Names 0).Name | toLowercase}}
						{{end}}
					{{end}}

				{{end}}
			{{end}}
			//Do SomeJob
			res, err := h.{{$rts.Method.Name.Name}}(r.Context(), params)
			{{apiErrorCheckTpl}}

			// прочие обработки
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			answer := resp{Error: "", Response: res}
			if err := json.NewEncoder(w).Encode(answer); err != nil {
				var ae *ApiError
				if errors.As(err, ae) {
					http.Error(w, err.Error(), ae.HTTPStatus)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
}

		{{end}}
	{{end}}`))
)

type PackageTmpl struct {
	NamePackage string
}

type ServeTmpl struct {
	Receiver string
	Routes   []*Route
}

type Route struct {
	Method *ast.FuncDecl
	Meta   *RouteMeta
}

type RouteMeta struct {
	URL    string `json:"url"`
	Auth   bool   `json:"auth"`
	Method string `json:"method"`
}

type StructTmpl struct {
	TypeSpec   *ast.TypeSpec
	StructType *ast.StructType
}

func getNameParamsStruct(params []*ast.Field) string {
	return fmt.Sprint(params[1].Type)
}

func parseTag(raw string) map[string]string {
	s := strings.Trim(raw, "`")
	tag := reflect.StructTag(s)

	val, ok := tag.Lookup("apivalidator")
	if !ok {
		return nil
	}

	out := make(map[string]string)
	for _, part := range strings.Split(val, ",") {
		if kv := strings.SplitN(part, "=", 2); len(kv) == 2 {
			out[kv[0]] = kv[1]
		} else {
			out[part] = ""
		}
	}
	return out
}

func hasKey(m map[string]string, k string) bool {
	_, ok := m[k]
	return ok
}

func getTypeField(field *ast.Field) string {
	var buf bytes.Buffer
	format.Node(&buf, token.NewFileSet(), field.Type)
	return buf.String()
}

func paramKey(m map[string]string, fieldName string) string {
	if k, ok := m["paramname"]; ok {
		return k
	}
	return strings.ToLower(fieldName)
}

func isMethod(fn *ast.FuncDecl) bool {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		// fmt.Printf("SKIP function %s is not a method\n", fn.Name.Name)
		return false
	}

	if fn.Doc == nil {
		// fmt.Printf("SKIP method %s has no comments\n", fn.Name.Name)
		return false
	}

	needCodegen := false
	for _, comment := range fn.Doc.List {
		needCodegen = needCodegen || strings.HasPrefix(comment.Text, "// apigen:api")
	}
	if !needCodegen {
		// fmt.Printf("SKIP method %#v doesnt have apigen mark\n", fn.Name.Name)
		return false
	}
	return true
}

func createdMethodList(fn *ast.FuncDecl, receiverList *[]ServeTmpl) {
	if !isMethod(fn) {
		return
	}

	nameReceiver := fmt.Sprint(fn.Recv.List[0].Type.(*ast.StarExpr).X)

	var sb strings.Builder
	sb.Reset()
	for _, comment := range fn.Doc.List {
		_, err := sb.WriteString(strings.TrimPrefix(comment.Text, "// apigen:api "))
		if err != nil {
			return
		}
	}
	resultComment := sb.String()

	resultJSON := &RouteMeta{}
	if err := json.Unmarshal([]byte(resultComment), resultJSON); err != nil {
		log.Fatal(err)
	}

	for i := range *receiverList {
		if (*receiverList)[i].Receiver == nameReceiver {
			(*receiverList)[i].Routes = append(
				(*receiverList)[i].Routes,
				&Route{
					Method: fn,
					Meta:   resultJSON,
				})
			return
		}
	}

	*receiverList = append(
		*receiverList,
		ServeTmpl{
			Receiver: nameReceiver,
			Routes: []*Route{
				{
					Method: fn,
					Meta:   resultJSON,
				},
			},
		},
	)
}

func createdStructList(sct *ast.GenDecl, structList *[]StructTmpl) {
	for _, spec := range sct.Specs {
		currType, ok := spec.(*ast.TypeSpec)
		if !ok {
			// fmt.Printf("SKIP %#T is not ast.TypeSpec\n", spec)
			continue
		}

		currStruct, ok := currType.Type.(*ast.StructType)
		if !ok {
			// fmt.Printf("SKIP %#T is not ast.StructType\n", currStruct)
			continue
		}
		*structList = append(*structList, StructTmpl{TypeSpec: currType, StructType: currStruct})

	}
}

func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	out, err := os.Create(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}

	receiverList := &[]ServeTmpl{}
	structList := &[]StructTmpl{}

	for _, f := range node.Decls {
		switch decl := f.(type) {
		case *ast.FuncDecl:
			createdMethodList(decl, receiverList)
		case *ast.GenDecl:
			createdStructList(decl, structList)
		default:
			// fmt.Printf("SKIP %#T is not *ast.FuncDecl or *ast.GenDecl\n", f)
			continue
		}
	}

	validList := &struct {
		Receivers *[]ServeTmpl
		Structs   *[]StructTmpl
	}{Receivers: receiverList,
		Structs: structList,
	}

	packageTpl.Execute(out, PackageTmpl{node.Name.Name})
	respTpl.Execute(out, struct{}{})
	serveTpl.Execute(out, validList)
	handlerTpl.Execute(out, validList)

}
