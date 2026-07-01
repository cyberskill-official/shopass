package main

import (
	"log"

	_ "shopass/services/track/internal/engine"
	// "shopass/services/track/internal/track"
	// "shopass/services/price/internal/price"
)

func main() {
	log.Println("Starting tracksvc")
	
	// Khởi tạo các repository và service
	// priceSvc := price.NewService(...)
	// ruleRepo := track.NewAlertRuleRepo(...)
	// dealSvc := deal.NewService(...)
	// stateRepo := engine.NewStateRepo(...)
	// handoffSvc := engine.NewHandoffService(...)
	
	// e := engine.NewEngine(ruleRepo, priceSvc, dealSvc, stateRepo, handoffSvc)
	
	// Đăng ký consumer (ví dụ: RabbitMQ / Kafka / Go Channel)
	// consumer.On("price.written", func(snap engine.Snapshot) {
	//     _ = e.EvaluateForProduct(context.Background(), snap)
	// })
}
