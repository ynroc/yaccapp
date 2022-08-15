import { IntlShape } from "react-intl";

// Typescript currently does not implement the intl Unit interface
type Unit =
  | "byte"
  | "kilobyte"
  | "megabyte"
  | "gigabyte"
  | "terabyte"
  | "petabyte";
const Units: Unit[] = [
  "byte",
  "kilobyte",
  "megabyte",
  "gigabyte",
  "terabyte",
  "petabyte",
];
const shortUnits = ["B", "KB", "MB", "GB", "TB", "PB"];

const fileSize = (bytes: number = 0) => {
  if (Number.isNaN(parseFloat(String(bytes))) || !Number.isFinite(bytes))
    return { size: 0, unit: Units[0] };

  let unit = 0;
  let count = bytes;
  while (count >= 1024 && unit + 1 < Units.length) {
    count /= 1024;
    unit++;
  }

  return {
    size: count,
    unit: Units[unit],
  };
};

class DurationUnit {
  static readonly SECOND: DurationUnit = new DurationUnit(
    "second",
    "seconds",
    "s",
    1
  );
  static readonly MINUTE: DurationUnit = new DurationUnit(
    "minute",
    "minutes",
    "m",
    60
  );
  static readonly HOUR: DurationUnit = new DurationUnit(
    "hour",
    "hours",
    "h",
    DurationUnit.MINUTE.secs * 60
  );
  static readonly DAY: DurationUnit = new DurationUnit(
    "day",
    "days",
    "D",
    DurationUnit.HOUR.secs * 24
  );
  static readonly WEEK: DurationUnit = new DurationUnit(
    "week",
    "weeks",
    "W",
    DurationUnit.DAY.secs * 7
  );
  static readonly MONTH: DurationUnit = new DurationUnit(
    "month",
    "months",
    "M",
    DurationUnit.DAY.secs * 30
  );
  static readonly YEAR: DurationUnit = new DurationUnit(
    "year",
    "years",
    "Y",
    DurationUnit.DAY.secs * 365
  );

  static readonly DURATIONS: DurationUnit[] = [
    DurationUnit.SECOND,
    DurationUnit.MINUTE,
    DurationUnit.HOUR,
    DurationUnit.DAY,
    DurationUnit.WEEK,
    DurationUnit.MONTH,
    DurationUnit.YEAR,
  ];

  private constructor(
    private readonly singular: string,
    private readonly plural: string,
    private readonly shortString: string,
    public secs: number
  ) {}

  toString() {
    return this.shortString;
  }
}

class DurationCount {
  public constructor(
    public readonly count: number,
    public readonly duration: DurationUnit
  ) {}

  toString() {
    return this.count.toString() + this.duration.toString();
  }
}

const secondsAsTime = (seconds: number = 0): DurationCount[] => {
  if (Number.isNaN(parseFloat(String(seconds))) || !Number.isFinite(seconds))
    return [new DurationCount(0, DurationUnit.DURATIONS[0])];

  const result = [];
  let remainingSeconds = seconds;
  // Run down the possible durations and pull them out
  for (let i = DurationUnit.DURATIONS.length - 1; i >= 0; i--) {
    const q = Math.floor(remainingSeconds / DurationUnit.DURATIONS[i].secs);
    if (q !== 0) {
      remainingSeconds %= DurationUnit.DURATIONS[i].secs;
      result.push(new DurationCount(q, DurationUnit.DURATIONS[i]));
    }
  }
  return result;
};

const timeAsString = (time: DurationCount[]): string => {
  return time.join(" ");
};

const secondsAsTimeString = (
  seconds: number = 0,
  maxUnitCount: number = 2
): string => {
  const timeArray = secondsAsTime(seconds).slice(0, maxUnitCount);
  return timeAsString(timeArray);
};

const formatFileSizeUnit = (u: Unit) => {
  const i = Units.indexOf(u);
  return shortUnits[i];
};

// returns the number of fractional digits to use when displaying file sizes
// returns 0 for MB and under, 1 for GB and over.
const fileSizeFractionalDigits = (unit: Unit) => {
  if (Units.indexOf(unit) >= 3) {
    return 1;
  }

  return 0;
};

