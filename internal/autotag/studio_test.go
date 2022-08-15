package autotag

import (
	"testing"

	"github.com/stashapp/stash/pkg/image"
	"github.com/stashapp/stash/pkg/models"
	"github.com/stashapp/stash/pkg/models/mocks"
	"github.com/stashapp/stash/pkg/scene"
	"github.com/stretchr/testify/assert"
)

type testStudioCase struct {
	studioName    string
	expectedRegex string
	aliasName     string
	aliasRegex    string
}

var testStudioCases = []testStudioCase{
	{
		"studio name",
		`(?i)(?:^|_|[^\p{L}\d])studio[.\-_ ]*name(?:$|_|[^\p{L}\d])`,
		"",
		"",
	},
	{
		"studio + name",
		`(?i)(?:^|_|[^\p{L}\d])studio[.\-_ ]*\+[.\-_ ]*name(?:$|_|[^\p{L}\d])`,
		"",
		"",
	},
	{
		`studio + name\`,
		`(?i)(?:^|_|[^\p{L}\d])studio[.\-_ ]*\+[.\-_ ]*name\\(?:$|_|[^\p{L}\d])`,
		"",
		"",
	},
	{
		"studio name",
		`(?i)(?:^|_|[^\p{L}\d])studio[.\-_ ]*name(?:$|_|[^\p{L}\d])`,
		"alias name",
		`(?i)(?:^|_|[^\p{L}\d])alias[.\-_ ]*name(?:$|_|[^\p{L}\d])`,
	},
	{
		"studio + name",
		`(?i)(?:^|_|[^\p{L}\d])studio[.\-_ ]*\+[.\-_ ]*name(?:$|_|[^\p{L}\d])`,
		"alias + name",
		`(?i)(?:^|_|[^\p{L}\d])alias[.\-_ ]*\+[.\-_ ]*name(?:$|_|[^\p{L}\d])`,
	},
	{
		`studio + name\`,
		`(?i)(?:^|_|[^\p{L}\d])studio[.\-_ ]*\+[.\-_ ]*name\\(?:$|_|[^\p{L}\d])`,
		`alias + name\`,
		`(?i)(?:^|_|[^\p{L}\d])alias[.\-_ ]*\+[.\-_ ]*name\\(?:$|_|[^\p{L}\d])`,
	},
}

func TestStudioScenes(t *testing.T) {
	t.Parallel()

	for _, p := range testStudioCases {
		testStudioScenes(t, p)
	}
}

func testStudioScenes(t *testing.T, tc testStudioCase) {
	studioName := tc.studioName
	expectedRegex := tc.expectedRegex
	aliasName := tc.aliasName
	aliasRegex := tc.aliasRegex

	mockSceneReader := &mocks.SceneReaderWriter{}

	const studioID = 2

	var aliases []string

	testPathName := studioName
	if aliasName != "" {
		aliases = []string{aliasName}
		testPathName = aliasName
	}

	matchingPaths, falsePaths := generateTestPaths(testPathName, "mp4")

	var scenes []*models.Scene
	for i, p := range append(matchingPaths, falsePaths...) {
		scenes = append(scenes, &models.Scene{
			ID:   i + 1,
			Path: p,
		})
	}

	studio := models.Studio{
		ID:   studioID,
		Name: models.NullString(studioName),
	}

	organized := false
	perPage := models.PerPageAll

	expectedSceneFilter := &models.SceneFilterType{
		Organized: &organized,
		Path: &models.StringCriterionInput{
			Value:    expectedRegex,
			Modifier: models.CriterionModifierMatchesRegex,
		},
	}

	expectedFindFilter := &models.FindFilterType{
		PerPage: &perPage,
	}

	// if alias provided, then don't find by name
	onNameQuery := mockSceneReader.On("Query", scene.QueryOptions(expectedSceneFilter, expectedFindFilter, false))

	if aliasName == "" {
		onNameQuery.Return(mocks.SceneQueryResult(scenes, len(scenes)), nil).Once()
	} else {
		onNameQuery.Return(mocks.SceneQueryResult(nil, 0), nil).Once()

		expectedAliasFilter := &models.SceneFilterType{
			Organized: &organized,
			Path: &models.StringCriterionInput{
				Value:    aliasRegex,
				Modifier: models.CriterionModifierMatchesRegex,
			},
		}

		mockSceneReader.On("Query", scene.QueryOptions(expectedAliasFilter, expectedFindFilter, false)).
			Return(mocks.SceneQueryResult(scenes, len(scenes)), nil).Once()
	}

	for i := range matchingPaths {
		sceneID := i + 1
		mockSceneReader.On("Find", sceneID).Return(&models.Scene{}, nil).Once()
		expectedStudioID := models.NullInt64(studioID)
		mockSceneReader.On("Update", models.ScenePartial{
			ID:       sceneID,
			StudioID: &expectedStudioID,
		}).Return(nil, nil).Once()
	}

	err := StudioScenes(&studio, nil, aliases, mockSceneReader, nil)

	assert := assert.New(t)

	assert.Nil(err)
	mockSceneReader.AssertExpectations(t)
}

func TestStudioImages(t *testing.T) {
	t.Parallel()

	for _, p := range testStudioCases {
		testStudioImages(t, p)
	}
}

func testStudioImages(t *testing.T, tc testStudioCase) {
	studioName := tc.studioName
	expectedRegex := tc.expectedRegex
	aliasName := tc.aliasName
	aliasRegex := tc.aliasRegex

	mockImageReader := &mocks.ImageReaderWriter{}

	const studioID = 2

	var aliases []string

	testPathName := studioName
	if aliasName != "" {
		aliases = []string{aliasName}
		testPathName = aliasName
	}

	var images []*models.Image
	matchingPaths, falsePaths := generateTestPaths(testPathName, imageExt)
	for i, p := range append(matchingPaths, falsePaths...) {
		images = append(images, &models.Image{
			ID:   i + 1,
			Path: p,
		})
	}

	studio := models.Studio{
		ID:   studioID,
		Name: models.NullString(studioName),
	}

	organized := false
	perPage := models.PerPageAll

	expectedImageFilter := &models.ImageFilterType{
		Organized: &organized,
		Path: &models.StringCriterionInput{
			Value:    expectedRegex,
			Modifier: models.CriterionModifierMatchesRegex,
		},
	}

	expectedFindFilter := &models.FindFilterType{
		PerPage: &perPage,
	}

	// if alias provided, then don't find by name
	onNameQuery := mockImageReader.On("Query", image.QueryOptions(expectedImageFilter, expectedFindFilter, false))
	if aliasName == "" {
		onNameQuery.Return(mocks.ImageQueryResult(images, len(images)), nil).Once()
	} else {
		onNameQuery.Return(mocks.ImageQueryResult(nil, 0), nil).Once()

		expectedAliasFilter := &models.ImageFilterType{
			Organized: &organized,
			Path: &models.StringCriterionInput{
				Value:    aliasRegex,
				Modifier: models.CriterionModifierMatchesRegex,
			},
		}

		mockImageReader.On("Query", image.QueryOptions(expectedAliasFilter, expectedFindFilter, false)).
			Return(mocks.ImageQueryResult(images, len(images)), nil).Once()
	}

	for i := range matchingPaths {
		imageID := i + 1
		mockImageReader.On("Find", imageID).Return(&models.Image{}, nil).Once()
		expectedStudioID := models.NullInt64(studioID)
		mockImageReader.On("Update", models.ImagePartial{
			ID:       imageID,
			StudioID: &expectedStudioID,
		}).Return(nil, nil).Once()
	}

	err := StudioImages(&studio, nil, aliases, mockImageReader, nil)

	assert := assert.New(t)

	assert.Nil(err)
	mockImageReader.AssertExpectations(t)
}

func TestStudioGalleries(t *testing.T) {
	t.Parallel()

	for _, p := range testStudioCases {
		testStudioGalleries(t, p)
	}
}

func testStudioGalleries(t *testing.T, tc testStudioCase) {
	studioName := tc.studioName
	expectedRegex := tc.expectedRegex
	aliasName := tc.aliasName
	aliasRegex := tc.aliasRegex
	mockGalleryReader := &mocks.GalleryReaderWriter{}

	const studioID = 2

	var aliases []string

	testPathName := studioName
	if aliasName != "" {
		aliases = []string{aliasName}
		testPathName = aliasName
	}

	var galleries []*models.Gallery
	matchingPaths, falsePaths := generateTestPaths(testPathName, galleryExt)
	for i, p := range append(matchingPaths, falsePaths...) {
		galleries = append(galleries, &models.Gallery{
			ID:   i + 1,
			Path: models.NullString(p),
		})
	}

	studio := models.Studio{
		ID:   studioID,
		Name: models.NullString(studioName),
	}

	organized := false
	perPage := models.PerPageAll

	expectedGalleryFilter := &models.GalleryFilterType{
		Organized: &organized,
		Path: &models.StringCriterionInput{
			Value:    expectedRegex,
			Modifier: models.CriterionModifierMatchesRegex,
		},
	}

	expectedFindFilter := &models.FindFilterType{
		PerPage: &perPage,
	}

	// if alias provided, then don't find by name
	onNameQuery := mockGalleryReader.On("Query", expectedGalleryFilter, expectedFindFilter)
	if aliasName == "" {
		onNameQuery.Return(galleries, len(galleries), nil).Once()
	} else {
		onNameQuery.Return(nil, 0, nil).Once()

		expectedAliasFilter := &models.GalleryFilterType{
			Organized: &organized,
			Path: &models.StringCriterionInput{
				Value:    aliasRegex,
				Modifier: models.CriterionModifierMatchesRegex,
			},
		}

		mockGalleryReader.On("Query", expectedAliasFilter, expectedFindFilter).Return(galleries, len(galleries), nil).Once()
	}

	for i := range matchingPaths {
		galleryID := i + 1
		mockGalleryReader.On("Find", galleryID).Return(&models.Gallery{}, nil).Once()
		expectedStudioID := models.NullInt64(studioID)
		mockGalleryReader.On("UpdatePartial", models.GalleryPartial{
			ID:       galleryID,
			StudioID: &expectedStudioID,
		}).Return(nil, nil).Once()
	}

	err := StudioGalleries(&studio, nil, aliases, mockGalleryReader, nil)

	assert := assert.New(t)

	assert.Nil(err)
	mockGalleryReader.AssertExpectations(t)
}
