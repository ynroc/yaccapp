import React, { useState, useEffect } from "react";
import { Button } from "react-bootstrap";
import { FormattedMessage } from "react-intl";
import Mousetrap from "mousetrap";
import * as GQL from "src/core/generated-graphql";
import { WallPanel } from "src/components/Wall/WallPanel";
import { PrimaryTags } from "./PrimaryTags";
import { SceneMarkerForm } from "./SceneMarkerForm";

interface ISceneMarkersPanelProps {
  sceneId: string;
  isVisible: boolean;
  onClickMarker: (marker: GQL.SceneMarkerDataFragment) => void;
}

export const SceneMarkersPanel: React.FC<ISceneMarkersPanelProps> = (
  props: ISceneMarkersPanelProps
) => {
  const { data, loading } = GQL.useFindSceneMarkerTagsQuery({
    variables: {
      id: props.sceneId,
    },
  });
  const [isEditorOpen, setIsEditorOpen] = useState<boolean>(false);
  const [
    editingMarker,
    setEditingMarker,
  ] = useState<GQL.SceneMarkerDataFragment>();

  // set up hotkeys
  useEffect(() => {
    if (props.isVisible) {
      Mousetrap.bind("n", () => onOpenEditor());

      return () => {
        Mousetrap.unbind("n");
      };
    }
  });

  if (loading) return null;

  function onOpenEditor(marker?: GQL.SceneMarkerDataFragment) {
    setIsEditorOpen(true);
    setEditingMarker(marker ?? undefined);
  }

  function onClickMarker(marker: GQL.SceneMarkerDataFragment) {
    props.onClickMarker(marker);
  }

  const closeEditor = () => {
    setEditingMarker(undefined);
    setIsEditorOpen(false);
  };

  if (isEditorOpen)
    return (
      <SceneMarkerForm
        sceneID={props.sceneId}
        editingMarker={editingMarker}
        onClose={closeEditor}
      />
    );

  const sceneMarkers = (
    data?.sceneMarkerTags.map((tag) => tag.scene_markers) ?? []
  ).reduce((prev, current) => [...prev, ...current], []);

  return (
    <div className="scene-markers-panel">
      <Button onClick={() => onOpenEditor()}>
        <FormattedMessage id="actions.create_marker" />
      </Button>
      <div className="container">
        <PrimaryTags
          sceneMarkers={sceneMarkers}
          onClickMarker={onClickMarker}
          onEdit={onOpenEditor}
        />
      </div>
      <WallPanel
        sceneMarkers={sceneMarkers}
        clickHandler={(marker) => {
          window.scrollTo(0, 0);
          onClickMarker(marker as GQL.SceneMarkerDataFragment);
        }}
      />
    </div>
  );
};

export default SceneMarkersPanel;
