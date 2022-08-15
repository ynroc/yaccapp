import React, { FunctionComponent } from "react";
import { useFindStudios } from "src/core/StashService";
import Slider from "react-slick";
import { StudioCard } from "./StudioCard";
import { ListFilterModel } from "src/models/list-filter/filter";
import { getSlickSliderSettings } from "src/core/recommendations";
import { RecommendationRow } from "../FrontPage/RecommendationRow";
import { FormattedMessage } from "react-intl";

interface IProps {
  isTouch: boolean;
  filter: ListFilterModel;
  header: string;
}

export const StudioRecommendationRow: FunctionComponent<IProps> = (
  props: IProps
) => {
  const result = useFindStudios(props.filter);
  const cardCount = result.data?.findStudios.count;

  if (!result.loading && !cardCount) {
    return null;
  }

  return (
    <RecommendationRow
      className="studio-recommendations"
      header={props.header}
      link={
        <a href={`/studios?${props.filter.makeQueryParameters()}`}>
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
                className="studio-skeleton skeleton-card"
              ></div>
            ))
          : result.data?.findStudios.studios.map((s) => (
              <StudioCard key={s.id} studio={s} hideParent={true} />
            ))}
      </Slider>
    </RecommendationRow>
  );
};
