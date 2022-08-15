package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/stashapp/stash/pkg/models"
	"github.com/stashapp/stash/pkg/utils"
)

const performerTable = "performers"
const performerIDColumn = "performer_id"
const performersTagsTable = "performers_tags"
const performersImageTable = "performers_image" // performer cover image

var countPerformersForTagQuery = `
SELECT tag_id AS id FROM performers_tags
WHERE performers_tags.tag_id = ?
GROUP BY performers_tags.performer_id
`

type performerQueryBuilder struct {
	repository
}

func NewPerformerReaderWriter(tx dbi) *performerQueryBuilder {
	return &performerQueryBuilder{
		repository{
			tx:        tx,
			tableName: performerTable,
			idColumn:  idColumn,
		},
	}
}

func (qb *performerQueryBuilder) Create(newObject models.Performer) (*models.Performer, error) {
	var ret models.Performer
	if err := qb.insertObject(newObject, &ret); err != nil {
		return nil, err
	}

	return &ret, nil
}

func (qb *performerQueryBuilder) Update(updatedObject models.PerformerPartial) (*models.Performer, error) {
	const partial = true
	if err := qb.update(updatedObject.ID, updatedObject, partial); err != nil {
		return nil, err
	}

	var ret models.Performer
	if err := qb.get(updatedObject.ID, &ret); err != nil {
		return nil, err
	}

	return &ret, nil
}

func (qb *performerQueryBuilder) UpdateFull(updatedObject models.Performer) (*models.Performer, error) {
	const partial = false
	if err := qb.update(updatedObject.ID, updatedObject, partial); err != nil {
		return nil, err
	}

	var ret models.Performer
	if err := qb.get(updatedObject.ID, &ret); err != nil {
		return nil, err
	}

	return &ret, nil
}

func (qb *performerQueryBuilder) Destroy(id int) error {
	// TODO - add on delete cascade to performers_scenes
	_, err := qb.tx.Exec("DELETE FROM performers_scenes WHERE performer_id = ?", id)
	if err != nil {
		return err
	}

	return qb.destroyExisting([]int{id})
}

func (qb *performerQueryBuilder) Find(id int) (*models.Performer, error) {
	var ret models.Performer
	if err := qb.get(id, &ret); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &ret, nil
}

func (qb *performerQueryBuilder) FindMany(ids []int) ([]*models.Performer, error) {
	var performers []*models.Performer
	for _, id := range ids {
		performer, err := qb.Find(id)
		if err != nil {
			return nil, err
		}

		if performer == nil {
			return nil, fmt.Errorf("performer with id %d not found", id)
		}

		performers = append(performers, performer)
	}

	return performers, nil
}

func (qb *performerQueryBuilder) FindBySceneID(sceneID int) ([]*models.Performer, error) {
	query := selectAll("performers") + `
		LEFT JOIN performers_scenes as scenes_join on scenes_join.performer_id = performers.id
		WHERE scenes_join.scene_id = ?
	`
	args := []interface{}{sceneID}
	return qb.queryPerformers(query, args)
}

func (qb *performerQueryBuilder) FindByImageID(imageID int) ([]*models.Performer, error) {
	query := selectAll("performers") + `
		LEFT JOIN performers_images as images_join on images_join.performer_id = performers.id
		WHERE images_join.image_id = ?
	`
	args := []interface{}{imageID}
	return qb.queryPerformers(query, args)
}

func (qb *performerQueryBuilder) FindByGalleryID(galleryID int) ([]*models.Performer, error) {
	query := selectAll("performers") + `
		LEFT JOIN performers_galleries as galleries_join on galleries_join.performer_id = performers.id
		WHERE galleries_join.gallery_id = ?
	`
	args := []interface{}{galleryID}
	return qb.queryPerformers(query, args)
}

func (qb *performerQueryBuilder) FindNamesBySceneID(sceneID int) ([]*models.Performer, error) {
	query := `
		SELECT performers.name FROM performers
		LEFT JOIN performers_scenes as scenes_join on scenes_join.performer_id = performers.id
		WHERE scenes_join.scene_id = ?
	`
	args := []interface{}{sceneID}
	return qb.queryPerformers(query, args)
}

func (qb *performerQueryBuilder) FindByNames(names []string, nocase bool) ([]*models.Performer, error) {
	query := "SELECT * FROM performers WHERE name"
	if nocase {
		query += " COLLATE NOCASE"
	}
	query += " IN " + getInBinding(len(names))

	var args []interface{}
	for _, name := range names {
		args = append(args, name)
	}
	return qb.queryPerformers(query, args)
}

