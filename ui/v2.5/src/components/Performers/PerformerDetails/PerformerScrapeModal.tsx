import React, { useEffect, useRef, useState } from "react";
import debounce from "lodash-es/debounce";
import { Button, Form } from "react-bootstrap";
import { useIntl } from "react-intl";

import * as GQL from "src/core/generated-graphql";
import { Modal, LoadingIndicator } from "src/components/Shared";
import { useScrapePerformerList } from "src/core/StashService";

const CLASSNAME = "PerformerScrapeModal";
const CLASSNAME_LIST = `${CLASSNAME}-list`;

interface IProps {
  scraper: GQL.Scraper;
  onHide: () => void;
  onSelectPerformer: (
    performer: GQL.ScrapedPerformerDataFragment,
    scraper: GQL.Scraper
  ) => void;
  name?: string;
}
const PerformerScrapeModal: React.FC<IProps> = ({
  scraper,
  name,
  onHide,
  onSelectPerformer,
}) => {
  const intl = useIntl();
  const inputRef = useRef<HTMLInputElement>(null);
  const [query, setQuery] = useState<string>(name ?? "");
  const { data, loading } = useScrapePerformerList(scraper.id, query);

  const performers = data?.scrapeSinglePerformer ?? [];

  const onInputChange = debounce((input: string) => {
    setQuery(input);
  }, 500);

  useEffect(() => inputRef.current?.focus(), []);

  return (
    <Modal
      show
      onHide={onHide}
      header={`Scrape performer from ${scraper.name}`}
      accept={{
        text: intl.formatMessage({ id: "actions.cancel" }),
        onClick: onHide,
        variant: "secondary",
      }}
    >
      <div className={CLASSNAME}>
        <Form.Control
          onChange={(e) => onInputChange(e.currentTarget.value)}
          defaultValue={name ?? ""}
          placeholder="Performer name..."
          className="text-input mb-4"
          ref={inputRef}
        />
        {loading ? (
          <div className="m-4 text-center">
            <LoadingIndicator inline />
          </div>
        ) : (
          <ul className={CLASSNAME_LIST}>
            {performers.map((p) => (
              <li key={p.url}>
                <Button
                  variant="link"
                  onClick={() => onSelectPerformer(p, scraper)}
                >
                  {p.name}
                </Button>
              </li>
            ))}
          </ul>
        )}
      </div>
    </Modal>
  );
};

export default PerformerScrapeModal;
