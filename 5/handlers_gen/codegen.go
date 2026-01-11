package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
	"text/template"
)

var (
	packageTpl = template.Must(template.New("packageTpl").Parse(
		`package {{.NamePackage}}
	
import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
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
	authTpl = `if r.Header.Get("X-Auth") != "100500" {
		w.Header().Set("WWW-Authenticate", "Basic realm='api'")
		http.Error(w, "unauthorized", http.StatusForbidden)
		return
	}`
	apiErrorCheckTpl = `if err != nil {
		var ae *ApiError
		if errors.As(err, ae) {
			http.Error(w, err.Error(), ae.HTTPStatus)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}`
	handlerTpl = template.Must(template.New("handlerTpl").
			Funcs(template.FuncMap{
			"authTpl":             func() string { return authTpl },
			"apiErrorCheckTpl":    func() string { return apiErrorCheckTpl },
			"getNameParamsStruct": getNameParamsStruct,
		}).
		Parse(
			`{{range $rec := .Receivers}}{{range $rts := $rec.Routes}}func (h *{{$rec.Receiver}}) handler{{$rts.Method.Name.Name}}(w http.ResponseWriter, r *http.Request) {
	{{if $rts.Meta.Auth}}{{authTpl}}
	{{end}}
	{{if eq $rts.Meta.Method "POST"}}values := r.URL.Query(){{else}}body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var params {{$rts.Method.Type.Params.List | getNameParamsStruct}}
	if err := json.Unmarshal(body, &params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if params.Login == "" {
		http.Error(w, "login must me not empty", http.StatusBadRequest)
		return
	}{{end}}

	res, err := h.{{$rts.Method.Name.Name}}(r.Context(), params)
	{{apiErrorCheckTpl}}
}

{{end}}{{end}}`))
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

type ValidTmpl struct {
	Struct *ast.TypeSpec
}

func getNameParamsStruct(params []*ast.Field) string {
	return fmt.Sprint(params[1].Type)
}

func isMethod(fn *ast.FuncDecl) bool {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		fmt.Printf("SKIP function %s is not a method\n", fn.Name.Name)
		return false
	}

	if fn.Doc == nil {
		fmt.Printf("SKIP method %s has no comments\n", fn.Name.Name)
		return false
	}

	needCodegen := false
	for _, comment := range fn.Doc.List {
		needCodegen = needCodegen || strings.HasPrefix(comment.Text, "// apigen:api")
	}
	if !needCodegen {
		fmt.Printf("SKIP method %#v doesnt have apigen mark\n", fn.Name.Name)
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

func createdStructList(sct *ast.GenDecl, structList *[]ValidTmpl) {
	for _, spec := range sct.Specs {
		currType, ok := spec.(*ast.TypeSpec)
		if !ok {
			fmt.Printf("SKIP %#T is not ast.TypeSpec\n", spec)
			continue
		}

		currStruct, ok := currType.Type.(*ast.StructType)
		if !ok {
			fmt.Printf("SKIP %#T is not ast.StructType\n", currStruct)
			continue
		}
		*structList = append(*structList, ValidTmpl{Struct: currType})
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
	structList := &[]ValidTmpl{}

	for _, f := range node.Decls {
		switch decl := f.(type) {
		case *ast.FuncDecl:
			createdMethodList(decl, receiverList)
		case *ast.GenDecl:
			createdStructList(decl, structList)
		default:
			fmt.Printf("SKIP %#T is not *ast.FuncDecl or *ast.GenDecl\n", f)
			continue
		}
	}

	receivers := &struct {
		Receivers *[]ServeTmpl
	}{Receivers: receiverList}

	packageTpl.Execute(out, PackageTmpl{node.Name.Name})
	serveTpl.Execute(out, receivers)
	handlerTpl.Execute(out, receivers)

}
