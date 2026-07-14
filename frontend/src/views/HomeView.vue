<template>
  <!-- Custom Home Content: Full Page Mode -->
  <div v-if="homeContent" class="flex min-h-screen flex-col">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="min-h-screen w-full flex-1 border-0"
      allowfullscreen
    ></iframe>
    <!-- HTML mode - SECURITY: homeContent is admin-only setting, XSS risk is acceptable -->
    <div v-else class="flex-1" v-html="homeContent"></div>
    <footer class="bg-white px-4 py-5 dark:bg-dark-950" dir="ltr">
      <div class="mx-auto flex max-w-6xl flex-col items-center justify-center gap-2 text-center text-sm text-slate-500 dark:text-dark-400">
        <img :src="companyIconUrl" alt="ISACAI" class="h-10 w-10 rounded-lg object-contain" />
        <p>© 2026 ISACAI. All rights reserved. 联系请加入 QQ 群：1027890648</p>
        <a
          href="https://beian.miit.gov.cn"
          target="_blank"
          rel="noopener noreferrer"
          class="transition-colors hover:text-slate-800 dark:hover:text-white"
        >
          粤ICP备2026050877号
        </a>
      </div>
    </footer>
  </div>

  <div
    v-else
    id="top"
    :dir="isRtl ? 'rtl' : 'ltr'"
    class="relative flex min-h-screen flex-col overflow-x-hidden bg-slate-50 text-slate-950 dark:bg-dark-950 dark:text-white"
  >
    <header class="sticky top-0 z-40 border-b border-slate-200/70 bg-white/85 px-4 backdrop-blur-xl dark:border-dark-800/80 dark:bg-dark-950/85">
      <nav class="mx-auto flex h-16 max-w-7xl items-center justify-between gap-4">
        <router-link to="/home" class="flex min-w-0 items-center gap-2.5" aria-label="ISACAI Home">
          <span class="flex h-10 w-10 shrink-0 items-center justify-center overflow-hidden rounded-xl border border-slate-200/80 bg-white p-1 shadow-sm dark:border-dark-700 dark:bg-white">
            <img :src="companyIconUrl" alt="ISACAI" class="h-full w-full object-contain" />
          </span>
          <span class="min-w-0">
            <span class="block truncate text-sm font-bold tracking-tight text-slate-950 dark:text-white">{{ siteName }}</span>
            <span class="hidden truncate text-[10px] text-slate-500 dark:text-dark-400 sm:block">{{ t('home.nav.tagline') }}</span>
          </span>
        </router-link>

        <div class="hidden items-center gap-1 lg:flex">
          <a href="#top" class="home-nav-link home-nav-link-active">{{ t('home.nav.home') }}</a>
          <router-link :to="dashboardPath" class="home-nav-link">{{ t('home.nav.dashboard') }}</router-link>
          <router-link to="/pricing" class="home-nav-link">{{ t('home.nav.pricing') }}</router-link>
          <a href="#integrations" class="home-nav-link">{{ t('home.nav.integrations') }}</a>
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="home-nav-link"
          >
            {{ t('home.docs') }}
          </a>
        </div>

        <div class="flex items-center gap-1 sm:gap-1.5">
          <LocaleSwitcher />

          <button
            type="button"
            class="home-icon-button"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
            @click="toggleTheme"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>

          <button
            type="button"
            class="home-icon-button relative"
            :title="t('home.notice.open')"
            data-testid="home-notice-bell"
            @click="openNotice"
          >
            <Icon name="bell" size="md" />
            <span class="absolute right-1.5 top-1.5 h-1.5 w-1.5 rounded-full bg-rose-500 ring-2 ring-white dark:ring-dark-950"></span>
          </button>

          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="ml-1 inline-flex items-center gap-1.5 rounded-full bg-sky-500 px-3.5 py-2 text-xs font-semibold text-white shadow-sm shadow-sky-500/20 transition-all hover:bg-sky-600"
          >
            <span class="flex h-5 w-5 items-center justify-center rounded-full bg-white/20 text-[10px] font-bold">{{ userInitial }}</span>
            <span class="hidden sm:inline">{{ t('home.dashboard') }}</span>
          </router-link>
          <template v-else>
            <router-link
              to="/login"
              class="ml-1 hidden rounded-full px-3 py-2 text-xs font-semibold text-slate-700 transition-colors hover:bg-slate-100 hover:text-slate-950 dark:text-dark-200 dark:hover:bg-dark-800 dark:hover:text-white sm:inline-flex"
            >
              {{ t('home.login') }}
            </router-link>
            <router-link
              to="/register"
              class="hidden rounded-full bg-sky-500 px-4 py-2 text-xs font-semibold text-white shadow-sm shadow-sky-500/25 transition-all hover:-translate-y-0.5 hover:bg-sky-600 md:inline-flex"
            >
              {{ t('home.register') }}
            </router-link>
          </template>

          <span class="lg:hidden">
            <button
              type="button"
              class="home-icon-button"
              :title="mobileMenuOpen ? t('home.nav.closeMenu') : t('home.nav.openMenu')"
              :aria-expanded="mobileMenuOpen"
              aria-controls="home-mobile-menu"
              @click="mobileMenuOpen = !mobileMenuOpen"
            >
              <Icon :name="mobileMenuOpen ? 'x' : 'menu'" size="md" />
            </button>
          </span>
        </div>
      </nav>

      <transition name="home-menu">
        <div id="home-mobile-menu" v-if="mobileMenuOpen" class="mx-auto max-w-7xl border-t border-slate-200/70 py-3 dark:border-dark-800 lg:hidden">
          <div class="grid gap-1 sm:grid-cols-2">
            <a href="#top" class="home-mobile-link" @click="mobileMenuOpen = false">{{ t('home.nav.home') }}</a>
            <router-link :to="dashboardPath" class="home-mobile-link" @click="mobileMenuOpen = false">{{ t('home.nav.dashboard') }}</router-link>
            <router-link to="/pricing" class="home-mobile-link" @click="mobileMenuOpen = false">{{ t('home.nav.pricing') }}</router-link>
            <a href="#integrations" class="home-mobile-link" @click="mobileMenuOpen = false">{{ t('home.nav.integrations') }}</a>
            <a
              :href="CC_SWITCH_DOWNLOAD_LINKS.officialSite"
              target="_blank"
              rel="noopener noreferrer"
              class="home-mobile-link"
              @click="mobileMenuOpen = false"
            >
              CC-Switch
            </a>
            <a
              v-if="docUrl"
              :href="docUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="home-mobile-link"
              @click="mobileMenuOpen = false"
            >
              {{ t('home.docs') }}
            </a>
          </div>
          <div v-if="!isAuthenticated" class="mt-3 grid grid-cols-2 gap-2 border-t border-slate-200/70 pt-3 dark:border-dark-800">
            <router-link to="/login" class="home-mobile-auth border border-slate-200 text-slate-700 dark:border-dark-700 dark:text-dark-200" @click="mobileMenuOpen = false">
              {{ t('home.login') }}
            </router-link>
            <router-link to="/register" class="home-mobile-auth bg-sky-500 text-white" @click="mobileMenuOpen = false">
              {{ t('home.register') }}
            </router-link>
          </div>
        </div>
      </transition>
    </header>

    <main class="relative z-10 flex-1">
      <section class="relative overflow-hidden border-b border-slate-200/70 bg-white dark:border-dark-800 dark:bg-dark-950">
        <div class="home-hero-grid pointer-events-none absolute inset-0 opacity-70 dark:opacity-20"></div>
        <div class="pointer-events-none absolute -left-32 top-0 h-[28rem] w-[28rem] rounded-full bg-sky-300/25 blur-3xl dark:bg-sky-500/10"></div>
        <div class="pointer-events-none absolute right-[-8rem] top-8 h-[32rem] w-[32rem] rounded-full bg-emerald-200/30 blur-3xl dark:bg-emerald-500/10"></div>
        <div class="pointer-events-none absolute bottom-[-12rem] left-1/3 h-[26rem] w-[26rem] rounded-full bg-violet-200/25 blur-3xl dark:bg-violet-500/10"></div>

        <div class="relative mx-auto grid max-w-7xl gap-12 px-4 pb-14 pt-16 md:px-6 md:pb-20 md:pt-20 lg:grid-cols-[0.92fr_1.08fr] lg:items-center lg:gap-14 lg:pb-24 lg:pt-24">
          <div :class="['text-center lg:text-left', isRtl ? 'lg:text-right' : '']">
            <div class="mb-6 inline-flex items-center gap-2 rounded-full border border-sky-200/80 bg-sky-50/80 px-3.5 py-2 text-xs font-semibold text-sky-700 shadow-sm dark:border-sky-900/70 dark:bg-sky-950/40 dark:text-sky-300">
              <span class="h-1.5 w-1.5 rounded-full bg-sky-500 shadow-[0_0_0_4px_rgba(14,165,233,0.12)]"></span>
              {{ t('home.heroEyebrow') }}
            </div>

            <h1 class="mx-auto max-w-3xl text-4xl font-bold leading-[1.1] tracking-[-0.035em] text-slate-950 md:text-5xl lg:mx-0 lg:text-[3.55rem] dark:text-white">
              <span class="block">{{ t('home.heroTitle') }}</span>
              <span class="mt-1 block bg-gradient-to-r from-sky-500 via-blue-500 to-violet-500 bg-clip-text text-transparent">
                {{ t('home.heroAccent') }}
              </span>
            </h1>
            <p class="mx-auto mt-6 max-w-2xl text-base leading-8 text-slate-600 md:text-lg lg:mx-0 dark:text-dark-300">
              {{ t('home.heroDescription') }}
            </p>

            <div class="mt-8 flex flex-col items-center justify-center gap-3 sm:flex-row lg:justify-start">
              <router-link
                :to="isAuthenticated ? dashboardPath : '/register'"
                class="inline-flex w-full items-center justify-center gap-2 rounded-xl bg-sky-500 px-6 py-3.5 text-sm font-bold text-white shadow-lg shadow-sky-500/20 transition-all hover:-translate-y-0.5 hover:bg-sky-600 sm:w-auto"
              >
                {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
                <Icon name="arrowRight" size="sm" />
              </router-link>
              <a
                v-if="docUrl"
                :href="docUrl"
                target="_blank"
                rel="noopener noreferrer"
                class="inline-flex w-full items-center justify-center gap-2 rounded-xl border border-slate-200 bg-white/80 px-6 py-3.5 text-sm font-semibold text-slate-700 shadow-sm transition-all hover:-translate-y-0.5 hover:border-slate-300 hover:bg-white sm:w-auto dark:border-dark-700 dark:bg-dark-900/80 dark:text-dark-200 dark:hover:border-dark-600"
              >
                <Icon name="book" size="sm" />
                {{ t('home.viewDocs') }}
              </a>
              <router-link
                v-else
                to="/pricing"
                class="inline-flex w-full items-center justify-center gap-2 rounded-xl border border-slate-200 bg-white/80 px-6 py-3.5 text-sm font-semibold text-slate-700 shadow-sm transition-all hover:-translate-y-0.5 hover:border-slate-300 hover:bg-white sm:w-auto dark:border-dark-700 dark:bg-dark-900/80 dark:text-dark-200 dark:hover:border-dark-600"
              >
                <Icon name="dollar" size="sm" />
                {{ t('home.nav.pricing') }}
              </router-link>
            </div>

            <p class="mx-auto mt-5 flex max-w-2xl items-start justify-center gap-2 text-sm leading-6 text-slate-500 lg:mx-0 lg:justify-start dark:text-dark-400">
              <Icon name="checkCircle" size="sm" class="mt-1 shrink-0 text-emerald-500" />
              {{ t('home.authPitch') }}
            </p>

            <div class="mt-8 border-t border-slate-200/80 pt-6 dark:border-dark-800">
              <p class="text-[10px] font-bold uppercase tracking-[0.22em] text-slate-400 dark:text-dark-500">
                {{ t('home.supportedAccess') }}
              </p>
              <div class="mt-3 flex flex-wrap items-center justify-center gap-2 lg:justify-start">
                <a
                v-for="mode in deploymentModes"
                :key="mode.titleKey"
                  :href="mode.titleKey === 'home.deployment.ccSwitch.title' ? CC_SWITCH_DOWNLOAD_LINKS.officialSite : '#integrations'"
                  :target="mode.titleKey === 'home.deployment.ccSwitch.title' ? '_blank' : undefined"
                  :rel="mode.titleKey === 'home.deployment.ccSwitch.title' ? 'noopener noreferrer' : undefined"
                  class="inline-flex items-center gap-2 rounded-full border border-slate-200 bg-white/70 px-3.5 py-2 text-xs font-semibold text-slate-600 shadow-sm transition-all hover:-translate-y-0.5 hover:border-sky-200 hover:text-sky-700 dark:border-dark-700 dark:bg-dark-900/70 dark:text-dark-300 dark:hover:border-sky-800 dark:hover:text-sky-300"
              >
                  <Icon :name="mode.icon" size="sm" />
                  {{ t(mode.titleKey) }}
                </a>
              </div>
            </div>
          </div>

          <div class="mx-auto w-full max-w-xl lg:max-w-none" dir="ltr">
            <div class="relative rounded-[1.5rem] border border-slate-200/90 bg-white/90 shadow-2xl shadow-slate-300/35 backdrop-blur-xl dark:border-dark-700 dark:bg-dark-900/90 dark:shadow-black/30">
              <div class="absolute -inset-4 -z-10 rounded-[2rem] bg-gradient-to-br from-sky-300/20 via-transparent to-violet-300/20 blur-2xl dark:from-sky-500/10 dark:to-violet-500/10"></div>

              <div class="flex items-center justify-between gap-3 border-b border-slate-200/80 px-3 pt-2 dark:border-dark-700">
                <div class="flex min-w-0 items-center gap-1 overflow-x-auto" role="tablist" :aria-label="t('home.apiCard.protocol')">
                  <button
                    v-for="protocol in protocolExamples"
                    :key="protocol.id"
                    type="button"
                    role="tab"
                    :aria-selected="activeProtocol === protocol.id"
                    aria-controls="home-api-demo-panel"
                    :class="[
                      'relative whitespace-nowrap px-3 py-3 text-xs font-semibold transition-colors',
                      activeProtocol === protocol.id
                        ? 'text-sky-600 dark:text-sky-300'
                        : 'text-slate-400 hover:text-slate-700 dark:text-dark-500 dark:hover:text-dark-200'
                    ]"
                    @click="activeProtocol = protocol.id"
                  >
                    {{ protocol.label }}
                    <span v-if="activeProtocol === protocol.id" class="absolute inset-x-2 bottom-0 h-0.5 rounded-full bg-sky-500"></span>
                  </button>
                </div>
                <span class="mr-3 inline-flex shrink-0 items-center gap-1.5 text-[10px] font-semibold uppercase tracking-wide text-emerald-600 dark:text-emerald-300">
                  <span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
                  DEMO · 200 OK
                </span>
              </div>

              <div class="flex items-center gap-2 border-b border-slate-200/80 bg-slate-50/80 px-5 py-3 font-mono text-[11px] text-slate-500 dark:border-dark-700 dark:bg-dark-950/60 dark:text-dark-400">
                <span class="rounded bg-violet-100 px-2 py-1 text-[9px] font-bold text-violet-700 dark:bg-violet-950/60 dark:text-violet-300">POST</span>
                <span class="truncate">{{ activeProtocolExample.path }}</span>
              </div>

              <div id="home-api-demo-panel" class="grid min-h-[26rem] md:grid-rows-2" role="tabpanel">
                <div class="border-b border-slate-200/80 p-5 dark:border-dark-700 md:p-6">
                  <div class="mb-3 flex items-center justify-between gap-3">
                    <p class="text-[10px] font-bold uppercase tracking-[0.18em] text-slate-400 dark:text-dark-500">{{ t('home.apiDemo.request') }}</p>
                    <span class="font-mono text-[10px] text-slate-400">curl</span>
                  </div>
                  <pre class="home-code-block"><code>{{ activeProtocolExample.request }}</code></pre>
                </div>
                <div class="p-5 md:p-6">
                  <div class="mb-3 flex items-center justify-between gap-3">
                    <p class="text-[10px] font-bold uppercase tracking-[0.18em] text-slate-400 dark:text-dark-500">{{ t('home.apiDemo.response') }}</p>
                    <span class="inline-flex items-center gap-1 text-[10px] font-semibold text-emerald-600 dark:text-emerald-300">
                      <span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
                      {{ t('home.apiDemo.routed') }}
                    </span>
                  </div>
                  <pre class="home-code-block"><code>{{ activeProtocolExample.response }}</code></pre>
                </div>
              </div>

              <div class="flex flex-wrap items-center gap-x-5 gap-y-2 border-t border-slate-200/80 bg-slate-50/80 px-5 py-3 font-mono text-[10px] uppercase tracking-wide text-slate-400 dark:border-dark-700 dark:bg-dark-950/60 dark:text-dark-500">
                <span>demo latency · 96 ms</span>
                <span>27 tokens</span>
                <span class="inline-flex items-center gap-1.5"><span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>stream · sse</span>
              </div>
            </div>
          </div>
        </div>

        <div class="relative border-t border-slate-200/70 bg-white/55 px-4 py-7 backdrop-blur-sm dark:border-dark-800 dark:bg-dark-950/45 md:px-6">
          <div class="mx-auto grid max-w-7xl grid-cols-2 gap-y-7 divide-slate-200 dark:divide-dark-800 md:grid-cols-4 md:divide-x">
            <div v-for="stat in heroStats" :key="stat.labelKey" class="px-3 text-center md:px-6">
              <p class="text-2xl font-bold tracking-tight text-slate-950 md:text-3xl dark:text-white">{{ stat.value }}</p>
              <p class="mt-1.5 text-xs font-medium text-slate-500 dark:text-dark-400">{{ t(stat.labelKey) }}</p>
            </div>
          </div>
        </div>
      </section>

      <section id="integrations" class="scroll-mt-20 bg-white px-4 py-16 dark:bg-dark-950 md:px-6 md:py-24">
        <div class="mx-auto max-w-7xl">
          <div class="mx-auto max-w-3xl text-center">
            <div class="inline-flex items-center gap-2 rounded-full bg-sky-50 px-3 py-1.5 text-xs font-semibold text-sky-700 dark:bg-sky-950/40 dark:text-sky-300">
              <Icon name="terminal" size="sm" />
              {{ t('home.integrations.eyebrow') }}
            </div>
            <h2 class="mt-4 text-3xl font-bold tracking-tight text-slate-950 dark:text-white md:text-4xl">
              {{ t('home.integrations.title') }}
            </h2>
            <p class="mt-3 text-sm leading-7 text-slate-500 dark:text-dark-400">
              {{ t('home.integrations.description') }}
            </p>
          </div>

          <div class="mt-10 grid gap-5 lg:grid-cols-[1.08fr_0.92fr]">
            <article class="relative overflow-hidden rounded-[1.75rem] bg-slate-950 p-6 text-white shadow-xl shadow-slate-300/30 dark:border dark:border-dark-700 dark:shadow-black/20 md:p-8">
              <div class="pointer-events-none absolute -right-16 -top-16 h-64 w-64 rounded-full bg-sky-500/25 blur-3xl"></div>
              <div class="pointer-events-none absolute -bottom-24 left-1/4 h-64 w-64 rounded-full bg-violet-500/20 blur-3xl"></div>

              <div class="relative">
                <div class="inline-flex items-center gap-2 rounded-full border border-sky-400/25 bg-sky-400/10 px-3 py-1.5 text-xs font-semibold text-sky-200">
                  <Icon name="bolt" size="sm" />
                  {{ t('home.ccSwitch.badge') }}
                </div>
                <h3 class="mt-5 max-w-xl text-2xl font-bold leading-tight md:text-3xl">{{ t('home.ccSwitch.title') }}</h3>
                <p class="mt-4 max-w-2xl text-sm leading-7 text-slate-300">{{ t('home.ccSwitch.description') }}</p>

                <div class="mt-7 flex flex-col gap-3 sm:flex-row">
                  <router-link
                    to="/keys"
                    class="inline-flex items-center justify-center gap-2 rounded-xl bg-sky-400 px-5 py-3 text-sm font-bold text-slate-950 transition-all hover:-translate-y-0.5 hover:bg-sky-300"
                  >
                    <Icon name="bolt" size="sm" />
                    {{ t('home.ccSwitch.primaryAction') }}
                  </router-link>
                  <router-link
                    to="/cc-switch"
                    class="inline-flex items-center justify-center gap-2 rounded-xl border border-white/15 bg-white/5 px-5 py-3 text-sm font-semibold text-white transition-colors hover:border-white/30 hover:bg-white/10"
                  >
                    {{ t('home.ccSwitch.guideAction') }}
                    <Icon name="arrowRight" size="sm" />
                  </router-link>
                </div>

                <div class="mt-8 grid grid-cols-2 gap-3">
                  <div
                    v-for="mode in deploymentModes"
                    :key="mode.titleKey"
                    class="flex items-center gap-3 rounded-xl border border-white/10 bg-white/[0.06] p-3.5"
                  >
                    <span :class="['flex h-9 w-9 shrink-0 items-center justify-center rounded-lg', mode.iconClass]">
                      <Icon :name="mode.icon" size="sm" />
                    </span>
                    <span class="min-w-0">
                      <span class="block truncate text-xs font-bold text-white">{{ t(mode.titleKey) }}</span>
                      <span class="mt-0.5 block truncate text-[10px] text-slate-400 sm:text-[11px]">{{ t(mode.descriptionKey) }}</span>
                    </span>
                  </div>
                </div>
              </div>
            </article>

            <div class="grid gap-4 sm:grid-cols-2">
              <article
                v-for="client in integrationClients"
                :key="client.name"
                class="group rounded-2xl border border-slate-200 bg-slate-50/70 p-5 transition-all hover:-translate-y-1 hover:border-sky-300 hover:bg-white hover:shadow-lg dark:border-dark-800 dark:bg-dark-900/70 dark:hover:border-sky-700 dark:hover:bg-dark-900"
              >
                <div class="flex items-start justify-between gap-3">
                  <div :class="['flex h-12 w-12 items-center justify-center rounded-xl', client.cardClass]">
                    <ModelIcon v-if="'model' in client" :model="client.model" size="29px" />
                    <Icon v-else :name="client.icon" size="lg" />
                  </div>
                  <span class="mt-1 inline-flex items-center gap-1 text-[10px] font-semibold text-emerald-600 dark:text-emerald-300">
                    <span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
                    {{ t('home.ccSwitch.ready') }}
                  </span>
                </div>
                <h3 class="mt-5 text-base font-bold text-slate-950 dark:text-white">{{ client.name }}</h3>
                <p class="mt-2 text-sm leading-6 text-slate-500 dark:text-dark-400">{{ t(client.descriptionKey) }}</p>
              </article>
            </div>
          </div>
        </div>
      </section>

      <section id="workflow" class="scroll-mt-20 border-y border-slate-200 bg-slate-50 px-4 py-16 dark:border-dark-800 dark:bg-dark-900/40 md:px-6 md:py-24">
        <div class="mx-auto max-w-7xl">
          <div class="mx-auto max-w-3xl text-center">
            <p class="text-xs font-bold uppercase tracking-[0.2em] text-sky-600 dark:text-sky-300">{{ t('home.workflow.eyebrow') }}</p>
            <h2 class="mt-4 text-3xl font-bold tracking-tight text-slate-950 dark:text-white md:text-4xl">{{ t('home.workflow.title') }}</h2>
            <p class="mt-4 text-sm leading-7 text-slate-500 dark:text-dark-400">{{ t('home.workflow.description') }}</p>
          </div>

          <div class="relative mt-12 grid gap-5 md:grid-cols-3">
            <div class="pointer-events-none absolute left-[16.66%] right-[16.66%] top-9 hidden h-px bg-gradient-to-r from-sky-200 via-violet-200 to-emerald-200 dark:from-sky-900 dark:via-violet-900 dark:to-emerald-900 md:block"></div>
            <article v-for="(step, index) in workflowSteps" :key="step.titleKey" class="relative rounded-2xl border border-slate-200 bg-white p-6 shadow-sm dark:border-dark-800 dark:bg-dark-900 md:p-7">
              <div class="flex items-center justify-between gap-4">
                <span :class="['relative z-10 flex h-14 w-14 items-center justify-center rounded-2xl ring-8 ring-slate-50 dark:ring-dark-900/40', step.iconClass]">
                  <Icon :name="step.icon" size="lg" />
                </span>
                <span class="font-mono text-3xl font-bold text-slate-100 dark:text-dark-700">0{{ index + 1 }}</span>
              </div>
              <h3 class="mt-6 text-lg font-bold text-slate-950 dark:text-white">{{ t(step.titleKey) }}</h3>
              <p class="mt-3 text-sm leading-7 text-slate-500 dark:text-dark-400">{{ t(step.descriptionKey) }}</p>
            </article>
          </div>
        </div>
      </section>

      <section id="pricing-preview" class="scroll-mt-20 bg-white px-4 py-16 dark:bg-dark-950 md:px-6 md:py-20">
        <div class="mx-auto max-w-7xl">
          <div class="grid gap-6 lg:grid-cols-[0.95fr_1.05fr] lg:items-center">
            <div>
              <p class="text-xs font-semibold uppercase tracking-wide text-emerald-600 dark:text-emerald-300">
                {{ t('home.pricing.eyebrow') }}
              </p>
              <h2 class="mt-3 text-3xl font-bold tracking-tight text-slate-950 dark:text-white md:text-4xl">
                {{ t('home.pricing.title') }}
              </h2>
              <p class="mt-3 max-w-2xl text-sm leading-6 text-slate-500 dark:text-dark-400">
                {{ t('home.pricing.description') }}
              </p>
              <router-link
                to="/pricing"
                class="mt-4 inline-flex items-center gap-1.5 text-sm font-semibold text-emerald-600 transition-colors hover:text-emerald-700 dark:text-emerald-300 dark:hover:text-emerald-200"
              >
                {{ t('home.pricing.fullPricingAction') }}
                <Icon name="arrowRight" size="sm" />
              </router-link>
            </div>
            <div class="grid gap-3 sm:grid-cols-2">
              <div class="rounded-2xl border border-emerald-200 bg-emerald-50 p-5 dark:border-emerald-900/60 dark:bg-emerald-950/20">
                <div class="flex items-center gap-2 text-sm font-semibold text-emerald-800 dark:text-emerald-200">
                  <Icon name="creditCard" size="sm" />
                  {{ t('home.pricing.rechargeLabel') }}
                </div>
                <p class="mt-3 text-2xl font-bold text-slate-950 dark:text-white">
                  {{ t('home.pricing.rechargeValue', { usd: formatCompactNumber(balanceRechargeMultiplier) }) }}
                </p>
                <p class="mt-2 text-xs leading-5 text-emerald-800/80 dark:text-emerald-200/80">
                  {{ t('home.pricing.rechargeHint') }}
                </p>
              </div>
              <div class="rounded-2xl border border-sky-200 bg-sky-50 p-5 dark:border-sky-900/60 dark:bg-sky-950/20">
                <div class="flex items-center gap-2 text-sm font-semibold text-sky-800 dark:text-sky-200">
                  <Icon name="calculator" size="sm" />
                  {{ t('home.pricing.tokenLabel') }}
                </div>
                <p class="mt-3 text-2xl font-bold text-slate-950 dark:text-white">
                  {{ t('home.pricing.tokenValue') }}
                </p>
                <p class="mt-2 text-xs leading-5 text-sky-800/80 dark:text-sky-200/80">
                  {{ t('home.pricing.tokenHint') }}
                </p>
              </div>
            </div>
          </div>
          <ModelPriceComparison :usd-per-cny="balanceRechargeMultiplier" class="mt-8" />
        </div>
      </section>

      <section class="border-y border-slate-200 bg-slate-50 px-4 py-16 dark:border-dark-800 dark:bg-dark-900/40 md:px-6 md:py-24">
        <div class="mx-auto max-w-7xl">
          <div class="mx-auto max-w-2xl text-center">
            <h2 class="text-3xl font-bold tracking-tight text-slate-950 dark:text-white md:text-4xl">
              {{ t('home.providers.title') }}
            </h2>
            <p class="mt-3 text-sm leading-7 text-slate-500 dark:text-dark-400">
              {{ t('home.providers.description') }}
            </p>
          </div>

          <div class="mt-9 grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-6">
            <div
              v-for="provider in providerLogos"
              :key="provider.labelKey"
              class="group flex min-h-28 flex-col items-center justify-center gap-3 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm transition-all hover:-translate-y-1 hover:border-sky-300 hover:shadow-lg dark:border-dark-800 dark:bg-dark-900 dark:hover:border-sky-700"
              :title="t(provider.labelKey)"
            >
              <div class="flex h-14 w-14 items-center justify-center rounded-xl bg-slate-50 text-slate-900 transition-colors group-hover:bg-sky-50 dark:bg-dark-950 dark:text-white dark:group-hover:bg-sky-950/40">
                <ModelIcon :model="provider.model" size="34px" />
              </div>
              <span class="max-w-full truncate text-sm font-semibold text-slate-700 dark:text-dark-200">
                {{ t(provider.labelKey) }}
              </span>
            </div>
          </div>
        </div>
      </section>
    </main>

    <footer class="relative z-10 bg-white px-4 py-5 dark:bg-dark-950" dir="ltr">
      <div class="mx-auto flex max-w-6xl flex-col items-center justify-center gap-2 text-center text-sm text-slate-500 dark:text-dark-400">
        <img :src="companyIconUrl" alt="ISACAI" class="h-10 w-10 rounded-lg object-contain" />
        <p>© 2026 ISACAI. All rights reserved. 联系请加入 QQ 群：1027890648</p>
        <div class="flex flex-wrap items-center justify-center gap-4">
          <a
            href="https://beian.miit.gov.cn"
            target="_blank"
            rel="noopener noreferrer"
            class="transition-colors hover:text-slate-800 dark:hover:text-white"
          >
            粤ICP备2026050877号
          </a>
          <router-link
            v-if="publicStatusEnabled"
            to="/status"
            class="inline-flex items-center gap-1.5 transition-colors hover:text-slate-800 dark:hover:text-white"
          >
            <span class="inline-flex h-2 w-2 rounded-full bg-emerald-500"></span>
            {{ t('publicStatus.title') }}
          </router-link>
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="transition-colors hover:text-slate-800 dark:hover:text-white"
          >
            {{ t('home.docs') }}
          </a>
        </div>
      </div>
    </footer>

    <transition name="notice-fade">
      <div
        v-if="showNotice"
        class="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-slate-950/60 px-4 py-8 backdrop-blur-sm sm:items-center"
        @keydown.esc="closeNoticeForSession"
      >
        <section
          ref="noticeDialog"
          role="dialog"
          aria-modal="true"
          aria-labelledby="home-notice-title"
          tabindex="-1"
          class="relative my-auto max-h-[calc(100vh-4rem)] w-full max-w-4xl overflow-y-auto rounded-lg border border-slate-200 bg-white shadow-2xl outline-none dark:border-dark-700 dark:bg-dark-900"
        >
          <div class="flex items-center justify-between border-b border-slate-200 px-5 py-4 dark:border-dark-700">
            <h2 id="home-notice-title" class="text-xl font-bold text-slate-950 dark:text-white">
              {{ t('home.notice.title') }}
            </h2>
            <div class="flex items-center gap-3">
              <span class="hidden items-center gap-1 rounded-lg bg-sky-50 px-3 py-2 text-xs font-semibold text-sky-700 dark:bg-sky-950/40 dark:text-sky-300 sm:inline-flex">
                <Icon name="bell" size="sm" />
                {{ t('home.notice.tabNotice') }}
              </span>
              <span class="hidden items-center gap-1 text-xs font-medium text-slate-500 dark:text-dark-400 sm:inline-flex">
                <Icon name="infoCircle" size="sm" />
                {{ t('home.notice.tabSystem') }}
              </span>
              <button
                @click="closeNoticeForSession"
                class="rounded-lg p-2 text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-900 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
                :title="t('common.close')"
              >
                <Icon name="x" size="md" />
              </button>
            </div>
          </div>

          <div class="grid gap-6 px-5 py-6 md:grid-cols-[1fr_190px] md:items-end">
            <div class="space-y-3 text-sm leading-7 text-slate-700 dark:text-dark-200">
              <p>
                <span class="font-semibold text-slate-950 dark:text-white">{{ t('home.notice.important') }}</span>
                {{ t('home.notice.qqGroupLabel') }}
                <span class="font-semibold text-red-500">1027890648</span>
                {{ t('home.notice.qqGroupSuffix') }}
              </p>
              <p>{{ t('home.notice.trust') }}</p>
              <p>
                {{ t('home.notice.status') }}
                <a
                  href="/"
                  class="font-medium text-sky-600 hover:text-sky-700 dark:text-sky-300 dark:hover:text-sky-200"
                >
                  {{ siteName }}
                </a>
              </p>
            </div>

            <div class="hidden justify-center md:flex">
              <div class="flex h-36 w-36 items-center justify-center rounded-lg border border-slate-200 bg-slate-50 p-4 dark:border-dark-700 dark:bg-dark-950">
                <img :src="companyIconUrl" alt="ISACAI" class="max-h-full max-w-full object-contain" />
              </div>
            </div>
          </div>

          <div class="flex flex-wrap items-center justify-end gap-3 border-t border-slate-200 px-5 py-4 dark:border-dark-700">
            <button
              @click="closeNoticeToday"
              class="rounded-lg bg-slate-100 px-4 py-2 text-sm font-semibold text-slate-700 transition-colors hover:bg-slate-200 dark:bg-dark-800 dark:text-dark-200 dark:hover:bg-dark-700"
            >
              {{ t('home.notice.closeToday') }}
            </button>
            <button
              @click="closeNoticeForSession"
              class="rounded-lg bg-sky-600 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-sky-700"
            >
              {{ t('home.notice.close') }}
            </button>
          </div>
        </section>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import ModelIcon from '@/components/common/ModelIcon.vue'
