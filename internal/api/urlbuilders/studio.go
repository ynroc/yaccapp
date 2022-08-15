package urlbuilders

import (
	"github.com/stashapp/stash/pkg/models"
	"strconv"
)

type StudioURLBuilder struct {
	BaseURL   string
	StudioID  string
	UpdatedAt string
}

func NewStudioURLBuilder(baseURL string, studio *models.Studio) StudioURLBuilder {
	return StudioURLBuilder{
		BaseURL:   baseURL,
		StudioID:  strconv.Itoa(studio.ID),
		UpdatedAt: strconv.FormatInt(studio.UpdatedAt.Timestamp.Unix(), 10),
	}
}

func (b StudioURLBuilder) GetStudioImageURL() string {
	return b.BaseURL + "/studio/" + b.StudioID + "/image?" + b.UpdatedAt
}
