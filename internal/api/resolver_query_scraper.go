package api

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/stashapp/stash/internal/manager"
	"github.com/stashapp/stash/internal/manager/config"
	"github.com/stashapp/stash/pkg/logger"
	"github.com/stashapp/stash/pkg/models"
	"github.com/stashapp/stash/pkg/scraper"
	"github.com/stashapp/stash/pkg/scraper/stashbox"
	"github.com/stashapp/stash/pkg/sliceutil/stringslice"
)

func (r *queryResolver) ScrapeURL(ctx context.Context, url string, ty models.ScrapeContentType) (models.ScrapedContent, error) {
	return r.scraperCache().ScrapeURL(ctx, url, ty)
}

// deprecated
func (r *queryResolver) ScrapeFreeonesPerformerList(ctx context.Context, query string) ([]string, error) {
	content, err := r.scraperCache().ScrapeName(ctx, scraper.FreeonesScraperID, query, models.ScrapeContentTypePerformer)

	if err != nil {
		return nil, err
	}

	performers, err := marshalScrapedPerformers(content)
	if err != nil {
		return nil, err
	}

	var ret []string
	for _, p := range performers {
		if p.Name != nil {
			ret = append(ret, *p.Name)
		}
	}

	return ret, nil
}

func (r *queryResolver) ListScrapers(ctx context.Context, types []models.ScrapeContentType) ([]*models.Scraper, error) {
	return r.scraperCache().ListScrapers(types), nil
}

func (r *queryResolver) ListPerformerScrapers(ctx context.Context) ([]*models.Scraper, error) {
	return r.scraperCache().ListScrapers([]models.ScrapeContentType{models.ScrapeContentTypePerformer}), nil
}

func (r *queryResolver) ListSceneScrapers(ctx context.Context) ([]*models.Scraper, error) {
	return r.scraperCache().ListScrapers([]models.ScrapeContentType{models.ScrapeContentTypeScene}), nil
}

func (r *queryResolver) ListGalleryScrapers(ctx context.Context) ([]*models.Scraper, error) {
	return r.scraperCache().ListScrapers([]models.ScrapeContentType{models.ScrapeContentTypeGallery}), nil
}

func (r *queryResolver) ListMovieScrapers(ctx context.Context) ([]*models.Scraper, error) {
	return r.scraperCache().ListScrapers([]models.ScrapeContentType{models.ScrapeContentTypeMovie}), nil
}

func (r *queryResolver) ScrapePerformerList(ctx context.Context, scraperID string, query string) ([]*models.ScrapedPerformer, error) {
	if query == "" {
		return nil, nil
	}

	content, err := r.scraperCache().ScrapeName(ctx, scraperID, query, models.ScrapeContentTypePerformer)
	if err != nil {
		return nil, err
	}

	return marshalScrapedPerformers(content)
}

func (r *queryResolver) ScrapePerformer(ctx context.Context, scraperID string, scrapedPerformer models.ScrapedPerformerInput) (*models.ScrapedPerformer, error) {
	content, err := r.scraperCache().ScrapeFragment(ctx, scraperID, scraper.Input{Performer: &scrapedPerformer})
	if err != nil {
		return nil, err
	}
	return marshalScrapedPerformer(content)
}

func (r *queryResolver) ScrapePerformerURL(ctx context.Context, url string) (*models.ScrapedPerformer, error) {
	content, err := r.scraperCache().ScrapeURL(ctx, url, models.ScrapeContentTypePerformer)
	if err != nil {
		return nil, err
	}

	return marshalScrapedPerformer(content)
}

func (r *queryResolver) ScrapeSceneQuery(ctx context.Context, scraperID string, query string) ([]*models.ScrapedScene, error) {
	if query == "" {
		return nil, nil
	}

	content, err := r.scraperCache().ScrapeName(ctx, scraperID, query, models.ScrapeContentTypeScene)
	if err != nil {
		return nil, err
	}

	ret, err := marshalScrapedScenes(content)
	if err != nil {
		return nil, err
	}

	filterSceneTags(ret)
	return ret, nil
}

func (r *queryResolver) ScrapeScene(ctx context.Context, scraperID string, scene models.SceneUpdateInput) (*models.ScrapedScene, error) {
	id, err := strconv.Atoi(scene.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: scene.ID is not an integer: '%s'", ErrInput, scene.ID)
	}

	content, err := r.scraperCache().ScrapeID(ctx, scraperID, id, models.ScrapeContentTypeScene)
	if err != nil {
		return nil, err
	}

	ret, err := marshalScrapedScene(content)
	if err != nil {
		return nil, err
	}

	filterSceneTags([]*models.ScrapedScene{ret})

	return ret, nil
}

