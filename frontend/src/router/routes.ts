import { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Movies',
    meta: {
      title: 'Movies',
      icon: 'fas fa-film',
      show: true,
    },
    component: () => import('../pages/Movies.vue'),
  },
  {
    path: '/movies/:id/subtitles',
    name: 'Subtitles',
    meta: {
      title: 'Subtitles',
      icon: 'fas fa-closed-captioning',
      show: false,
    },
    component: () => import('../pages/Subtitle.vue'),
  },
  {
    path: '/movie-queue',
    name: 'MovieQueue',
    meta: {
      title: 'Queues',
      icon: 'fas fa-layer-group',
      show: true,
    },
    component: () => import('../pages/MovieQueue.vue'),
  },
  {
    path: '/languages',
    name: 'Languages',
    meta: {
      title: 'Languages',
      icon: 'fas fa-language',
      show: true,
    },
    component: () => import('../pages/Languages.vue'),
  },
  {
    path: '/settings',
    name: 'Settings',
    meta: {
      title: 'Settings',
      icon: 'fas fa-cog',
      show: true,
    },
    component: () => import('../pages/Settings.vue'),
  },
];

export default routes;
