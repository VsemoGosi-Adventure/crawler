package service

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/VsemoGosi-Adventure/crawler/internal/producer"
	"github.com/VsemoGosi-Adventure/crawler/internal/provider"
)

type CrawlerService struct {
	providers []provider.EventProvider
	producer  producer.EventProducer
	interval  time.Duration
}

func NewCrawlerService(
	prod producer.EventProducer,
	interval time.Duration,
	providers ...provider.EventProvider,
) *CrawlerService {
	return &CrawlerService{
		providers: providers,
		producer:  prod,
		interval:  interval,
	}
}

func (s *CrawlerService) Start(ctx context.Context) {
	log.Printf("Crawler service started with interval: %v", s.interval)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.RunIteration(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("Crawler service stopping...")
			return
		case <-ticker.C:
			s.RunIteration(ctx)
		}
	}
}

func (s *CrawlerService) RunIteration(ctx context.Context) {
	log.Println("Starting data collection iteration across all providers...")

	var wg sync.WaitGroup

	for _, p := range s.providers {
		wg.Add(1)
		go func(prov provider.EventProvider) {
			defer wg.Done()

			name := prov.Name()
			events, err := prov.FetchEvents(ctx)
			if err != nil {
				log.Printf("[%s] Failed to fetch events: %v", name, err)
				return
			}

			for _, ev := range events {
				if err := s.producer.PublishEvent(ctx, ev); err != nil {
					log.Printf("[%s] Failed to publish event to Redpanda: %v", name, err)
				}
			}

			log.Printf("[%s] Successfully processed and published %d events", name, len(events))
		}(p)
	}

	wg.Wait()
	log.Println("All providers finished iteration.")
}
