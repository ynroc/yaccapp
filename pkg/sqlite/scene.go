package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/stashapp/stash/pkg/models"
	"github.com/stashapp/stash/pkg/utils"
)

const sceneTable = "scenes"
const sceneIDColumn = "scene_id"
const performersScenesTable = "performers_scenes"
const scenesTagsTable = "scenes_tags"
const scenesGalleriesTable = "scenes_galleries"
const moviesScenesTable = "movies_scenes"

const sceneCaptionsTable = "scene_captions"
const sceneCaptionCodeColumn = "language_code"
const sceneCaptionFilenameColumn = "filename"
const sceneCaptionTypeColumn = "caption_type"

var scenesForPerformerQuery = selectAll(sceneTable) + `
LEFT JOIN performers_scenes as performers_join on performers_join.scene_id = scenes.id
WHERE performers_join.performer_id = ?
GROUP BY scenes.id
`

var countScenesForPerformerQuery = `
SELECT performer_id FROM performers_scenes as performers_join
WHERE performer_id = ?
GROUP BY scene_id
`

var scenesForStudioQuery = selectAll(sceneTable) + `
JOIN studios ON studios.id = scenes.studio_id
WHERE studios.id = ?
GROUP BY scenes.id
`
var scenesForMovieQuery = selectAll(sceneTable) + `
LEFT JOIN movies_scenes as movies_join on movies_join.scene_id = scenes.id
WHERE movies_join.movie_id = ?
GROUP BY scenes.id
`

var countScenesForTagQuery = `
SELECT tag_id AS id FROM scenes_tags
WHERE scenes_tags.tag_id = ?
GROUP BY scenes_tags.scene_id
`

var scenesForGalleryQuery = selectAll(sceneTable) + `
LEFT JOIN scenes_galleries as galleries_join on galleries_join.scene_id = scenes.id
WHERE galleries_join.gallery_id = ?
GROUP BY scenes.id
`

var countScenesForMissingChecksumQuery = `
SELECT id FROM scenes
WHERE scenes.checksum is null
`

var countScenesForMissingOSHashQuery = `
SELECT id FROM scenes
WHERE scenes.oshash is null
`

var findExactDuplicateQuery = `
SELECT GROUP_CONCAT(id) as ids
FROM scenes
WHERE phash IS NOT NULL
GROUP BY phash
HAVING COUNT(phash) > 1
ORDER BY SUM(size) DESC;
`

var findAllPhashesQuery = `
SELECT id, phash
FROM scenes
WHERE phash IS NOT NULL
ORDER BY size DESC
`

type sceneQueryBuilder struct {
	repository
}

func NewSceneReaderWriter(tx dbi) *sceneQueryBuilder {
	return &sceneQueryBuilder{
		repository{
			tx:        tx,
			tableName: sceneTable,
			idColumn:  idColumn,
		},
	}
}

func (qb *sceneQueryBuilder) Create(newObject models.Scene) (*models.Scene, error) {
	var ret models.Scene
	if err := qb.insertObject(newObject, &ret); err != nil {
		return nil, err
	}

	return &ret, nil
}

func (qb *sceneQueryBuilder) Update(updatedObject models.ScenePartial) (*models.Scene, error) {
	const partial = true
	if err := qb.update(updatedObject.ID, updatedObject, partial); err != nil {
		return nil, err
	}

	return qb.find(updatedObject.ID)
}

func (qb *sceneQueryBuilder) UpdateFull(updatedObject models.Scene) (*models.Scene, error) {
	const partial = false
	if err := qb.update(updatedObject.ID, updatedObject, partial); err != nil {
		return nil, err
	}

	return qb.find(updatedObject.ID)
}

func (qb *sceneQueryBuilder) UpdateFileModTime(id int, modTime models.NullSQLiteTimestamp) error {
	return qb.updateMap(id, map[string]interface{}{
		"file_mod_time": modTime,
	})
}

func (qb *sceneQueryBuilder) captionRepository() *captionRepository {
	return &captionRepository{
		repository: repository{
			tx:        qb.tx,
			tableName: sceneCaptionsTable,
			idColumn:  sceneIDColumn,
		},
	}
}

func (qb *sceneQueryBuilder) GetCaptions(sceneID int) ([]*models.SceneCaption, error) {
	return qb.captionRepository().get(sceneID)
}

func (qb *sceneQueryBuilder) UpdateCaptions(sceneID int, captions []*models.SceneCaption) error {
	return qb.captionRepository().replace(sceneID, captions)

}

