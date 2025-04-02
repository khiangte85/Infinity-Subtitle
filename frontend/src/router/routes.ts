import { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Dashboard',
    meta: {
      title: 'Dashboard',
      icon: 'fas fa-tv',
    },
    component: () => import('../pages/Dashboard.vue'),
  },
  {
    path: '/movies',
    name: 'Movies',
    meta: {
      title: 'Movies',
      icon: 'fas fa-film',
    },
    component: () => import('../pages/Movies.vue'),
  },
  {
    path: '/languages',
    name: 'Languages',
    meta: {
      title: 'Languages',
      icon: 'fas fa-language',
    },
    component: () => import('../pages/Languages.vue'),
  },
];

export default routes;
