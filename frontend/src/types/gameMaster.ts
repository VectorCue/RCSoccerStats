import type { StandingEntry, Team } from './standings'

export interface TournamentMetadata {
  tournament_abbrev: string
  tournament_name: string
  tournament_publishing_timezone: string
  tournament_header_type: string
}

export interface LeagueStage {
  name: string
  standings_publishing_state: 'PUBLISHED' | 'UNPUBLISHED'
  standings_last_published: string
}

export interface League {
  abbrev: string
  name: string
  league_stages: LeagueStage[]
  main_league_abbrev: string
  main_league_name: string
  sub_league_abbrev: string | null
  sub_league_name: string | null
}

export interface MainLeague {
  abbrev: string
  name: string
  league_stages: LeagueStage[]
  sub_leagues: League[]
}

export interface LeaguesResponse extends TournamentMetadata {
  leagues: League[]
  main_leagues: MainLeague[]
}

export interface ExactLeague {
  abbrev: string
  name: string
  main: MainLeague
  isSubLeague: boolean
  stages: LeagueStage[]
}

export interface Selection {
  main: string
  sub: string
  league: string
  stage: string
}

export type MatchStatus = 'draft' | 'scheduled' | 'in_progress' | 'submitted' | 'approved'

export interface GameMasterMatch {
  number: number
  team1: Team | null
  team2: Team | null
  league: string
  league_abbrev: string
  main_league_abbrev: string
  main_league_name: string
  sub_league_abbrev: string | null
  sub_league_name: string | null
  league_stage: string
  stage_type: 'round_robin' | 'swiss' | 'knockout' | 'individual' | 'manual'
  is_placeholder: boolean
  is_bye: boolean
  status: MatchStatus
  group_name: string
  start: string
  duration: string
  pitch: string
  goals1: number
  goals2: number
  points1: number
  points2: number
  referees: { first_name: string; last_name: string }[]
  last_published: string
}

export interface MatchesResponse extends TournamentMetadata {
  matches: GameMasterMatch[]
}

export interface GameMasterStandings extends TournamentMetadata {
  league_abbrev: string
  league_name: string
  league_stage_number: string
  league_stage_name: string
  last_published: string
  standings: StandingEntry[]
}
