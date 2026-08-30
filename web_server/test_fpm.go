package main

import (
	"fmt"
	"net/http/httptest"
	"github.com/yookoala/gofast"
)

func main() {
	connFactory := gofast.SimpleConnFactory("unix", "/run/php/php-fpm.sock")
	clientFactory := gofast.SimpleClientFactory(connFactory)
	phpHandler := gofast.NewHandler(
		gofast.NewPHPFS("/home/aurnob/Documents/go_projects/web_server/static")(gofast.BasicSession),
		clientFactory,
	)

	req := httptest.NewRequest("GET", "/index.php", nil)
	w := httptest.NewRecorder()
	
	phpHandler.ServeHTTP(w, req)
	
	fmt.Printf("Headers from FPM: %v\n", w.Header())
}