func (qb *sceneQueryBuilder) IncrementOCounter(id int) (int, error) {
	_, err := qb.tx.Exec(
		`UPDATE scenes SET o_counter = o_counter + 1 WHERE scenes.id = ?`,
		id,
	)
	if err != nil {
		return 0, err
	}

	scene, err := qb.find(id)
	if err != nil {
		return 0, err
	}

	return scene.OCounter, nil
}

func (qb *sceneQueryBuilder) DecrementOCounter(id int) (int, error) {
	_, err := qb.tx.Exec(
		`UPDATE scenes SET o_counter = o_counter - 1 WHERE scenes.id = ? and scenes.o_counter > 0`,
		id,
	)
	if err != nil {
		return 0, err
	}

	scene, err := qb.find(id)
	if err != nil {
		return 0, err
	}

	return scene.OCounter, nil
}

func (qb *sceneQueryBuilder) ResetOCounter(id int) (int, error) {
	_, err := qb.tx.Exec(
		`UPDATE scenes SET o_counter = 0 WHERE scenes.id = ?`,
		id,
	)
	if err != nil {
		return 0, err
	}

	scene, err := qb.find(id)
	if err != nil {
		return 0, err
	}

	return scene.OCounter, nil
}

func (qb *sceneQueryBuilder) Destroy(id int) error {
	// delete all related table rows
	// TODO - this should be handled by a delete cascade
	if err := qb.performersRepository().destroy([]int{id}); err != nil {
		return err
	}

	// scene markers should be handled prior to calling destroy
	// galleries should be handled prior to calling destroy

	return qb.destroyExisting([]int{id})
}

func (qb *sceneQueryBuilder) Find(id int) (*models.Scene, error) {
	return qb.find(id)
}

func (qb *sceneQueryBuilder) FindMany(ids []int) ([]*models.Scene, error) {
	var scenes []*models.Scene
	for _, id := range ids {
		scene, err := qb.Find(id)
		if err != nil {
			return nil, err
		}

		if scene == nil {
			return nil, fmt.Errorf("scene with id %d not found", id)
		}

		scenes = append(scenes, scene)
	}

	return scenes, nil
}

func (qb *sceneQueryBuilder) find(id int) (*models.Scene, error) {
	var ret models.Scene
	if err := qb.get(id, &ret); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &ret, nil
}

func (qb *sceneQueryBuilder) FindByChecksum(checksum string) (*models.Scene, error) {
	query := "SELECT * FROM scenes WHERE checksum = ? LIMIT 1"
	args := []interface{}{checksum}
	return qb.queryScene(query, args)
}

func (qb *sceneQueryBuilder) FindByOSHash(oshash string) (*models.Scene, error) {
	query := "SELECT * FROM scenes WHERE oshash = ? LIMIT 1"
	args := []interface{}{oshash}
	return qb.queryScene(query, args)
}

func (qb *sceneQueryBuilder) FindByPath(path string) (*models.Scene, error) {
	query := selectAll(sceneTable) + "WHERE path = ? LIMIT 1"
	args := []interface{}{path}
	return qb.queryScene(query, args)
}

func (qb *sceneQueryBuilder) FindByPerformerID(performerID int) ([]*models.Scene, error) {
	args := []interface{}{performerID}
	return qb.queryScenes(scenesForPerformerQuery, args)
}

func (qb *sceneQueryBuilder) FindByGalleryID(galleryID int) ([]*models.Scene, error) {
	args := []interface{}{galleryID}
	return qb.queryScenes(scenesForGalleryQuery, args)
}

func (qb *sceneQueryBuilder) CountByPerformerID(performerID int) (int, error) {
	args := []interface{}{performerID}
	return qb.runCountQuery(qb.buildCountQuery(countScenesForPerformerQuery), args)
}

func (qb *sceneQueryBuilder) FindByMovieID(movieID int) ([]*models.Scene, error) {
	args := []interface{}{movieID}
	return qb.queryScenes(scenesForMovieQuery, args)
}

func (qb *sceneQueryBuilder) CountByMovieID(movieID int) (int, error) {
	args := []interface{}{movieID}
	return qb.runCountQuery(qb.buildCountQuery(scenesForMovieQuery), args)
}

func (qb *sceneQueryBuilder) Count() (int, error) {
	return qb.runCountQuery(qb.buildCountQuery("SELECT scenes.id FROM scenes"), nil)
}

func (qb *sceneQueryBuilder) Size() (float64, error) {
	return qb.runSumQuery("SELECT SUM(cast(size as double)) as sum FROM scenes", nil)
}

func (qb *sceneQueryBuilder) Duration() (float64, error) {
	return qb.runSumQuery("SELECT SUM(cast(duration as double)) as sum FROM scenes", nil)
}

func (qb *sceneQueryBuilder) CountByStudioID(studioID int) (int, error) {
	args := []interface{}{studioID}
	return qb.runCountQuery(qb.buildCountQuery(scenesForStudioQuery), args)
}

