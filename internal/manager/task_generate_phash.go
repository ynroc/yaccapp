package manager

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/stashapp/stash/pkg/hash/videophash"
	"github.com/stashapp/stash/pkg/logger"
	"github.com/stashapp/stash/pkg/models"
)

type GeneratePhashTask struct {
	Scene               models.Scene
	Overwrite           bool
	fileNamingAlgorithm models.HashAlgorithm
	txnManager          models.TransactionManager
}

func (t *GeneratePhashTask) GetDescription() string {
	return fmt.Sprintf("Generating phash for %s", t.Scene.Path)
}

func (t *GeneratePhashTask) Start(ctx context.Context) {
	if !t.shouldGenerate() {
		return
	}

	ffprobe := instance.FFProbe
	videoFile, err := ffprobe.NewVideoFile(t.Scene.Path)
	if err != nil {
		logger.Errorf("error reading video file: %s", err.Error())
		return
	}

	hash, err := videophash.Generate(instance.FFMPEG, videoFile)
	if err != nil {
		logger.Errorf("error generating phash: %s", err.Error())
		logErrorOutput(err)
		return
	}

	if err := t.txnManager.WithTxn(ctx, func(r models.Repository) error {
		qb := r.Scene()
		hashValue := sql.NullInt64{Int64: int64(*hash), Valid: true}
		scenePartial := models.ScenePartial{
			ID:    t.Scene.ID,
			Phash: &hashValue,
		}
		_, err := qb.Update(scenePartial)
		return err
	}); err != nil {
		logger.Error(err.Error())
	}
}

func (t *GeneratePhashTask) shouldGenerate() bool {
	return t.Overwrite || !t.Scene.Phash.Valid
}
