import { createApp } from 'vue';
import router  from './router';
import { Quasar, Notify, LocalStorage } from 'quasar';
import App from './App.vue';
import './style.css';

import '@quasar/extras/material-icons/material-icons.css';
import '@quasar/extras/fontawesome-v6/fontawesome-v6.css';

import 'quasar/src/css/index.sass';

import { createI18n } from 'vue-i18n';

import messages from './i18n';

export type MessageLanguages = keyof typeof messages;
// Type-define 'en-US' as the master schema for the resource
export type MessageSchema = (typeof messages)['en-US'];

// See https://vue-i18n.intlify.dev/guide/advanced/typescript.html#global-resource-schema-type-definition
/* eslint-disable @typescript-eslint/no-empty-interface */
declare module 'vue-i18n' {
  // define the locale messages schema
  export interface DefineLocaleMessage extends MessageSchema {}

  // define the datetime format schema
  export interface DefineDateTimeFormat {}

  // define the number format schema
  export interface DefineNumberFormat {}
}
/* eslint-enable @typescript-eslint/no-empty-interface */

const defaultLocale = LocalStorage.getItem('locale');

LocalStorage.set('locale', defaultLocale ?? 'zh-CN');

const i18n = createI18n({
  locale: defaultLocale?.toString(),
  legacy: false,
  globalInjection: true,
  messages,
});


const app = createApp(App);

app.use(i18n);
app.use(router);
app.use(Quasar, {
  plugins: {
    Notify,
    LocalStorage
  },
});

app.mount('#app');