func (qb *sceneQueryBuilder) CountByTagID(tagID int) (int, error) {
	args := []interface{}{tagID}
	return qb.runCountQuery(qb.buildCountQuery(countScenesForTagQuery), args)
}

// CountMissingChecksum returns the number of scenes missing a checksum value.
func (qb *sceneQueryBuilder) CountMissingChecksum() (int, error) {
	return qb.runCountQuery(qb.buildCountQuery(countScenesForMissingChecksumQuery), []interface{}{})
}

// CountMissingOSHash returns the number of scenes missing an oshash value.
func (qb *sceneQueryBuilder) CountMissingOSHash() (int, error) {
	return qb.runCountQuery(qb.buildCountQuery(countScenesForMissingOSHashQuery), []interface{}{})
}

func (qb *sceneQueryBuilder) Wall(q *string) ([]*models.Scene, error) {
	s := ""
	if q != nil {
		s = *q
	}
	query := selectAll(sceneTable) + "WHERE scenes.details LIKE '%" + s + "%' ORDER BY RANDOM() LIMIT 80"
	return qb.queryScenes(query, nil)
}

func (qb *sceneQueryBuilder) All() ([]*models.Scene, error) {
	return qb.queryScenes(selectAll(sceneTable)+qb.getDefaultSceneSort(), nil)
}

func illegalFilterCombination(type1, type2 string) error {
	return fmt.Errorf("cannot have %s and %s in the same filter", type1, type2)
}

func (qb *sceneQueryBuilder) validateFilter(sceneFilter *models.SceneFilterType) error {
	const and = "AND"
	const or = "OR"
	const not = "NOT"

	if sceneFilter.And != nil {
		if sceneFilter.Or != nil {
			return illegalFilterCombination(and, or)
		}
		if sceneFilter.Not != nil {
			return illegalFilterCombination(and, not)
		}

		return qb.validateFilter(sceneFilter.And)
	}

	if sceneFilter.Or != nil {
		if sceneFilter.Not != nil {
			return illegalFilterCombination(or, not)
		}

		return qb.validateFilter(sceneFilter.Or)
	}

	if sceneFilter.Not != nil {
		return qb.validateFilter(sceneFilter.Not)
	}

	return nil
}

func (qb *sceneQueryBuilder) makeFilter(sceneFilter *models.SceneFilterType) *filterBuilder {
	query := &filterBuilder{}

	if sceneFilter.And != nil {
		query.and(qb.makeFilter(sceneFilter.And))
	}
	if sceneFilter.Or != nil {
		query.or(qb.makeFilter(sceneFilter.Or))
	}
	if sceneFilter.Not != nil {
		query.not(qb.makeFilter(sceneFilter.Not))
	}

	query.handleCriterion(stringCriterionHandler(sceneFilter.Path, "scenes.path"))
	query.handleCriterion(stringCriterionHandler(sceneFilter.Title, "scenes.title"))
	query.handleCriterion(stringCriterionHandler(sceneFilter.Details, "scenes.details"))
	query.handleCriterion(stringCriterionHandler(sceneFilter.Oshash, "scenes.oshash"))
	query.handleCriterion(stringCriterionHandler(sceneFilter.Checksum, "scenes.checksum"))
	query.handleCriterion(phashCriterionHandler(sceneFilter.Phash))
	query.handleCriterion(intCriterionHandler(sceneFilter.Rating, "scenes.rating"))
	query.handleCriterion(intCriterionHandler(sceneFilter.OCounter, "scenes.o_counter"))
	query.handleCriterion(boolCriterionHandler(sceneFilter.Organized, "scenes.organized"))
	query.handleCriterion(durationCriterionHandler(sceneFilter.Duration, "scenes.duration"))
	query.handleCriterion(resolutionCriterionHandler(sceneFilter.Resolution, "scenes.height", "scenes.width"))
	query.handleCriterion(hasMarkersCriterionHandler(sceneFilter.HasMarkers))
	query.handleCriterion(sceneIsMissingCriterionHandler(qb, sceneFilter.IsMissing))
	query.handleCriterion(stringCriterionHandler(sceneFilter.URL, "scenes.url"))

	query.handleCriterion(criterionHandlerFunc(func(f *filterBuilder) {
		if sceneFilter.StashID != nil {
			qb.stashIDRepository().join(f, "scene_stash_ids", "scenes.id")
			stringCriterionHandler(sceneFilter.StashID, "scene_stash_ids.stash_id")(f)
		}
	}))

	query.handleCriterion(boolCriterionHandler(sceneFilter.Interactive, "scenes.interactive"))
	query.handleCriterion(intCriterionHandler(sceneFilter.InteractiveSpeed, "scenes.interactive_speed"))

	query.handleCriterion(sceneCaptionCriterionHandler(qb, sceneFilter.Captions))

	query.handleCriterion(sceneTagsCriterionHandler(qb, sceneFilter.Tags))
	query.handleCriterion(sceneTagCountCriterionHandler(qb, sceneFilter.TagCount))
	query.handleCriterion(scenePerformersCriterionHandler(qb, sceneFilter.Performers))
	query.handleCriterion(scenePerformerCountCriterionHandler(qb, sceneFilter.PerformerCount))
	query.handleCriterion(sceneStudioCriterionHandler(qb, sceneFilter.Studios))
	query.handleCriterion(sceneMoviesCriterionHandler(qb, sceneFilter.Movies))
	query.handleCriterion(scenePerformerTagsCriterionHandler(qb, sceneFilter.PerformerTags))
	query.handleCriterion(scenePerformerFavoriteCriterionHandler(sceneFilter.PerformerFavorite))
	query.handleCriterion(scenePerformerAgeCriterionHandler(sceneFilter.PerformerAge))
	query.handleCriterion(scenePhashDuplicatedCriterionHandler(sceneFilter.Duplicated))

	return query
}

