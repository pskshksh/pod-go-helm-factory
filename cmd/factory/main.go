package main

import (
	"context"
	"fmt"
	"log"

	"github.com/pskshksh/pod-go-helm-factory/charts"
)

func main() {
	api := charts.Service{
		Name:        "api",
		Description: "Demo api srv",
	}

	dir, err := api.Generate(context.Background(), "dist", "0.1.0")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Generated: ", dir)
}
