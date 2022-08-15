package autotag

import (
	"testing"

	"github.com/stashapp/stash/pkg/models"
	"github.com/stashapp/stash/pkg/models/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const imageExt = "jpg"

func TestImagePerformers(t *testing.T) {
	t.Parallel()

	const imageID = 1
	const performerName = "performer name"
	const performerID = 2
	performer := models.Performer{
		ID:   performerID,
		Name: models.NullString(performerName),
	}

	const reversedPerformerName = "name performer"
	const reversedPerformerID = 3
	reversedPerformer := models.Performer{
		ID:   reversedPerformerID,
		Name: models.NullString(reversedPerformerName),
	}

	testTables := generateTestTable(performerName, imageExt)

	assert := assert.New(t)

	for _, test := range testTables {
		mockPerformerReader := &mocks.PerformerReaderWriter{}
		mockImageReader := &mocks.ImageReaderWriter{}

		mockPerformerReader.On("Query", mock.Anything, mock.Anything).Return(nil, 0, nil)
		mockPerformerReader.On("QueryForAutoTag", mock.Anything).Return([]*models.Performer{&performer, &reversedPerformer}, nil).Once()

		if test.Matches {
			mockImageReader.On("GetPerformerIDs", imageID).Return(nil, nil).Once()
			mockImageReader.On("UpdatePerformers", imageID, []int{performerID}).Return(nil).Once()
		}

		image := models.Image{
			ID:   imageID,
			Path: test.Path,
		}
		err := ImagePerformers(&image, mockImageReader, mockPerformerReader, nil)

		assert.Nil(err)
		mockPerformerReader.AssertExpectations(t)
		mockImageReader.AssertExpectations(t)
	}
}

func TestImageStudios(t *testing.T) {
	t.Parallel()

	const imageID = 1
	const studioName = "studio name"
	const studioID = 2
	studio := models.Studio{
		ID:   studioID,
		Name: models.NullString(studioName),
	}

	const reversedStudioName = "name studio"
	const reversedStudioID = 3
	reversedStudio := models.Studio{
		ID:   reversedStudioID,
		Name: models.NullString(reversedStudioName),
	}

	testTables := generateTestTable(studioName, imageExt)

	assert := assert.New(t)

	doTest := func(mockStudioReader *mocks.StudioReaderWriter, mockImageReader *mocks.ImageReaderWriter, test pathTestTable) {
		if test.Matches {
			mockImageReader.On("Find", imageID).Return(&models.Image{}, nil).Once()
			expectedStudioID := models.NullInt64(studioID)
			mockImageReader.On("Update", models.ImagePartial{
				ID:       imageID,
				StudioID: &expectedStudioID,
			}).Return(nil, nil).Once()
		}

		image := models.Image{
			ID:   imageID,
			Path: test.Path,
		}
		err := ImageStudios(&image, mockImageReader, mockStudioReader, nil)

		assert.Nil(err)
		mockStudioReader.AssertExpectations(t)
		mockImageReader.AssertExpectations(t)
	}

	for _, test := range testTables {
		mockStudioReader := &mocks.StudioReaderWriter{}
		mockImageReader := &mocks.ImageReaderWriter{}

		mockStudioReader.On("Query", mock.Anything, mock.Anything).Return(nil, 0, nil)
		mockStudioReader.On("QueryForAutoTag", mock.Anything).Return([]*models.Studio{&studio, &reversedStudio}, nil).Once()
		mockStudioReader.On("GetAliases", mock.Anything).Return([]string{}, nil).Maybe()

		doTest(mockStudioReader, mockImageReader, test)
	}

	// test against aliases
	const unmatchedName = "unmatched"
	studio.Name.String = unmatchedName

	for _, test := range testTables {
		mockStudioReader := &mocks.StudioReaderWriter{}
		mockImageReader := &mocks.ImageReaderWriter{}

		mockStudioReader.On("Query", mock.Anything, mock.Anything).Return(nil, 0, nil)
		mockStudioReader.On("QueryForAutoTag", mock.Anything).Return([]*models.Studio{&studio, &reversedStudio}, nil).Once()
		mockStudioReader.On("GetAliases", studioID).Return([]string{
			studioName,
		}, nil).Once()
		mockStudioReader.On("GetAliases", reversedStudioID).Return([]string{}, nil).Once()

		doTest(mockStudioReader, mockImageReader, test)
	}
}

func TestImageTags(t *testing.T) {
	t.Parallel()

	const imageID = 1
	const tagName = "tag name"
	const tagID = 2
	tag := models.Tag{
		ID:   tagID,
		Name: tagName,
	}

	const reversedTagName = "name tag"
	const reversedTagID = 3
	reversedTag := models.Tag{
		ID:   reversedTagID,
		Name: reversedTagName,
	}

	testTables := generateTestTable(tagName, imageExt)

	assert := assert.New(t)

	doTest := func(mockTagReader *mocks.TagReaderWriter, mockImageReader *mocks.ImageReaderWriter, test pathTestTable) {
		if test.Matches {
			mockImageReader.On("GetTagIDs", imageID).Return(nil, nil).Once()
			mockImageReader.On("UpdateTags", imageID, []int{tagID}).Return(nil).Once()
		}

		image := models.Image{
			ID:   imageID,
			Path: test.Path,
		}
		err := ImageTags(&image, mockImageReader, mockTagReader, nil)

		assert.Nil(err)
		mockTagReader.AssertExpectations(t)
		mockImageReader.AssertExpectations(t)
	}

	for _, test := range testTables {
		mockTagReader := &mocks.TagReaderWriter{}
		mockImageReader := &mocks.ImageReaderWriter{}

		mockTagReader.On("Query", mock.Anything, mock.Anything).Return(nil, 0, nil)
		mockTagReader.On("QueryForAutoTag", mock.Anything).Return([]*models.Tag{&tag, &reversedTag}, nil).Once()
		mockTagReader.On("GetAliases", mock.Anything).Return([]string{}, nil).Maybe()

		doTest(mockTagReader, mockImageReader, test)
	}

	// test against aliases
	const unmatchedName = "unmatched"
	tag.Name = unmatchedName

	for _, test := range testTables {
		mockTagReader := &mocks.TagReaderWriter{}
		mockImageReader := &mocks.ImageReaderWriter{}

		mockTagReader.On("Query", mock.Anything, mock.Anything).Return(nil, 0, nil)
		mockTagReader.On("QueryForAutoTag", mock.Anything).Return([]*models.Tag{&tag, &reversedTag}, nil).Once()
		mockTagReader.On("GetAliases", tagID).Return([]string{
			tagName,
		}, nil).Once()
		mockTagReader.On("GetAliases", reversedTagID).Return([]string{}, nil).Once()

		doTest(mockTagReader, mockImageReader, test)
	}
}
