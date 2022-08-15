import * as GQL from "src/core/generated-graphql";
import { PerformersCriterion } from "src/models/list-filter/criteria/performers";
import { CountryCriterion } from "src/models/list-filter/criteria/country";
import {
  StudiosCriterion,
  ParentStudiosCriterion,
} from "src/models/list-filter/criteria/studios";
import {
  ChildTagsCriterionOption,
  ParentTagsCriterionOption,
  TagsCriterion,
  TagsCriterionOption,
} from "src/models/list-filter/criteria/tags";
import { ListFilterModel } from "src/models/list-filter/filter";
import { MoviesCriterion } from "src/models/list-filter/criteria/movies";
import {
  Criterion,
  CriterionValue,
} from "src/models/list-filter/criteria/criterion";
import { GalleriesCriterion } from "src/models/list-filter/criteria/galleries";
import { PhashCriterion } from "src/models/list-filter/criteria/phash";

function addExtraCriteria(
  dest: Criterion<CriterionValue>[],
  src?: Criterion<CriterionValue>[]
) {
  if (src && src.length > 0) {
    dest.push(...src);
  }
}

const makePerformerScenesUrl = (
  performer: Partial<GQL.PerformerDataFragment>,
  extraCriteria?: Criterion<CriterionValue>[]
) => {
  if (!performer.id) return "#";
  const filter = new ListFilterModel(GQL.FilterMode.Scenes);
  const criterion = new PerformersCriterion();
  criterion.value = [
    { id: performer.id, label: performer.name || `Performer ${performer.id}` },
  ];
  filter.criteria.push(criterion);
  addExtraCriteria(filter.criteria, extraCriteria);
  return `/scenes?${filter.makeQueryParameters()}`;
};

const makePerformerImagesUrl = (
  performer: Partial<GQL.PerformerDataFragment>,
  extraCriteria?: Criterion<CriterionValue>[]
) => {
  if (!performer.id) return "#";
  const filter = new ListFilterModel(GQL.FilterMode.Images);
  const criterion = new PerformersCriterion();
  criterion.value = [
    { id: performer.id, label: performer.name || `Performer ${performer.id}` },
  ];
  filter.criteria.push(criterion);
  addExtraCriteria(filter.criteria, extraCriteria);
  return `/images?${filter.makeQueryParameters()}`;
};

const makePerformerGalleriesUrl = (
  performer: Partial<GQL.PerformerDataFragment>,
  extraCriteria?: Criterion<CriterionValue>[]
) => {
  if (!performer.id) return "#";
  const filter = new ListFilterModel(GQL.FilterMode.Galleries);
  const criterion = new PerformersCriterion();
  criterion.value = [
    { id: performer.id, label: performer.name || `Performer ${performer.id}` },
  ];
  filter.criteria.push(criterion);
  addExtraCriteria(filter.criteria, extraCriteria);
  return `/galleries?${filter.makeQueryParameters()}`;
};

const makePerformerMoviesUrl = (
  performer: Partial<GQL.PerformerDataFragment>,
  extraCriteria?: Criterion<CriterionValue>[]
) => {
  if (!performer.id) return "#";
  const filter = new ListFilterModel(GQL.FilterMode.Movies);
  const criterion = new PerformersCriterion();
  criterion.value = [
    { id: performer.id, label: performer.name || `Performer ${performer.id}` },
  ];
  filter.criteria.push(criterion);
  addExtraCriteria(filter.criteria, extraCriteria);
  return `/movies?${filter.makeQueryParameters()}`;
};

const makePerformersCountryUrl = (
  performer: Partial<GQL.PerformerDataFragment>
) => {
  if (!performer.id) return "#";
  const filter = new ListFilterModel(GQL.FilterMode.Performers);
  const criterion = new CountryCriterion();
  criterion.value = `${performer.country}`;
  filter.criteria.push(criterion);
  return `/performers?${filter.makeQueryParameters()}`;
};

const makeStudioScenesUrl = (studio: Partial<GQL.StudioDataFragment>) => {
  if (!studio.id) return "#";
  const filter = new ListFilterModel(GQL.FilterMode.Scenes);
  const criterion = new StudiosCriterion();
  criterion.value = {
    items: [{ id: studio.id, label: studio.name || `Studio ${studio.id}` }],
    depth: 0,
  };
  filter.criteria.push(criterion);
  return `/scenes?${filter.makeQueryParameters()}`;
};

