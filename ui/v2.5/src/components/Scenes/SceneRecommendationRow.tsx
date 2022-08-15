import React, { FunctionComponent, useMemo } from "react";
import { useFindScenes } from "src/core/StashService";
import Slider from "react-slick";
import { SceneCard } from "./SceneCard";
import { SceneQueue } from "src/models/sceneQueue";
import { ListFilterModel } from "src/models/list-filter/filter";
import { getSlickSliderSettings } from "src/core/recommendations";
import { RecommendationRow } from "../FrontPage/RecommendationRow";
import { FormattedMessage } from "react-intl";

interface IProps {
  isTouch: boolean;
  filter: ListFilterModel;
  header: string;
}

export const SceneRecommendationRow: FunctionComponent<IProps> = (
  props: IProps
) => {
  const result = useFindScenes(props.filter);
  const cardCount = result.data?.findScenes.count;

  const queue = useMemo(() => {
    return SceneQueue.fromListFilterModel(props.filter);
  }, [props.filter]);

  if (!result.loading && !cardCount) {
    return null;
  }

  return (
    <RecommendationRow
      className="scene-recommendations"
      header={props.header}
      link={
        <a href={`/scenes?${props.filter.makeQueryParameters()}`}>
          <FormattedMessage id="view_all" />
        </a>
      }
    >
      <Slider
        {...getSlickSliderSettings(
          cardCount ? cardCount : props.filter.itemsPerPage,
          props.isTouch
        )}
      >
        {result.loading
          ? [...Array(props.filter.itemsPerPage)].map((i) => (
              <div key={`_${i}`} className="scene-skeleton skeleton-card"></div>
            ))
          : result.data?.findScenes.scenes.map((scene, index) => (
              <SceneCard
                key={scene.id}
                scene={scene}
                queue={queue}
                index={index}
                zoomIndex={1}
              />
            ))}
      </Slider>
    </RecommendationRow>
  );
};
