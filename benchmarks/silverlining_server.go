package main

import (
	"log"

	"github.com/go-www/silverlining"
)

func main() {
	addr := ":18080"
	log.Printf("silverlining benchmark server listening on %s", addr)
	if err := silverlining.ListenAndServe(addr, func(r *silverlining.Context) {
		r.WriteFullBodyString(200, "Hello, World!")
	}); err != nil {
		log.Fatal(err)
	}
}