func (qb *performerQueryBuilder) CountByTagID(tagID int) (int, error) {
	args := []interface{}{tagID}
	return qb.runCountQuery(qb.buildCountQuery(countPerformersForTagQuery), args)
}

func (qb *performerQueryBuilder) Count() (int, error) {
	return qb.runCountQuery(qb.buildCountQuery("SELECT performers.id FROM performers"), nil)
}

func (qb *performerQueryBuilder) All() ([]*models.Performer, error) {
	return qb.queryPerformers(selectAll("performers")+qb.getPerformerSort(nil), nil)
}

func (qb *performerQueryBuilder) QueryForAutoTag(words []string) ([]*models.Performer, error) {
	// TODO - Query needs to be changed to support queries of this type, and
	// this method should be removed
	query := selectAll(performerTable)

	var whereClauses []string
	var args []interface{}

	for _, w := range words {
		whereClauses = append(whereClauses, "name like ?")
		args = append(args, w+"%")
		// TODO - commented out until alias matching works both ways
		// whereClauses = append(whereClauses, "aliases like ?")
		// args = append(args, w+"%")
	}

	whereOr := "(" + strings.Join(whereClauses, " OR ") + ")"
	where := strings.Join([]string{
		"ignore_auto_tag = 0",
		whereOr,
	}, " AND ")
	return qb.queryPerformers(query+" WHERE "+where, args)
}

func (qb *performerQueryBuilder) validateFilter(filter *models.PerformerFilterType) error {
	const and = "AND"
	const or = "OR"
	const not = "NOT"

	if filter.And != nil {
		if filter.Or != nil {
			return illegalFilterCombination(and, or)
		}
		if filter.Not != nil {
			return illegalFilterCombination(and, not)
		}

		return qb.validateFilter(filter.And)
	}

	if filter.Or != nil {
		if filter.Not != nil {
			return illegalFilterCombination(or, not)
		}

		return qb.validateFilter(filter.Or)
	}

	if filter.Not != nil {
		return qb.validateFilter(filter.Not)
	}

	return nil
}

func (qb *performerQueryBuilder) makeFilter(filter *models.PerformerFilterType) *filterBuilder {
	query := &filterBuilder{}

	if filter.And != nil {
		query.and(qb.makeFilter(filter.And))
	}
	if filter.Or != nil {
		query.or(qb.makeFilter(filter.Or))
	}
	if filter.Not != nil {
		query.not(qb.makeFilter(filter.Not))
	}

	const tableName = performerTable
	query.handleCriterion(stringCriterionHandler(filter.Name, tableName+".name"))
	query.handleCriterion(stringCriterionHandler(filter.Details, tableName+".details"))

	query.handleCriterion(boolCriterionHandler(filter.FilterFavorites, tableName+".favorite"))
	query.handleCriterion(boolCriterionHandler(filter.IgnoreAutoTag, tableName+".ignore_auto_tag"))

	query.handleCriterion(yearFilterCriterionHandler(filter.BirthYear, tableName+".birthdate"))
	query.handleCriterion(yearFilterCriterionHandler(filter.DeathYear, tableName+".death_date"))

	query.handleCriterion(performerAgeFilterCriterionHandler(filter.Age))

	query.handleCriterion(criterionHandlerFunc(func(f *filterBuilder) {
		if gender := filter.Gender; gender != nil {
			f.addWhere(tableName+".gender = ?", gender.Value.String())
		}
	}))

	query.handleCriterion(performerIsMissingCriterionHandler(qb, filter.IsMissing))
	query.handleCriterion(stringCriterionHandler(filter.Ethnicity, tableName+".ethnicity"))
	query.handleCriterion(stringCriterionHandler(filter.Country, tableName+".country"))
	query.handleCriterion(stringCriterionHandler(filter.EyeColor, tableName+".eye_color"))
	query.handleCriterion(stringCriterionHandler(filter.Height, tableName+".height"))
	query.handleCriterion(stringCriterionHandler(filter.Measurements, tableName+".measurements"))
	query.handleCriterion(stringCriterionHandler(filter.FakeTits, tableName+".fake_tits"))
	query.handleCriterion(stringCriterionHandler(filter.CareerLength, tableName+".career_length"))
	query.handleCriterion(stringCriterionHandler(filter.Tattoos, tableName+".tattoos"))
	query.handleCriterion(stringCriterionHandler(filter.Piercings, tableName+".piercings"))
	query.handleCriterion(intCriterionHandler(filter.Rating, tableName+".rating"))
	query.handleCriterion(stringCriterionHandler(filter.HairColor, tableName+".hair_color"))
	query.handleCriterion(stringCriterionHandler(filter.URL, tableName+".url"))
	query.handleCriterion(intCriterionHandler(filter.Weight, tableName+".weight"))
	query.handleCriterion(criterionHandlerFunc(func(f *filterBuilder) {
		if filter.StashID != nil {
			qb.stashIDRepository().join(f, "performer_stash_ids", "performers.id")
			stringCriterionHandler(filter.StashID, "performer_stash_ids.stash_id")(f)
		}
	}))

	// TODO - need better handling of aliases
	query.handleCriterion(stringCriterionHandler(filter.Aliases, tableName+".aliases"))

	query.handleCriterion(performerTagsCriterionHandler(qb, filter.Tags))

	query.handleCriterion(performerStudiosCriterionHandler(qb, filter.Studios))

	query.handleCriterion(performerTagCountCriterionHandler(qb, filter.TagCount))
	query.handleCriterion(performerSceneCountCriterionHandler(qb, filter.SceneCount))
	query.handleCriterion(performerImageCountCriterionHandler(qb, filter.ImageCount))
	query.handleCriterion(performerGalleryCountCriterionHandler(qb, filter.GalleryCount))

	return query
}

