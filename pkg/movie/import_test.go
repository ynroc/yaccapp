package movie

import (
	"errors"
	"testing"

	"github.com/stashapp/stash/pkg/models"
	"github.com/stashapp/stash/pkg/models/jsonschema"
	"github.com/stashapp/stash/pkg/models/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const invalidImage = "aW1hZ2VCeXRlcw&&"

const (
	movieNameErr      = "movieNameErr"
	existingMovieName = "existingMovieName"

	existingMovieID  = 100
	existingStudioID = 101

	existingStudioName = "existingStudioName"
	existingStudioErr  = "existingStudioErr"
	missingStudioName  = "existingStudioName"

	errImageID = 3
)

func TestImporterName(t *testing.T) {
	i := Importer{
		Input: jsonschema.Movie{
			Name: movieName,
		},
	}

	assert.Equal(t, movieName, i.Name())
}

func TestImporterPreImport(t *testing.T) {
	i := Importer{
		Input: jsonschema.Movie{
			Name:       movieName,
			FrontImage: invalidImage,
		},
	}

	err := i.PreImport()
	assert.NotNil(t, err)

	i.Input.FrontImage = frontImage
	i.Input.BackImage = invalidImage

	err = i.PreImport()
	assert.NotNil(t, err)

	i.Input.BackImage = ""

	err = i.PreImport()
	assert.Nil(t, err)

	i.Input.BackImage = backImage

	err = i.PreImport()
	assert.Nil(t, err)
}

func TestImporterPreImportWithStudio(t *testing.T) {
	studioReaderWriter := &mocks.StudioReaderWriter{}

	i := Importer{
		StudioWriter: studioReaderWriter,
		Input: jsonschema.Movie{
			Name:       movieName,
			FrontImage: frontImage,
			Studio:     existingStudioName,
			Rating:     5,
			Duration:   10,
		},
	}

	studioReaderWriter.On("FindByName", existingStudioName, false).Return(&models.Studio{
		ID: existingStudioID,
	}, nil).Once()
	studioReaderWriter.On("FindByName", existingStudioErr, false).Return(nil, errors.New("FindByName error")).Once()

	err := i.PreImport()
	assert.Nil(t, err)
	assert.Equal(t, int64(existingStudioID), i.movie.StudioID.Int64)

	i.Input.Studio = existingStudioErr
	err = i.PreImport()
	assert.NotNil(t, err)

	studioReaderWriter.AssertExpectations(t)
}

func TestImporterPreImportWithMissingStudio(t *testing.T) {
	studioReaderWriter := &mocks.StudioReaderWriter{}

	i := Importer{
		StudioWriter: studioReaderWriter,
		Input: jsonschema.Movie{
			Name:       movieName,
			FrontImage: frontImage,
			Studio:     missingStudioName,
		},
		MissingRefBehaviour: models.ImportMissingRefEnumFail,
	}

	studioReaderWriter.On("FindByName", missingStudioName, false).Return(nil, nil).Times(3)
	studioReaderWriter.On("Create", mock.AnythingOfType("models.Studio")).Return(&models.Studio{
		ID: existingStudioID,
	}, nil)

	err := i.PreImport()
	assert.NotNil(t, err)

	i.MissingRefBehaviour = models.ImportMissingRefEnumIgnore
	err = i.PreImport()
	assert.Nil(t, err)

	i.MissingRefBehaviour = models.ImportMissingRefEnumCreate
	err = i.PreImport()
	assert.Nil(t, err)
	assert.Equal(t, int64(existingStudioID), i.movie.StudioID.Int64)

	studioReaderWriter.AssertExpectations(t)
}

func TestImporterPreImportWithMissingStudioCreateErr(t *testing.T) {
	studioReaderWriter := &mocks.StudioReaderWriter{}

	i := Importer{
		StudioWriter: studioReaderWriter,
		Input: jsonschema.Movie{
			Name:       movieName,
			FrontImage: frontImage,
			Studio:     missingStudioName,
		},
		MissingRefBehaviour: models.ImportMissingRefEnumCreate,
	}

	studioReaderWriter.On("FindByName", missingStudioName, false).Return(nil, nil).Once()
	studioReaderWriter.On("Create", mock.AnythingOfType("models.Studio")).Return(nil, errors.New("Create error"))

	err := i.PreImport()
	assert.NotNil(t, err)
}

func TestImporterPostImport(t *testing.T) {
	readerWriter := &mocks.MovieReaderWriter{}

	i := Importer{
		ReaderWriter:   readerWriter,
		frontImageData: frontImageBytes,
		backImageData:  backImageBytes,
	}

	updateMovieImageErr := errors.New("UpdateImages error")

	readerWriter.On("UpdateImages", movieID, frontImageBytes, backImageBytes).Return(nil).Once()
	readerWriter.On("UpdateImages", errImageID, frontImageBytes, backImageBytes).Return(updateMovieImageErr).Once()

	err := i.PostImport(movieID)
	assert.Nil(t, err)

	err = i.PostImport(errImageID)
	assert.NotNil(t, err)

	readerWriter.AssertExpectations(t)
}

func TestImporterFindExistingID(t *testing.T) {
	readerWriter := &mocks.MovieReaderWriter{}

	i := Importer{
		ReaderWriter: readerWriter,
		Input: jsonschema.Movie{
			Name: movieName,
		},
	}

	errFindByName := errors.New("FindByName error")
	readerWriter.On("FindByName", movieName, false).Return(nil, nil).Once()
	readerWriter.On("FindByName", existingMovieName, false).Return(&models.Movie{
		ID: existingMovieID,
	}, nil).Once()
	readerWriter.On("FindByName", movieNameErr, false).Return(nil, errFindByName).Once()

	id, err := i.FindExistingID()
	assert.Nil(t, id)
	assert.Nil(t, err)

	i.Input.Name = existingMovieName
	id, err = i.FindExistingID()
	assert.Equal(t, existingMovieID, *id)
	assert.Nil(t, err)

	i.Input.Name = movieNameErr
	id, err = i.FindExistingID()
	assert.Nil(t, id)
	assert.NotNil(t, err)

	readerWriter.AssertExpectations(t)
}

func TestCreate(t *testing.T) {
	readerWriter := &mocks.MovieReaderWriter{}

	movie := models.Movie{
		Name: models.NullString(movieName),
	}

	movieErr := models.Movie{
		Name: models.NullString(movieNameErr),
	}

	i := Importer{
		ReaderWriter: readerWriter,
		movie:        movie,
	}

	errCreate := errors.New("Create error")
	readerWriter.On("Create", movie).Return(&models.Movie{
		ID: movieID,
	}, nil).Once()
	readerWriter.On("Create", movieErr).Return(nil, errCreate).Once()

	id, err := i.Create()
	assert.Equal(t, movieID, *id)
	assert.Nil(t, err)

	i.movie = movieErr
	id, err = i.Create()
	assert.Nil(t, id)
	assert.NotNil(t, err)

	readerWriter.AssertExpectations(t)
}

func TestUpdate(t *testing.T) {
	readerWriter := &mocks.MovieReaderWriter{}

	movie := models.Movie{
		Name: models.NullString(movieName),
	}

	movieErr := models.Movie{
		Name: models.NullString(movieNameErr),
	}

	i := Importer{
		ReaderWriter: readerWriter,
		movie:        movie,
	}

	errUpdate := errors.New("Update error")

	// id needs to be set for the mock input
	movie.ID = movieID
	readerWriter.On("UpdateFull", movie).Return(nil, nil).Once()

	err := i.Update(movieID)
	assert.Nil(t, err)

	i.movie = movieErr

	// need to set id separately
	movieErr.ID = errImageID
	readerWriter.On("UpdateFull", movieErr).Return(nil, errUpdate).Once()

	err = i.Update(errImageID)
	assert.NotNil(t, err)

	readerWriter.AssertExpectations(t)
}