const makeStudioImagesUrl = (studio: Partial<GQL.StudioDataFragment>) => {
  if (!studio.id) return "#";
  const filter = new ListFilterModel(GQL.FilterMode.Images);
  const criterion = new StudiosCriterion();
  criterion.value = {
    items: [{ id: studio.id, label: studio.name || `Studio ${studio.id}` }],
    depth: 0,
  };
  filter.criteria.push(criterion);
  return `/images?${filter.makeQueryParameters()}`;
};

const makeStudioGalleriesUrl = (studio: Partial<GQL.StudioDataFragment>) => {
  if (!studio.id) return "#";
  const filter = new ListFilterModel(GQL.FilterMode.Galleries);
  const criterion = new StudiosCriterion();
  criterion.value = {
    items: [{ id: studio.id, label: studio.name || `Studio ${studio.id}` }],
    depth: 0,
  };
  filter.criteria.push(criterion);
  return `/galleries?${filter.makeQueryParameters()}`;
};

const makeStudioMoviesUrl = (studio: Partial<GQL.StudioDataFragment>) => {
  if (!studio.id) return "#";
  const filter = new ListFilterModel(GQL.FilterMode.Movies);
  const criterion = new StudiosCriterion();
  criterion.value = {
    items: [{ id: studio.id, label: studio.name || `Studio ${studio.id}` }],
    depth: 0,
  };
  filter.criteria.push(criterion);
  return `/movies?${filter.makeQueryParameters()}`;
};

const makeChildStudiosUrl = (studio: Partial<GQL.StudioDataFragment>) => {
  if (!studio.id) return "#";
  const filter = new ListFilterModel(GQL.FilterMode.Studios);
  const criterion = new ParentStudiosCriterion();
  criterion.value = [
    { id: studio.id, label: studio.name || `Studio ${studio.id}` },
  ];
  filter.criteria.push(criterion);
  return `/studios?${filter.makeQueryParameters()}`;
};

const makeMovieScenesUrl = (movie: Partial<GQL.MovieDataFragment>) => {
  if (!movie.id) return "#";
  const filter = new ListFilterModel(GQL.FilterMode.Scenes);
  const criterion = new MoviesCriterion();
  criterion.value = [
    { id: movie.id, label: movie.name || `Movie ${movie.id}` },
  ];
  filter.criteria.push(criterion);
  return `/scenes?${filter.makeQueryParameters()}`;
};

const makeParentTagsUrl = (tag: Partial<GQL.TagDataFragment>) => {
  if (!tag.id) return "#";
  const filter = new ListFilterModel(GQL.FilterMode.Tags);
  const criterion = new TagsCriterion(ChildTagsCriterionOption);
  criterion.value = {
    items: [
      {
        id: tag.id,
        label: tag.name || `Tag ${tag.id}`,
      },
    ],
    depth: 0,
  };
  filter.criteria.push(criterion);
  return `/tags?${filter.makeQueryParameters()}`;
};

const makeChildTagsUrl = (tag: Partial<GQL.TagDataFragment>) => {
  if (!tag.id) return "#";
  const filter = new ListFilterModel(GQL.FilterMode.Tags);
  const criterion = new TagsCriterion(ParentTagsCriterionOption);
  criterion.value = {
    items: [
      {
        id: tag.id,
        label: tag.name || `Tag ${tag.id}`,
      },
    ],
    depth: 0,
  };
  filter.criteria.push(criterion);
  return `/tags?${filter.makeQueryParameters()}`;
};

const makeTagScenesUrl = (tag: Partial<GQL.TagDataFragment>) => {
  if (!tag.id) return "#";
  const filter = new ListFilterModel(GQL.FilterMode.Scenes);
  const criterion = new TagsCriterion(TagsCriterionOption);
  criterion.value = {
    items: [{ id: tag.id, label: tag.name || `Tag ${tag.id}` }],
    depth: 0,
  };
  filter.criteria.push(criterion);
  return `/scenes?${filter.makeQueryParameters()}`;
};