func (qb *sceneQueryBuilder) Query(options models.SceneQueryOptions) (*models.SceneQueryResult, error) {
	sceneFilter := options.SceneFilter
	findFilter := options.FindFilter

	if sceneFilter == nil {
		sceneFilter = &models.SceneFilterType{}
	}
	if findFilter == nil {
		findFilter = &models.FindFilterType{}
	}

	query := qb.newQuery()
	distinctIDs(&query, sceneTable)

	if q := findFilter.Q; q != nil && *q != "" {
		query.join("scene_markers", "", "scene_markers.scene_id = scenes.id")
		searchColumns := []string{"scenes.title", "scenes.details", "scenes.path", "scenes.oshash", "scenes.checksum", "scene_markers.title"}
		query.parseQueryString(searchColumns, *q)
	}

	if err := qb.validateFilter(sceneFilter); err != nil {
		return nil, err
	}
	filter := qb.makeFilter(sceneFilter)

	query.addFilter(filter)

	qb.setSceneSort(&query, findFilter)
	query.sortAndPagination += getPagination(findFilter)

	result, err := qb.queryGroupedFields(options, query)
	if err != nil {
		return nil, fmt.Errorf("error querying aggregate fields: %w", err)
	}

	idsResult, err := query.findIDs()
	if err != nil {
		return nil, fmt.Errorf("error finding IDs: %w", err)
	}

	result.IDs = idsResult
	return result, nil
}

func (qb *sceneQueryBuilder) queryGroupedFields(options models.SceneQueryOptions, query queryBuilder) (*models.SceneQueryResult, error) {
	if !options.Count && !options.TotalDuration && !options.TotalSize {
		// nothing to do - return empty result
		return models.NewSceneQueryResult(qb), nil
	}

	aggregateQuery := qb.newQuery()

	if options.Count {
		aggregateQuery.addColumn("COUNT(temp.id) as total")
	}

	if options.TotalDuration {
		query.addColumn("COALESCE(scenes.duration, 0) as duration")
		aggregateQuery.addColumn("COALESCE(SUM(temp.duration), 0) as duration")
	}

	if options.TotalSize {
		query.addColumn("COALESCE(scenes.size, 0) as size")
		aggregateQuery.addColumn("COALESCE(SUM(temp.size), 0) as size")
	}

	const includeSortPagination = false
	aggregateQuery.from = fmt.Sprintf("(%s) as temp", query.toSQL(includeSortPagination))

	out := struct {
		Total    int
		Duration float64
		Size     float64
	}{}
	if err := qb.repository.queryStruct(aggregateQuery.toSQL(includeSortPagination), query.args, &out); err != nil {
		return nil, err
	}

	ret := models.NewSceneQueryResult(qb)
	ret.Count = out.Total
	ret.TotalDuration = out.Duration
	ret.TotalSize = out.Size
	return ret, nil
}

func phashCriterionHandler(phashFilter *models.StringCriterionInput) criterionHandlerFunc {
	return func(f *filterBuilder) {
		if phashFilter != nil {
			// convert value to int from hex
			// ignore errors
			value, _ := utils.StringToPhash(phashFilter.Value)

			if modifier := phashFilter.Modifier; phashFilter.Modifier.IsValid() {
				switch modifier {
				case models.CriterionModifierEquals:
					f.addWhere("scenes.phash = ?", value)
				case models.CriterionModifierNotEquals:
					f.addWhere("scenes.phash != ?", value)
				case models.CriterionModifierIsNull:
					f.addWhere("scenes.phash IS NULL")
				case models.CriterionModifierNotNull:
					f.addWhere("scenes.phash IS NOT NULL")
				}
			}
		}
	}
}

