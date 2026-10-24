package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/VsemoGosi-Adventure/crawler/internal/config"
	"github.com/VsemoGosi-Adventure/crawler/internal/producer"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	producer := producer.NewProducer(cfg.ProducerConfig)
	defer producer.Close()

	// crawlerSvc := service.NewCrawlerService(
	// 	producer,

	// )

	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutdown signal received...")
		// cancel()
	}()
}
