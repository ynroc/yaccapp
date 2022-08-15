import React from "react";
import {
  Form,
  Col,
  Row,
  InputGroup,
  Button,
  FormControl,
  Badge,
} from "react-bootstrap";
import { CollapseButton } from "src/components/Shared/CollapseButton";
import Icon from "src/components/Shared/Icon";
import Modal from "src/components/Shared/Modal";
import isEqual from "lodash-es/isEqual";
import clone from "lodash-es/clone";
import { FormattedMessage, useIntl } from "react-intl";
import {
  faCheck,
  faPencilAlt,
  faPlus,
  faTimes,
} from "@fortawesome/free-solid-svg-icons";

export class ScrapeResult<T> {
  public newValue?: T;
  public originalValue?: T;
  public scraped: boolean = false;
  public useNewValue: boolean = false;

  public constructor(originalValue?: T | null, newValue?: T | null) {
    this.originalValue = originalValue ?? undefined;
    this.newValue = newValue ?? undefined;

    const valuesEqual = isEqual(originalValue, newValue);
    this.useNewValue = !!this.newValue && !valuesEqual;
    this.scraped = this.useNewValue;
  }

  public setOriginalValue(value?: T) {
    this.originalValue = value;
    this.newValue = value;
  }

  public cloneWithValue(value?: T) {
    const ret = clone(this);

    ret.newValue = value;
    ret.useNewValue = !isEqual(ret.newValue, ret.originalValue);

    // #2691 - if we're setting the value, assume it should be treated as
    // scraped
    ret.scraped = true;

    return ret;
  }

  public getNewValue() {
    if (this.useNewValue) {
      return this.newValue;
    }
  }
}

interface IHasName {
  name: string | undefined;
}

interface IScrapedFieldProps<T> {
  result: ScrapeResult<T>;
}

interface IScrapedRowProps<T, V extends IHasName>
  extends IScrapedFieldProps<T> {
  title: string;
  renderOriginalField: (result: ScrapeResult<T>) => JSX.Element | undefined;
  renderNewField: (result: ScrapeResult<T>) => JSX.Element | undefined;
  onChange: (value: ScrapeResult<T>) => void;
  newValues?: V[];
  onCreateNew?: (index: number) => void;
}

function renderButtonIcon(selected: boolean) {
  const className = selected ? "text-success" : "text-muted";

  return (
    <Icon
      className={`fa-fw ${className}`}
      icon={selected ? faCheck : faTimes}
    />
  );
}

export const ScrapeDialogRow = <T, V extends IHasName>(
  props: IScrapedRowProps<T, V>
) => {
  function handleSelectClick(isNew: boolean) {
    const ret = clone(props.result);
    ret.useNewValue = isNew;
    props.onChange(ret);
  }

  function hasNewValues() {
    return props.newValues && props.newValues.length > 0 && props.onCreateNew;
  }

  if (!props.result.scraped && !hasNewValues()) {
    return <></>;
  }

  function renderNewValues() {
    if (!hasNewValues()) {
      return;
    }

    const ret = (
      <>
        {props.newValues!.map((t, i) => (
          <Badge
            className="tag-item"
            variant="secondary"
            key={t.name}
            onClick={() => props.onCreateNew!(i)}
          >
            {t.name}
            <Button className="minimal ml-2">
              <Icon className="fa-fw" icon={faPlus} />
            </Button>
          </Badge>
        ))}
      </>
    );

    const minCollapseLength = 10;

    if (props.newValues!.length >= minCollapseLength) {
      return (
        <CollapseButton text={`Missing (${props.newValues!.length})`}>
          {ret}
        </CollapseButton>
      );
    }

    return ret;
  }

  return (
    <Row className="px-3 pt-3">
      <Form.Label column lg="3">
        {props.title}
      </Form.Label>

      <Col lg="9">
        <Row>
          <Col xs="6">
            <InputGroup>
              <InputGroup.Prepend className="bg-secondary text-white border-secondary">
                <Button
                  variant="secondary"
                  onClick={() => handleSelectClick(false)}
                >
                  {renderButtonIcon(!props.result.useNewValue)}
                </Button>
              </InputGroup.Prepend>
              {props.renderOriginalField(props.result)}
            </InputGroup>
          </Col>
          <Col xs="6">
            <InputGroup>
              <InputGroup.Prepend>
                <Button
                  variant="secondary"
                  onClick={() => handleSelectClick(true)}
                >
                  {renderButtonIcon(props.result.useNewValue)}
                </Button>
              </InputGroup.Prepend>
              {props.renderNewField(props.result)}
            </InputGroup>
            {renderNewValues()}
          </Col>
        </Row>
      </Col>
    </Row>
  );
};

interface IScrapedInputGroupProps {
  isNew?: boolean;
  placeholder?: string;
  locked?: boolean;
  result: ScrapeResult<string>;
  onChange?: (value: string) => void;
}

