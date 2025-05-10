package main

import (
	"github.com/zakharkaverin1/calculator/internal/application"
)

func main() {
	Orchestrator := application.NewOrchestrator()
	Orchestrator.Run()
}
