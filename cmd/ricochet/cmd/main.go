package main

import (
	"log"

	"github.com/grik-ai/ricochet-task/cmd/ricochet"
)

func main() {
	if err := ricochet.Execute(); err != nil {
		log.Fatalf("Ошибка выполнения команды: %v", err)
	}
}
