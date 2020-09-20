package health

import (
	"fmt"
	"net/http"
)

func Handler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, "OK")
}
