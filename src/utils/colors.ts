// ANSI color codes for terminal output
const codes = {
  reset: "\x1b[0m",
  bold: "\x1b[1m",
  red: "\x1b[31m",
  green: "\x1b[32m",
  yellow: "\x1b[33m",
  blue: "\x1b[34m",
  cyan: "\x1b[36m",
  gray: "\x1b[90m",
};

const colorize =
  (code: string) =>
  (text: string): string =>
    `${code}${text}${codes.reset}`;

export const colors = {
  red: colorize(codes.red),
  green: colorize(codes.green),
  yellow: colorize(codes.yellow),
  blue: colorize(codes.blue),
  cyan: colorize(codes.cyan),
  gray: colorize(codes.gray),
  bold: colorize(codes.bold),
  cyanBold: colorize(codes.cyan + codes.bold),
};
