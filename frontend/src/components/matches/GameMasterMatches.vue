<script setup lang="ts">
import type { GameMasterMatch, MatchStatus } from '@/types/gameMaster'
import { matchScore, tournamentTime } from '@/lib/gameMaster'

defineProps<{ matches: GameMasterMatch[] }>()

const statuses: Record<MatchStatus, string> = {
  draft: 'Draft',
  scheduled: 'Scheduled',
  in_progress: 'In progress',
  submitted: 'Awaiting approval',
  approved: 'Final',
}
</script>

<template>
  <div class="overflow-x-auto">
    <table class="w-full min-w-[900px] border-collapse text-left">
      <thead class="bg-gray-50 text-sm text-gray-600">
        <tr>
          <th scope="col" class="px-4 py-3">#</th>
          <th scope="col" class="px-4 py-3">Time</th>
          <th scope="col" class="px-4 py-3">Team 1</th>
          <th scope="col" class="px-4 py-3 text-center">Score</th>
          <th scope="col" class="px-4 py-3">Team 2</th>
          <th scope="col" class="px-4 py-3">Field</th>
          <th scope="col" class="px-4 py-3 text-center">Points</th>
          <th scope="col" class="px-4 py-3">Status</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="match in matches"
          :key="match.number"
          class="border-t border-gray-100 text-base"
          :class="{
            'bg-amber-50': match.status === 'in_progress',
            'text-gray-500': match.status === 'approved',
          }"
        >
          <td class="px-4 py-4 tabular-nums">{{ match.number }}</td>
          <td class="whitespace-nowrap px-4 py-4 tabular-nums">
            {{ tournamentTime(match.start) }}
          </td>
          <td class="px-4 py-4 font-medium">
            {{ match.is_placeholder ? 'Awaiting teams' : match.team1?.name || '—' }}
          </td>
          <td class="whitespace-nowrap px-4 py-4 text-center font-semibold tabular-nums">
            {{ matchScore(match) }}
          </td>
          <td class="px-4 py-4 font-medium">
            {{
              match.is_placeholder
                ? 'Awaiting teams'
                : match.is_bye
                  ? 'BYE'
                  : match.team2?.name || 'Solo run'
            }}
          </td>
          <td class="px-4 py-4">{{ match.pitch || '—' }}</td>
          <td class="whitespace-nowrap px-4 py-4 text-center tabular-nums">
            <template v-if="match.status === 'approved' && !match.is_placeholder">
              {{ match.points1
              }}<template v-if="match.team2 !== null"> : {{ match.points2 }}</template>
            </template>
            <template v-else>—</template>
          </td>
          <td class="whitespace-nowrap px-4 py-4 text-sm">
            {{ match.is_placeholder ? 'Awaiting teams' : statuses[match.status] }}
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