const makeTagPerformersUrl = (tag: Partial<GQL.TagDataFragment>) => {
  if (!tag.id) return "#";
  const filter = new ListFilterModel(GQL.FilterMode.Performers);
  const criterion = new TagsCriterion(TagsCriterionOption);
  criterion.value = {
    items: [{ id: tag.id, label: tag.name || `Tag ${tag.id}` }],
    depth: 0,
  };
  filter.criteria.push(criterion);
  return `/performers?${filter.makeQueryParameters()}`;
};

const makeTagSceneMarkersUrl = (tag: Partial<GQL.TagDataFragment>) => {
  if (!tag.id) return "#";
  const filter = new ListFilterModel(GQL.FilterMode.SceneMarkers);
  const criterion = new TagsCriterion(TagsCriterionOption);
  criterion.value = {
    items: [{ id: tag.id, label: tag.name || `Tag ${tag.id}` }],
    depth: 0,
  };
  filter.criteria.push(criterion);
  return `/scenes/markers?${filter.makeQueryParameters()}`;
};

const makeTagGalleriesUrl = (tag: Partial<GQL.TagDataFragment>) => {
  if (!tag.id) return "#";
  const filter = new ListFilterModel(GQL.FilterMode.Galleries);
  const criterion = new TagsCriterion(TagsCriterionOption);
  criterion.value = {
    items: [{ id: tag.id, label: tag.name || `Tag ${tag.id}` }],
    depth: 0,
  };
  filter.criteria.push(criterion);
  return `/galleries?${filter.makeQueryParameters()}`;
};

const makeTagImagesUrl = (tag: Partial<GQL.TagDataFragment>) => {
  if (!tag.id) return "#";
  const filter = new ListFilterModel(GQL.FilterMode.Images);
  const criterion = new TagsCriterion(TagsCriterionOption);
  criterion.value = {
    items: [{ id: tag.id, label: tag.name || `Tag ${tag.id}` }],
    depth: 0,
  };
  filter.criteria.push(criterion);
  return `/images?${filter.makeQueryParameters()}`;
};

const makeSceneMarkerUrl = (
  sceneMarker: Partial<GQL.SceneMarkerDataFragment>
) => {
  if (!sceneMarker.id || !sceneMarker.scene) return "#";
  return `/scenes/${sceneMarker.scene.id}?t=${sceneMarker.seconds}`;
};

const makeScenesPHashMatchUrl = (phash: GQL.Maybe<string> | undefined) => {
  if (!phash) return "#";
  const filter = new ListFilterModel(GQL.FilterMode.Scenes);
  const criterion = new PhashCriterion();
  criterion.value = phash;
  filter.criteria.push(criterion);
  return `/scenes?${filter.makeQueryParameters()}`;
};

const makeGalleryImagesUrl = (
  gallery: Partial<GQL.GalleryDataFragment | GQL.SlimGalleryDataFragment>,
  extraCriteria?: Criterion<CriterionValue>[]
) => {
  if (!gallery.id) return "#";
  const filter = new ListFilterModel(GQL.FilterMode.Images);
  const criterion = new GalleriesCriterion();
  criterion.value = [
    { id: gallery.id, label: gallery.title || `Gallery ${gallery.id}` },
  ];
  filter.criteria.push(criterion);
  addExtraCriteria(filter.criteria, extraCriteria);
  return `/images?${filter.makeQueryParameters()}`;
};

export default {
  makePerformerScenesUrl,
  makePerformerImagesUrl,
  makePerformerGalleriesUrl,
  makePerformerMoviesUrl,
  makePerformersCountryUrl,
  makeStudioScenesUrl,
  makeStudioImagesUrl,
  makeStudioGalleriesUrl,
  makeStudioMoviesUrl,
  makeParentTagsUrl,
  makeChildTagsUrl,
  makeTagSceneMarkersUrl,
  makeTagScenesUrl,
  makeTagPerformersUrl,
  makeTagGalleriesUrl,
  makeTagImagesUrl,
  makeScenesPHashMatchUrl,
  makeSceneMarkerUrl,
  makeMovieScenesUrl,
  makeChildStudiosUrl,
  makeGalleryImagesUrl,
};
