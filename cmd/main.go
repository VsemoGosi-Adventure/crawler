package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/VsemoGosi-Adventure/crawler/internal/config"
	"github.com/VsemoGosi-Adventure/crawler/internal/producer"
	"github.com/VsemoGosi-Adventure/crawler/internal/provider"
	"github.com/VsemoGosi-Adventure/crawler/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	prod := producer.NewProducer(cfg.ProducerConfig)
	defer prod.Close()

	timepadClient := provider.NewTimepadClient(cfg.TimepadCity, cfg.TimepadToken)

	crawlerSvc := service.NewCrawlerService(
		prod,
		cfg.CrawlerInterval,
		timepadClient,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutdown signal received...")
		cancel()
	}()

	crawlerSvc.Start(ctx)
}
