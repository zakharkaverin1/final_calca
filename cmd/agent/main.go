package main

import (
	"log"
	"github.com/zakharkaverin1/final_calca/internal/application"
)

func main() {
	agent := application.NewAgent()
	log.Println("Агент запущен")
	agent.Run()
}