import ModelPriceComparison from '@/components/common/ModelPriceComparison.vue'
import Icon from '@/components/icons/Icon.vue'
import { FeatureFlags, isFeatureFlagEnabled } from '@/utils/featureFlags'
import { CC_SWITCH_DOWNLOAD_LINKS } from '@/utils/ccswitchImport'
import { formatCompactNumber, normalizeRechargeUsdPerCny } from '@/utils/pricing'
import { sanitizeUrl } from '@/utils/url'

const { t, locale } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

// Site settings - directly from appStore (already initialized from injected config)
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const companyIconUrl = '/logo.png'
const balanceRechargeMultiplier = computed(() =>
  normalizeRechargeUsdPerCny(appStore.cachedPublicSettings?.balance_recharge_multiplier),
)

const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isDark = ref(document.documentElement.classList.contains('dark'))
const isRtl = computed(() => locale.value === 'ar')
const showNotice = ref(false)
const mobileMenuOpen = ref(false)
const noticeDialog = ref<HTMLElement | null>(null)

const protocolExamples = [
  {
    id: 'chat',
    label: 'Chat',
    path: '/v1/chat/completions',
    request: `curl -X POST https://isacai.space/v1/chat/completions
  -H "Authorization: Bearer sk-••••"
  -H "Content-Type: application/json"
  -d '{ "model": "gpt-5.6-sol",
        "messages": [{ "role": "user", "content": "Hello" }] }'`,
    response: `{
  "choices": [{
    "message": { "content": "Hello from ISACAI." }
  }],
  "usage": { "total_tokens": 27 }
}`
  },
  {
    id: 'responses',
    label: 'Responses',
    path: '/v1/responses',
    request: `curl -X POST https://isacai.space/v1/responses
  -H "Authorization: Bearer sk-••••"
  -H "Content-Type: application/json"
  -d '{ "model": "gpt-5.5",
        "input": "Explain this repository" }'`,
    response: `{
  "status": "completed",
  "output_text": "Repository analysis ready.",
  "usage": { "total_tokens": 31 }
}`
  },
  {
    id: 'claude',
    label: 'Claude',
    path: '/v1/messages',
    request: `curl -X POST https://isacai.space/v1/messages
  -H "x-api-key: sk-••••"
  -H "Content-Type: application/json"
  -d '{ "model": "claude-sonnet-4-6",
        "messages": [{ "role": "user", "content": "Hello" }] }'`,
    response: `{
  "type": "message",
  "content": [{ "type": "text", "text": "Hello." }],
  "stop_reason": "end_turn"
}`
  },
  {
    id: 'gemini',
    label: 'Gemini',
    path: '/v1beta/models/gemini-2.5-pro:generateContent',
    request: `curl -X POST https://isacai.space/v1beta/models/gemini-2.5-pro:generateContent
  -H "x-goog-api-key: sk-••••"
  -H "Content-Type: application/json"
  -d '{ "contents": [{ "parts": [{ "text": "Hello" }] }] }'`,
    response: `{
  "candidates": [{
    "content": { "parts": [{ "text": "Hello." }] }
  }],
  "usageMetadata": { "totalTokenCount": 25 }
}`
  }
] as const

