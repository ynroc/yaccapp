package scene

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/stashapp/stash/pkg/models"
	"github.com/stashapp/stash/pkg/models/jsonschema"
)

type MarkerImporter struct {
	SceneID             int
	ReaderWriter        models.SceneMarkerReaderWriter
	TagWriter           models.TagReaderWriter
	Input               jsonschema.SceneMarker
	MissingRefBehaviour models.ImportMissingRefEnum

	tags   []*models.Tag
	marker models.SceneMarker
}

func (i *MarkerImporter) PreImport() error {
	seconds, _ := strconv.ParseFloat(i.Input.Seconds, 64)
	i.marker = models.SceneMarker{
		Title:     i.Input.Title,
		Seconds:   seconds,
		SceneID:   sql.NullInt64{Int64: int64(i.SceneID), Valid: true},
		CreatedAt: models.SQLiteTimestamp{Timestamp: i.Input.CreatedAt.GetTime()},
		UpdatedAt: models.SQLiteTimestamp{Timestamp: i.Input.UpdatedAt.GetTime()},
	}

	if err := i.populateTags(); err != nil {
		return err
	}

	return nil
}

func (i *MarkerImporter) populateTags() error {
	// primary tag cannot be ignored
	mrb := i.MissingRefBehaviour
	if mrb == models.ImportMissingRefEnumIgnore {
		mrb = models.ImportMissingRefEnumFail
	}

	primaryTag, err := importTags(i.TagWriter, []string{i.Input.PrimaryTag}, mrb)
	if err != nil {
		return err
	}

	i.marker.PrimaryTagID = primaryTag[0].ID

	if len(i.Input.Tags) > 0 {
		tags, err := importTags(i.TagWriter, i.Input.Tags, i.MissingRefBehaviour)
		if err != nil {
			return err
		}

		i.tags = tags
	}

	return nil
}

func (i *MarkerImporter) PostImport(id int) error {
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

func (i *MarkerImporter) Name() string {
	return fmt.Sprintf("%s (%s)", i.Input.Title, i.Input.Seconds)
}

func (i *MarkerImporter) FindExistingID() (*int, error) {
	existingMarkers, err := i.ReaderWriter.FindBySceneID(i.SceneID)

	if err != nil {
		return nil, err
	}

	for _, m := range existingMarkers {
		if m.Seconds == i.marker.Seconds {
			id := m.ID
			return &id, nil
		}
	}

	return nil, nil
}

func (i *MarkerImporter) Create() (*int, error) {
	created, err := i.ReaderWriter.Create(i.marker)
	if err != nil {
		return nil, fmt.Errorf("error creating marker: %v", err)
	}

	id := created.ID
	return &id, nil
}

func (i *MarkerImporter) Update(id int) error {
	marker := i.marker
	marker.ID = id
	_, err := i.ReaderWriter.Update(marker)
	if err != nil {
		return fmt.Errorf("error updating existing marker: %v", err)
	}

	return nil
}
