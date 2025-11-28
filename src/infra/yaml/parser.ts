import yaml from 'js-yaml'

export interface YamlDumpOptions {
  indent?: number
  lineWidth?: number
  sortKeys?: boolean
}

const defaultDumpOptions: YamlDumpOptions = {
  indent: 2,
  lineWidth: -1,
  sortKeys: false,
}

export function parse<T = unknown>(content: string): T {
  return yaml.load(content) as T
}

export function stringify(data: unknown, options?: YamlDumpOptions): string {
  const opts = { ...defaultDumpOptions, ...options }
  return yaml.dump(data, {
    indent: opts.indent,
    lineWidth: opts.lineWidth,
    noRefs: true,
    sortKeys: opts.sortKeys,
  })
}
