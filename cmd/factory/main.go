package main

import (
	"context"
	"fmt"
	"log"

	"github.com/pskshksh/pod-go-helm-factory/charts"
)

func main() {
	app := charts.Service{
		Name:        "sampleapp",
		Description: "Sample HTTP service",
		Container: charts.Container{
			Image: "ghcr.io/pskshksh/pod-go-helm-factory/sampleapp",
		},
		NetworkPolicy:       &charts.NetworkPolicy{AllowSameNamespace: true},
		PodDisruptionBudget: &charts.PodDisruptionBudget{},
	}

	dir, err := app.Generate(context.Background(), "dist", "0.1.0")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Generated:", dir)
}
