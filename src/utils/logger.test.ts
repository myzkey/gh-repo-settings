import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { colors, getLogLevel, logger, setLogLevel } from './logger'

describe('colors', () => {
  it('should wrap text with red color codes', () => {
    const result = colors.red('test')
    expect(result).toContain('\x1b[31m')
    expect(result).toContain('test')
    expect(result).toContain('\x1b[0m')
  })

  it('should wrap text with green color codes', () => {
    const result = colors.green('test')
    expect(result).toContain('\x1b[32m')
  })

  it('should wrap text with yellow color codes', () => {
    const result = colors.yellow('test')
    expect(result).toContain('\x1b[33m')
  })

  it('should wrap text with blue color codes', () => {
    const result = colors.blue('test')
    expect(result).toContain('\x1b[34m')
  })

  it('should wrap text with cyan color codes', () => {
    const result = colors.cyan('test')
    expect(result).toContain('\x1b[36m')
  })

  it('should wrap text with gray color codes', () => {
    const result = colors.gray('test')
    expect(result).toContain('\x1b[90m')
  })

  it('should wrap text with bold codes', () => {
    const result = colors.bold('test')
    expect(result).toContain('\x1b[1m')
  })

  it('should wrap text with cyan and bold codes', () => {
    const result = colors.cyanBold('test')
    expect(result).toContain('\x1b[36m')
    expect(result).toContain('\x1b[1m')
  })
})

describe('setLogLevel / getLogLevel', () => {
  afterEach(() => {
    setLogLevel('normal')
  })

  it('should default to normal level', () => {
    setLogLevel('normal')
    expect(getLogLevel()).toBe('normal')
  })

  it('should set verbose level', () => {
    setLogLevel('verbose')
    expect(getLogLevel()).toBe('verbose')
  })

  it('should set quiet level', () => {
    setLogLevel('quiet')
    expect(getLogLevel()).toBe('quiet')
  })
})

describe('logger', () => {
  let consoleSpy: ReturnType<typeof vi.spyOn>
  let consoleErrorSpy: ReturnType<typeof vi.spyOn>

  beforeEach(() => {
    consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
    consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    setLogLevel('normal')
  })

  afterEach(() => {
    consoleSpy.mockRestore()
    consoleErrorSpy.mockRestore()
    setLogLevel('normal')
  })

  describe('debug', () => {
    it('should not log in normal mode', () => {
      setLogLevel('normal')
      logger.debug('test message')
      expect(consoleSpy).not.toHaveBeenCalled()
    })

    it('should log in verbose mode', () => {
      setLogLevel('verbose')
      logger.debug('test message')
      expect(consoleSpy).toHaveBeenCalled()
      expect(consoleSpy.mock.calls[0][0]).toContain('[debug]')
    })

    it('should not log in quiet mode', () => {
      setLogLevel('quiet')
      logger.debug('test message')
      expect(consoleSpy).not.toHaveBeenCalled()
    })
  })

  describe('info', () => {
    it('should log in normal mode', () => {
      setLogLevel('normal')
      logger.info('test message')
      expect(consoleSpy).toHaveBeenCalled()
    })

    it('should log in verbose mode', () => {
      setLogLevel('verbose')
      logger.info('test message')
      expect(consoleSpy).toHaveBeenCalled()
    })

    it('should not log in quiet mode', () => {
      setLogLevel('quiet')
      logger.info('test message')
      expect(consoleSpy).not.toHaveBeenCalled()
    })
  })

  describe('success', () => {
    it('should log in normal mode', () => {
      logger.success('test message')
      expect(consoleSpy).toHaveBeenCalled()
    })

    it('should not log in quiet mode', () => {
      setLogLevel('quiet')
      logger.success('test message')
      expect(consoleSpy).not.toHaveBeenCalled()
    })
  })

  describe('warn', () => {
    it('should log in normal mode', () => {
      logger.warn('test message')
      expect(consoleSpy).toHaveBeenCalled()
    })

    it('should not log in quiet mode', () => {
      setLogLevel('quiet')
      logger.warn('test message')
      expect(consoleSpy).not.toHaveBeenCalled()
    })
  })

  describe('error', () => {
    it('should always log even in quiet mode', () => {
      setLogLevel('quiet')
      logger.error('test message')
      expect(consoleErrorSpy).toHaveBeenCalled()
    })
  })

  describe('heading', () => {
    it('should log in normal mode', () => {
      logger.heading('test heading')
      expect(consoleSpy).toHaveBeenCalled()
    })

    it('should not log in quiet mode', () => {
      setLogLevel('quiet')
      logger.heading('test heading')
      expect(consoleSpy).not.toHaveBeenCalled()
    })
  })

  describe('section', () => {
    it('should log in normal mode', () => {
      logger.section('test section')
      expect(consoleSpy).toHaveBeenCalled()
    })

    it('should not log in quiet mode', () => {
      setLogLevel('quiet')
      logger.section('test section')
      expect(consoleSpy).not.toHaveBeenCalled()
    })
  })

  describe('log', () => {
    it('should log in normal mode', () => {
      logger.log('plain text')
      expect(consoleSpy).toHaveBeenCalledWith('plain text')
    })

    it('should not log in quiet mode', () => {
      setLogLevel('quiet')
      logger.log('plain text')
      expect(consoleSpy).not.toHaveBeenCalled()
    })
  })
})
