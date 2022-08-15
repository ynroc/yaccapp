import React, { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import cx from "classnames";
import * as GQL from "src/core/generated-graphql";
import TextUtils from "src/utils/text";
import { Button, Form, Spinner } from "react-bootstrap";
import Icon from "src/components/Shared/Icon";
import { useIntl } from "react-intl";
import {
  faChevronDown,
  faChevronUp,
  faRandom,
  faStepBackward,
  faStepForward,
} from "@fortawesome/free-solid-svg-icons";

export interface IPlaylistViewer {
  scenes?: GQL.SlimSceneDataFragment[];
  currentID?: string;
  start?: number;
  continue?: boolean;
  hasMoreScenes: boolean;
  setContinue: (v: boolean) => void;
  onSceneClicked: (id: string) => void;
  onNext: () => void;
  onPrevious: () => void;
  onRandom: () => void;
  onMoreScenes: () => void;
  onLessScenes: () => void;
}

export const QueueViewer: React.FC<IPlaylistViewer> = ({
  scenes,
  currentID,
  start,
  continue: continuePlaylist = false,
  hasMoreScenes,
  setContinue,
  onNext,
  onPrevious,
  onRandom,
  onSceneClicked,
  onMoreScenes,
  onLessScenes,
}) => {
  const intl = useIntl();
  const [lessLoading, setLessLoading] = useState(false);
  const [moreLoading, setMoreLoading] = useState(false);

  const currentIndex = scenes?.findIndex((s) => s.id === currentID);

  useEffect(() => {
    setLessLoading(false);
    setMoreLoading(false);
  }, [scenes]);

  function isCurrentScene(scene: GQL.SlimSceneDataFragment) {
    return scene.id === currentID;
  }

  function handleSceneClick(
    event: React.MouseEvent<HTMLAnchorElement, MouseEvent>,
    id: string
  ) {
    onSceneClicked(id);
    event.preventDefault();
  }

  function lessClicked() {
    setLessLoading(true);
    onLessScenes();
  }

  function moreClicked() {
    setMoreLoading(true);
    onMoreScenes();
  }

  function renderPlaylistEntry(scene: GQL.SlimSceneDataFragment) {
    return (
      <li
        className={cx("my-2", { current: isCurrentScene(scene) })}
        key={scene.id}
      >
        <Link
          to={`/scenes/${scene.id}`}
          onClick={(e) => handleSceneClick(e, scene.id)}
        >
          <div className="ml-1 d-flex align-items-center">
            <div className="thumbnail-container">
              <img alt={scene.title ?? ""} src={scene.paths.screenshot ?? ""} />
            </div>
            <div>
              <span className="align-middle text-break">
                {scene.title ?? TextUtils.fileNameFromPath(scene.path)}
              </span>
            </div>
          </div>
        </Link>
      </li>
    );
  }

  return (
    <div id="queue-viewer">
      <div className="queue-controls">
        <div>
          <Form.Check
            checked={continuePlaylist}
            label={intl.formatMessage({ id: "actions.continue" })}
            onChange={() => {
              setContinue(!continuePlaylist);
            }}
          />
        </div>
        <div>
          {(currentIndex ?? 0) > 0 ? (
            <Button
              className="minimal"
              variant="secondary"
              onClick={() => onPrevious()}
            >
              <Icon icon={faStepBackward} />
            </Button>
          ) : (
            ""
          )}
          {(currentIndex ?? 0) < (scenes ?? []).length - 1 ? (
            <Button
              className="minimal"
              variant="secondary"
              onClick={() => onNext()}
            >
              <Icon icon={faStepForward} />
            </Button>
          ) : (
            ""
          )}
          <Button
            className="minimal"
            variant="secondary"
            onClick={() => onRandom()}
          >
            <Icon icon={faRandom} />
          </Button>
        </div>
      </div>
      <div id="queue-content">
        {(start ?? 0) > 1 ? (
          <div className="d-flex justify-content-center">
            <Button onClick={() => lessClicked()} disabled={lessLoading}>
              {!lessLoading ? (
                <Icon icon={faChevronUp} />
              ) : (
                <Spinner animation="border" role="status" />
              )}
            </Button>
          </div>
        ) : undefined}
        <ol start={start}>{(scenes ?? []).map(renderPlaylistEntry)}</ol>
        {hasMoreScenes ? (
          <div className="d-flex justify-content-center">
            <Button onClick={() => moreClicked()} disabled={moreLoading}>
              {!moreLoading ? (
                <Icon icon={faChevronDown} />
              ) : (
                <Spinner animation="border" role="status" />
              )}
            </Button>
          </div>
        ) : undefined}
      </div>
    </div>
  );
};

export default QueueViewer;