func (qb *performerQueryBuilder) Query(performerFilter *models.PerformerFilterType, findFilter *models.FindFilterType) ([]*models.Performer, int, error) {
	if performerFilter == nil {
		performerFilter = &models.PerformerFilterType{}
	}
	if findFilter == nil {
		findFilter = &models.FindFilterType{}
	}

	query := qb.newQuery()
	distinctIDs(&query, performerTable)

	if q := findFilter.Q; q != nil && *q != "" {
		searchColumns := []string{"performers.name", "performers.aliases"}
		query.parseQueryString(searchColumns, *q)
	}

	if err := qb.validateFilter(performerFilter); err != nil {
		return nil, 0, err
	}
	filter := qb.makeFilter(performerFilter)

	query.addFilter(filter)

	query.sortAndPagination = qb.getPerformerSort(findFilter) + getPagination(findFilter)
	idsResult, countResult, err := query.executeFind()
	if err != nil {
		return nil, 0, err
	}

	var performers []*models.Performer
	for _, id := range idsResult {
		performer, err := qb.Find(id)
		if err != nil {
			return nil, 0, err
		}
		performers = append(performers, performer)
	}

	return performers, countResult, nil
}

func performerIsMissingCriterionHandler(qb *performerQueryBuilder, isMissing *string) criterionHandlerFunc {
	return func(f *filterBuilder) {
		if isMissing != nil && *isMissing != "" {
			switch *isMissing {
			case "scenes": // Deprecated: use `scene_count == 0` filter instead
				f.addLeftJoin(performersScenesTable, "scenes_join", "scenes_join.performer_id = performers.id")
				f.addWhere("scenes_join.scene_id IS NULL")
			case "image":
				f.addLeftJoin(performersImageTable, "image_join", "image_join.performer_id = performers.id")
				f.addWhere("image_join.performer_id IS NULL")
			case "stash_id":
				qb.stashIDRepository().join(f, "performer_stash_ids", "performers.id")
				f.addWhere("performer_stash_ids.performer_id IS NULL")
			default:
				f.addWhere("(performers." + *isMissing + " IS NULL OR TRIM(performers." + *isMissing + ") = '')")
			}
		}
	}
}

func yearFilterCriterionHandler(year *models.IntCriterionInput, col string) criterionHandlerFunc {
	return func(f *filterBuilder) {
		if year != nil && year.Modifier.IsValid() {
			clause, args := getIntCriterionWhereClause("cast(strftime('%Y', "+col+") as int)", *year)
			f.addWhere(clause, args...)
		}
	}
}

func performerAgeFilterCriterionHandler(age *models.IntCriterionInput) criterionHandlerFunc {
	return func(f *filterBuilder) {
		if age != nil && age.Modifier.IsValid() {
			clause, args := getIntCriterionWhereClause(
				"cast(IFNULL(strftime('%Y.%m%d', performers.death_date), strftime('%Y.%m%d', 'now')) - strftime('%Y.%m%d', performers.birthdate) as int)",
				*age,
			)
			f.addWhere(clause, args...)
		}
	}
}

