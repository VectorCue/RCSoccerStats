<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from 'axios'
import GameMasterMatches from '@/components/matches/GameMasterMatches.vue'
import StandingsTable from '@/components/standings/StandingsTable.vue'
import {
  exactLeagueParams,
  exactLeagues,
  groupMatches,
  matchParams,
  resolveSelection,
  tournamentTime,
} from '@/lib/gameMaster'
import type {
  ExactLeague,
  GameMasterMatch,
  GameMasterStandings,
  LeaguesResponse,
  MatchesResponse,
  Selection,
} from '@/types/gameMaster'

const props = defineProps<{ view: 'matches' | 'standings' }>()
const route = useRoute()
const router = useRouter()

const catalog = ref<LeaguesResponse | null>(null)
const matches = ref<GameMasterMatch[]>([])
const tables = ref<{ league: ExactLeague; data: GameMasterStandings | null; error: string }[]>([])
const error = ref('')
const loading = ref(true)
const refreshing = ref(false)
const updated = ref('')
let controller: AbortController | undefined
let refreshInterval: number | undefined

function queryValue(key: string) {
  const value = route.query[key]
  return (Array.isArray(value) ? value[0] : value) || ''
}

const selection = computed<Selection>(() => ({
  main: queryValue('main_league'),
  sub: queryValue('sub_league'),
  league: queryValue('league'),
  stage: queryValue('league_stage'),
}))

const resolved = computed(() => {
  try {
    return resolveSelection(catalog.value?.main_leagues ?? [], selection.value)
  } catch {
    return undefined
  }
})
const exact = computed(() => resolved.value?.exact)
const leagueOptions = computed(() => {
  const leagues = exactLeagues(catalog.value?.main_leagues ?? []).filter(
    (league) => !selection.value.main || league.main.abbrev === selection.value.main,
  )
  if (exact.value && !leagues.some((league) => league.abbrev === exact.value?.abbrev))
    leagues.push(exact.value)
  return leagues
})
const matchGroups = computed(() => groupMatches(matches.value))
const title = computed(() => (props.view === 'matches' ? 'Matches' : 'Standings'))

function selectMain(value: string) {
  void router.replace({
    query: {
      ...route.query,
      main_league: value || undefined,
      sub_league: undefined,
      league: undefined,
      league_stage: undefined,
    },
  })
}

function selectLeague(value: string) {
  const league = leagueOptions.value.find((item) => item.abbrev === value)
  void router.replace({
    query: {
      ...route.query,
      main_league: league?.main.abbrev || selection.value.main || undefined,
      sub_league: league?.isSubLeague ? league.abbrev : undefined,
      league: league && !league.isSubLeague ? league.abbrev : undefined,
      league_stage: undefined,
    },
  })
}

function selectStage(value: string) {
  void router.replace({ query: { ...route.query, league_stage: value || undefined } })
}

const mainSelection = computed({ get: () => selection.value.main, set: selectMain })
const leagueSelection = computed({
  get: () => selection.value.sub || selection.value.league,
  set: selectLeague,
})
const stageSelection = computed({ get: () => selection.value.stage, set: selectStage })

function errorMessage(reason: unknown): string {
  if (axios.isAxiosError(reason)) {
    const message = reason.response?.data?.error
    return typeof message === 'string'
      ? message
      : 'Could not load tournament data. Please try again.'
  }
  return reason instanceof Error ? reason.message : 'Could not load tournament data.'
}

