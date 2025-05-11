package main

import (
	"log"

	"github.com/zakharkaverin1/final_calculator/internal/application"
)

func main() {
	Orchestrator := application.NewOrchestrator()
	log.Println("Оркестратор запущен")
	Orchestrator.Run()
}