// filterSceneTags removes tags matching excluded tag patterns from the provided scraped scenes
func filterSceneTags(scenes []*models.ScrapedScene) {
	excludePatterns := manager.GetInstance().Config.GetScraperExcludeTagPatterns()
	var excludeRegexps []*regexp.Regexp

	for _, excludePattern := range excludePatterns {
		reg, err := regexp.Compile(strings.ToLower(excludePattern))
		if err != nil {
			logger.Errorf("Invalid tag exclusion pattern: %v", err)
		} else {
			excludeRegexps = append(excludeRegexps, reg)
		}
	}

	if len(excludeRegexps) == 0 {
		return
	}

	var ignoredTags []string

	for _, s := range scenes {
		var newTags []*models.ScrapedTag
		for _, t := range s.Tags {
			ignore := false
			for _, reg := range excludeRegexps {
				if reg.MatchString(strings.ToLower(t.Name)) {
					ignore = true
					ignoredTags = stringslice.StrAppendUnique(ignoredTags, t.Name)
					break
				}
			}

			if !ignore {
				newTags = append(newTags, t)
			}
		}

		s.Tags = newTags
	}

	if len(ignoredTags) > 0 {
		logger.Debugf("Scraping ignored tags: %s", strings.Join(ignoredTags, ", "))
	}
}

func (r *queryResolver) ScrapeSceneURL(ctx context.Context, url string) (*models.ScrapedScene, error) {
	content, err := r.scraperCache().ScrapeURL(ctx, url, models.ScrapeContentTypeScene)
	if err != nil {
		return nil, err
	}

	ret, err := marshalScrapedScene(content)
	if err != nil {
		return nil, err
	}

	filterSceneTags([]*models.ScrapedScene{ret})

	return ret, nil
}

func (r *queryResolver) ScrapeGallery(ctx context.Context, scraperID string, gallery models.GalleryUpdateInput) (*models.ScrapedGallery, error) {
	id, err := strconv.Atoi(gallery.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: gallery id is not an integer: '%s'", ErrInput, gallery.ID)
	}

	content, err := r.scraperCache().ScrapeID(ctx, scraperID, id, models.ScrapeContentTypeGallery)
	if err != nil {
		return nil, err
	}

	return marshalScrapedGallery(content)
}

func (r *queryResolver) ScrapeGalleryURL(ctx context.Context, url string) (*models.ScrapedGallery, error) {
	content, err := r.scraperCache().ScrapeURL(ctx, url, models.ScrapeContentTypeGallery)
	if err != nil {
		return nil, err
	}

	return marshalScrapedGallery(content)
}

func (r *queryResolver) ScrapeMovieURL(ctx context.Context, url string) (*models.ScrapedMovie, error) {
	content, err := r.scraperCache().ScrapeURL(ctx, url, models.ScrapeContentTypeMovie)
	if err != nil {
		return nil, err
	}

	return marshalScrapedMovie(content)
}

func (r *queryResolver) getStashBoxClient(index int) (*stashbox.Client, error) {
	boxes := config.GetInstance().GetStashBoxes()

	if index < 0 || index >= len(boxes) {
		return nil, fmt.Errorf("%w: invalid stash_box_index %d", ErrInput, index)
	}

	return stashbox.NewClient(*boxes[index], r.txnManager), nil
}

func (r *queryResolver) ScrapeSingleScene(ctx context.Context, source models.ScraperSourceInput, input models.ScrapeSingleSceneInput) ([]*models.ScrapedScene, error) {
	var ret []*models.ScrapedScene

	var sceneID int
	if input.SceneID != nil {
		var err error
		sceneID, err = strconv.Atoi(*input.SceneID)
		if err != nil {
			return nil, fmt.Errorf("%w: sceneID is not an integer: '%s'", ErrInput, *input.SceneID)
		}
	}

	switch {
	case source.ScraperID != nil:
		var err error
		var c models.ScrapedContent
		var content []models.ScrapedContent

		switch {
		case input.SceneID != nil:
			c, err = r.scraperCache().ScrapeID(ctx, *source.ScraperID, sceneID, models.ScrapeContentTypeScene)
			if c != nil {
				content = []models.ScrapedContent{c}
			}
		case input.SceneInput != nil:
			c, err = r.scraperCache().ScrapeFragment(ctx, *source.ScraperID, scraper.Input{Scene: input.SceneInput})
			if c != nil {
				content = []models.ScrapedContent{c}
			}
		case input.Query != nil:
			content, err = r.scraperCache().ScrapeName(ctx, *source.ScraperID, *input.Query, models.ScrapeContentTypeScene)
		default:
			err = fmt.Errorf("%w: scene_id, scene_input, or query must be set", ErrInput)
		}

		if err != nil {
			return nil, err
		}

		ret, err = marshalScrapedScenes(content)
		if err != nil {
			return nil, err
		}
	case source.StashBoxIndex != nil:
		client, err := r.getStashBoxClient(*source.StashBoxIndex)
		if err != nil {
			return nil, err
		}

		switch {
		case input.SceneID != nil:
			ret, err = client.FindStashBoxSceneByFingerprints(ctx, sceneID)
		case input.Query != nil:
			ret, err = client.QueryStashBoxScene(ctx, *input.Query)
		default:
			return nil, fmt.Errorf("%w: scene_id or query must be set", ErrInput)
		}

		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("%w: scraper_id or stash_box_index must be set", ErrInput)
	}

	filterSceneTags(ret)

	return ret, nil
}

