package main

import (
	"log"

	"github.com/zakharkaverin1/final_calca/internal/application"
)

func main() {
	Orchestrator := application.NewOrchestrator()
	log.Println("Оркестратор запущен")
	Orchestrator.Run()
}
