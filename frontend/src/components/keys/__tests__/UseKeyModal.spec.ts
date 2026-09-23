import { afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import type { GroupPlatform } from '@/types'

const { copyToClipboardMock, saveAsMock } = vi.hoisted(() => ({
  copyToClipboardMock: vi.fn().mockResolvedValue(true),
  saveAsMock: vi.fn()
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: copyToClipboardMock
  })
}))

vi.mock('file-saver', () => ({
  saveAs: saveAsMock
}))

import UseKeyModal from '../UseKeyModal.vue'

function readBlobAsText(blob: Blob): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.addEventListener('load', () => resolve(String(reader.result || '')))
    reader.addEventListener('error', () => reject(reader.error))
    reader.readAsText(blob)
  })
}

describe('UseKeyModal', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
    saveAsMock.mockClear()
  })

  it('omits the attribution override from every standard Claude Code setup form', async () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-anthropic-test',
        baseUrl: 'https://example.com/v1',
        platform: 'anthropic'
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

    for (const [shell, trafficSetting] of [
      ['macOS / Linux', 'export CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1'],
      ['Windows CMD', 'set CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1'],
      ['PowerShell', '$env:CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1']
    ]) {
      if (shell !== 'macOS / Linux') {
        const shellTab = wrapper.findAll('button').find(
          (button) => button.text().trim() === shell
        )
        expect(shellTab).toBeDefined()
        await shellTab!.trigger('click')
        await nextTick()
      }

      const codeBlocks = wrapper.findAll('pre code').map((code) => code.text())
      const allCode = codeBlocks.join('\n')
      const settings = JSON.parse(codeBlocks.find((content) => content.includes('"$schema"'))!)

      expect(allCode).not.toContain('CLAUDE_CODE_ATTRIBUTION_HEADER')
      expect(allCode).toContain(trafficSetting)
      expect(settings.env.CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC).toBe('1')
      expect(settings.env).not.toHaveProperty('CLAUDE_CODE_ATTRIBUTION_HEADER')
    }
  })

  it('renders Grok Build and OpenCode setup for Grok groups', async () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-grok-test',
        baseUrl: 'https://example.com/v1',
        platform: 'grok'
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

    const grokTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.grokCli')
    )
    expect(grokTab).toBeDefined()

    const allCode = wrapper.findAll('pre code').map((code) => code.text()).join('\n')
    expect(allCode).toContain('GROK_MODELS_BASE_URL')
    expect(allCode).toContain('XAI_API_KEY')
    expect(allCode).toContain('[model."grok-4.5"]')
    expect(allCode).toContain('[model."grok-build-0.1"]')
    expect(allCode).toContain('[model."grok-4.20-multi-agent-0309"]')
    expect(allCode).toContain('[model."grok-4.3"]')
    expect(allCode).toContain('default = "grok-4.5"')
    expect(allCode).toContain('models_base_url = "https://example.com/v1"')
    expect(allCode).toContain('models_list_url = "https://example.com/v1/models"')
    expect(allCode).toContain('xai_api_base_url = "https://example.com/v1"')
    expect(allCode).toContain('cli_chat_proxy_base_url = "https://example.com/v1"')
    expect(allCode).toContain('preferred_method = "api_key"')
    expect(allCode).toContain('image_description = "grok-4.5"')
    expect(allCode).toContain('auto_compact_threshold_percent = 80')
    expect(allCode).toContain('image_gen = true')
    expect(allCode).toContain('video_gen = true')
    expect(allCode).toContain('image_gen_model_override = "grok-imagine-image-quality"')
    expect(allCode).toContain('image_edit_model_override = "grok-imagine-edit"')
    expect(allCode).toContain('env_key = "XAI_API_KEY"')
    expect(allCode).toContain('Keep api_backend = "responses" on every model entry.')
    expect(allCode).toContain('grok-imagine-image')
    expect(allCode).toContain('grok-imagine-edit')
    expect(allCode).toMatch(/\[model\."grok-4\.5"\][\s\S]*?context_window = 500000/)
    expect(allCode).toMatch(/\[model\."grok-build-0\.1"\][\s\S]*?context_window = 256000/)
    // Prefer env_key; hardcode api_key only as commented alternative
    expect(allCode).not.toMatch(/^api_key = "sk-grok-test"$/m)

    const modelBlocks = allCode
      .split(/(?=^\[model\.)/m)
      .filter((block) => block.startsWith('[model."'))
    expect(modelBlocks.length).toBeGreaterThanOrEqual(4)
    for (const block of modelBlocks) {
      if (block.includes('# [model.')) continue
      expect(block).toContain('api_backend = "responses"')
    }

    const windowsTab = wrapper.findAll('button').find(
      (button) => button.text().trim() === 'Windows'
    )
    expect(windowsTab).toBeDefined()
    await windowsTab!.trigger('click')
    await nextTick()
    expect(wrapper.text().toLowerCase()).toContain('%userprofile%\\.grok\\config.toml')

    const opencodeTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.opencode')
    )
    expect(opencodeTab).toBeDefined()
    await opencodeTab!.trigger('click')
    await nextTick()

    const opencodeConfig = wrapper
      .findAll('pre code')
      .map((code) => code.text())
      .find((content) =>
        content.trim().startsWith('{') && content.includes('"provider"') && content.includes('"grok"')
      )
    expect(opencodeConfig).toBeDefined()
    const parsed = JSON.parse(opencodeConfig!)
    expect(parsed.provider.grok.npm).toBe('@ai-sdk/openai-compatible')
    expect(parsed.provider.grok.name).toBe('Grok via Sub2API')
    expect(parsed.provider.grok.options).toEqual({
      baseURL: 'https://example.com/v1',
      apiKey: 'sk-grok-test'
    })
    expect(parsed.provider.grok.models['grok-4.5']).toBeDefined()
    expect(parsed.provider.grok.models['grok-4.5'].limit.context).toBe(500000)
    expect(parsed.provider.grok.models['grok-build-0.1']).toBeDefined()
    expect(parsed.provider.grok.models['grok-4.20-multi-agent-0309']).toBeDefined()
    expect(parsed.provider.grok.models['grok-composer-2.5-fast']).toBeDefined()
    expect(parsed.provider.grok.models['gpt-5.6']).toBeUndefined()
  })

  it('renders copyable Claude Code setup through the Grok Messages gateway', async () => {
    copyToClipboardMock.mockClear()
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-grok-claude-test',
        baseUrl: 'https://example.com/v1',
        platform: 'grok'
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

    const claudeTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.claudeCode')
    )
    expect(claudeTab).toBeDefined()
    await claudeTab!.trigger('click')
    await nextTick()

    let codeBlocks = wrapper.findAll('pre code').map((code) => code.text())
    expect(codeBlocks.join('\n')).toContain('ANTHROPIC_BASE_URL="https://example.com"')
    expect(codeBlocks.join('\n')).toContain('ANTHROPIC_AUTH_TOKEN="sk-grok-claude-test"')
    const unixConfig = codeBlocks.find((content) => content.startsWith('export ANTHROPIC_BASE_URL'))
    expect(unixConfig).toBeDefined()
    for (const name of [
      'ANTHROPIC_MODEL',
      'ANTHROPIC_DEFAULT_OPUS_MODEL',
      'ANTHROPIC_DEFAULT_SONNET_MODEL',
      'ANTHROPIC_DEFAULT_HAIKU_MODEL',
      'ANTHROPIC_DEFAULT_FABLE_MODEL',
      'CLAUDE_CODE_SUBAGENT_MODEL'
    ]) {
      expect(unixConfig).toContain(`export ${name}="grok-4.5"`)
    }
    const settingsConfig = codeBlocks.find((content) => content.includes('"$schema"'))
    expect(settingsConfig).toBeDefined()
    const parsedSettings = JSON.parse(settingsConfig!)
    expect(parsedSettings.$schema).toBe('https://json.schemastore.org/claude-code-settings.json')
    expect(parsedSettings.env.ANTHROPIC_MODEL).toBe('grok-4.5')
    expect(codeBlocks.join('\n')).not.toContain('CLAUDE_CODE_ATTRIBUTION_HEADER')
    expect(codeBlocks.join('\n')).toContain('CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC')
    expect(parsedSettings.env.CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC).toBe('1')
    expect(parsedSettings.env).not.toHaveProperty('CLAUDE_CODE_ATTRIBUTION_HEADER')
    expect(wrapper.text()).toContain('keys.useKeyModal.claudeSettingsHint')
    expect(wrapper.text()).toContain('keys.useKeyModal.grok.claudeNote')
    expect(wrapper.find('nav[aria-label="Client"]').classes()).toContain('min-w-max')
    expect(wrapper.find('nav[aria-label="Client"]').element.parentElement?.classList.contains('overflow-x-auto')).toBe(true)

    const cmdTab = wrapper.findAll('button').find(
      (button) => button.text().trim() === 'Windows CMD'
    )
    expect(cmdTab).toBeDefined()
    await cmdTab!.trigger('click')
    await nextTick()

    codeBlocks = wrapper.findAll('pre code').map((code) => code.text())
    expect(codeBlocks.join('\n')).toContain('set ANTHROPIC_MODEL=grok-4.5')
    expect(codeBlocks.join('\n')).toContain('set ANTHROPIC_DEFAULT_FABLE_MODEL=grok-4.5')
    expect(codeBlocks.join('\n')).toContain('set CLAUDE_CODE_SUBAGENT_MODEL=grok-4.5')
    expect(codeBlocks.join('\n')).toContain('set CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1')
    expect(codeBlocks.join('\n')).not.toContain('CLAUDE_CODE_ATTRIBUTION_HEADER')
    const cmdSettings = JSON.parse(codeBlocks.find((content) => content.includes('"$schema"'))!)
    expect(cmdSettings.env.CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC).toBe('1')
    expect(cmdSettings.env).not.toHaveProperty('CLAUDE_CODE_ATTRIBUTION_HEADER')

    const powershellTab = wrapper.findAll('button').find(
      (button) => button.text().trim() === 'PowerShell'
    )
    expect(powershellTab).toBeDefined()
    await powershellTab!.trigger('click')
    await nextTick()

    codeBlocks = wrapper.findAll('pre code').map((code) => code.text())
    expect(codeBlocks.join('\n')).toContain('$env:ANTHROPIC_BASE_URL="https://example.com"')
    expect(codeBlocks.join('\n')).toContain('$env:ANTHROPIC_MODEL="grok-4.5"')
    expect(codeBlocks.join('\n')).toContain('$env:ANTHROPIC_DEFAULT_FABLE_MODEL="grok-4.5"')
    expect(codeBlocks.join('\n')).toContain('$env:CLAUDE_CODE_SUBAGENT_MODEL="grok-4.5"')
    expect(codeBlocks.join('\n')).toContain('$env:CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC="1"')
    expect(codeBlocks.join('\n')).not.toContain('CLAUDE_CODE_ATTRIBUTION_HEADER')
    const powershellSettings = JSON.parse(codeBlocks.find((content) => content.includes('"$schema"'))!)
    expect(powershellSettings.env.CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC).toBe('1')
    expect(powershellSettings.env).not.toHaveProperty('CLAUDE_CODE_ATTRIBUTION_HEADER')
    expect(wrapper.text()).toContain('%USERPROFILE%\\.claude\\settings.json')

    const copyButton = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.copy')
    )
    expect(copyButton).toBeDefined()
    await copyButton!.trigger('click')
    expect(copyToClipboardMock).toHaveBeenCalledWith(
      expect.stringContaining('ANTHROPIC_AUTH_TOKEN="sk-grok-claude-test"'),
      'keys.copied'
    )
  })

  it('renders Codex custom provider setup through the Grok Responses gateway', async () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-grok-codex-test',
        baseUrl: 'https://example.com/v1',
        platform: 'grok'
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

    const codexTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.codexCli')
    )
    expect(codexTab).toBeDefined()
    await codexTab!.trigger('click')
    await nextTick()

    let codeBlocks = wrapper.findAll('pre code').map((code) => code.text())
    const configToml = codeBlocks.find((content) => content.includes('[model_providers.sub2api]'))
    expect(configToml).toBeDefined()
    expect(configToml).toContain('model_provider = "sub2api"')
    expect(configToml).toContain('model = "grok-4.5"')
    expect(configToml).toContain('base_url = "https://example.com/v1"')
    expect(configToml).toContain('env_key = "SUB2API_API_KEY"')
    expect(configToml).toContain('wire_api = "responses"')
    // API-key provider: Codex must not require a ChatGPT OAuth login.
    expect(configToml).toContain('requires_openai_auth = false')
    expect(configToml).toContain('supports_websockets = false')
    expect(configToml).toContain('grok-4.20-multi-agent-0309 (text / web_search)')
    expect(configToml).toContain('grok-imagine-image')
    expect(configToml).toContain('grok-imagine-video')
    // Hardcoded bearer is only a commented fallback when env cannot be set.
    expect(configToml).toMatch(/# experimental_bearer_token = "sk-grok-codex-test"/)
    expect(configToml).not.toContain('supports_websockets = true')
    expect(configToml).not.toContain('responses_websockets_v2')
    expect(wrapper.text()).not.toContain('auth.json')
    expect(codeBlocks.join('\n')).toContain('SUB2API_API_KEY')

    const windowsTab = wrapper.findAll('button').find(
      (button) => button.text().trim() === 'Windows'
    )
    expect(windowsTab).toBeDefined()
    await windowsTab!.trigger('click')
    await nextTick()

    codeBlocks = wrapper.findAll('pre code').map((code) => code.text())
    expect(wrapper.text().toLowerCase()).toContain('%userprofile%\\.codex\\config.toml'.toLowerCase())
    expect(codeBlocks.join('\n')).toContain('experimental_bearer_token = "sk-grok-codex-test"')
  })

  it('keeps legacy OpenAI Codex config as the default with GPT-5.6 and goals', () => {
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
    expect(configToml).toContain('requires_openai_auth = true')
    expect(configToml).not.toContain('experimental_bearer_token')
    expect(configToml).not.toContain('x-openai-actor-authorization')
    expect(configToml).not.toContain('env_key')
    expect(configToml).not.toContain('image_generation')
    expect(configToml).not.toContain('supports_websockets')
    expect(configToml).not.toContain('responses_websockets_v2')
    expect(configToml).toContain('[features]\ngoals = true')
    expect(configToml).not.toContain('model_reasoning_effort = "xhigh"')
    expect(codeBlocks).toContain('{\n  "OPENAI_API_KEY": "sk-test"\n}')
    expect(wrapper.text()).toContain('auth.json')
    expect(wrapper.find('[data-testid="codex-api-key-restart-notice"]').exists()).toBe(false)
  })

  it('normalizes OpenAI Codex and Claude base URLs independently', async () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-test',
        baseUrl: 'https://example.com',
        platform: 'openai',
        allowMessagesDispatch: true
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

    const codexConfig = wrapper.findAll('pre code')
      .map((code) => code.text())
      .find((content) => content.includes('model_provider = "OpenAI"'))
    expect(codexConfig).toContain('base_url = "https://example.com/v1"')

    const claudeTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.claudeCode')
    )
    expect(claudeTab).toBeDefined()
    await claudeTab!.trigger('click')
    await nextTick()

    const claudeConfig = wrapper.findAll('pre code')
      .map((code) => code.text())
      .find((content) => content.startsWith('export ANTHROPIC_BASE_URL'))
    expect(claudeConfig).toContain('ANTHROPIC_BASE_URL="https://example.com"')
  })

  it('renders API Key Mode authorization in OpenAI Codex config', async () => {
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

    const apiKeyMode = wrapper.get('[data-testid="codex-auth-mode-api-key"]')
    await apiKeyMode.trigger('click')
    await nextTick()

    const codeBlocks = wrapper.findAll('pre code').map((code) => code.text())
    const configToml = codeBlocks.find((content) => content.includes('model_provider = "OpenAI"'))

    expect(apiKeyMode.attributes('aria-checked')).toBe('true')
    expect(configToml).toBeDefined()
    expect(configToml).toContain('requires_openai_auth = false')
    expect(configToml).toContain('experimental_bearer_token = "sk-test"')
    expect(configToml).toContain('http_headers = { "x-openai-actor-authorization" = "local-image-extension" }')
    expect(configToml).not.toContain('env_key')
    expect(configToml).not.toContain('image_generation')
    expect(codeBlocks).not.toContain('{\n  "OPENAI_API_KEY": "sk-test"\n}')
    expect(wrapper.text()).not.toContain('auth.json')

    const restartNotice = wrapper.get('[data-testid="codex-api-key-restart-notice"]')
    expect(restartNotice.text()).toContain(
      'keys.useKeyModal.openai.authModeApiKeyRestartNotice'
    )

    await wrapper.get('[data-testid="codex-auth-mode-legacy"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="codex-api-key-restart-notice"]').exists()).toBe(false)
    expect(wrapper.findAll('pre code').map((code) => code.text()).join('\n')).not.toContain(
      'x-openai-actor-authorization'
    )
  })

  it('keeps legacy OpenAI Codex WebSocket config as the default with GPT-5.6 and goals', async () => {
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
    expect(configToml).toContain('requires_openai_auth = true')
    expect(configToml).not.toContain('experimental_bearer_token')
    expect(configToml).not.toContain('x-openai-actor-authorization')
    expect(configToml).not.toContain('env_key')
    expect(configToml).not.toContain('image_generation')
    expect(configToml).toContain('supports_websockets = true')
    expect(configToml).toContain('[features]\nresponses_websockets_v2 = true\ngoals = true')
    expect(codeBlocks).toContain('{\n  "OPENAI_API_KEY": "sk-test"\n}')
    expect(wrapper.text()).toContain('auth.json')
  })

  it('preserves API Key Mode when switching to OpenAI Codex WebSocket config', async () => {
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

    const apiKeyMode = wrapper.get('[data-testid="codex-auth-mode-api-key"]')
    await apiKeyMode.trigger('click')

    const wsTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.codexCliWs')
    )
    expect(wsTab).toBeDefined()
    await wsTab!.trigger('click')
    await nextTick()

    const codeBlocks = wrapper.findAll('pre code').map((code) => code.text())
    const configToml = codeBlocks.find((content) => content.includes('supports_websockets = true'))

    expect(wrapper.get('[data-testid="codex-auth-mode-api-key"]').attributes('aria-checked')).toBe('true')
    expect(configToml).toBeDefined()
    expect(configToml).toContain('requires_openai_auth = false')
    expect(configToml).toContain('experimental_bearer_token = "sk-test"')
    expect(configToml).toContain('http_headers = { "x-openai-actor-authorization" = "local-image-extension" }')
    expect(configToml).not.toContain('env_key')
    expect(configToml).not.toContain('image_generation')
    expect(configToml).toContain('supports_websockets = true')
    expect(configToml).toContain('[features]\nresponses_websockets_v2 = true\ngoals = true')
    expect(codeBlocks).not.toContain('{\n  "OPENAI_API_KEY": "sk-test"\n}')
    expect(wrapper.text()).not.toContain('auth.json')
  })

  it('resets Codex authentication mode when the modal reopens or platform changes', async () => {
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

    await wrapper.get('[data-testid="codex-auth-mode-api-key"]').trigger('click')
    await wrapper.setProps({ show: false })
    await wrapper.setProps({ show: true })
    await nextTick()

    expect(wrapper.get('[data-testid="codex-auth-mode-legacy"]').attributes('aria-checked')).toBe('true')
    expect(wrapper.findAll('pre code').map((code) => code.text()).join('\n')).toContain('requires_openai_auth = true')

    await wrapper.get('[data-testid="codex-auth-mode-api-key"]').trigger('click')
    await wrapper.setProps({ platform: 'gemini' })
    await wrapper.setProps({ platform: 'openai' })
    await nextTick()

    expect(wrapper.get('[data-testid="codex-auth-mode-legacy"]').attributes('aria-checked')).toBe('true')
    expect(wrapper.findAll('pre code').map((code) => code.text()).join('\n')).not.toContain('x-openai-actor-authorization')
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

  function mountEndpointExport(platform: GroupPlatform, baseUrl: string) {
    return mount(UseKeyModal, {
      props: { show: true, apiKey: 'sk-endpoint-test', baseUrl, platform, allowMessagesDispatch: true },
      global: { stubs: oneClickStubs }
    })
  }

  function installedUnixFile(script: string, path: string): string {
    const marker = `cat > "$HOME/${path}" <<'SUB2API_EOF'\n`
    expect(script).toContain(marker)
    return script.split(marker)[1]!.split('\nSUB2API_EOF')[0]!
  }

  it.each([
    ['https://example.com', 'https://example.com/v1'],
    ['https://example.com/v1/', 'https://example.com/v1'],
    ['https://example.com/gateway/team/', 'https://example.com/gateway/team/v1'],
    ['https://example.com/gateway/team/v1///', 'https://example.com/gateway/team/v1'],
    ['https://example.com/gateway/team/v1/v1/', 'https://example.com/gateway/team/v1'],
    ['https://example.com/v1/tenant/v1', 'https://example.com/v1/tenant/v1']
  ])('keeps Codex manual and one-click endpoints consistent for %s', async (baseUrl, expectedBase) => {
    const wrapper = mountEndpointExport('openai', baseUrl)
    for (const label of ['codexCli', 'codexCliWs']) {
      const tab = wrapper.find('nav[aria-label="Client"]').findAll('button').find((button) =>
        button.text().includes(`keys.useKeyModal.cliTabs.${label}`)
      )!
      await tab.trigger('click')
      const blocks = wrapper.findAll('pre code').map((code) => code.text())
      const script = blocks.find((content) => content.includes('cat > "$HOME/.codex/config.toml"'))!
      const manualConfig = blocks.find((content) => content.startsWith('# ISACAPI Codex default model:'))!
      const installedConfig = installedUnixFile(script, '.codex/config.toml')

      expect(installedConfig).toBe(manualConfig)
      expect(installedConfig).toContain(`base_url = "${expectedBase}"`)
      expect(script).toContain(`export OPENAI_BASE_URL='${expectedBase}'`)
      expect(script).toContain(`export OPENAI_API_BASE='${expectedBase}'`)
    }
    wrapper.unmount()
  })

  it.each(['anthropic', 'openai', 'kimi', 'grok', 'antigravity'] as const)(
    'uses the site root in every Claude Code export for %s',
    async (platform) => {
      const wrapper = mountEndpointExport(platform, 'https://example.com/gateway/team/v1///')
      const claudeTab = wrapper.find('nav[aria-label="Client"]').findAll('button').find((button) =>
        button.text().includes('keys.useKeyModal.cliTabs.claudeCode')
      )!
      await claudeTab.trigger('click')
      const expectedBase = `https://example.com/gateway/team${platform === 'antigravity' ? '/antigravity' : ''}`
      const blocks = wrapper.findAll('pre code').map((code) => code.text())
      const script = blocks.find((content) => content.includes('cat > "$HOME/.claude/settings.json"'))!
      const installedSettings = JSON.parse(installedUnixFile(script, '.claude/settings.json'))
      const manualSettings = JSON.parse(blocks.find((content) => content.includes('"$schema"'))!)

      expect(installedSettings.env.ANTHROPIC_BASE_URL).toBe(expectedBase)
      expect(manualSettings.env.ANTHROPIC_BASE_URL).toBe(expectedBase)
      expect(blocks).toContainEqual(expect.stringContaining(`export ANTHROPIC_BASE_URL="${expectedBase}"`))
      expect(script).toContain(`export ANTHROPIC_BASE_URL='${expectedBase}'`)

      const windowsButton = wrapper.findAll('button').find((button) => button.text() === 'Windows')!
      await windowsButton.trigger('click')
      const windowsScript = wrapper.findAll('pre code').map((code) => code.text())
        .find((content) => content.includes('Set-Content'))!
      expect(windowsScript).toContain(`"ANTHROPIC_BASE_URL": "${expectedBase}"`)
      expect(windowsScript).toContain(`$env:ANTHROPIC_BASE_URL='${expectedBase}'`)
      wrapper.unmount()
    }
  )

  it.each([
    ['gemini', 'https://example.com/gateway/v1/', 'https://example.com/gateway'],
    ['gemini', 'https://example.com/gateway/v1beta///', 'https://example.com/gateway'],
    ['antigravity', 'https://example.com/gateway/v1/', 'https://example.com/gateway/antigravity'],
    ['antigravity', 'https://example.com/gateway/antigravity/v1beta/', 'https://example.com/gateway/antigravity']
  ] as const)('uses the same Gemini CLI root for %s at %s', async (platform, baseUrl, expectedBase) => {
    const wrapper = mountEndpointExport(platform, baseUrl)
    const geminiTab = wrapper.find('nav[aria-label="Client"]').findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.geminiCli')
    )!
    await geminiTab.trigger('click')
    const blocks = wrapper.findAll('pre code').map((code) => code.text())
    const script = blocks.find((content) => content.includes('cat > "$HOME/.gemini/.env"'))!

    expect(installedUnixFile(script, '.gemini/.env').split('\n')[0]).toBe(`GOOGLE_GEMINI_BASE_URL=${expectedBase}`)
    expect(blocks).toContainEqual(expect.stringContaining(`export GOOGLE_GEMINI_BASE_URL="${expectedBase}"`))
    expect(script).toContain(`export GOOGLE_GEMINI_BASE_URL='${expectedBase}'`)
    wrapper.unmount()
  })

  it.each([
    ['openai', 'openai', 'v1', 'OPENAI_BASE_URL'],
    ['anthropic', 'anthropic', 'v1', 'ANTHROPIC_BASE_URL'],
    ['gemini', 'gemini', 'v1beta', 'GOOGLE_GEMINI_BASE_URL'],
    ['kimi', 'openai', 'v1', 'OPENAI_BASE_URL'],
    ['grok', 'grok', 'v1', 'OPENAI_BASE_URL']
  ] as const)('keeps the OpenCode file and SDK environment compatible for %s', async (platform, provider, version, envName) => {
    const wrapper = mountEndpointExport(platform, 'https://example.com/gateway/v1beta/')
    const opencodeTab = wrapper.find('nav[aria-label="Client"]').findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.opencode')
    )!
    await opencodeTab.trigger('click')
    const blocks = wrapper.findAll('pre code').map((code) => code.text())
    const script = blocks.find((content) => content.includes('cat > "$HOME/.config/opencode/opencode.json"'))!
    const installedConfig = installedUnixFile(script, '.config/opencode/opencode.json')
    const manualConfig = blocks.find((content) => content.trim().startsWith('{') && content.includes('"provider"'))!

    expect(installedConfig).toBe(manualConfig)
    expect(JSON.parse(installedConfig).provider[provider].options.baseURL).toBe(`https://example.com/gateway/${version}`)
    const envBase = envName === 'OPENAI_BASE_URL' ? 'https://example.com/gateway/v1' : 'https://example.com/gateway'
    expect(script).toContain(`export ${envName}='${envBase}'`)
    wrapper.unmount()
  })

  it.each(['openai', 'anthropic', 'gemini', 'antigravity', 'grok', 'kimi', 'composite'] as const)(
    'does not require an undelivered model catalog in %s Codex exports',
    async (platform) => {
      const wrapper = mountEndpointExport(platform, 'https://example.com/gateway')
      const codexTab = wrapper.find('nav[aria-label="Client"]').findAll('button').find((button) =>
        button.text().includes('keys.useKeyModal.cliTabs.codexCli')
      )!
      await codexTab.trigger('click')

      const blocks = wrapper.findAll('pre code').map((code) => code.text())
      const configs = blocks.filter((content) => content.includes('model_provider = '))
      expect(configs.length).toBeGreaterThan(0)
      for (const config of configs) {
        expect(config).not.toMatch(/^model_catalog_json\s*=/m)
        expect(config).toContain('# model_catalog_json = "~/.codex/codex-models.json"')
        expect(config).not.toContain('cat > "$HOME/.codex/codex-models.json"')
      }
      if (platform !== 'openai' && platform !== 'grok') {
        const script = blocks.find((content) => content.includes('cat > "$HOME/.codex/config.toml"'))!
        const installedConfig = installedUnixFile(script, '.codex/config.toml')
        expect(installedConfig).toContain('model_provider = "sub2api"')
        expect(installedConfig).toContain('env_key = "SUB2API_API_KEY"')
        expect(script).toContain("export SUB2API_API_KEY='sk-endpoint-test'")
        expect(script).not.toContain('cat > "$HOME/.codex/auth.json"')
      }

      const windowsTab = wrapper.find('nav[aria-label="Tabs"]').findAll('button').find((button) =>
        button.text().trim() === 'Windows'
      )!
      await windowsTab.trigger('click')
      const windowsConfig = wrapper.findAll('pre code').map((code) => code.text())
        .find((content) => content.includes('model_provider = ') && !content.includes('mkdir -p'))!
      expect(windowsConfig).not.toMatch(/^model_catalog_json\s*=/m)
      expect(windowsConfig).toContain('# model_catalog_json = "%userprofile%\\\\.codex\\\\codex-models.json"')
      wrapper.unmount()
    }
  )

  it.each(['openai', 'composite'] as const)(
    'bundles the fetched %s catalog before enabling it in one-click scripts',
    async (platform) => {
      const manifest = {
        models: [{
          slug: 'gpt-5.6-sol',
          default_reasoning_level: 'high',
          supported_reasoning_levels: [{ effort: 'high', description: 'High reasoning' }],
          input_modalities: ['text', 'image'],
          model_messages: { instructions_template: 'Keep the complete model descriptor.' }
        }],
        metadata: { source: 'routed-group' }
      }
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        json: async () => manifest
      }))
      const wrapper = mountEndpointExport(platform, 'https://example.com/gateway/v1')
      const codexTab = wrapper.find('nav[aria-label="Client"]').findAll('button').find((button) =>
        button.text().includes('keys.useKeyModal.cliTabs.codexCli')
      )!
      await codexTab.trigger('click')
      await wrapper.get('[data-testid="codex-model-catalog-fetch"]').trigger('click')
      await flushPromises()

      const blocks = wrapper.findAll('pre code').map((code) => code.text())
      const script = blocks.find((content) => content.includes('cat >> "$HOME/.codex/config.toml"'))!
      expect(JSON.parse(installedUnixFile(script, '.codex/codex-models.json'))).toEqual(manifest)
      expect(script).toContain('isacapi_catalog_path=$(printf \'%s\' "$HOME/.codex/codex-models.json"')
      expect(script).toContain('printf \'model_catalog_json = "%s"\\n\' "$isacapi_catalog_path" > "$HOME/.codex/config.toml"')
      expect(script.indexOf('cat > "$HOME/.codex/codex-models.json"'))
        .toBeLessThan(script.indexOf('cat >> "$HOME/.codex/config.toml"'))
      const manualConfig = blocks.find((content) => content.includes('model_provider = ') && !content.includes('mkdir -p'))!
      expect(manualConfig).not.toMatch(/^model_catalog_json\s*=/m)

      const windowsButton = wrapper.findAll('button').find((button) => button.text() === 'Windows')!
      await windowsButton.trigger('click')
      const windowsScript = wrapper.findAll('pre code').map((code) => code.text())
        .find((content) => content.includes('$isacapiCodexConfig'))!
      const catalogWrite = '\n\'@\n[System.IO.File]::WriteAllText("$env:USERPROFILE\\.codex\\codex-models.json"'
      const catalogText = windowsScript.split(catalogWrite)[0]!.split("@'\n").pop()!
      expect(JSON.parse(catalogText)).toEqual(manifest)
      expect(windowsScript).toContain("Join-Path $env:USERPROFILE '.codex\\codex-models.json'")
      expect(windowsScript).toContain("$isacapiCodexConfig = 'model_catalog_json = \"' + $isacapiCatalogPath")
      expect(windowsScript).toContain('[System.IO.File]::WriteAllText("$env:USERPROFILE\\.codex\\config.toml", $isacapiCodexConfig, (New-Object System.Text.UTF8Encoding($false)))')
      expect(windowsScript).toContain('[System.IO.File]::WriteAllText("$env:USERPROFILE\\.codex\\codex-models.json", $isacapiCodexFile, (New-Object System.Text.UTF8Encoding($false)))')
      if (platform === 'openai') {
        expect(windowsScript).toContain('[System.IO.File]::WriteAllText("$env:USERPROFILE\\.codex\\auth.json", $isacapiCodexFile, (New-Object System.Text.UTF8Encoding($false)))')
      }
      expect(windowsScript).not.toMatch(/Set-Content -Path "\$env:USERPROFILE\\\.codex\\/)
      expect(windowsScript).not.toMatch(/^model_catalog_json\s*=\s*"[%~]/m)
      wrapper.unmount()
    }
  )

  it('renders GPT-5.6 and GPT-6 Astra capabilities in OpenCode config', async () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-test',
        baseUrl: 'https://example.com/v1',
        platform: 'openai'
      },
      global: { stubs: oneClickStubs }
    })

    const opencodeTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.opencode')
    )
    expect(opencodeTab).toBeDefined()
    await opencodeTab!.trigger('click')
    await nextTick()

    const config = wrapper
      .findAll('pre code')
      .map((code) => code.text())
      .find((content) => content.trim().startsWith('{') && content.includes('"gpt-5.6-sol"'))

    expect(config).toBeDefined()
    const parsed = JSON.parse(config!)
    const models = parsed.provider.openai.models
    for (const model of ['gpt-5.6', 'gpt-5.6-sol', 'gpt-5.6-terra', 'gpt-5.6-luna']) {
      expect(models[model]).toBeDefined()
      expect(models[model].variants).toHaveProperty('max')
      expect(models[model].variants).toHaveProperty('xhigh')
    }
    expect(models['gpt-5.6'].name).toBe('GPT-5.6 (Sol)')
    expect(models['gpt-6']).toEqual({
      name: 'GPT-6 (Astra)',
      limit: { context: 1050000, output: 128000 },
      options: { store: false },
      variants: { low: {}, medium: {}, high: {}, xhigh: {}, max: {} }
    })
    expect(models['gpt-6-astra']).toEqual({
      name: 'GPT-6 Astra',
      limit: { context: 1050000, output: 128000 },
      options: { store: false },
      variants: { low: {}, medium: {}, high: {}, xhigh: {}, max: {} }
    })
  })

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
    const fable51 = parsed.provider['antigravity-claude'].models['claude-fable-5-1']
    const fable = parsed.provider['antigravity-claude'].models['claude-fable-5']

    expect(fable51.name).toBe('Claude Fable 5.1')
    expect(fable51.limit).toEqual({ context: 1048576, output: 128000 })
    expect(fable51.options.thinking).toEqual({ type: 'adaptive' })
    expect(fable51.options.thinking).not.toHaveProperty('budgetTokens')
    expect(fable.name).toBe('Claude Fable 5')
    expect(fable.limit).toEqual({ context: 1048576, output: 128000 })
    expect(fable.options.thinking).toEqual({ type: 'adaptive' })
    expect(fable.options.thinking).not.toHaveProperty('budgetTokens')
  })

  // Scenario: API Key users can fetch a routed group catalog and reference it from config.toml.
  it('offers a downloadable Codex catalog for Composite API keys', async () => {
    const manifest = {
      models: [
        {
          slug: 'claude-opus-4-8',
          default_reasoning_level: 'medium',
          supported_reasoning_levels: [{ effort: 'max', description: 'Maximum reasoning depth' }],
          input_modalities: ['text'],
          model_messages: { instructions_template: 'Use the routed model.' }
        },
        {
          slug: 'grok-4.6',
          default_reasoning_level: 'high',
          supported_reasoning_levels: [{ effort: 'xhigh', description: 'Extra-high reasoning depth' }],
          input_modalities: ['text'],
          model_messages: { instructions_template: 'Use the routed model.' }
        }
      ]
    }
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => manifest
    })
    vi.stubGlobal('fetch', fetchMock)

    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-composite-test',
        baseUrl: 'https://example.com/v1',
        platform: 'composite'
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

    const codexTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.codexCli')
    )
    expect(codexTab).toBeDefined()
    await codexTab!.trigger('click')
    await nextTick()

    const unixConfig = wrapper.findAll('pre code')
      .map((code) => code.text())
      .find((content) => content.includes('[model_providers.sub2api]'))
    expect(unixConfig).toContain('model_catalog_json = "~/.codex/codex-models.json"')
    expect(unixConfig).toContain('env_key = "SUB2API_API_KEY"')

    await wrapper.get('[data-testid="codex-model-catalog-fetch"]').trigger('click')
    await flushPromises()

    expect(fetchMock).toHaveBeenCalledWith(
      'https://example.com/v1/models?client_version=0.147.0',
      expect.objectContaining({
        headers: expect.objectContaining({ Authorization: 'Bearer sk-composite-test' })
      })
    )
    expect(wrapper.get('[data-testid="codex-model-catalog"]').text())
      .toContain('keys.useKeyModal.codexModelCatalog.download')

    const loadedUnixConfig = wrapper.findAll('pre code')
      .map((code) => code.text())
      .find((content) => content.includes('[model_providers.sub2api]'))
    expect(loadedUnixConfig).toContain('model = "claude-opus-4-8"')
    expect(loadedUnixConfig).toContain('review_model = "claude-opus-4-8"')
    expect(loadedUnixConfig).not.toContain('model = "gpt-5.5"')

    const downloadButton = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.codexModelCatalog.download')
    )
    expect(downloadButton).toBeDefined()
    await downloadButton!.trigger('click')
    expect(saveAsMock).toHaveBeenCalledWith(expect.any(Blob), 'codex-models.json')
    const downloadedBlob = saveAsMock.mock.calls[0]?.[0] as Blob
    expect(JSON.parse(await readBlobAsText(downloadedBlob))).toEqual(manifest)

    const windowsTab = wrapper.find('nav[aria-label="Tabs"]').findAll('button').find((button) => button.text().trim() === 'Windows')
    expect(windowsTab).toBeDefined()
    await windowsTab!.trigger('click')
    await nextTick()

    const windowsConfig = wrapper.findAll('pre code')
      .map((code) => code.text())
      .find((content) => content.includes('[model_providers.sub2api]'))
    expect(windowsConfig).toContain(
      'model_catalog_json = "%userprofile%\\\\.codex\\\\codex-models.json"'
    )
  })

  it.each(['anthropic', 'gemini', 'antigravity', 'kimi', 'zhipu', 'minimax'] as const)(
    'offers Codex catalog configuration for the %s routed group',
    async (platform) => {
      const wrapper = mount(UseKeyModal, {
        props: {
          show: true,
          apiKey: `sk-${platform}-test`,
          baseUrl: 'https://example.com/v1',
          platform
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

      const codexTab = wrapper.findAll('button').find((button) =>
        button.text().includes('keys.useKeyModal.cliTabs.codexCli')
      )
      expect(codexTab).toBeDefined()
      await codexTab!.trigger('click')
      await nextTick()

      expect(wrapper.find('[data-testid="codex-model-catalog"]').exists()).toBe(true)
      const config = wrapper.findAll('pre code')
        .map((code) => code.text())
        .find((content) => content.includes('[model_providers.sub2api]'))
      expect(config).toContain('model_catalog_json = "~/.codex/codex-models.json"')
      expect(config).toContain('base_url = "https://example.com/v1"')
      expect(config).toContain('wire_api = "responses"')
    }
  )

  // Scenario: the platform-preferred model remains selected when the downloaded catalog contains it.
  it('keeps the preferred Composite default when it exists in the catalog', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({
        models: [
          { slug: 'claude-opus-4-8' },
          { slug: 'gpt-5.5' }
        ]
      })
    }))

    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-composite-test',
        baseUrl: 'https://example.com/v1',
        platform: 'composite'
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

    const codexTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.codexCli')
    )
    expect(codexTab).toBeDefined()
    await codexTab!.trigger('click')
    await wrapper.get('[data-testid="codex-model-catalog-fetch"]').trigger('click')
    await flushPromises()

    const config = wrapper.findAll('pre code')
      .map((code) => code.text())
      .find((content) => content.includes('[model_providers.sub2api]'))
    expect(config).toContain('model = "gpt-5.5"')
    expect(config).toContain('review_model = "gpt-5.5"')
  })

  it('derives OpenAI Codex reasoning effort from the selected catalog descriptor', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({
        models: [
          {
            slug: 'glm-5.3',
            default_reasoning_level: 'none',
            supported_reasoning_levels: [{ effort: 'none' }]
          }
        ]
      })
    }))

    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-openai-test',
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

    await wrapper.get('[data-testid="codex-model-catalog-fetch"]').trigger('click')
    await flushPromises()

    const configToml = wrapper.findAll('pre code')
      .map((code) => code.text())
      .find((content) => content.includes('model_provider = "OpenAI"'))
    expect(configToml).toContain('model = "glm-5.3"')
    expect(configToml).not.toContain('model_reasoning_effort')
  })
})