type ProtocolId = (typeof protocolExamples)[number]['id']
const activeProtocol = ref<ProtocolId>('chat')
const activeProtocolExample = computed(
  () => protocolExamples.find((protocol) => protocol.id === activeProtocol.value) || protocolExamples[0]
)

const heroStats = computed(() => [
  { value: `1 : ${formatCompactNumber(balanceRechargeMultiplier.value)}`, labelKey: 'home.heroStats.recharge' },
  { value: '×1', labelKey: 'home.heroStats.groupRate' },
  { value: '7', labelKey: 'home.heroStats.models' },
  { value: '4', labelKey: 'home.heroStats.deployments' }
] as const)

const workflowSteps = [
  {
    icon: 'userPlus',
    iconClass: 'bg-sky-50 text-sky-600 dark:bg-sky-950/50 dark:text-sky-300',
    titleKey: 'home.workflow.steps.account.title',
    descriptionKey: 'home.workflow.steps.account.description'
  },
  {
    icon: 'bolt',
    iconClass: 'bg-violet-50 text-violet-600 dark:bg-violet-950/50 dark:text-violet-300',
    titleKey: 'home.workflow.steps.configure.title',
    descriptionKey: 'home.workflow.steps.configure.description'
  },
  {
    icon: 'terminal',
    iconClass: 'bg-emerald-50 text-emerald-600 dark:bg-emerald-950/50 dark:text-emerald-300',
    titleKey: 'home.workflow.steps.connect.title',
    descriptionKey: 'home.workflow.steps.connect.description'
  }
] as const

