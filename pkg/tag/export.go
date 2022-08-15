package tag

import (
	"fmt"

	"github.com/stashapp/stash/pkg/models"
	"github.com/stashapp/stash/pkg/models/json"
	"github.com/stashapp/stash/pkg/models/jsonschema"
	"github.com/stashapp/stash/pkg/utils"
)

// ToJSON converts a Tag object into its JSON equivalent.
func ToJSON(reader models.TagReader, tag *models.Tag) (*jsonschema.Tag, error) {
	newTagJSON := jsonschema.Tag{
		Name:          tag.Name,
		IgnoreAutoTag: tag.IgnoreAutoTag,
		CreatedAt:     json.JSONTime{Time: tag.CreatedAt.Timestamp},
		UpdatedAt:     json.JSONTime{Time: tag.UpdatedAt.Timestamp},
	}

	aliases, err := reader.GetAliases(tag.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting tag aliases: %v", err)
	}

	newTagJSON.Aliases = aliases

	image, err := reader.GetImage(tag.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting tag image: %v", err)
	}

	if len(image) > 0 {
		newTagJSON.Image = utils.GetBase64StringFromData(image)
	}

	parents, err := reader.FindByChildTagID(tag.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting parents: %v", err)
	}

	newTagJSON.Parents = GetNames(parents)

	return &newTagJSON, nil
}

func GetIDs(tags []*models.Tag) []int {
	var results []int
	for _, tag := range tags {
		results = append(results, tag.ID)
	}

	return results
}

func GetNames(tags []*models.Tag) []string {
	var results []string
	for _, tag := range tags {
		results = append(results, tag.Name)
	}

	return results
}
