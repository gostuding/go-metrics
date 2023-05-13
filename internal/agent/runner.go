package agent

import (
	"log"
	"time"
)

// интерфейс для отправки и обновления данных
type Storager interface {
	UpdateMetrics()
	SendMetrics(string, int)
}

// бесконечный цикл отправки данных
func StartAgent(IP string, port int, updateTime int, sendTime int, storage Storager) {
	update := time.Now().Add(time.Duration(updateTime) * time.Second)
	send := time.Now().Add(time.Duration(sendTime) * time.Second)
	for {
		time.Sleep(1 * time.Second)
		if time.Now().After(update) {
			storage.UpdateMetrics()
			update = time.Now().Add(time.Duration(updateTime) * time.Second)
			log.Printf("Update %v \n", time.Duration(updateTime)*time.Second)
		}
		if time.Now().After(send) {
			storage.SendMetrics(IP, port)
			send = time.Now().Add(time.Duration(sendTime) * time.Second)
			log.Printf("Send %v\n", time.Duration(sendTime)*time.Second)
		}
	}
}
