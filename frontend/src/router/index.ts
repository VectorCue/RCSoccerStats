import { createRouter, createWebHashHistory } from 'vue-router'

const router = createRouter({
  history: createWebHashHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/v2/matches',
      name: 'v2-matches',
      component: () => import('@/views/V2TournamentView.vue'),
      props: { view: 'matches' },
    },
    {
      path: '/v2/standings',
      name: 'v2-standings',
      component: () => import('@/views/V2TournamentView.vue'),
      props: { view: 'standings' },
    },
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/HomeView.vue'),
    },
    {
      path: '/standings',
      name: 'standings',
      component: () => import('@/views/StandingsView.vue'),
    },
    {
      path: '/matches',
      name: 'matches',
      component: () => import('@/views/MatchesView.vue'),
    },
    {
      path: '/referee',
      name: 'referee',
      component: () => import('@/views/RefereeView.vue'),
    },
  ],
})

export default router
