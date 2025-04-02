<script lang="ts" setup>
  import { ref } from 'vue';
  import SideBar from './components/SideBar.vue';

  const leftDrawerOpen = ref(false);

  function toggleLeftDrawer() {
    leftDrawerOpen.value = !leftDrawerOpen.value;
  }
</script>
<template>
  <q-layout view="lHh Lpr lFf">
    <q-header
      bordered
    >
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
