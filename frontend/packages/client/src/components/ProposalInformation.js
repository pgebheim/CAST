import React, { useEffect, useRef, useState } from "react";
import Blockies from "react-blockies";
import { LinkOut } from "./Svg";
import { parseDateFromServer } from "../utils";
import { useVotingResults, useWindowDimensions } from "../hooks";
import Tooltip from "./Tooltip";

const InfoBlock = ({ title, content }) => {
  return (
    <div className="columns is-mobile p-0 m-0 mb-5 small-text">
      <div
        className="column p-0 is-flex is-align-items-center flex-1 has-text-grey"
        style={{ maxWidth: "40%" }}
      >
        {title}
      </div>
      <div
        className="column p-0 is-flex flex-1 is-align-items-center"
        style={{ height: "1.5rem" }}
      >
        {content}
      </div>
    </div>
  );
};

const Results = ({ voteResults }) => {
  const showViewMore = false;
  const options = Object.keys(voteResults);

  const totalVotes = options.reduce(
    (previousValue, currentValue) => previousValue + voteResults[currentValue],
    0
  );

  return (
    <>
      {options.map((option, index) => {
        const percentage =
          totalVotes === 0 || voteResults[option] === 0
            ? 0
            : ((100 * voteResults[option]) / totalVotes).toFixed(2);
        return (
          <div key={`result-item-${index}`} style={{ marginBottom: "2.5rem" }}>
            <div className="is-flex mb-2">
              <div className="flex-1 small-text has-text-grey">{option}</div>
              <div className="is-flex flex-1 is-justify-content-flex-end small-text has-text-grey">
                {`${percentage}%`}
              </div>
            </div>
            <div
              style={{ height: 8, background: "#DCDCDC" }}
              className="has-background-grey-light rounded-lg"
            >
              <div
                className="rounded-lg"
                style={{
                  width: `${percentage}%`,
                  height: "100%",
                  background: "#747474",
                }}
              />
            </div>
          </div>
        );
      })}
      {showViewMore && (
        <div className="is-flex is-justify-content-start is-align-items-center">
          <button className="button is-white has-background-white-ter p-0">
            View more
          </button>
        </div>
      )}
    </>
  );
};

const WrapperSpacingTop = ({ children }) => {
  return (
    <>
      {/* Desktop and bigger screens */}
      <div className="is-hidden-tablet-only is-hidden-mobile px-6 pb-0 pt-6">
        {children}
      </div>
      {/* Tablet view */}
      <div className="is-hidden-desktop is-hidden-mobile px-5 pb-0 pt-5">
        {children}
      </div>
      {/* Mobile only  */}
      <div className="is-hidden-tablet px-1 pb-0 pt-1">{children}</div>
    </>
  );
};

const WrapperSpacingBottom = ({ children }) => {
  return (
    <>
      <div className="is-hidden-desktop is-hidden-mobile px-5 pt-3 pb-4">
        {children}
      </div>
      <div className="is-hidden-tablet-only is-hidden-mobile px-6 pt-2 pb-6">
        {children}
      </div>
      <div className="is-hidden-tablet px-1 pt-1 pb-1">{children}</div>
    </>
  );
};