const NOTICE_VERSION = '2026-07-pricing-v2'
const NOTICE_SESSION_KEY = `isacai_home_notice_session_${NOTICE_VERSION}`
const NOTICE_DATE_KEY = `isacai_home_notice_date_${NOTICE_VERSION}`

const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))
const userInitial = computed(() => {
  const user = authStore.user
  if (!user || !user.email) return 'I'
  return user.email.charAt(0).toUpperCase()
})

const integrationClients = [
  {
    name: 'Codex',
    model: 'gpt-4.1',
    icon: 'terminal',
    panelClass: 'bg-emerald-400/15 text-emerald-300',
    cardClass: 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300',
    descriptionKey: 'home.integrations.codexDescription'
  },
  {
    name: 'Claude Code',
    model: 'claude-3-5-sonnet',
    icon: 'terminal',
    panelClass: 'bg-orange-400/15 text-orange-300',
    cardClass: 'bg-orange-50 text-orange-700 dark:bg-orange-950/40 dark:text-orange-300',
    descriptionKey: 'home.integrations.claudeCodeDescription'
  },
  {
    name: 'Gemini CLI',
    model: 'gemini-2.5-pro',
    icon: 'sparkles',
    panelClass: 'bg-blue-400/15 text-blue-300',
    cardClass: 'bg-blue-50 text-blue-700 dark:bg-blue-950/40 dark:text-blue-300',
    descriptionKey: 'home.integrations.geminiDescription'
  },
  {
    name: 'VS Code / Cursor / IDE',
    panelName: 'IDE / Terminal',
    icon: 'grid',
    panelClass: 'bg-violet-400/15 text-violet-300',
    cardClass: 'bg-violet-50 text-violet-700 dark:bg-violet-950/40 dark:text-violet-300',
    descriptionKey: 'home.integrations.ideDescription'
  }
] as const

