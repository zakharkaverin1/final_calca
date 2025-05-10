package application

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/zakharkaverin1/calculator/pkg/calculation"
)

type Response struct {
	Id  int     `json:"id"`
	Res float64 `json:"res"`
}

type Agent struct {
	power int
	url   string
}

func NewAgent() *Agent {
	p, err := strconv.Atoi(os.Getenv("COMPUTING_POWER"))
	if err != nil {
		p = 1
	}
	return &Agent{power: p, url: "http://localhost:8080"}
}

func (a *Agent) Run() {
	for i := 0; i < a.power; i++ {
		log.Printf("Начинаем работу демона номер %d", i)
		go a.worker(i)
	}
	select {}
}
func (a *Agent) worker(id int) {
	for {
		resp, err := http.Get(a.url + "/internal/task")
		if err != nil {
			log.Printf("Демон %d: ошибка получения задачи: %v", id, err)
			time.Sleep(1 * time.Second)
			continue
		}
		log.Printf("Демон %d: GET /internal/task → %d", id, resp.StatusCode)
		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			time.Sleep(1 * time.Second)
			continue
		}

		var taskResponse struct {
			Task struct {
				ID            string  `json:"id"`
				Arg1          float64 `json:"arg1"`
				Arg2          float64 `json:"arg2"`
				Operation     string  `json:"operation"`
				OperationTime int     `json:"operation_time"`
			} `json:"task"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&taskResponse); err != nil {
			log.Printf("Демон %d: ошибка парсинга задачи: %v", id, err)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		time.Sleep(time.Duration(taskResponse.Task.OperationTime) * time.Millisecond)

		result, err := calculation.Compute(
			taskResponse.Task.Operation,
			taskResponse.Task.Arg1,
			taskResponse.Task.Arg2,
		)
		if err != nil {
			log.Printf("Демон %d: ошибка вычисления: %v", id, err)
			continue
		}

		response := map[string]interface{}{
			"task_id": taskResponse.Task.ID,
			"result":  result,
		}

		jsonResp, _ := json.Marshal(response)
		fmt.Println(result)

		respPost, err := http.Post(
			a.url+"/internal/task",
			"application/json",
			bytes.NewBuffer(jsonResp),
		)
		if err != nil {
			log.Printf("Демон %d: ошибка POST результата: %v", id, err)
		} else {
			log.Printf("Демон %d: POST /internal/task → %d", id, respPost.StatusCode)
		}
		
		log.Printf("Демон %d: POST /internal/task → %d", id, respPost.StatusCode)
		if err != nil {
			log.Printf("Демон %d: ошибка отправки результата: %v", id, err)
			continue
		}

		if respPost.StatusCode != http.StatusOK {
			log.Printf("Демон %d: сервер вернул статус %d", id, respPost.StatusCode)
			continue
		}
		log.Printf("Демон %d: успешно обработал задачу %s", id, taskResponse.Task.ID)
	}
}
