import chalk from "chalk";
import type { DiffItem } from "../types.js";

export function printDiff(diffs: DiffItem[]): void {
  if (diffs.length === 0) {
    console.log(chalk.green("No changes detected."));
    return;
  }

  console.log(chalk.bold("\nPlanned changes:\n"));

  const grouped = groupBy(diffs, (d) => d.type);

  for (const [type, items] of Object.entries(grouped)) {
    console.log(chalk.cyan.bold(`[${type}]`));

    for (const item of items) {
      const icon = getActionIcon(item.action);
      const color = getActionColor(item.action);
      console.log(`  ${icon} ${color(item.details)}`);
      if (item.apiCall) {
        console.log(chalk.gray(`    API: ${item.apiCall}`));
      }
    }
    console.log();
  }
}

function getActionIcon(action: DiffItem["action"]): string {
  switch (action) {
    case "create":
      return chalk.green("+");
    case "update":
      return chalk.yellow("~");
    case "delete":
      return chalk.red("-");
    case "check":
      return chalk.blue("?");
  }
}

function getActionColor(action: DiffItem["action"]): (text: string) => string {
  switch (action) {
    case "create":
      return chalk.green;
    case "update":
      return chalk.yellow;
    case "delete":
      return chalk.red;
    case "check":
      return chalk.blue;
  }
}

function groupBy<T>(arr: T[], fn: (item: T) => string): Record<string, T[]> {
  return arr.reduce(
    (acc, item) => {
      const key = fn(item);
      if (!acc[key]) {
        acc[key] = [];
      }
      acc[key].push(item);
      return acc;
    },
    {} as Record<string, T[]>
  );
}
