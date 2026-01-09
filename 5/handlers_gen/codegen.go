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
)

type receiverCodoGen struct {
	receivers map[string][]*apiMethod
}

type parametrsCodoGen struct {
	parametrs map[string][]*ast.StructType
}

type apiMethod struct {
	method        *ast.FuncDecl
	commentMethod *CommentJSON
}

type CommentJSON struct {
	URL    string `json:"url"`
	Auth   bool   `json:"auth"`
	Method string `json:"method"`
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

func isStructParams(g *ast.GenDecl) {
	for _, spec := range g.Specs {
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

		// if g.Doc == nil {
		// 	fmt.Printf("SKIP struct %#v doesnt have comments\n", currType.Name.Name)
		// 	continue
		// }

		// needCodegen := false
		// for _, comment := range g.Doc.List {
		// 	needCodegen = needCodegen || strings.HasPrefix(comment.Text, "// cgen: binpack")
		// }
		// if !needCodegen {
		// 	fmt.Printf("SKIP struct %#v doesnt have cgen mark\n", currType.Name.Name)
		// 	continue SPECS_LOOP
		// }
	}
}

func createdMethodList(fn *ast.FuncDecl, receiverList *receiverCodoGen) {
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

	resultJSON := &CommentJSON{}
	if err := json.Unmarshal([]byte(resultComment), resultJSON); err != nil {
		log.Fatal(err)
	}
	receiverList.receivers[nameReceiver] = append(receiverList.receivers[nameReceiver],
		&apiMethod{
			method:        fn,
			commentMethod: resultJSON,
		})
}

func generateServeHTTP(out *os.File, receiverList *receiverCodoGen) {
	for nameReceiver, elem := range receiverList.receivers {
		fmt.Println(nameReceiver)

		fmt.Fprintln(out, "func (h *"+nameReceiver+") ServeHTTP(w http.ResponseWriter, r *http.Request) {")
		fmt.Fprintln(out, "\tswitch r.URL.Path {")

		for _, v := range elem {
			fmt.Fprintln(out, "\tcase \""+v.commentMethod.URL+"\":")
			fmt.Fprintln(out, "\t\th.handler"+v.method.Name.Name+"(w, r)")

			fmt.Println(v.method.Name.Name)
			fmt.Println(v.commentMethod)
			fmt.Printf("parametrs %s\n", v.method.Type.Params.List[1].Type)
		}

		fmt.Fprintln(out, "\tdefault:")
		fmt.Fprintln(out, "\t\thttp.Error(w, \"Not Found\", http.StatusNotFound)")
		fmt.Fprintln(out, "\t}")
		fmt.Fprintln(out, "}") // end of ServeHTTP method
		fmt.Fprintln(out)
	}
}

func generateWrapperDoMethod(out *os.File, receiverList *receiverCodoGen) {
	for nameReceiver, elem := range receiverList.receivers {
		for _, v := range elem {
			fmt.Fprintln(out, "func (h *"+nameReceiver+") handler"+v.method.Name.Name+"(w http.ResponseWriter, r *http.Request) {")

			// nameWrapperoSomeJob := fmt.Sprint(v.method.Type.Params.List[1].Type)
			fmt.Fprintln(out, "\tres, err := h."+v.method.Name.Name+"(ctx, params)")
			fmt.Fprintln(out, "}") // wrapperDoSomeJob end of  method
			fmt.Fprintln(out)
		}
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

	fmt.Fprintln(out, `package `+node.Name.Name)
	fmt.Fprintln(out) // empty line
	fmt.Fprintln(out, `import "net/http"`)
	fmt.Fprintln(out) // empty line

	receiverList := &receiverCodoGen{
		receivers: make(map[string][]*apiMethod),
	}

	for _, f := range node.Decls {
		switch decl := f.(type) {
		case *ast.FuncDecl:
			if isMethod(decl) {
				createdMethodList(decl, receiverList)
			}
		case *ast.GenDecl:
			isStructParams(decl)
		default:
			fmt.Printf("SKIP %#T is not *ast.FuncDecl or *ast.GenDecl\n", f)
			continue
		}
	}

	generateServeHTTP(out, receiverList)
	generateWrapperDoMethod(out, receiverList)
}
