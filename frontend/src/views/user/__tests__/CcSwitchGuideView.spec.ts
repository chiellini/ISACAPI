import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const viewPath = resolve(dirname(fileURLToPath(import.meta.url)), '../CcSwitchGuideView.vue')
const viewSource = readFileSync(viewPath, 'utf8')

describe('CcSwitchGuideView', () => {
  it('documents local CC-Switch and remote SSH setup paths', () => {
    expect(viewSource).toContain('CC_SWITCH_DOWNLOAD_LINKS')
    expect(viewSource).toContain("ccSwitchGuide.scenarios.local.points")
    expect(viewSource).toContain("ccSwitchGuide.scenarios.remote.points")
    expect(viewSource).toContain('~/.claude/settings.json')
    expect(viewSource).toContain('~/.codex/config.toml')
    expect(viewSource).toContain('ANTHROPIC_BASE_URL')
    expect(viewSource).toContain('OPENAI_BASE_URL')
  })
})