const secondsToTimestamp = (seconds: number) => {
  let ret = new Date(seconds * 1000).toISOString().substr(11, 8);

  if (ret.startsWith("00")) {
    // strip hours if under one hour
    ret = ret.substr(3);
  }
  if (ret.startsWith("0")) {
    // for duration under a minute, leave one leading zero
    ret = ret.substr(1);
  }
  return ret;
};

const fileNameFromPath = (path: string) => {
  if (!!path === false) return "No File Name";
  return path.replace(/^.*[\\/]/, "");
};

const stringToDate = (dateString: string) => {
  if (!dateString) return null;

  const parts = dateString.split("-");
  // Invalid date string
  if (parts.length !== 3) return null;

  const year = Number(parts[0]);
  const monthIndex = Math.max(0, Number(parts[1]) - 1);
  const day = Number(parts[2]);

  return new Date(year, monthIndex, day, 0, 0, 0, 0);
};

const getAge = (dateString?: string | null, fromDateString?: string | null) => {
  if (!dateString) return 0;

  const birthdate = stringToDate(dateString);
  const fromDate = fromDateString ? stringToDate(fromDateString) : new Date();

  if (!birthdate || !fromDate) return 0;

  let age = fromDate.getFullYear() - birthdate.getFullYear();
  if (
    birthdate.getMonth() > fromDate.getMonth() ||
    (birthdate.getMonth() >= fromDate.getMonth() &&
      birthdate.getDate() > fromDate.getDate())
  ) {
    age -= 1;
  }

  return age;
};

const bitRate = (bitrate: number) => {
  const megabits = bitrate / 1000000;
  return `${megabits.toFixed(2)} megabits per second`;
};

const resolution = (width: number, height: number) => {
  const number = width > height ? height : width;
  if (number >= 4320) {
    return "8K";
  }
  if (number >= 3384) {
    return "6K";
  }
  if (number >= 2880) {
    return "5K";
  }
  if (number >= 2160) {
    return "4K";
  }
  if (number >= 1920) {
    return "1920p";
  }
  if (number >= 1440) {
    return "1440p";
  }
  if (number >= 1080) {
    return "1080p";
  }
  if (number >= 720) {
    return "720p";
  }
  if (number >= 540) {
    return "540p";
  }
  if (number >= 480) {
    return "480p";
  }
  if (number >= 360) {
    return "360p";
  }
  if (number >= 240) {
    return "240p";
  }
  if (number >= 144) {
    return "144p";
  }
};

const twitterURL = new URL("https://www.twitter.com");
const instagramURL = new URL("https://www.instagram.com");

const sanitiseURL = (url?: string, siteURL?: URL) => {
  if (!url) {
    return url;
  }

  if (url.startsWith("http://") || url.startsWith("https://")) {
    // just return the entire URL
    return url;
  }

  if (siteURL) {
    // if url starts with the site host, then prepend the protocol
    if (url.startsWith(siteURL.host)) {
      return `${siteURL.protocol}//${url}`;
    }

    // otherwise, construct the url from the protocol, host and passed url
    return `${siteURL.protocol}//${siteURL.host}/${url}`;
  }

  // just prepend the protocol - assume https
  return `https://${url}`;
};

const formatDate = (intl: IntlShape, date?: string, utc = true) => {
  if (!date) {
    return "";
  }

  return intl.formatDate(date, {
    format: "long",
    timeZone: utc ? "utc" : undefined,
  });
};

const formatDateTime = (intl: IntlShape, dateTime?: string, utc = false) =>
  `${formatDate(intl, dateTime, utc)} ${intl.formatTime(dateTime, {
    timeZone: utc ? "utc" : undefined,
  })}`;

const capitalize = (val: string) =>
  val
    .replace(/^[-_]*(.)/, (_, c) => c.toUpperCase())
    .replace(/[-_]+(.)/g, (_, c) => ` ${c.toUpperCase()}`);

const TextUtils = {
  fileSize,
  formatFileSizeUnit,
  fileSizeFractionalDigits,
  secondsToTimestamp,
  fileNameFromPath,
  stringToDate,
  age: getAge,
  bitRate,
  resolution,
  sanitiseURL,
  twitterURL,
  instagramURL,
  formatDate,
  formatDateTime,
  capitalize,
  secondsAsTimeString,
};

export default TextUtils;
