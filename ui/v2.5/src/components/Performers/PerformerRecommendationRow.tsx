import React, { FunctionComponent } from "react";
import { useFindPerformers } from "src/core/StashService";
import Slider from "react-slick";
import { PerformerCard } from "./PerformerCard";
import { ListFilterModel } from "src/models/list-filter/filter";
import { getSlickSliderSettings } from "src/core/recommendations";
import { RecommendationRow } from "../FrontPage/RecommendationRow";
import { FormattedMessage } from "react-intl";

interface IProps {
  isTouch: boolean;
  filter: ListFilterModel;
  header: string;
}

export const PerformerRecommendationRow: FunctionComponent<IProps> = (
  props: IProps
) => {
  const result = useFindPerformers(props.filter);
  const cardCount = result.data?.findPerformers.count;

  if (!result.loading && !cardCount) {
    return null;
  }

  return (
    <RecommendationRow
      className="performer-recommendations"
      header={props.header}
      link={
        <a href={`/performers?${props.filter.makeQueryParameters()}`}>
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
              <div
                key={`_${i}`}
                className="performer-skeleton skeleton-card"
              ></div>
            ))
          : result.data?.findPerformers.performers.map((p) => (
              <PerformerCard key={p.id} performer={p} />
            ))}
      </Slider>
    </RecommendationRow>
  );
};