const ProposalInformation = ({
  creatorAddr = "",
  strategies = [],
  isCoreCreator = false,
  ipfs = "",
  ipfsUrl = "",
  startTime = "",
  endTime = "",
  proposalId = "",
  openStrategyModal = () => {},
  className = "",
}) => {
  const dateFormatConf = {
    day: "numeric",
    hour: "numeric",
    minute: "numeric",
    month: "short",
    year: "numeric",
    hour12: true,
  };
  // stores navbar height calculated after component is mounted
  const [navbarHeight, setNavbarHeight] = useState(0);

  useEffect(() => {
    setNavbarHeight(document.querySelector("header").offsetHeight);
  }, []);

  const { height: windowHeight, width: windowWidth } = useWindowDimensions();

  // stores when user scrolls
  const [scroll, setScroll] = useState(0);

  // stores style to apply when info bar needs to be fixed
  const [fixedStyle, setFixedStyle] = useState(null);
  // ref of the panel info component
  const ref = useRef(null);
  // ref of the parent component
  const parentRef = useRef(null);
  // used to store return point
  const topRef = useRef({ pointStatic: null });

  const { loading: loadingVotingResults, data: votingResults } =
    useVotingResults(proposalId);

  // this effect watches for user scroll to make info panel fixed to navbar
  useEffect(() => {
    if (ref?.current && parentRef?.current) {
      const { top, height: infoPanelHeightSize } =
        ref?.current.getBoundingClientRect() || {};

      const { width } = parentRef?.current.getBoundingClientRect() || {};

      const winScroll =
        document.body.scrollTop || document.documentElement.scrollTop;

      // if window size is bigger that navbar size + info panel then apply fixed
      if (windowHeight > navbarHeight + infoPanelHeightSize) {
        // user scrolled down and panel is next to navbar
        // adding 4px so the nav bar sticks more smoothly
        if (top < navbarHeight + 4 && !fixedStyle) {
          // save reference where the element needs to go back
          if (!topRef.current.returnTop) {
            topRef.current.pointStatic = winScroll;
          }
          setFixedStyle({
            className: " is-panel-fixed",
            // use width of parent component
            style: {
              width,
              top: navbarHeight,
            },
          });
        }
      }

      if (
        fixedStyle &&
        topRef.current.pointStatic &&
        topRef.current.pointStatic > winScroll
      ) {
        topRef.current.pointStatic = null;
        setFixedStyle(null);
      }
    }
  }, [scroll, fixedStyle, windowHeight, navbarHeight]);

  // this effect watches for window width resizes to change info panel width
  useEffect(() => {
    if (parentRef?.current) {
      const { width } = parentRef?.current.getBoundingClientRect() || {};

      // window was resized then change width of the component if fixed style was applied
      if (fixedStyle && fixedStyle.windowWidth !== windowWidth) {
        setFixedStyle((state) => ({
          ...state,
          windowWidth,
          style: { ...state.style, width },
        }));
      }
    }
  }, [fixedStyle, windowWidth, navbarHeight]);

  const handleScroll = () => {
    setScroll(document.body.scrollTop || document.documentElement.scrollTop);
  };

  // this effect watches for window scrolling
  useEffect(() => {
    document.addEventListener("scroll", handleScroll);
    return () => document.removeEventListener("scroll", handleScroll);
  }, []);

  return (
    <div ref={parentRef}>
      <div
        className={`${className} rounded-sm${fixedStyle?.className || ""}`}
        ref={ref}
        style={fixedStyle?.style || {}}
      >
        <WrapperSpacingTop>
          <p className="mb-5">Information</p>
          <InfoBlock
            title={"Author"}
            content={
              <div className="columns is-mobile m-0">
                <div className="column is-narrow is-flex is-align-items-center p-0">
                  <Blockies
                    seed={creatorAddr}
                    size={6}
                    scale={4}
                    className="blockies"
                  />
                </div>
                <div className="column px-2 py-0 is-flex flex-1 is-align-items-center">
                  {creatorAddr}
                </div>
                {isCoreCreator && (
                  <div
                    className="column p-0 is-flex is-align-items-center is-justify-content-center-tablet subtitle is-size-7"
                    style={{ fontFamily: "Roboto Mono" }}
                  >
                    Core
                  </div>
                )}
              </div>
            }
          />
          {ipfs && (
            <InfoBlock
              title={"IPFS"}
              content={
                <a
                  href={ipfsUrl}
                  rel="noopener noreferrer"
                  target="_blank"
                  className="button is-text p-0 small-text"
                  style={{ height: "2rem !important" }}
                >
                  <Tooltip
                    classNames="is-flex is-flex-grow-1 is-align-items-center"
                    position="top"
                    text="Open Ipfs link"
                  >
                    <p className="mr-2">{`${ipfs.substring(0, 8)}`}</p>
                    <LinkOut width="12" height="12" />
                  </Tooltip>
                </a>
              }
            />
          )}
          <InfoBlock
            title={"Start date"}
            content={parseDateFromServer(startTime).date.toLocaleString(
              undefined,
              dateFormatConf
            )}
          />
          <InfoBlock
            title={"End date"}
            content={parseDateFromServer(endTime).date.toLocaleString(
              undefined,
              dateFormatConf
            )}
          />
        </WrapperSpacingTop>
        <hr />
        <WrapperSpacingBottom>
          <p className="mb-5">Current Results</p>
          {!loadingVotingResults && (
            <Results voteResults={votingResults?.results || []} />
          )}
        </WrapperSpacingBottom>
      </div>
    </div>
  );
};

export default ProposalInformation;