const ScrapedInputGroup: React.FC<IScrapedInputGroupProps> = (props) => {
  return (
    <FormControl
      placeholder={props.placeholder}
      value={props.isNew ? props.result.newValue : props.result.originalValue}
      readOnly={!props.isNew || props.locked}
      onChange={(e) => {
        if (props.isNew && props.onChange) {
          props.onChange(e.target.value);
        }
      }}
      className="bg-secondary text-white border-secondary"
    />
  );
};

interface IScrapedInputGroupRowProps {
  title: string;
  placeholder?: string;
  result: ScrapeResult<string>;
  locked?: boolean;
  onChange: (value: ScrapeResult<string>) => void;
}

export const ScrapedInputGroupRow: React.FC<IScrapedInputGroupRowProps> = (
  props
) => {
  return (
    <ScrapeDialogRow
      title={props.title}
      result={props.result}
      renderOriginalField={() => (
        <ScrapedInputGroup
          placeholder={props.placeholder || props.title}
          result={props.result}
        />
      )}
      renderNewField={() => (
        <ScrapedInputGroup
          placeholder={props.placeholder || props.title}
          result={props.result}
          isNew
          locked={props.locked}
          onChange={(value) =>
            props.onChange(props.result.cloneWithValue(value))
          }
        />
      )}
      onChange={props.onChange}
    />
  );
};

const ScrapedTextArea: React.FC<IScrapedInputGroupProps> = (props) => {
  return (
    <FormControl
      as="textarea"
      placeholder={props.placeholder}
      value={props.isNew ? props.result.newValue : props.result.originalValue}
      readOnly={!props.isNew}
      onChange={(e) => {
        if (props.isNew && props.onChange) {
          props.onChange(e.target.value);
        }
      }}
      className="bg-secondary text-white border-secondary scene-description"
    />
  );
};

export const ScrapedTextAreaRow: React.FC<IScrapedInputGroupRowProps> = (
  props
) => {
  return (
    <ScrapeDialogRow
      title={props.title}
      result={props.result}
      renderOriginalField={() => (
        <ScrapedTextArea
          placeholder={props.placeholder || props.title}
          result={props.result}
        />
      )}
      renderNewField={() => (
        <ScrapedTextArea
          placeholder={props.placeholder || props.title}
          result={props.result}
          isNew
          onChange={(value) =>
            props.onChange(props.result.cloneWithValue(value))
          }
        />
      )}
      onChange={props.onChange}
    />
  );
};

interface IScrapedImageProps {
  isNew?: boolean;
  className?: string;
  placeholder?: string;
  result: ScrapeResult<string>;
}

const ScrapedImage: React.FC<IScrapedImageProps> = (props) => {
  const value = props.isNew
    ? props.result.newValue
    : props.result.originalValue;

  if (!value) {
    return <></>;
  }

  return (
    <img className={props.className} src={value} alt={props.placeholder} />
  );
};

interface IScrapedImageRowProps {
  title: string;
  className?: string;
  result: ScrapeResult<string>;
  onChange: (value: ScrapeResult<string>) => void;
}

export const ScrapedImageRow: React.FC<IScrapedImageRowProps> = (props) => {
  return (
    <ScrapeDialogRow
      title={props.title}
      result={props.result}
      renderOriginalField={() => (
        <ScrapedImage
          result={props.result}
          className={props.className}
          placeholder={props.title}
        />
      )}
      renderNewField={() => (
        <ScrapedImage
          result={props.result}
          className={props.className}
          placeholder={props.title}
          isNew
        />
      )}
      onChange={props.onChange}
    />
  );
};

interface IScrapeDialogProps {
  title: string;
  renderScrapeRows: () => JSX.Element;
  onClose: (apply?: boolean) => void;
}

export const ScrapeDialog: React.FC<IScrapeDialogProps> = (
  props: IScrapeDialogProps
) => {
  const intl = useIntl();
  return (
    <Modal
      show
      icon={faPencilAlt}
      header={props.title}
      accept={{
        onClick: () => {
          props.onClose(true);
        },
        text: intl.formatMessage({ id: "actions.apply" }),
      }}
      cancel={{
        onClick: () => props.onClose(),
        text: intl.formatMessage({ id: "actions.cancel" }),
        variant: "secondary",
      }}
      modalProps={{ size: "lg", dialogClassName: "scrape-dialog" }}
    >
      <div className="dialog-container">
        <Form>
          <Row className="px-3 pt-3">
            <Col lg={{ span: 9, offset: 3 }}>
              <Row>
                <Form.Label column xs="6">
                  <FormattedMessage id="dialogs.scrape_results_existing" />
                </Form.Label>
                <Form.Label column xs="6">
                  <FormattedMessage id="dialogs.scrape_results_scraped" />
                </Form.Label>
              </Row>
            </Col>
          </Row>

          {props.renderScrapeRows()}
        </Form>
      </div>
    </Modal>
  );
};