async function refresh(clear = false) {
  controller?.abort()
  const request = new AbortController()
  controller = request
  const signal = request.signal
  const currentSelection = { ...selection.value }
  const currentView = props.view
  refreshing.value = true
  error.value = ''
  if (clear) {
    loading.value = true
    matches.value = []
    tables.value = []
    updated.value = ''
  }
  const options = { signal, timeout: 15000 }
  try {
    const { data } = await axios.get<LeaguesResponse>('/api/v2/leagues', options)
    if (signal.aborted) return
    if (!Array.isArray(data.main_leagues))
      throw new Error('The tournament API did not return a main-league catalog.')
    catalog.value = data
    const scope = resolveSelection(data.main_leagues, currentSelection)
    if (currentView === 'matches') {
      const response = await axios.get<MatchesResponse>('/api/v2/matches', {
        ...options,
        params: matchParams(currentSelection),
      })
      if (!signal.aborted) matches.value = response.data.matches
    } else {
      // Rankings cannot be merged: request one exact league at a time, including
      // the root's own table when it has stages, and keep empty child leagues.
      const results = await Promise.all(
        scope.leagues.map(async (league) => {
          try {
            const response = await axios.get<GameMasterStandings>('/api/v2/standings', {
              ...options,
              params: exactLeagueParams(league, currentSelection.stage),
            })
            return {
              league,
              data: {
                ...response.data,
                standings: [...response.data.standings].sort(
                  (a, b) => a.group - b.group || a.rank - b.rank,
                ),
              },
              error: '',
            }
          } catch (reason) {
            return { league, data: null, error: errorMessage(reason) }
          }
        }),
      )
      if (!signal.aborted) tables.value = results
    }
    if (!signal.aborted) updated.value = new Date().toLocaleTimeString()
  } catch (reason) {
    if (!signal.aborted) error.value = errorMessage(reason)
  } finally {
    if (!signal.aborted) {
      loading.value = false
      refreshing.value = false
    }
  }
}

watch(
  () => route.fullPath,
  () => {
    void refresh(true)
  },
)
onMounted(() => {
  void refresh(true)
  refreshInterval = window.setInterval(() => {
    if (!refreshing.value) void refresh()
  }, 20000)
})
onUnmounted(() => {
  controller?.abort()
  window.clearInterval(refreshInterval)
})
</script>

