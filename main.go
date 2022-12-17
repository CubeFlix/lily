// main.go
// Lily server main program.

package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/cubeflix/lily/cmd"
)

func main() {
	http.ListenAndServe("localhost:8080", nil)
	cmd.Execute()
}