func (r *queryResolver) ScrapeMultiScenes(ctx context.Context, source models.ScraperSourceInput, input models.ScrapeMultiScenesInput) ([][]*models.ScrapedScene, error) {
	if source.ScraperID != nil {
		return nil, ErrNotImplemented
	} else if source.StashBoxIndex != nil {
		client, err := r.getStashBoxClient(*source.StashBoxIndex)
		if err != nil {
			return nil, err
		}

		sceneIDs, err := stringslice.StringSliceToIntSlice(input.SceneIds)
		if err != nil {
			return nil, err
		}

		return client.FindStashBoxScenesByFingerprints(ctx, sceneIDs)
	}

	return nil, errors.New("scraper_id or stash_box_index must be set")
}

func (r *queryResolver) ScrapeSinglePerformer(ctx context.Context, source models.ScraperSourceInput, input models.ScrapeSinglePerformerInput) ([]*models.ScrapedPerformer, error) {
	if source.ScraperID != nil {
		if input.PerformerInput != nil {
			performer, err := r.scraperCache().ScrapeFragment(ctx, *source.ScraperID, scraper.Input{Performer: input.PerformerInput})
			if err != nil {
				return nil, err
			}

			return marshalScrapedPerformers([]models.ScrapedContent{performer})
		}

		if input.Query != nil {
			content, err := r.scraperCache().ScrapeName(ctx, *source.ScraperID, *input.Query, models.ScrapeContentTypePerformer)
			if err != nil {
				return nil, err
			}

			return marshalScrapedPerformers(content)
		}

		return nil, ErrNotImplemented
	} else if source.StashBoxIndex != nil {
		client, err := r.getStashBoxClient(*source.StashBoxIndex)
		if err != nil {
			return nil, err
		}

		var ret []*models.StashBoxPerformerQueryResult
		switch {
		case input.PerformerID != nil:
			ret, err = client.FindStashBoxPerformersByNames(ctx, []string{*input.PerformerID})
		case input.Query != nil:
			ret, err = client.QueryStashBoxPerformer(ctx, *input.Query)
		default:
			return nil, ErrNotImplemented
		}

		if err != nil {
			return nil, err
		}

		if len(ret) > 0 {
			return ret[0].Results, nil
		}

		return nil, nil
	}

	return nil, errors.New("scraper_id or stash_box_index must be set")
}

func (r *queryResolver) ScrapeMultiPerformers(ctx context.Context, source models.ScraperSourceInput, input models.ScrapeMultiPerformersInput) ([][]*models.ScrapedPerformer, error) {
	if source.ScraperID != nil {
		return nil, ErrNotImplemented
	} else if source.StashBoxIndex != nil {
		client, err := r.getStashBoxClient(*source.StashBoxIndex)
		if err != nil {
			return nil, err
		}

		return client.FindStashBoxPerformersByPerformerNames(ctx, input.PerformerIds)
	}

	return nil, errors.New("scraper_id or stash_box_index must be set")
}

func (r *queryResolver) ScrapeSingleGallery(ctx context.Context, source models.ScraperSourceInput, input models.ScrapeSingleGalleryInput) ([]*models.ScrapedGallery, error) {
	if source.StashBoxIndex != nil {
		return nil, ErrNotSupported
	}

	if source.ScraperID == nil {
		return nil, fmt.Errorf("%w: scraper_id must be set", ErrInput)
	}

	var c models.ScrapedContent

	switch {
	case input.GalleryID != nil:
		galleryID, err := strconv.Atoi(*input.GalleryID)
		if err != nil {
			return nil, fmt.Errorf("%w: gallery id is not an integer: '%s'", ErrInput, *input.GalleryID)
		}
		c, err = r.scraperCache().ScrapeID(ctx, *source.ScraperID, galleryID, models.ScrapeContentTypeGallery)
		if err != nil {
			return nil, err
		}
		return marshalScrapedGalleries([]models.ScrapedContent{c})
	case input.GalleryInput != nil:
		c, err := r.scraperCache().ScrapeFragment(ctx, *source.ScraperID, scraper.Input{Gallery: input.GalleryInput})
		if err != nil {
			return nil, err
		}
		return marshalScrapedGalleries([]models.ScrapedContent{c})
	default:
		return nil, ErrNotImplemented
	}
}

func (r *queryResolver) ScrapeSingleMovie(ctx context.Context, source models.ScraperSourceInput, input models.ScrapeSingleMovieInput) ([]*models.ScrapedMovie, error) {
	return nil, ErrNotSupported
}
