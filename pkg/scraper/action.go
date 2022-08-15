package scraper

import (
	"context"
	"net/http"

	"github.com/stashapp/stash/pkg/models"
)

type scraperAction string

const (
	scraperActionScript scraperAction = "script"
	scraperActionStash  scraperAction = "stash"
	scraperActionXPath  scraperAction = "scrapeXPath"
	scraperActionJson   scraperAction = "scrapeJson"
)

func (e scraperAction) IsValid() bool {
	switch e {
	case scraperActionScript, scraperActionStash, scraperActionXPath, scraperActionJson:
		return true
	}
	return false
}

type scraperActionImpl interface {
	scrapeByURL(ctx context.Context, url string, ty models.ScrapeContentType) (models.ScrapedContent, error)
	scrapeByName(ctx context.Context, name string, ty models.ScrapeContentType) ([]models.ScrapedContent, error)
	scrapeByFragment(ctx context.Context, input Input) (models.ScrapedContent, error)

	scrapeSceneByScene(ctx context.Context, scene *models.Scene) (*models.ScrapedScene, error)
	scrapeGalleryByGallery(ctx context.Context, gallery *models.Gallery) (*models.ScrapedGallery, error)
}

func (c config) getScraper(scraper scraperTypeConfig, client *http.Client, txnManager models.TransactionManager, globalConfig GlobalConfig) scraperActionImpl {
	switch scraper.Action {
	case scraperActionScript:
		return newScriptScraper(scraper, c, globalConfig)
	case scraperActionStash:
		return newStashScraper(scraper, client, txnManager, c, globalConfig)
	case scraperActionXPath:
		return newXpathScraper(scraper, client, txnManager, c, globalConfig)
	case scraperActionJson:
		return newJsonScraper(scraper, client, txnManager, c, globalConfig)
	}

	panic("unknown scraper action: " + scraper.Action)
}
