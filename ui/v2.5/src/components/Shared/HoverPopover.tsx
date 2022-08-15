import React, { useState, useCallback, useEffect, useRef } from "react";
import { Overlay, Popover, OverlayProps } from "react-bootstrap";

interface IHoverPopover {
  enterDelay?: number;
  leaveDelay?: number;
  content: JSX.Element[] | JSX.Element | string;
  className?: string;
  placement?: OverlayProps["placement"];
  onOpen?: () => void;
  onClose?: () => void;
}

export const HoverPopover: React.FC<IHoverPopover> = ({
  enterDelay = 0,
  leaveDelay = 400,
  content,
  children,
  className,
  placement = "top",
  onOpen,
  onClose,
}) => {
  const [show, setShow] = useState(false);
  const triggerRef = useRef<HTMLDivElement>(null);
  const enterTimer = useRef<number>();
  const leaveTimer = useRef<number>();

  const handleMouseEnter = useCallback(() => {
    window.clearTimeout(leaveTimer.current);
    enterTimer.current = window.setTimeout(() => {
      setShow(true);
      onOpen?.();
    }, enterDelay);
  }, [enterDelay, onOpen]);

  const handleMouseLeave = useCallback(() => {
    window.clearTimeout(enterTimer.current);
    leaveTimer.current = window.setTimeout(() => {
      setShow(false);
      onClose?.();
    }, leaveDelay);
  }, [leaveDelay, onClose]);

  useEffect(
    () => () => {
      window.clearTimeout(enterTimer.current);
      window.clearTimeout(leaveTimer.current);
    },
    []
  );

  return (
    <>
      <div
        className={className}
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
        ref={triggerRef}
      >
        {children}
      </div>
      {triggerRef.current && (
        <Overlay show={show} placement={placement} target={triggerRef.current}>
          <Popover
            onMouseEnter={handleMouseEnter}
            onMouseLeave={handleMouseLeave}
            id="popover"
            className="hover-popover-content"
          >
            {content}
          </Popover>
        </Overlay>
      )}
    </>
  );
};
