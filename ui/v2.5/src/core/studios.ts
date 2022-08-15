import * as GQL from "src/core/generated-graphql";
import { StudiosCriterion } from "src/models/list-filter/criteria/studios";
import { ListFilterModel } from "src/models/list-filter/filter";

export const studioFilterHook = (studio: GQL.StudioDataFragment) => {
  return (filter: ListFilterModel) => {
    const studioValue = { id: studio.id, label: studio.name };
    // if studio is already present, then we modify it, otherwise add
    let studioCriterion = filter.criteria.find((c) => {
      return c.criterionOption.type === "studios";
    }) as StudiosCriterion;

    if (
      studioCriterion &&
      (studioCriterion.modifier === GQL.CriterionModifier.IncludesAll ||
        studioCriterion.modifier === GQL.CriterionModifier.Includes)
    ) {
      // we should be showing studio only. Remove other values
      studioCriterion.value.items = studioCriterion.value.items.filter(
        (v) => v.id === studio.id
      );

      if (studioCriterion.value.items.length === 0) {
        studioCriterion.value.items.push(studioValue);
      }
    } else {
      // overwrite
      studioCriterion = new StudiosCriterion();
      studioCriterion.value = {
        items: [studioValue],
        depth: 0,
      };
      filter.criteria.push(studioCriterion);
    }

    return filter;
  };
};
