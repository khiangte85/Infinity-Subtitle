import { createRouter, createMemoryHistory, RouteRecordRaw } from 'vue-router';

import routes from './routes';

const router = createRouter({
  history: createMemoryHistory(),
  routes,
});

export default router;
