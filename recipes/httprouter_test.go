// (c) Copyright IBM Corp. 2022

package recipes

import (
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"testing"
)

func TestHttpRouterRecipe(t *testing.T) {

	src := `package main

	import (
			"fmt"
			"net/http"
			"log"
	
			"github.com/julienschmidt/httprouter"
	)
	
	func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			fmt.Fprint(w, "Welcome!\n")
	}
	
	func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
	}
	
	func main() {
			var router *httprouter.Router
			router = httprouter.New()
			router.GET("/", Index)
			router.GET("/hello/:name", Hello)
	
			log.Fatal(http.ListenAndServe(":8080", router))
	}
	`
	// src := `package main

	// import (
	// 		"fmt"
	// 		"net/http"
	// 		"log"

	// 		"github.com/julienschmidt/httprouter"
	// )

	// func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// 		fmt.Fprint(w, "Welcome!\n")
	// }

	// func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// 		fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
	// }

	// func main() {
	// 		router := httprouter.New()
	// 		router.GET("/", Index)
	// 		router.GET("/hello/:name", Hello)

	// 		log.Fatal(http.ListenAndServe(":8080", router))
	// }
	// `

	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, "", src, 0)
	fmt.Println(node, err)

	recipe := NewHttpRouter()

	res, changed := recipe.Instrument(fset, node, "httprouter", "__instanaSensor")
	fmt.Println(res, changed)

	format.Node(os.Stdout, fset, res)

}
