import React, { lazy, Suspense, useCallback, useState } from "react";
import { ILightboxImage } from "./types";

const LightboxComponent = lazy(() => import("./Lightbox"));

export interface IState {
  images: ILightboxImage[];
  isVisible: boolean;
  isLoading: boolean;
  showNavigation: boolean;
  initialIndex?: number;
  pageCallback?: (direction: number) => void;
  pageHeader?: string;
  slideshowEnabled: boolean;
  onClose?: () => void;
}
interface IContext {
  setLightboxState: (state: Partial<IState>) => void;
}

export const LightboxContext = React.createContext<IContext>({
  setLightboxState: () => {},
});
const Lightbox: React.FC = ({ children }) => {
  const [lightboxState, setLightboxState] = useState<IState>({
    images: [],
    isVisible: false,
    isLoading: false,
    showNavigation: true,
    slideshowEnabled: false,
  });

  const setPartialState = useCallback(
    (state: Partial<IState>) => {
      setLightboxState((currentState: IState) => ({
        ...currentState,
        ...state,
      }));
    },
    [setLightboxState]
  );

  const onHide = () => {
    setLightboxState({ ...lightboxState, isVisible: false });
    if (lightboxState.onClose) {
      lightboxState.onClose();
    }
  };

  return (
    <LightboxContext.Provider value={{ setLightboxState: setPartialState }}>
      {children}
      <Suspense fallback={<></>}>
        {lightboxState.isVisible && (
          <LightboxComponent {...lightboxState} hide={onHide} />
        )}
      </Suspense>
    </LightboxContext.Provider>
  );
};

export default Lightbox;
