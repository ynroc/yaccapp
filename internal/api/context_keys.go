package api

// https://stackoverflow.com/questions/40891345/fix-should-not-use-basic-type-string-as-key-in-context-withvalue-golint

type key int

const (
	// galleryKey key = 0
	performerKey key = iota + 1
	sceneKey
	studioKey
	movieKey
	tagKey
	downloadKey
	imageKey
)
