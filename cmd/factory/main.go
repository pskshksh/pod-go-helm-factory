package main

import (
	"context"
	"fmt"
	"log"

	"github.com/pskshksh/pod-go-helm-factory/catalog"
)

const chartVersion = "0.1.0"

func main() {
	ctx := context.Background()
	for _, svc := range catalog.Services() {
		dir, err := svc.Generate(ctx, "dist", chartVersion)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Generated:", dir)
	}
}
