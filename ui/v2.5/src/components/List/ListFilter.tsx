import debounce from "lodash-es/debounce";
import cloneDeep from "lodash-es/cloneDeep";
import React, { HTMLAttributes, useEffect, useRef, useState } from "react";
import cx from "classnames";
import Mousetrap from "mousetrap";
import { SortDirectionEnum } from "src/core/generated-graphql";
import {
  Button,
  ButtonGroup,
  Dropdown,
  Form,
  OverlayTrigger,
  Tooltip,
  InputGroup,
  FormControl,
  Popover,
  Overlay,
} from "react-bootstrap";

import { Icon } from "src/components/Shared";
import { ListFilterModel } from "src/models/list-filter/filter";
import { useFocus } from "src/utils";
import { ListFilterOptions } from "src/models/list-filter/filter-options";
import { FormattedMessage, useIntl } from "react-intl";
import { PersistanceLevel } from "src/hooks/ListHook";
import { SavedFilterList } from "./SavedFilterList";
import {
  faBookmark,
  faCaretDown,
  faCaretUp,
  faCheck,
  faFilter,
  faRandom,
  faTimes,
} from "@fortawesome/free-solid-svg-icons";

const maxPageSize = 1000;
interface IListFilterProps {
  onFilterUpdate: (newFilter: ListFilterModel) => void;
  filter: ListFilterModel;
  filterOptions: ListFilterOptions;
  filterDialogOpen?: boolean;
  persistState?: PersistanceLevel;
  openFilterDialog: () => void;
}

const PAGE_SIZE_OPTIONS = ["20", "40", "60", "120", "250", "500", "1000"];

