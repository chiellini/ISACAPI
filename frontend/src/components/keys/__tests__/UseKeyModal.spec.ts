import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn().mockResolvedValue(true)
  })
}))

import UseKeyModal from '../UseKeyModal.vue'

describe('UseKeyModal', () => {
  it('renders GPT-5.6 and goals feature in OpenAI Codex config', () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-test',
        baseUrl: 'https://example.com/v1',
        platform: 'openai'
      },
      global: {
        stubs: {
          BaseDialog: {
            template: '<div><slot /><slot name="footer" /></div>'
          },
          Icon: {
            template: '<span />'
          }
        }
      }
    })

    const codeBlocks = wrapper.findAll('pre code').map((code) => code.text())
    const configToml = codeBlocks.find((content) => content.includes('model_provider = "OpenAI"'))

    expect(configToml).toBeDefined()
    expect(configToml).toContain('# ISACAPI Codex default model: gpt-5.6-sol')
    expect(configToml).toContain('model = "gpt-5.6-sol"')
    expect(configToml).toContain('review_model = "gpt-5.6-sol"')
    expect(configToml).not.toContain('model = "gpt-5.5"')
    expect(wrapper.text()).toContain('gpt-5.6-sol')
    expect(configToml).not.toContain('model = "gpt-5.4"')
    expect(configToml).not.toContain('model_context_window')
    expect(configToml).not.toContain('model_auto_compact_token_limit')
    expect(configToml).toContain('[features]\ngoals = true')
  })

  it('renders GPT-5.6 and goals feature in OpenAI Codex WebSocket config', async () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-test',
        baseUrl: 'https://example.com/v1',
        platform: 'openai'
      },
      global: {
        stubs: {
          BaseDialog: {
            template: '<div><slot /><slot name="footer" /></div>'
          },
          Icon: {
            template: '<span />'
          }
        }
      }
    })

    const wsTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.codexCliWs')
    )

    expect(wsTab).toBeDefined()
    await wsTab!.trigger('click')
    await nextTick()

    const codeBlocks = wrapper.findAll('pre code').map((code) => code.text())
    const configToml = codeBlocks.find((content) => content.includes('supports_websockets = true'))

    expect(configToml).toBeDefined()
    expect(configToml).toContain('# ISACAPI Codex default model: gpt-5.6-sol')
    expect(configToml).toContain('model = "gpt-5.6-sol"')
    expect(configToml).toContain('review_model = "gpt-5.6-sol"')
    expect(configToml).not.toContain('model = "gpt-5.5"')
    expect(configToml).not.toContain('model = "gpt-5.4"')
    expect(configToml).not.toContain('model_context_window')
    expect(configToml).not.toContain('model_auto_compact_token_limit')
    expect(configToml).toContain('[features]\nresponses_websockets_v2 = true\ngoals = true')
  })

  it('renders GPT-5.4 mini entry in OpenCode config', async () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-test',
        baseUrl: 'https://example.com/v1',
        platform: 'openai'
      },
      global: {
        stubs: {
          BaseDialog: {
            template: '<div><slot /><slot name="footer" /></div>'
          },
          Icon: {
            template: '<span />'
          }
        }
      }
    })

    const opencodeTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.opencode')
    )

    expect(opencodeTab).toBeDefined()
    await opencodeTab!.trigger('click')
    await nextTick()

    const codeBlock = wrapper.find('pre code')
    expect(codeBlock.exists()).toBe(true)
    expect(codeBlock.text()).toContain('"name": "GPT-5.4 Mini"')
    expect(codeBlock.text()).not.toContain('"name": "GPT-5.4 Nano"')
  })

  const oneClickStubs = {
    BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' },
    Icon: { template: '<span />' }
  }

  it('builds a one-click install command writing ~/.claude/settings.json for Anthropic', () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-test',
        baseUrl: 'https://example.com',
        platform: 'anthropic'
      },
      global: { stubs: oneClickStubs }
    })

    const script = wrapper
      .findAll('pre code')
      .map((code) => code.text())
      .find((content) => content.includes('.claude/settings.json'))

    expect(script).toBeDefined()
    expect(script).toContain('mkdir -p "$HOME/.claude"')
    expect(script).toContain(`cat > "$HOME/.claude/settings.json" <<'SUB2API_EOF'`)
    expect(script).toContain('"ANTHROPIC_BASE_URL": "https://example.com"')
    expect(script).toContain('"ANTHROPIC_AUTH_TOKEN": "sk-test"')
    expect(script).toContain('"ANTHROPIC_API_KEY": "sk-test"')
    expect(script).toContain('for rc in "$HOME/.bashrc" "$HOME/.bash_profile" "$HOME/.zshrc" "$HOME/.profile"; do')
    expect(script).toContain('# >>> ISACAPI API env >>>')
    expect(script).toContain("export ANTHROPIC_API_KEY='sk-test'")
    expect(script).toContain('terminal.integrated.env.osx')
    expect(script).toContain('terminal.integrated.env.linux')
  })

  it('builds a one-click install command writing both Codex files for OpenAI', () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-test',
        baseUrl: 'https://example.com/v1',
        platform: 'openai'
      },
      global: { stubs: oneClickStubs }
    })

    const script = wrapper
      .findAll('pre code')
      .map((code) => code.text())
      .find((content) => content.includes('.codex/config.toml'))

    expect(script).toBeDefined()
    expect(script).toContain('# keys.useKeyModal.oneClick.scriptComment (Codex: gpt-5.6-sol)')
    expect(script).toContain('.codex/config.toml')
    expect(script).toContain('.codex/auth.json')
    expect(script).toContain('"OPENAI_API_KEY": "sk-test"')
    expect(script).toContain('model_provider = "OpenAI"')
    expect(script).toContain('# ISACAPI Codex default model: gpt-5.6-sol')
    expect(script).toContain('model = "gpt-5.6-sol"')
    expect(script).toContain('review_model = "gpt-5.6-sol"')
    expect(script).not.toContain('model = "gpt-5.5"')
    expect(script).toContain("export OPENAI_BASE_URL='https://example.com/v1'")
    expect(script).toContain("export OPENAI_API_BASE='https://example.com/v1'")
    expect(script).toContain('VS Code terminal env updated')
  })

  it('switches the one-click command to PowerShell on Windows', async () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-test',
        baseUrl: 'https://example.com',
        platform: 'anthropic'
      },
      global: { stubs: oneClickStubs }
    })

    const winButton = wrapper.findAll('button').find((b) => b.text() === 'Windows')
    expect(winButton).toBeDefined()
    await winButton!.trigger('click')
    await nextTick()

    const script = wrapper
      .findAll('pre code')
      .map((code) => code.text())
      .find((content) => content.includes('Set-Content'))

    expect(script).toBeDefined()
    expect(script).toContain('$env:USERPROFILE\\.claude\\settings.json')
    expect(script).toContain('New-Item -ItemType Directory -Force')
    expect(script).toContain('[Environment]::SetEnvironmentVariable')
    expect(script).toContain('$PROFILE.CurrentUserAllHosts')
    expect(script).toContain('terminal.integrated.env.windows')
  })

  it('renders Claude Fable 5 OpenCode config with adaptive thinking', async () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-test',
        baseUrl: 'https://example.com/v1',
        platform: 'antigravity'
      },
      global: {
        stubs: {
          BaseDialog: {
            template: '<div><slot /><slot name="footer" /></div>'
          },
          Icon: {
            template: '<span />'
          }
        }
      }
    })

    const opencodeTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.opencode')
    )

    expect(opencodeTab).toBeDefined()
    await opencodeTab!.trigger('click')
    await nextTick()

    const claudeConfig = wrapper.findAll('pre code')
      .map((code) => code.text())
      .find((content) => content.includes('"antigravity-claude"'))

    expect(claudeConfig).toBeDefined()
    const parsed = JSON.parse(claudeConfig!)
    const fable = parsed.provider['antigravity-claude'].models['claude-fable-5']

    expect(fable.name).toBe('Claude Fable 5')
    expect(fable.limit).toEqual({ context: 1048576, output: 128000 })
    expect(fable.options.thinking).toEqual({ type: 'adaptive' })
    expect(fable.options.thinking).not.toHaveProperty('budgetTokens')
  })
})
