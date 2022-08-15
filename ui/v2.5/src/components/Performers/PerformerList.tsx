import cloneDeep from "lodash-es/cloneDeep";
import React, { useState } from "react";
import { useIntl } from "react-intl";
import { useHistory } from "react-router-dom";
import Mousetrap from "mousetrap";
import {
  FindPerformersQueryResult,
  SlimPerformerDataFragment,
} from "src/core/generated-graphql";
import {
  queryFindPerformers,
  usePerformersDestroy,
} from "src/core/StashService";
import { usePerformersList } from "src/hooks";
import { showWhenSelected, PersistanceLevel } from "src/hooks/ListHook";
import { ListFilterModel } from "src/models/list-filter/filter";
import { DisplayMode } from "src/models/list-filter/types";
import { PerformerTagger } from "src/components/Tagger";
import { ExportDialog, DeleteEntityDialog } from "src/components/Shared";
import { IPerformerCardExtraCriteria, PerformerCard } from "./PerformerCard";
import { PerformerListTable } from "./PerformerListTable";
import { EditPerformersDialog } from "./EditPerformersDialog";

interface IPerformerList {
  filterHook?: (filter: ListFilterModel) => ListFilterModel;
  persistState?: PersistanceLevel;
  extraCriteria?: IPerformerCardExtraCriteria;
}

export const PerformerList: React.FC<IPerformerList> = ({
  filterHook,
  persistState,
  extraCriteria,
}) => {
  const intl = useIntl();
  const history = useHistory();
  const [isExportDialogOpen, setIsExportDialogOpen] = useState(false);
  const [isExportAll, setIsExportAll] = useState(false);

  const otherOperations = [
    {
      text: intl.formatMessage({ id: "actions.open_random" }),
      onClick: getRandom,
    },
    {
      text: intl.formatMessage({ id: "actions.export" }),
      onClick: onExport,
      isDisplayed: showWhenSelected,
    },
    {
      text: intl.formatMessage({ id: "actions.export_all" }),
      onClick: onExportAll,
    },
  ];

  const addKeybinds = (
    result: FindPerformersQueryResult,
    filter: ListFilterModel
  ) => {
    Mousetrap.bind("p r", () => {
      getRandom(result, filter);
    });

    return () => {
      Mousetrap.unbind("p r");
    };
  };

  async function onExport() {
    setIsExportAll(false);
    setIsExportDialogOpen(true);
  }

  async function onExportAll() {
    setIsExportAll(true);
    setIsExportDialogOpen(true);
  }

  function maybeRenderPerformerExportDialog(selectedIds: Set<string>) {
    if (isExportDialogOpen) {
      return (
        <>
          <ExportDialog
            exportInput={{
              performers: {
                ids: Array.from(selectedIds.values()),
                all: isExportAll,
              },
            }}
            onClose={() => {
              setIsExportDialogOpen(false);
            }}
          />
        </>
      );
    }
  }

  function renderEditPerformersDialog(
    selectedPerformers: SlimPerformerDataFragment[],
    onClose: (applied: boolean) => void
  ) {
    return (
      <>
        <EditPerformersDialog selected={selectedPerformers} onClose={onClose} />
      </>
    );
  }

  const renderDeleteDialog = (
    selectedPerformers: SlimPerformerDataFragment[],
    onClose: (confirmed: boolean) => void
  ) => (
    <DeleteEntityDialog
      selected={selectedPerformers}
      onClose={onClose}
      singularEntity={intl.formatMessage({ id: "performer" })}
      pluralEntity={intl.formatMessage({ id: "performers" })}
      destroyMutation={usePerformersDestroy}
    />
  );

  const listData = usePerformersList({
    otherOperations,
    renderContent,
    renderEditDialog: renderEditPerformersDialog,
    filterHook,
    addKeybinds,
    selectable: true,
    persistState,
    renderDeleteDialog,
  });

  async function getRandom(
    result: FindPerformersQueryResult,
    filter: ListFilterModel
  ) {
    if (result.data?.findPerformers) {
      const { count } = result.data.findPerformers;
      const index = Math.floor(Math.random() * count);
      const filterCopy = cloneDeep(filter);
      filterCopy.itemsPerPage = 1;
      filterCopy.currentPage = index + 1;
      const singleResult = await queryFindPerformers(filterCopy);
      if (
        singleResult &&
        singleResult.data &&
        singleResult.data.findPerformers &&
        singleResult.data.findPerformers.performers.length === 1
      ) {
        const { id } = singleResult!.data!.findPerformers!.performers[0]!;
        history.push(`/performers/${id}`);
      }
    }
  }

  function renderContent(
    result: FindPerformersQueryResult,
    filter: ListFilterModel,
    selectedIds: Set<string>
  ) {
    if (!result.data?.findPerformers) {
      return;
    }
    if (filter.displayMode === DisplayMode.Grid) {
      return (
        <>
          {maybeRenderPerformerExportDialog(selectedIds)}
          <div className="row justify-content-center">
            {result.data.findPerformers.performers.map((p) => (
              <PerformerCard
                key={p.id}
                performer={p}
                selecting={selectedIds.size > 0}
                selected={selectedIds.has(p.id)}
                onSelectedChanged={(selected: boolean, shiftKey: boolean) =>
                  listData.onSelectChange(p.id, selected, shiftKey)
                }
                extraCriteria={extraCriteria}
              />
            ))}
          </div>
        </>
      );
    }
    if (filter.displayMode === DisplayMode.List) {
      return (
        <PerformerListTable
          performers={result.data.findPerformers.performers}
        />
      );
    }
    if (filter.displayMode === DisplayMode.Tagger) {
      return (
        <PerformerTagger performers={result.data.findPerformers.performers} />
      );
    }
  }

  return listData.template;
};