export const ListFilter: React.FC<IListFilterProps> = ({
  onFilterUpdate,
  filter,
  filterOptions,
  filterDialogOpen,
  openFilterDialog,
  persistState,
}) => {
  const [customPageSizeShowing, setCustomPageSizeShowing] = useState(false);
  const [queryRef, setQueryFocus] = useFocus();
  const [queryClearShowing, setQueryClearShowing] = useState(
    !!filter.searchTerm
  );
  const perPageSelect = useRef(null);
  const [perPageInput, perPageFocus] = useFocus();

  const searchCallback = debounce((value: string) => {
    const newFilter = cloneDeep(filter);
    newFilter.searchTerm = value;
    newFilter.currentPage = 1;
    onFilterUpdate(newFilter);
  }, 500);

  const intl = useIntl();

  useEffect(() => {
    Mousetrap.bind("/", (e) => {
      setQueryFocus();
      e.preventDefault();
    });

    Mousetrap.bind("r", () => onReshuffleRandomSort());

    return () => {
      Mousetrap.unbind("/");
      Mousetrap.unbind("r");
    };
  });

  useEffect(() => {
    if (customPageSizeShowing) {
      perPageFocus();
    }
  }, [customPageSizeShowing, perPageFocus]);

  function onChangePageSize(val: string) {
    if (val === "custom") {
      // added timeout since Firefox seems to trigger the rootClose immediately
      // without it
      setTimeout(() => setCustomPageSizeShowing(true), 0);
      return;
    }

    setCustomPageSizeShowing(false);

    let pp = parseInt(val, 10);
    if (Number.isNaN(pp) || pp <= 0) {
      return;
    }

    // don't allow page sizes over 1000
    if (pp > maxPageSize) {
      pp = maxPageSize;
    }

    const newFilter = cloneDeep(filter);
    newFilter.itemsPerPage = pp;
    newFilter.currentPage = 1;
    onFilterUpdate(newFilter);
  }

  function onChangeQuery(event: React.FormEvent<HTMLInputElement>) {
    searchCallback(event.currentTarget.value);
    setQueryClearShowing(!!event.currentTarget.value);
  }

  function onClearQuery() {
    queryRef.current.value = "";
    searchCallback("");
    setQueryFocus();
    setQueryClearShowing(false);
  }

  function onChangeSortDirection() {
    const newFilter = cloneDeep(filter);
    if (filter.sortDirection === SortDirectionEnum.Asc) {
      newFilter.sortDirection = SortDirectionEnum.Desc;
    } else {
      newFilter.sortDirection = SortDirectionEnum.Asc;
    }

    onFilterUpdate(newFilter);
  }

  function onChangeSortBy(eventKey: string | null) {
    const newFilter = cloneDeep(filter);
    newFilter.sortBy = eventKey ?? undefined;
    newFilter.currentPage = 1;
    onFilterUpdate(newFilter);
  }

  function onReshuffleRandomSort() {
    const newFilter = cloneDeep(filter);
    newFilter.currentPage = 1;
    newFilter.randomSeed = -1;
    onFilterUpdate(newFilter);
  }

  function renderSortByOptions() {
    return filterOptions.sortByOptions
      .map((o) => {
        return {
          message: intl.formatMessage({ id: o.messageID }),
          value: o.value,
        };
      })
      .sort((a, b) => a.message.localeCompare(b.message))
      .map((option) => (
        <Dropdown.Item
          onSelect={onChangeSortBy}
          key={option.value}
          className="bg-secondary text-white"
          eventKey={option.value}
        >
          {option.message}
        </Dropdown.Item>
      ));
  }

  const SavedFilterDropdown = React.forwardRef<
    HTMLDivElement,
    HTMLAttributes<HTMLDivElement>
  >(({ style, className }: HTMLAttributes<HTMLDivElement>, ref) => (
    <div ref={ref} style={style} className={className}>
      <SavedFilterList
        filter={filter}
        onSetFilter={(f) => {
          onFilterUpdate(f);
        }}
        persistState={persistState}
      />
    </div>
  ));
  SavedFilterDropdown.displayName = "SavedFilterDropdown";

  function render() {
    const currentSortBy = filterOptions.sortByOptions.find(
      (o) => o.value === filter.sortBy
    );

    const pageSizeOptions = PAGE_SIZE_OPTIONS.map((o) => {
      return {
        label: o,
        value: o,
      };
    });
    const currentPerPage = filter.itemsPerPage.toString();
    if (!pageSizeOptions.find((o) => o.value === currentPerPage)) {
      pageSizeOptions.push({ label: currentPerPage, value: currentPerPage });
      pageSizeOptions.sort(
        (a, b) => parseInt(a.value, 10) - parseInt(b.value, 10)
      );
    }

    pageSizeOptions.push({
      label: `${intl.formatMessage({ id: "custom" })}...`,
      value: "custom",
    });

    return (
      <>
        <div className="mb-2 mr-2 d-flex">
          <div className="flex-grow-1 query-text-field-group">
            <FormControl
              ref={queryRef}
              placeholder={`${intl.formatMessage({ id: "actions.search" })}…`}
              defaultValue={filter.searchTerm}
              onInput={onChangeQuery}
              className="query-text-field bg-secondary text-white border-secondary"
            />
            <Button
              variant="secondary"
              onClick={onClearQuery}
              title={intl.formatMessage({ id: "actions.clear" })}
              className={cx(
                "query-text-field-clear",
                queryClearShowing ? "" : "d-none"
              )}
            >
              <Icon icon={faTimes} />
            </Button>
          </div>
        </div>

        <ButtonGroup className="mr-2 mb-2">
          <Dropdown>
            <OverlayTrigger
              placement="top"
              overlay={
                <Tooltip id="filter-tooltip">
                  <FormattedMessage id="search_filter.saved_filters" />
                </Tooltip>
              }
            >
              <Dropdown.Toggle variant="secondary">
                <Icon icon={faBookmark} />
              </Dropdown.Toggle>
            </OverlayTrigger>
            <Dropdown.Menu
              as={SavedFilterDropdown}
              className="saved-filter-list-menu"
            />
          </Dropdown>
          <OverlayTrigger
            placement="top"
            overlay={
              <Tooltip id="filter-tooltip">
                <FormattedMessage id="search_filter.name" />
              </Tooltip>
            }
          >
            <Button
              variant="secondary"
              onClick={() => openFilterDialog()}
              active={filterDialogOpen}
            >
              <Icon icon={faFilter} />
            </Button>
          </OverlayTrigger>
        </ButtonGroup>

        <Dropdown as={ButtonGroup} className="mr-2 mb-2">
          <InputGroup.Prepend>
            <Dropdown.Toggle variant="secondary">
              {currentSortBy
                ? intl.formatMessage({ id: currentSortBy.messageID })
                : ""}
            </Dropdown.Toggle>
          </InputGroup.Prepend>
          <Dropdown.Menu className="bg-secondary text-white">
            {renderSortByOptions()}
          </Dropdown.Menu>
          <OverlayTrigger
            overlay={
              <Tooltip id="sort-direction-tooltip">
                {filter.sortDirection === SortDirectionEnum.Asc
                  ? intl.formatMessage({ id: "ascending" })
                  : intl.formatMessage({ id: "descending" })}
              </Tooltip>
            }
          >
            <Button variant="secondary" onClick={onChangeSortDirection}>
              <Icon
                icon={
                  filter.sortDirection === SortDirectionEnum.Asc
                    ? faCaretUp
                    : faCaretDown
                }
              />
            </Button>
          </OverlayTrigger>
          {filter.sortBy === "random" && (
            <OverlayTrigger
              overlay={
                <Tooltip id="sort-reshuffle-tooltip">
                  {intl.formatMessage({ id: "actions.reshuffle" })}
                </Tooltip>
              }
            >
              <Button variant="secondary" onClick={onReshuffleRandomSort}>
                <Icon icon={faRandom} />
              </Button>
            </OverlayTrigger>
          )}
        </Dropdown>

        <div className="mb-2">
          <Form.Control
            as="select"
            ref={perPageSelect}
            onChange={(e) => onChangePageSize(e.target.value)}
            value={filter.itemsPerPage.toString()}
            className="btn-secondary"
          >
            {pageSizeOptions.map((s) => (
              <option value={s.value} key={s.value}>
                {s.label}
              </option>
            ))}
          </Form.Control>
          <Overlay
            target={perPageSelect.current}
            show={customPageSizeShowing}
            placement="bottom"
            rootClose
            onHide={() => setCustomPageSizeShowing(false)}
          >
            <Popover id="custom_pagesize_popover">
              <Form inline>
                <InputGroup>
                  <Form.Control
                    type="number"
                    min={1}
                    max={maxPageSize}
                    className="text-input"
                    ref={perPageInput}
                    onKeyPress={(e: React.KeyboardEvent<HTMLInputElement>) => {
                      if (e.key === "Enter") {
                        onChangePageSize(
                          (perPageInput.current as HTMLInputElement)?.value ??
                            ""
                        );
                        e.preventDefault();
                      }
                    }}
                  />
                  <InputGroup.Append>
                    <Button
                      variant="primary"
                      onClick={() =>
                        onChangePageSize(
                          (perPageInput.current as HTMLInputElement)?.value ??
                            ""
                        )
                      }
                    >
                      <Icon icon={faCheck} />
                    </Button>
                  </InputGroup.Append>
                </InputGroup>
              </Form>
            </Popover>
          </Overlay>
        </div>
      </>
    );
  }

  return render();
};