func scenePhashDuplicatedCriterionHandler(duplicatedFilter *models.PHashDuplicationCriterionInput) criterionHandlerFunc {
	return func(f *filterBuilder) {
		// TODO: Wishlist item: Implement Distance matching
		if duplicatedFilter != nil {
			var v string
			if *duplicatedFilter.Duplicated {
				v = ">"
			} else {
				v = "="
			}
			f.addInnerJoin("(SELECT id FROM scenes JOIN (SELECT phash FROM scenes GROUP BY phash HAVING COUNT(phash) "+v+" 1) dupes on scenes.phash = dupes.phash)", "scph", "scenes.id = scph.id")
		}
	}
}

func durationCriterionHandler(durationFilter *models.IntCriterionInput, column string) criterionHandlerFunc {
	return func(f *filterBuilder) {
		if durationFilter != nil {
			clause, args := getIntCriterionWhereClause("cast("+column+" as int)", *durationFilter)
			f.addWhere(clause, args...)
		}
	}
}

func resolutionCriterionHandler(resolution *models.ResolutionCriterionInput, heightColumn string, widthColumn string) criterionHandlerFunc {
	return func(f *filterBuilder) {
		if resolution != nil && resolution.Value.IsValid() {
			min := resolution.Value.GetMinResolution()
			max := resolution.Value.GetMaxResolution()

			widthHeight := fmt.Sprintf("MIN(%s, %s)", widthColumn, heightColumn)

			switch resolution.Modifier {
			case models.CriterionModifierEquals:
				f.addWhere(fmt.Sprintf("%s BETWEEN %d AND %d", widthHeight, min, max))
			case models.CriterionModifierNotEquals:
				f.addWhere(fmt.Sprintf("%s NOT BETWEEN %d AND %d", widthHeight, min, max))
			case models.CriterionModifierLessThan:
				f.addWhere(fmt.Sprintf("%s < %d", widthHeight, min))
			case models.CriterionModifierGreaterThan:
				f.addWhere(fmt.Sprintf("%s > %d", widthHeight, max))
			}
		}
	}
}

func hasMarkersCriterionHandler(hasMarkers *string) criterionHandlerFunc {
	return func(f *filterBuilder) {
		if hasMarkers != nil {
			f.addLeftJoin("scene_markers", "", "scene_markers.scene_id = scenes.id")
			if *hasMarkers == "true" {
				f.addHaving("count(scene_markers.scene_id) > 0")
			} else {
				f.addWhere("scene_markers.id IS NULL")
			}
		}
	}
}

func sceneIsMissingCriterionHandler(qb *sceneQueryBuilder, isMissing *string) criterionHandlerFunc {
	return func(f *filterBuilder) {
		if isMissing != nil && *isMissing != "" {
			switch *isMissing {
			case "galleries":
				qb.galleriesRepository().join(f, "galleries_join", "scenes.id")
				f.addWhere("galleries_join.scene_id IS NULL")
			case "studio":
				f.addWhere("scenes.studio_id IS NULL")
			case "movie":
				qb.moviesRepository().join(f, "movies_join", "scenes.id")
				f.addWhere("movies_join.scene_id IS NULL")
			case "performers":
				qb.performersRepository().join(f, "performers_join", "scenes.id")
				f.addWhere("performers_join.scene_id IS NULL")
			case "date":
				f.addWhere(`scenes.date IS NULL OR scenes.date IS "" OR scenes.date IS "0001-01-01"`)
			case "tags":
				qb.tagsRepository().join(f, "tags_join", "scenes.id")
				f.addWhere("tags_join.scene_id IS NULL")
			case "stash_id":
				qb.stashIDRepository().join(f, "scene_stash_ids", "scenes.id")
				f.addWhere("scene_stash_ids.scene_id IS NULL")
			default:
				f.addWhere("(scenes." + *isMissing + " IS NULL OR TRIM(scenes." + *isMissing + ") = '')")
			}
		}
	}
}

func (qb *sceneQueryBuilder) getMultiCriterionHandlerBuilder(foreignTable, joinTable, foreignFK string, addJoinsFunc func(f *filterBuilder)) multiCriterionHandlerBuilder {
	return multiCriterionHandlerBuilder{
		primaryTable: sceneTable,
		foreignTable: foreignTable,
		joinTable:    joinTable,
		primaryFK:    sceneIDColumn,
		foreignFK:    foreignFK,
		addJoinsFunc: addJoinsFunc,
	}
}

