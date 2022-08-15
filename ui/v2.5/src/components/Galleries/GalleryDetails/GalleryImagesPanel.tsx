import React from "react";
import * as GQL from "src/core/generated-graphql";
import { GalleriesCriterion } from "src/models/list-filter/criteria/galleries";
import { ListFilterModel } from "src/models/list-filter/filter";
import { ImageList } from "src/components/Images/ImageList";
import { mutateRemoveGalleryImages } from "src/core/StashService";
import { showWhenSelected, PersistanceLevel } from "src/hooks/ListHook";
import { useToast } from "src/hooks";
import { TextUtils } from "src/utils";
import { useIntl } from "react-intl";
import { faMinus } from "@fortawesome/free-solid-svg-icons";

interface IGalleryDetailsProps {
  gallery: GQL.GalleryDataFragment;
}

export const GalleryImagesPanel: React.FC<IGalleryDetailsProps> = ({
  gallery,
}) => {
  const intl = useIntl();
  const Toast = useToast();

  function filterHook(filter: ListFilterModel) {
    const galleryValue = {
      id: gallery.id!,
      label: gallery.title ?? TextUtils.fileNameFromPath(gallery.path ?? ""),
    };
    // if galleries is already present, then we modify it, otherwise add
    let galleryCriterion = filter.criteria.find((c) => {
      return c.criterionOption.type === "galleries";
    }) as GalleriesCriterion;

    if (
      galleryCriterion &&
      (galleryCriterion.modifier === GQL.CriterionModifier.IncludesAll ||
        galleryCriterion.modifier === GQL.CriterionModifier.Includes)
    ) {
      // add the gallery if not present
      if (
        !galleryCriterion.value.find((p) => {
          return p.id === gallery.id;
        })
      ) {
        galleryCriterion.value.push(galleryValue);
      }

      galleryCriterion.modifier = GQL.CriterionModifier.IncludesAll;
    } else {
      // overwrite
      galleryCriterion = new GalleriesCriterion();
      galleryCriterion.value = [galleryValue];
      filter.criteria.push(galleryCriterion);
    }

    return filter;
  }

  async function removeImages(
    result: GQL.FindImagesQueryResult,
    filter: ListFilterModel,
    selectedIds: Set<string>
  ) {
    try {
      await mutateRemoveGalleryImages({
        gallery_id: gallery.id!,
        image_ids: Array.from(selectedIds.values()),
      });

      Toast.success({
        content: intl.formatMessage(
          { id: "toast.removed_entity" },
          {
            count: selectedIds.size,
            singularEntity: intl.formatMessage({ id: "image" }),
            pluralEntity: intl.formatMessage({ id: "images" }),
          }
        ),
      });
    } catch (e) {
      Toast.error(e);
    }
  }

  const otherOperations = [
    {
      text: intl.formatMessage({ id: "actions.remove_from_gallery" }),
      onClick: removeImages,
      isDisplayed: showWhenSelected,
      postRefetch: true,
      icon: faMinus,
      buttonVariant: "danger",
    },
  ];

  return (
    <ImageList
      filterHook={filterHook}
      extraOperations={otherOperations}
      persistState={PersistanceLevel.VIEW}
      persistanceKey="galleryimages"
    />
  );
};
