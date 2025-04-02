import { createApp } from 'vue';
import router  from './router';
import { Quasar } from 'quasar';
import App from './App.vue';
import './style.css';

import '@quasar/extras/material-icons/material-icons.css';
import '@quasar/extras/fontawesome-v6/fontawesome-v6.css';

import 'quasar/src/css/index.sass';

const app = createApp(App);

app.use(router);
app.use(Quasar, {
  plugins: {},
});

app.mount('#app');