func sceneCaptionCriterionHandler(qb *sceneQueryBuilder, captions *models.StringCriterionInput) criterionHandlerFunc {
	h := stringListCriterionHandlerBuilder{
		joinTable:    sceneCaptionsTable,
		stringColumn: sceneCaptionCodeColumn,
		addJoinTable: func(f *filterBuilder) {
			qb.captionRepository().join(f, "", "scenes.id")
		},
	}

	return h.handler(captions)
}

func sceneTagsCriterionHandler(qb *sceneQueryBuilder, tags *models.HierarchicalMultiCriterionInput) criterionHandlerFunc {
	h := joinedHierarchicalMultiCriterionHandlerBuilder{
		tx: qb.tx,

		primaryTable: sceneTable,
		foreignTable: tagTable,
		foreignFK:    "tag_id",

		relationsTable: "tags_relations",
		joinAs:         "scene_tag",
		joinTable:      scenesTagsTable,
		primaryFK:      sceneIDColumn,
	}

	return h.handler(tags)
}

func sceneTagCountCriterionHandler(qb *sceneQueryBuilder, tagCount *models.IntCriterionInput) criterionHandlerFunc {
	h := countCriterionHandlerBuilder{
		primaryTable: sceneTable,
		joinTable:    scenesTagsTable,
		primaryFK:    sceneIDColumn,
	}

	return h.handler(tagCount)
}

func scenePerformersCriterionHandler(qb *sceneQueryBuilder, performers *models.MultiCriterionInput) criterionHandlerFunc {
	h := joinedMultiCriterionHandlerBuilder{
		primaryTable: sceneTable,
		joinTable:    performersScenesTable,
		joinAs:       "performers_join",
		primaryFK:    sceneIDColumn,
		foreignFK:    performerIDColumn,

		addJoinTable: func(f *filterBuilder) {
			qb.performersRepository().join(f, "performers_join", "scenes.id")
		},
	}

	return h.handler(performers)
}

func scenePerformerCountCriterionHandler(qb *sceneQueryBuilder, performerCount *models.IntCriterionInput) criterionHandlerFunc {
	h := countCriterionHandlerBuilder{
		primaryTable: sceneTable,
		joinTable:    performersScenesTable,
		primaryFK:    sceneIDColumn,
	}

	return h.handler(performerCount)
}

func scenePerformerFavoriteCriterionHandler(performerfavorite *bool) criterionHandlerFunc {
	return func(f *filterBuilder) {
		if performerfavorite != nil {
			f.addLeftJoin("performers_scenes", "", "scenes.id = performers_scenes.scene_id")

			if *performerfavorite {
				// contains at least one favorite
				f.addLeftJoin("performers", "", "performers.id = performers_scenes.performer_id")
				f.addWhere("performers.favorite = 1")
			} else {
				// contains zero favorites
				f.addLeftJoin(`(SELECT performers_scenes.scene_id as id FROM performers_scenes
JOIN performers ON performers.id = performers_scenes.performer_id
GROUP BY performers_scenes.scene_id HAVING SUM(performers.favorite) = 0)`, "nofaves", "scenes.id = nofaves.id")
				f.addWhere("performers_scenes.scene_id IS NULL OR nofaves.id IS NOT NULL")
			}
		}
	}
}

func scenePerformerAgeCriterionHandler(performerAge *models.IntCriterionInput) criterionHandlerFunc {
	return func(f *filterBuilder) {
		if performerAge != nil {
			f.addInnerJoin("performers_scenes", "", "scenes.id = performers_scenes.scene_id")
			f.addInnerJoin("performers", "", "performers_scenes.performer_id = performers.id")

			f.addWhere("scenes.date != '' AND performers.birthdate != ''")
			f.addWhere("scenes.date IS NOT NULL AND performers.birthdate IS NOT NULL")
			f.addWhere("scenes.date != '0001-01-01' AND performers.birthdate != '0001-01-01'")

			ageCalc := "cast(strftime('%Y.%m%d', scenes.date) - strftime('%Y.%m%d', performers.birthdate) as int)"
			whereClause, args := getIntWhereClause(ageCalc, performerAge.Modifier, performerAge.Value, performerAge.Value2)
			f.addWhere(whereClause, args...)
		}
	}
}

func sceneStudioCriterionHandler(qb *sceneQueryBuilder, studios *models.HierarchicalMultiCriterionInput) criterionHandlerFunc {
	h := hierarchicalMultiCriterionHandlerBuilder{
		tx: qb.tx,

		primaryTable: sceneTable,
		foreignTable: studioTable,
		foreignFK:    studioIDColumn,
		derivedTable: "studio",
		parentFK:     "parent_id",
	}

	return h.handler(studios)
}