const deploymentModes = [
  {
    titleKey: 'home.deployment.ccSwitch.title',
    descriptionKey: 'home.deployment.ccSwitch.description',
    icon: 'bolt',
    iconClass: 'bg-sky-400/15 text-sky-300',
  },
  {
    titleKey: 'home.deployment.terminal.title',
    descriptionKey: 'home.deployment.terminal.description',
    icon: 'terminal',
    iconClass: 'bg-violet-400/15 text-violet-300',
  },
  {
    titleKey: 'home.deployment.server.title',
    descriptionKey: 'home.deployment.server.description',
    icon: 'server',
    iconClass: 'bg-amber-400/15 text-amber-300',
  },
  {
    titleKey: 'home.deployment.client.title',
    descriptionKey: 'home.deployment.client.description',
    icon: 'monitor',
    iconClass: 'bg-emerald-400/15 text-emerald-300',
  },
] as const

const providerLogos = [
  { model: 'gpt-4.1', labelKey: 'home.providers.items.openai' },
  { model: 'claude-3-5-sonnet', labelKey: 'home.providers.items.claude' },
  { model: 'gemini-2.5-pro', labelKey: 'home.providers.items.gemini' },
  { model: 'grok-4', labelKey: 'home.providers.items.xai' },
  { model: 'deepseek-chat', labelKey: 'home.providers.items.deepseek' },
  { model: 'glm-4-plus', labelKey: 'home.providers.items.zhipu' }
] as const

