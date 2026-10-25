<script setup lang="ts"></script>

<template>
  <div class="mx-auto max-w-3xl space-y-6 p-8">
    <h1 class="text-3xl font-bold">RCJV Soccer Stats</h1>
    <p class="text-gray-600">
      Tournament matches and standings with main-league and sub-league selection.
    </p>
    <nav aria-label="Tournament screens" class="flex flex-wrap gap-4">
      <RouterLink to="/v2/matches" class="rounded-lg bg-blue-700 px-6 py-4 font-semibold text-white"
        >Matches v2</RouterLink
      >
      <RouterLink
        to="/v2/standings"
        class="rounded-lg bg-blue-700 px-6 py-4 font-semibold text-white"
        >Standings v2</RouterLink
      >
    </nav>
    <nav aria-label="Legacy screens" class="flex gap-4 text-sm text-blue-700">
      <RouterLink to="/matches">Legacy matches</RouterLink>
      <RouterLink to="/standings">Legacy standings</RouterLink>
      <RouterLink to="/referee">Referees</RouterLink>
    </nav>
  </div>
</template>

<style scoped></style>