<template>
  <div class="min-h-screen bg-gray-50 px-4 py-6 text-gray-900 sm:px-6 lg:px-10">
    <div class="mx-auto max-w-screen-2xl space-y-6">
      <header class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <RouterLink to="/" class="text-sm font-semibold text-blue-700"
            >RCJV Soccer Stats</RouterLink
          >
          <h1 class="mt-2 text-3xl font-bold">{{ title }}</h1>
          <p v-if="catalog" class="mt-1 text-gray-600">{{ catalog.tournament_name }}</p>
        </div>
        <nav
          aria-label="Tournament views"
          class="flex gap-1 rounded-lg border border-gray-200 bg-white p-1"
        >
          <RouterLink
            v-for="tab in ['matches', 'standings']"
            :key="tab"
            :to="{ path: `/v2/${tab}`, query: route.query }"
            :aria-current="view === tab ? 'page' : undefined"
            class="rounded-md px-4 py-2 text-sm font-semibold capitalize"
            :class="view === tab ? 'bg-blue-700 text-white' : 'text-gray-600 hover:bg-gray-100'"
            >{{ tab }}</RouterLink
          >
        </nav>
      </header>

      <section
        aria-label="League filters"
        class="flex flex-wrap items-end gap-4 rounded-xl border border-gray-200 bg-white p-5 shadow-sm"
      >
        <div class="min-w-48 flex-1 text-sm font-medium">
          <label for="v2-main-league">Main league</label>
          <select
            id="v2-main-league"
            v-model="mainSelection"
            class="mt-2 block w-full rounded-md border border-gray-300 bg-white px-3 py-2.5 text-base"
          >
            <option value="">All main leagues</option>
            <option v-for="main in catalog?.main_leagues" :key="main.abbrev" :value="main.abbrev">
              {{ main.name }}
            </option>
          </select>
        </div>
        <div class="min-w-48 flex-1 text-sm font-medium">
          <label for="v2-sub-league">Sub-league</label>
          <select
            id="v2-sub-league"
            v-model="leagueSelection"
            class="mt-2 block w-full rounded-md border border-gray-300 bg-white px-3 py-2.5 text-base"
          >
            <option value="">All sub-leagues</option>
            <option v-for="league in leagueOptions" :key="league.abbrev" :value="league.abbrev">
              {{ !selection.main && league.isSubLeague ? `${league.main.name} / ` : ''
              }}{{ league.name }}{{ league.isSubLeague ? '' : ' (main league only)' }}
            </option>
          </select>
        </div>
        <div class="min-w-48 flex-1 text-sm font-medium">
          <label for="v2-stage">Stage</label>
          <select
            id="v2-stage"
            v-model="stageSelection"
            :disabled="!exact"
            class="mt-2 block w-full rounded-md border border-gray-300 bg-white px-3 py-2.5 text-base disabled:bg-gray-100 disabled:text-gray-500"
          >
            <option value="">
              {{ view === 'matches' ? 'All stages' : 'Seeding (cumulative)' }}
            </option>
            <option v-for="(stage, index) in exact?.stages" :key="index" :value="String(index)">
              {{ stage.name }}
            </option>
          </select>
        </div>
        <button
          type="button"
          class="rounded-md bg-blue-700 px-5 py-2.5 font-semibold text-white hover:bg-blue-800 disabled:opacity-50"
          :disabled="refreshing"
          @click="refresh()"
        >
          {{ refreshing ? 'Refreshing…' : 'Refresh' }}
        </button>
      </section>

      <div class="flex flex-wrap justify-between gap-2 text-sm text-gray-500">
        <p>
          <template v-if="catalog?.tournament_publishing_timezone"
            >Schedule timezone: {{ catalog.tournament_publishing_timezone }} ·
          </template>
          Updates every 20 seconds<span v-if="updated">
            · Refreshed {{ updated }} (local time)</span
          >
        </p>
        <p v-if="view === 'matches'">Scores are final only after approval.</p>
        <p v-else>Each league has its own standings table.</p>
      </div>

      <div
        v-if="error"
        role="alert"
        class="rounded-lg border border-red-200 bg-red-50 p-5 text-red-800"
      >
        {{ error }}
      </div>
      <p v-if="loading" role="status" class="py-12 text-center text-gray-500">
        Loading {{ title.toLowerCase() }}…
      </p>

      <template v-else-if="view === 'matches'">
        <p
          v-if="!matchGroups.length && !error"
          role="status"
          class="rounded-xl border border-gray-200 bg-white p-10 text-center text-gray-500"
        >
          No published matches for this selection.
        </p>
        <section
          v-for="group in matchGroups"
          :key="group.key"
          class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm"
        >
          <div class="border-b border-gray-200 px-5 py-4">
            <p v-if="group.main !== group.league" class="text-sm text-gray-500">{{ group.main }}</p>
            <h2 class="text-xl font-semibold">
              {{ group.league }} <span class="font-normal text-gray-500">· {{ group.stage }}</span>
            </h2>
          </div>
          <GameMasterMatches :matches="group.matches" />
        </section>
      </template>

      <template v-else>
        <p
          v-if="!tables.length && !error"
          role="status"
          class="rounded-xl border border-gray-200 bg-white p-10 text-center text-gray-500"
        >
          No leagues with standings for this selection.
        </p>
        <section
          v-for="table in tables"
          :key="table.league.abbrev"
          class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm"
        >
          <div class="border-b border-gray-200 px-5 py-4">
            <p v-if="table.league.isSubLeague" class="text-sm text-gray-500">
              {{ table.league.main.name }}
            </p>
            <h2 class="text-xl font-semibold">
              {{ table.league.name }}
              <span class="font-normal text-gray-500"
                >· {{ table.data?.league_stage_name || 'Seeding' }}</span
              >
            </h2>
          </div>
          <p v-if="table.error" role="alert" class="p-5 text-red-700">{{ table.error }}</p>
          <template v-else-if="table.data">
            <div v-if="table.data.standings.length" class="overflow-x-auto p-3">
              <StandingsTable :standings="table.data.standings" class="min-w-[900px]" />
            </div>
            <p v-else class="p-8 text-center text-gray-500">No published standings yet.</p>
            <p class="border-t border-gray-100 px-5 py-3 text-xs text-gray-500">
              Last published: {{ tournamentTime(table.data.last_published) }}
            </p>
          </template>
        </section>
      </template>
    </div>
  </div>
</template>