const publicStatusEnabled = computed(() => isFeatureFlagEnabled(FeatureFlags.publicStatus))

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

function todayKey() {
  const now = new Date()
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}`
}

function focusNotice() {
  nextTick(() => noticeDialog.value?.focus())
}

function initNotice() {
  const closedForSession = sessionStorage.getItem(NOTICE_SESSION_KEY) === '1'
  const closedToday = localStorage.getItem(NOTICE_DATE_KEY) === todayKey()
  showNotice.value = !closedForSession && !closedToday
  if (showNotice.value) focusNotice()
}

function closeNoticeToday() {
  localStorage.setItem(NOTICE_DATE_KEY, todayKey())
  showNotice.value = false
}

function openNotice() {
  mobileMenuOpen.value = false
  showNotice.value = true
  focusNotice()
}

function closeNoticeForSession() {
  sessionStorage.setItem(NOTICE_SESSION_KEY, '1')
  showNotice.value = false
}

onMounted(() => {
  initTheme()
  initNotice()

  authStore.checkAuth()

  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
.home-nav-link {
  @apply rounded-full px-3 py-2 text-xs font-medium text-slate-600 transition-colors hover:bg-slate-100 hover:text-slate-950 dark:text-dark-300 dark:hover:bg-dark-800 dark:hover:text-white;
}

.home-nav-link-active {
  @apply bg-sky-50 font-semibold text-sky-700 dark:bg-sky-950/40 dark:text-sky-300;
}

.home-icon-button {
  @apply inline-flex h-9 w-9 items-center justify-center rounded-full text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-900 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white;
}

.home-mobile-link {
  @apply rounded-xl px-3 py-2.5 text-sm font-medium text-slate-600 transition-colors hover:bg-slate-100 hover:text-slate-950 dark:text-dark-300 dark:hover:bg-dark-800 dark:hover:text-white;
}

.home-mobile-auth {
  @apply inline-flex items-center justify-center rounded-xl px-4 py-2.5 text-sm font-semibold;
}

.home-code-block {
  @apply m-0 whitespace-pre-wrap break-words font-mono text-[11px] leading-6 text-slate-600 dark:text-dark-300;
  overflow-wrap: anywhere;
}

.home-hero-grid {
  background-image:
    linear-gradient(rgba(148, 163, 184, 0.08) 1px, transparent 1px),
    linear-gradient(90deg, rgba(148, 163, 184, 0.08) 1px, transparent 1px);
  background-size: 48px 48px;
  mask-image: linear-gradient(to bottom, black, transparent 78%);
}

.home-menu-enter-active,
.home-menu-leave-active {
  transition: opacity 160ms ease, transform 160ms ease;
}

.home-menu-enter-from,
.home-menu-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

.notice-fade-enter-active,
.notice-fade-leave-active {
  transition: opacity 180ms ease;
}

.notice-fade-enter-from,
.notice-fade-leave-to {
  opacity: 0;
}
</style>
