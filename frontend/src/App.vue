<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { LocalStorage } from 'quasar';
  import { localeOptions as locales } from './constants/locales';
  import SideBar from './components/SideBar.vue';

  const leftDrawerOpen = ref(false);
  const { locale } = useI18n({ useScope: 'global' });
  const localeOptions = ref(locales);

  watch(locale, (value) => {
    LocalStorage.set('locale', value);
  });

  function toggleLeftDrawer() {
    leftDrawerOpen.value = !leftDrawerOpen.value;
  }
</script>
<template>
  <q-layout view="lHh Lpr lFf">
    <q-header bordered>
      <q-toolbar>
        <q-btn
          dense
          flat
          round
          icon="menu"
          @click="toggleLeftDrawer"
        />

        <!-- <q-toolbar-title>
          Infinity Subtitle
        </q-toolbar-title> -->
        <q-space />
        <q-select
          v-model="locale"
          :options="localeOptions"
          class="select-white q-ml-lg"
          dense
          emit-value
          map-options
          option-label="name"
          option-value="locale"
          outlined
          :style="{ minWidth: '150px' }"
        >
          <template v-slot:prepend>
            <q-icon
              name="fas fa-language"
              size="sm"
              class="q-pr-sm"
            />
          </template>
        </q-select>
      </q-toolbar>
    </q-header>

    <q-drawer
      show-if-above
      v-model="leftDrawerOpen"
      side="left"
      behavior="desktop"
      bordered
    >
      <!-- drawer content -->
      <SideBar />
    </q-drawer>

    <q-page-container>
      <q-page class="q-px-lg q-pt-sm text-left">
        <router-view :key="$route.fullPath" />
      </q-page>
    </q-page-container>
  </q-layout>
</template>
