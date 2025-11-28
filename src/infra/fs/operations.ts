import {
  existsSync as nodeExistsSync,
  mkdirSync as nodeMkdirSync,
  readdirSync as nodeReaddirSync,
  readFileSync as nodeReadFileSync,
  writeFileSync as nodeWriteFileSync,
} from 'node:fs'
import { join as nodeJoin } from 'node:path'

export function existsSync(path: string): boolean {
  return nodeExistsSync(path)
}

export function readFileSync(path: string, encoding: BufferEncoding): string {
  return nodeReadFileSync(path, encoding)
}

export function writeFileSync(path: string, data: string): void {
  nodeWriteFileSync(path, data)
}

export function readdirSync(path: string): string[] {
  return nodeReaddirSync(path)
}

export function mkdirSync(path: string, options?: { recursive?: boolean }): void {
  nodeMkdirSync(path, options)
}

export function join(...paths: string[]): string {
  return nodeJoin(...paths)
}
