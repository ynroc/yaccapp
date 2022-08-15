package image

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/stashapp/stash/pkg/models"
	"github.com/stashapp/stash/pkg/models/jsonschema"
	"github.com/stashapp/stash/pkg/sliceutil/stringslice"
)

type Importer struct {
	ReaderWriter        models.ImageReaderWriter
	StudioWriter        models.StudioReaderWriter
	GalleryWriter       models.GalleryReaderWriter
	PerformerWriter     models.PerformerReaderWriter
	TagWriter           models.TagReaderWriter
	Input               jsonschema.Image
	Path                string
	MissingRefBehaviour models.ImportMissingRefEnum

	ID         int
	image      models.Image
	galleries  []*models.Gallery
	performers []*models.Performer
	tags       []*models.Tag
}

func (i *Importer) PreImport() error {
	i.image = i.imageJSONToImage(i.Input)

	if err := i.populateStudio(); err != nil {
		return err
	}

	if err := i.populateGalleries(); err != nil {
		return err
	}

	if err := i.populatePerformers(); err != nil {
		return err
	}

	if err := i.populateTags(); err != nil {
		return err
	}

	return nil
}

func (i *Importer) imageJSONToImage(imageJSON jsonschema.Image) models.Image {
	newImage := models.Image{
		Checksum: imageJSON.Checksum,
		Path:     i.Path,
	}

	if imageJSON.Title != "" {
		newImage.Title = sql.NullString{String: imageJSON.Title, Valid: true}
	}
	if imageJSON.Rating != 0 {
		newImage.Rating = sql.NullInt64{Int64: int64(imageJSON.Rating), Valid: true}
	}

	newImage.Organized = imageJSON.Organized
	newImage.OCounter = imageJSON.OCounter
	newImage.CreatedAt = models.SQLiteTimestamp{Timestamp: imageJSON.CreatedAt.GetTime()}
	newImage.UpdatedAt = models.SQLiteTimestamp{Timestamp: imageJSON.UpdatedAt.GetTime()}

	if imageJSON.File != nil {
		if imageJSON.File.Size != 0 {
			newImage.Size = sql.NullInt64{Int64: int64(imageJSON.File.Size), Valid: true}
		}
		if imageJSON.File.Width != 0 {
			newImage.Width = sql.NullInt64{Int64: int64(imageJSON.File.Width), Valid: true}
		}
		if imageJSON.File.Height != 0 {
			newImage.Height = sql.NullInt64{Int64: int64(imageJSON.File.Height), Valid: true}
		}
	}

	return newImage
}

func (i *Importer) populateStudio() error {
	if i.Input.Studio != "" {
		studio, err := i.StudioWriter.FindByName(i.Input.Studio, false)
		if err != nil {
			return fmt.Errorf("error finding studio by name: %v", err)
		}

		if studio == nil {
			if i.MissingRefBehaviour == models.ImportMissingRefEnumFail {
				return fmt.Errorf("image studio '%s' not found", i.Input.Studio)
			}

			if i.MissingRefBehaviour == models.ImportMissingRefEnumIgnore {
				return nil
			}

			if i.MissingRefBehaviour == models.ImportMissingRefEnumCreate {
				studioID, err := i.createStudio(i.Input.Studio)
				if err != nil {
					return err
				}
				i.image.StudioID = sql.NullInt64{
					Int64: int64(studioID),
					Valid: true,
				}
			}
		} else {
			i.image.StudioID = sql.NullInt64{Int64: int64(studio.ID), Valid: true}
		}
	}

	return nil
}

func (i *Importer) createStudio(name string) (int, error) {
	newStudio := *models.NewStudio(name)

	created, err := i.StudioWriter.Create(newStudio)
	if err != nil {
		return 0, err
	}

	return created.ID, nil
}

func (i *Importer) populateGalleries() error {
	for _, checksum := range i.Input.Galleries {
		gallery, err := i.GalleryWriter.FindByChecksum(checksum)
		if err != nil {
			return fmt.Errorf("error finding gallery: %v", err)
		}

		if gallery == nil {
			if i.MissingRefBehaviour == models.ImportMissingRefEnumFail {
				return fmt.Errorf("image gallery '%s' not found", i.Input.Studio)
			}

			// we don't create galleries - just ignore
			if i.MissingRefBehaviour == models.ImportMissingRefEnumIgnore || i.MissingRefBehaviour == models.ImportMissingRefEnumCreate {
				continue
			}
		} else {
			i.galleries = append(i.galleries, gallery)
		}
	}

	return nil
}