func sceneMoviesCriterionHandler(qb *sceneQueryBuilder, movies *models.MultiCriterionInput) criterionHandlerFunc {
	addJoinsFunc := func(f *filterBuilder) {
		qb.moviesRepository().join(f, "", "scenes.id")
		f.addLeftJoin("movies", "", "movies_scenes.movie_id = movies.id")
	}
	h := qb.getMultiCriterionHandlerBuilder(movieTable, moviesScenesTable, "movie_id", addJoinsFunc)
	return h.handler(movies)
}

func scenePerformerTagsCriterionHandler(qb *sceneQueryBuilder, tags *models.HierarchicalMultiCriterionInput) criterionHandlerFunc {
	return func(f *filterBuilder) {
		if tags != nil {
			if tags.Modifier == models.CriterionModifierIsNull || tags.Modifier == models.CriterionModifierNotNull {
				var notClause string
				if tags.Modifier == models.CriterionModifierNotNull {
					notClause = "NOT"
				}

				f.addLeftJoin("performers_scenes", "", "scenes.id = performers_scenes.scene_id")
				f.addLeftJoin("performers_tags", "", "performers_scenes.performer_id = performers_tags.performer_id")

				f.addWhere(fmt.Sprintf("performers_tags.tag_id IS %s NULL", notClause))
				return
			}

			if len(tags.Value) == 0 {
				return
			}

			valuesClause := getHierarchicalValues(qb.tx, tags.Value, tagTable, "tags_relations", "", tags.Depth)

			f.addWith(`performer_tags AS (
SELECT ps.scene_id, t.column1 AS root_tag_id FROM performers_scenes ps
INNER JOIN performers_tags pt ON pt.performer_id = ps.performer_id
INNER JOIN (` + valuesClause + `) t ON t.column2 = pt.tag_id
)`)

			f.addLeftJoin("performer_tags", "", "performer_tags.scene_id = scenes.id")

			addHierarchicalConditionClauses(f, tags, "performer_tags", "root_tag_id")
		}
	}
}

func (qb *sceneQueryBuilder) getDefaultSceneSort() string {
	return " ORDER BY scenes.path, scenes.date ASC "
}

func (qb *sceneQueryBuilder) setSceneSort(query *queryBuilder, findFilter *models.FindFilterType) {
	if findFilter == nil || findFilter.Sort == nil || *findFilter.Sort == "" {
		return
	}
	sort := findFilter.GetSort("title")
	direction := findFilter.GetDirection()
	switch sort {
	case "movie_scene_number":
		query.join(moviesScenesTable, "movies_join", "scenes.id = movies_join.scene_id")
		query.sortAndPagination += fmt.Sprintf(" ORDER BY movies_join.scene_index %s", getSortDirection(direction))
	case "tag_count":
		query.sortAndPagination += getCountSort(sceneTable, scenesTagsTable, sceneIDColumn, direction)
	case "performer_count":
		query.sortAndPagination += getCountSort(sceneTable, performersScenesTable, sceneIDColumn, direction)
	default:
		query.sortAndPagination += getSort(sort, direction, "scenes")
	}
}

func (qb *sceneQueryBuilder) queryScene(query string, args []interface{}) (*models.Scene, error) {
	results, err := qb.queryScenes(query, args)
	if err != nil || len(results) < 1 {
		return nil, err
	}
	return results[0], nil
}

func (qb *sceneQueryBuilder) queryScenes(query string, args []interface{}) ([]*models.Scene, error) {
	var ret models.Scenes
	if err := qb.query(query, args, &ret); err != nil {
		return nil, err
	}

	return []*models.Scene(ret), nil
}

func (qb *sceneQueryBuilder) imageRepository() *imageRepository {
	return &imageRepository{
		repository: repository{
			tx:        qb.tx,
			tableName: "scenes_cover",
			idColumn:  sceneIDColumn,
		},
		imageColumn: "cover",
	}
}

func (qb *sceneQueryBuilder) GetCover(sceneID int) ([]byte, error) {
	return qb.imageRepository().get(sceneID)
}

func (qb *sceneQueryBuilder) UpdateCover(sceneID int, image []byte) error {
	return qb.imageRepository().replace(sceneID, image)
}

func (qb *sceneQueryBuilder) DestroyCover(sceneID int) error {
	return qb.imageRepository().destroy([]int{sceneID})
}

func (qb *sceneQueryBuilder) moviesRepository() *repository {
	return &repository{
		tx:        qb.tx,
		tableName: moviesScenesTable,
		idColumn:  sceneIDColumn,
	}
}

