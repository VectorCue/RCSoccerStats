import type { ExactLeague, GameMasterMatch, MainLeague, Selection } from '../types/gameMaster'

export function exactLeagues(catalog: MainLeague[]): ExactLeague[] {
  return catalog.flatMap((main) => [
    // A root with no direct stages has no separate table to display.
    ...(main.league_stages.length
      ? [
          {
            abbrev: main.abbrev,
            name: main.name,
            main,
            isSubLeague: false,
            stages: main.league_stages,
          },
        ]
      : []),
    ...main.sub_leagues.map((sub) => ({
      abbrev: sub.abbrev,
      name: sub.name,
      main,
      isSubLeague: true,
      stages: sub.league_stages,
    })),
  ])
}

export function resolveSelection(catalog: MainLeague[], selection: Selection) {
  const main = catalog.find((item) => item.abbrev === selection.main)
  if (selection.main && !main) throw new Error(`Unknown main league: ${selection.main}`)
  if (selection.sub && selection.league && selection.sub !== selection.league) {
    throw new Error('The league and sub-league selectors must name the same league.')
  }

  const abbrev = selection.sub || selection.league
  const leagues = exactLeagues(catalog)
  // Exact roots are valid even when their own stage list is empty.
  const root = catalog.find((item) => item.abbrev === abbrev)
  const exact =
    leagues.find((item) => item.abbrev === abbrev) ??
    (root
      ? {
          abbrev: root.abbrev,
          name: root.name,
          main: root,
          isSubLeague: false,
          stages: root.league_stages,
        }
      : undefined)
  if (abbrev && !exact) throw new Error(`Unknown league: ${abbrev}`)
  if (selection.sub && !exact?.isSubLeague)
    throw new Error('The sub-league must be a child league.')
  if (main && exact && exact.main.abbrev !== main.abbrev) {
    throw new Error('The selected sub-league does not belong to this main league.')
  }
  if (selection.stage) {
    if (!exact) throw new Error('Select one exact league before selecting a stage.')
    if (!/^\d+$/.test(selection.stage) || !exact.stages[Number(selection.stage)]) {
      throw new Error('The selected stage does not exist in this league.')
    }
  }
  return {
    exact,
    leagues: exact ? [exact] : leagues.filter((item) => !main || item.main.abbrev === main.abbrev),
  }
}

export function exactLeagueParams(league: ExactLeague, stage = ''): Record<string, string> {
  const params: Record<string, string> = { main_league: league.main.abbrev }
  params[league.isSubLeague ? 'sub_league' : 'league'] = league.abbrev
  if (stage !== '') params.league_stage = stage
  return params
}

export function matchParams(selection: Selection): Record<string, string> {
  const params: Record<string, string> = {}
  if (selection.main) params.main_league = selection.main
  if (selection.sub) params.sub_league = selection.sub
  if (selection.league) params.league = selection.league
  if (selection.stage !== '') params.league_stage = selection.stage
  return params
}

export function groupMatches(matches: GameMasterMatch[]) {
  const groups = new Map<
    string,
    { key: string; league: string; main: string; stage: string; matches: GameMasterMatch[] }
  >()
  // All timestamps share the supplied tournament timezone, so lexical order is chronological.
  const sorted = [...matches].sort(
    (a, b) => (a.start || '\uffff').localeCompare(b.start || '\uffff') || a.number - b.number,
  )
  for (const match of sorted) {
    const key = JSON.stringify([match.league_abbrev, match.league_stage])
    let group = groups.get(key)
    if (!group) {
      group = {
        key,
        league: match.league,
        main: match.main_league_name,
        stage: match.league_stage,
        matches: [],
      }
      groups.set(key, group)
    }
    group.matches.push(match)
  }
  return [...groups.values()]
}

// These are tournament wall times, not browser-local dates or UTC timestamps.
export function tournamentTime(value: string) {
  return value ? value.replace(' ', ' · ') : '—'
}

export function matchScore(match: GameMasterMatch) {
  if (match.status !== 'approved' || match.is_placeholder) return '—'
  if (match.is_bye) return 'BYE'
  return match.team2 === null ? String(match.goals1) : `${match.goals1} : ${match.goals2}`
}