func (i *Importer) populatePerformers() error {
	if len(i.Input.Performers) > 0 {
		names := i.Input.Performers
		performers, err := i.PerformerWriter.FindByNames(names, false)
		if err != nil {
			return err
		}

		var pluckedNames []string
		for _, performer := range performers {
			if !performer.Name.Valid {
				continue
			}
			pluckedNames = append(pluckedNames, performer.Name.String)
		}

		missingPerformers := stringslice.StrFilter(names, func(name string) bool {
			return !stringslice.StrInclude(pluckedNames, name)
		})

		if len(missingPerformers) > 0 {
			if i.MissingRefBehaviour == models.ImportMissingRefEnumFail {
				return fmt.Errorf("image performers [%s] not found", strings.Join(missingPerformers, ", "))
			}

			if i.MissingRefBehaviour == models.ImportMissingRefEnumCreate {
				createdPerformers, err := i.createPerformers(missingPerformers)
				if err != nil {
					return fmt.Errorf("error creating image performers: %v", err)
				}

				performers = append(performers, createdPerformers...)
			}

			// ignore if MissingRefBehaviour set to Ignore
		}

		i.performers = performers
	}

	return nil
}

func (i *Importer) createPerformers(names []string) ([]*models.Performer, error) {
	var ret []*models.Performer
	for _, name := range names {
		newPerformer := *models.NewPerformer(name)

		created, err := i.PerformerWriter.Create(newPerformer)
		if err != nil {
			return nil, err
		}

		ret = append(ret, created)
	}

	return ret, nil
}

func (i *Importer) populateTags() error {
	if len(i.Input.Tags) > 0 {

		tags, err := importTags(i.TagWriter, i.Input.Tags, i.MissingRefBehaviour)
		if err != nil {
			return err
		}

		i.tags = tags
	}

	return nil
}

func (i *Importer) PostImport(id int) error {
	if len(i.galleries) > 0 {
		var galleryIDs []int
		for _, g := range i.galleries {
			galleryIDs = append(galleryIDs, g.ID)
		}

		if err := i.ReaderWriter.UpdateGalleries(id, galleryIDs); err != nil {
			return fmt.Errorf("failed to associate galleries: %v", err)
		}
	}

	if len(i.performers) > 0 {
		var performerIDs []int
		for _, performer := range i.performers {
			performerIDs = append(performerIDs, performer.ID)
		}

		if err := i.ReaderWriter.UpdatePerformers(id, performerIDs); err != nil {
			return fmt.Errorf("failed to associate performers: %v", err)
		}
	}

	if len(i.tags) > 0 {
		var tagIDs []int
		for _, t := range i.tags {
			tagIDs = append(tagIDs, t.ID)
		}
		if err := i.ReaderWriter.UpdateTags(id, tagIDs); err != nil {
			return fmt.Errorf("failed to associate tags: %v", err)
		}
	}

	return nil
}

func (i *Importer) Name() string {
	return i.Path
}

func (i *Importer) FindExistingID() (*int, error) {
	var existing *models.Image
	var err error
	existing, err = i.ReaderWriter.FindByChecksum(i.Input.Checksum)

	if err != nil {
		return nil, err
	}

	if existing != nil {
		id := existing.ID
		return &id, nil
	}

	return nil, nil
}

func (i *Importer) Create() (*int, error) {
	created, err := i.ReaderWriter.Create(i.image)
	if err != nil {
		return nil, fmt.Errorf("error creating image: %v", err)
	}

	id := created.ID
	i.ID = id
	return &id, nil
}

func (i *Importer) Update(id int) error {
	image := i.image
	image.ID = id
	i.ID = id
	_, err := i.ReaderWriter.UpdateFull(image)
	if err != nil {
		return fmt.Errorf("error updating existing image: %v", err)
	}

	return nil
}

func importTags(tagWriter models.TagReaderWriter, names []string, missingRefBehaviour models.ImportMissingRefEnum) ([]*models.Tag, error) {
	tags, err := tagWriter.FindByNames(names, false)
	if err != nil {
		return nil, err
	}

	var pluckedNames []string
	for _, tag := range tags {
		pluckedNames = append(pluckedNames, tag.Name)
	}

	missingTags := stringslice.StrFilter(names, func(name string) bool {
		return !stringslice.StrInclude(pluckedNames, name)
	})

	if len(missingTags) > 0 {
		if missingRefBehaviour == models.ImportMissingRefEnumFail {
			return nil, fmt.Errorf("tags [%s] not found", strings.Join(missingTags, ", "))
		}

		if missingRefBehaviour == models.ImportMissingRefEnumCreate {
			createdTags, err := createTags(tagWriter, missingTags)
			if err != nil {
				return nil, fmt.Errorf("error creating tags: %v", err)
			}

			tags = append(tags, createdTags...)
		}

		// ignore if MissingRefBehaviour set to Ignore
	}

	return tags, nil
}

func createTags(tagWriter models.TagWriter, names []string) ([]*models.Tag, error) {
	var ret []*models.Tag
	for _, name := range names {
		newTag := *models.NewTag(name)

		created, err := tagWriter.Create(newTag)
		if err != nil {
			return nil, err
		}

		ret = append(ret, created)
	}

	return ret, nil
}