func performerTagsCriterionHandler(qb *performerQueryBuilder, tags *models.HierarchicalMultiCriterionInput) criterionHandlerFunc {
	h := joinedHierarchicalMultiCriterionHandlerBuilder{
		tx: qb.tx,

		primaryTable: performerTable,
		foreignTable: tagTable,
		foreignFK:    "tag_id",

		relationsTable: "tags_relations",
		joinAs:         "image_tag",
		joinTable:      performersTagsTable,
		primaryFK:      performerIDColumn,
	}

	return h.handler(tags)
}

func performerTagCountCriterionHandler(qb *performerQueryBuilder, count *models.IntCriterionInput) criterionHandlerFunc {
	h := countCriterionHandlerBuilder{
		primaryTable: performerTable,
		joinTable:    performersTagsTable,
		primaryFK:    performerIDColumn,
	}

	return h.handler(count)
}

func performerSceneCountCriterionHandler(qb *performerQueryBuilder, count *models.IntCriterionInput) criterionHandlerFunc {
	h := countCriterionHandlerBuilder{
		primaryTable: performerTable,
		joinTable:    performersScenesTable,
		primaryFK:    performerIDColumn,
	}

	return h.handler(count)
}

func performerImageCountCriterionHandler(qb *performerQueryBuilder, count *models.IntCriterionInput) criterionHandlerFunc {
	h := countCriterionHandlerBuilder{
		primaryTable: performerTable,
		joinTable:    performersImagesTable,
		primaryFK:    performerIDColumn,
	}

	return h.handler(count)
}

func performerGalleryCountCriterionHandler(qb *performerQueryBuilder, count *models.IntCriterionInput) criterionHandlerFunc {
	h := countCriterionHandlerBuilder{
		primaryTable: performerTable,
		joinTable:    performersGalleriesTable,
		primaryFK:    performerIDColumn,
	}

	return h.handler(count)
}

func performerStudiosCriterionHandler(qb *performerQueryBuilder, studios *models.HierarchicalMultiCriterionInput) criterionHandlerFunc {
	return func(f *filterBuilder) {
		if studios != nil {
			formatMaps := []utils.StrFormatMap{
				{
					"primaryTable": sceneTable,
					"joinTable":    performersScenesTable,
					"primaryFK":    sceneIDColumn,
				},
				{
					"primaryTable": imageTable,
					"joinTable":    performersImagesTable,
					"primaryFK":    imageIDColumn,
				},
				{
					"primaryTable": galleryTable,
					"joinTable":    performersGalleriesTable,
					"primaryFK":    galleryIDColumn,
				},
			}

			if studios.Modifier == models.CriterionModifierIsNull || studios.Modifier == models.CriterionModifierNotNull {
				var notClause string
				if studios.Modifier == models.CriterionModifierNotNull {
					notClause = "NOT"
				}

				var conditions []string
				for _, c := range formatMaps {
					f.addLeftJoin(c["joinTable"].(string), "", fmt.Sprintf("%s.performer_id = performers.id", c["joinTable"]))
					f.addLeftJoin(c["primaryTable"].(string), "", fmt.Sprintf("%s.%s = %s.id", c["joinTable"], c["primaryFK"], c["primaryTable"]))

					conditions = append(conditions, fmt.Sprintf("%s.studio_id IS NULL", c["primaryTable"]))
				}

				f.addWhere(fmt.Sprintf("%s (%s)", notClause, strings.Join(conditions, " AND ")))
				return
			}

			if len(studios.Value) == 0 {
				return
			}

			var clauseCondition string

			switch studios.Modifier {
			case models.CriterionModifierIncludes:
				// return performers who appear in scenes/images/galleries with any of the given studios
				clauseCondition = "NOT"
			case models.CriterionModifierExcludes:
				// exclude performers who appear in scenes/images/galleries with any of the given studios
				clauseCondition = ""
			default:
				return
			}

			const derivedPerformerStudioTable = "performer_studio"
			valuesClause := getHierarchicalValues(qb.tx, studios.Value, studioTable, "", "parent_id", studios.Depth)
			f.addWith("studio(root_id, item_id) AS (" + valuesClause + ")")

			templStr := `SELECT performer_id FROM {primaryTable}
	INNER JOIN {joinTable} ON {primaryTable}.id = {joinTable}.{primaryFK}
	INNER JOIN studio ON {primaryTable}.studio_id = studio.item_id`

			var unions []string
			for _, c := range formatMaps {
				unions = append(unions, utils.StrFormat(templStr, c))
			}

			f.addWith(fmt.Sprintf("%s AS (%s)", derivedPerformerStudioTable, strings.Join(unions, " UNION ")))

			f.addLeftJoin(derivedPerformerStudioTable, "", fmt.Sprintf("performers.id = %s.performer_id", derivedPerformerStudioTable))
			f.addWhere(fmt.Sprintf("%s.performer_id IS %s NULL", derivedPerformerStudioTable, clauseCondition))
		}
	}
}