func (qb *sceneQueryBuilder) GetMovies(id int) (ret []models.MoviesScenes, err error) {
	if err := qb.moviesRepository().getAll(id, func(rows *sqlx.Rows) error {
		var ms models.MoviesScenes
		if err := rows.StructScan(&ms); err != nil {
			return err
		}

		ret = append(ret, ms)
		return nil
	}); err != nil {
		return nil, err
	}

	return ret, nil
}

func (qb *sceneQueryBuilder) UpdateMovies(sceneID int, movies []models.MoviesScenes) error {
	// destroy existing joins
	r := qb.moviesRepository()
	if err := r.destroy([]int{sceneID}); err != nil {
		return err
	}

	for _, m := range movies {
		m.SceneID = sceneID
		if _, err := r.insert(m); err != nil {
			return err
		}
	}

	return nil
}

func (qb *sceneQueryBuilder) performersRepository() *joinRepository {
	return &joinRepository{
		repository: repository{
			tx:        qb.tx,
			tableName: performersScenesTable,
			idColumn:  sceneIDColumn,
		},
		fkColumn: performerIDColumn,
	}
}

func (qb *sceneQueryBuilder) GetPerformerIDs(id int) ([]int, error) {
	return qb.performersRepository().getIDs(id)
}

func (qb *sceneQueryBuilder) UpdatePerformers(id int, performerIDs []int) error {
	// Delete the existing joins and then create new ones
	return qb.performersRepository().replace(id, performerIDs)
}

func (qb *sceneQueryBuilder) tagsRepository() *joinRepository {
	return &joinRepository{
		repository: repository{
			tx:        qb.tx,
			tableName: scenesTagsTable,
			idColumn:  sceneIDColumn,
		},
		fkColumn: tagIDColumn,
	}
}

func (qb *sceneQueryBuilder) GetTagIDs(id int) ([]int, error) {
	return qb.tagsRepository().getIDs(id)
}

func (qb *sceneQueryBuilder) UpdateTags(id int, tagIDs []int) error {
	// Delete the existing joins and then create new ones
	return qb.tagsRepository().replace(id, tagIDs)
}

func (qb *sceneQueryBuilder) galleriesRepository() *joinRepository {
	return &joinRepository{
		repository: repository{
			tx:        qb.tx,
			tableName: scenesGalleriesTable,
			idColumn:  sceneIDColumn,
		},
		fkColumn: galleryIDColumn,
	}
}

func (qb *sceneQueryBuilder) GetGalleryIDs(id int) ([]int, error) {
	return qb.galleriesRepository().getIDs(id)
}

func (qb *sceneQueryBuilder) UpdateGalleries(id int, galleryIDs []int) error {
	// Delete the existing joins and then create new ones
	return qb.galleriesRepository().replace(id, galleryIDs)
}

func (qb *sceneQueryBuilder) stashIDRepository() *stashIDRepository {
	return &stashIDRepository{
		repository{
			tx:        qb.tx,
			tableName: "scene_stash_ids",
			idColumn:  sceneIDColumn,
		},
	}
}

func (qb *sceneQueryBuilder) GetStashIDs(sceneID int) ([]*models.StashID, error) {
	return qb.stashIDRepository().get(sceneID)
}

func (qb *sceneQueryBuilder) UpdateStashIDs(sceneID int, stashIDs []models.StashID) error {
	return qb.stashIDRepository().replace(sceneID, stashIDs)
}

func (qb *sceneQueryBuilder) FindDuplicates(distance int) ([][]*models.Scene, error) {
	var dupeIds [][]int
	if distance == 0 {
		var ids []string
		if err := qb.tx.Select(&ids, findExactDuplicateQuery); err != nil {
			return nil, err
		}

		for _, id := range ids {
			strIds := strings.Split(id, ",")
			var sceneIds []int
			for _, strId := range strIds {
				if intId, err := strconv.Atoi(strId); err == nil {
					sceneIds = append(sceneIds, intId)
				}
			}
			dupeIds = append(dupeIds, sceneIds)
		}
	} else {
		var hashes []*utils.Phash

		if err := qb.queryFunc(findAllPhashesQuery, nil, false, func(rows *sqlx.Rows) error {
			phash := utils.Phash{
				Bucket: -1,
			}
			if err := rows.StructScan(&phash); err != nil {
				return err
			}

			hashes = append(hashes, &phash)
			return nil
		}); err != nil {
			return nil, err
		}

		dupeIds = utils.FindDuplicates(hashes, distance)
	}

	var duplicates [][]*models.Scene
	for _, sceneIds := range dupeIds {
		if scenes, err := qb.FindMany(sceneIds); err == nil {
			duplicates = append(duplicates, scenes)
		}
	}

	return duplicates, nil
}
