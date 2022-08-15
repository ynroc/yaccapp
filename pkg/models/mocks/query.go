package mocks

import "github.com/stashapp/stash/pkg/models"

type sceneResolver struct {
	scenes []*models.Scene
}

func (s *sceneResolver) Find(id int) (*models.Scene, error) {
	panic("not implemented")
}

func (s *sceneResolver) FindMany(ids []int) ([]*models.Scene, error) {
	return s.scenes, nil
}

func SceneQueryResult(scenes []*models.Scene, count int) *models.SceneQueryResult {
	ret := models.NewSceneQueryResult(&sceneResolver{
		scenes: scenes,
	})

	ret.Count = count
	return ret
}

type imageResolver struct {
	images []*models.Image
}

func (s *imageResolver) FindMany(ids []int) ([]*models.Image, error) {
	return s.images, nil
}

func ImageQueryResult(images []*models.Image, count int) *models.ImageQueryResult {
	ret := models.NewImageQueryResult(&imageResolver{
		images: images,
	})

	ret.Count = count
	return ret
}