func (qb *performerQueryBuilder) getPerformerSort(findFilter *models.FindFilterType) string {
	var sort string
	var direction string
	if findFilter == nil {
		sort = "name"
		direction = "ASC"
	} else {
		sort = findFilter.GetSort("name")
		direction = findFilter.GetDirection()
	}

	if sort == "tag_count" {
		return getCountSort(performerTable, performersTagsTable, performerIDColumn, direction)
	}
	if sort == "scenes_count" {
		return getCountSort(performerTable, performersScenesTable, performerIDColumn, direction)
	}
	if sort == "images_count" {
		return getCountSort(performerTable, performersImagesTable, performerIDColumn, direction)
	}
	if sort == "galleries_count" {
		return getCountSort(performerTable, performersGalleriesTable, performerIDColumn, direction)
	}

	return getSort(sort, direction, "performers")
}

func (qb *performerQueryBuilder) queryPerformers(query string, args []interface{}) ([]*models.Performer, error) {
	var ret models.Performers
	if err := qb.query(query, args, &ret); err != nil {
		return nil, err
	}

	return []*models.Performer(ret), nil
}

func (qb *performerQueryBuilder) tagsRepository() *joinRepository {
	return &joinRepository{
		repository: repository{
			tx:        qb.tx,
			tableName: performersTagsTable,
			idColumn:  performerIDColumn,
		},
		fkColumn: tagIDColumn,
	}
}

func (qb *performerQueryBuilder) GetTagIDs(id int) ([]int, error) {
	return qb.tagsRepository().getIDs(id)
}

func (qb *performerQueryBuilder) UpdateTags(id int, tagIDs []int) error {
	// Delete the existing joins and then create new ones
	return qb.tagsRepository().replace(id, tagIDs)
}

func (qb *performerQueryBuilder) imageRepository() *imageRepository {
	return &imageRepository{
		repository: repository{
			tx:        qb.tx,
			tableName: "performers_image",
			idColumn:  performerIDColumn,
		},
		imageColumn: "image",
	}
}

func (qb *performerQueryBuilder) GetImage(performerID int) ([]byte, error) {
	return qb.imageRepository().get(performerID)
}

func (qb *performerQueryBuilder) UpdateImage(performerID int, image []byte) error {
	return qb.imageRepository().replace(performerID, image)
}

func (qb *performerQueryBuilder) DestroyImage(performerID int) error {
	return qb.imageRepository().destroy([]int{performerID})
}

func (qb *performerQueryBuilder) stashIDRepository() *stashIDRepository {
	return &stashIDRepository{
		repository{
			tx:        qb.tx,
			tableName: "performer_stash_ids",
			idColumn:  performerIDColumn,
		},
	}
}

func (qb *performerQueryBuilder) GetStashIDs(performerID int) ([]*models.StashID, error) {
	return qb.stashIDRepository().get(performerID)
}

func (qb *performerQueryBuilder) UpdateStashIDs(performerID int, stashIDs []models.StashID) error {
	return qb.stashIDRepository().replace(performerID, stashIDs)
}

func (qb *performerQueryBuilder) FindByStashID(stashID models.StashID) ([]*models.Performer, error) {
	query := selectAll("performers") + `
		LEFT JOIN performer_stash_ids on performer_stash_ids.performer_id = performers.id
		WHERE performer_stash_ids.stash_id = ?
		AND performer_stash_ids.endpoint = ?
	`
	args := []interface{}{stashID.StashID, stashID.Endpoint}
	return qb.queryPerformers(query, args)
}

func (qb *performerQueryBuilder) FindByStashIDStatus(hasStashID bool, stashboxEndpoint string) ([]*models.Performer, error) {
	query := selectAll("performers") + `
		LEFT JOIN performer_stash_ids on performer_stash_ids.performer_id = performers.id
	`

	if hasStashID {
		query += `
			WHERE performer_stash_ids.stash_id IS NOT NULL
			AND performer_stash_ids.endpoint = ?
		`
	} else {
		query += `
			WHERE performer_stash_ids.stash_id IS NULL
		`
	}

	args := []interface{}{stashboxEndpoint}
	return qb.queryPerformers(query, args)
}
