<script setup lang="ts">
  import routes from '../router/routes';
</script>

<template>
  <q-list class="q-pa-md text-left">
    <q-item class="q-pa-none">
      <q-img
        src="../assets/images/logo.png"
        :style="{ maxWidth: '50px' }"
      />
      <q-item-section class="q-ml-sm">
        <q-item-label
          :style="{
            fontSize: '14px',
            fontWeight: 'bold',
          }"
        >
          Infinity Subtitle
        </q-item-label>
      </q-item-section>
    </q-item>

    <q-separator color="grey-8" class="q-mt-lg" />

    <template
      v-for="(route, key) in routes"
      :key="key"
    >
      <q-item
        :to="route.path"
        clickable
        active-class="active-route-link active-menu-link"
        class="q-my-sm nav-item-wrapper q-pt-sm q-pb-sm"
        exact
        :style="{
          borderRadius: '8px',
        }"
      >
        <q-item-section
          avatar
          :style="{ minWidth: 'auto' }"
        >
          <q-icon
            :name="route.meta?.icon as string"
            size="xs"
          />
        </q-item-section>

        <q-item-section>
          <q-item-label
            :style="{
              fontSize: '14px',
            }"
          >
            {{ route.meta?.title }}
          </q-item-label>
        </q-item-section>
      </q-item>
    </template>
  </q-list>
</template>
