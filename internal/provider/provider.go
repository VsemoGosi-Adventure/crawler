package provider

import (
	"context"

	"github.com/VsemoGosi-Adventure/crawler/internal/dto"
)

type EventProvider interface {
	Name() string
	FetchEvents(ctx context.Context) ([]dto.CrawledEventMessage, error)
}
