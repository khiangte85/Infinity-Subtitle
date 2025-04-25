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
    path: '/languages',
    name: 'Languages',
    meta: {
      title: 'Languages',
      icon: 'fas fa-language',
      show: true,
    },
    component: () => import('../pages/Languages.vue'),
  },
];

export default routes;
