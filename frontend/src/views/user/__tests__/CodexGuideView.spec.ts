import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const viewPath = resolve(dirname(fileURLToPath(import.meta.url)), '../CodexGuideView.vue')
const viewSource = readFileSync(viewPath, 'utf8')

describe('CodexGuideView', () => {
  it('covers installation, generated configuration, and everyday CLI usage', () => {
    expect(viewSource).toContain('npm install -g @openai/codex')
    expect(viewSource).toContain("codexGuide.configure.steps")
    expect(viewSource).toContain('~/.codex/config.toml')
    expect(viewSource).toContain('%USERPROFILE%\\\\.codex\\\\config.toml')
    expect(viewSource).toContain('codex resume --last')
    expect(viewSource).toContain('codex exec')
  })

  it('links users back to API keys and the CC-Switch alternative', () => {
    expect(viewSource).toContain('to="/keys"')
    expect(viewSource).toContain('to="/cc-switch"')
    expect(viewSource).toContain('https://learn.chatgpt.com/docs/codex/cli')
  })
})
