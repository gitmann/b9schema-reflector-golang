package main

import (
	"fmt"
	"github.com/gitmann/b9schema-reflector-golang/b9schema/reflector"
	"github.com/gitmann/b9schema-reflector-golang/b9schema/renderer"
	"strings"
)

// Hello, World for b9schema.
func main() {
	type HelloStruct struct {
		Hello string
		World float64
	}

	r := reflector.NewReflector()
	schema := r.DeriveSchema(HelloStruct{})

	render := renderer.NewOpenAPIRenderer("/hello/world", nil)

	lines, err := render.ProcessResult(schema)
	if err != nil {
		fmt.Println("ERROR", err)
	} else {
		fmt.Println(strings.Join(lines, "\n"))
	}
}
