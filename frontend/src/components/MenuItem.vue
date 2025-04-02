<script setup lang="ts">
  const props = defineProps({
    title: {
      type: String,
      required: true,
    },
    path: {
      type: String,
      required: true,
    },
    icon: {
      type: String,
      required: false,
    },
    size: {
      type: String,
      required: false,
    },
  });
</script>

<template>
  <q-item
    :to="props.path"
    clickable
    active-class="active-menu-link"
    class="q-py-none nav-item-wrapper"
    exact
  >
    <q-item-section avatar>
      <q-icon
        :name="props.icon"
        :size="props.size || '16px'"
      />
    </q-item-section>

    <q-item-section>
      <q-item-label>{{ props.title }}</q-item-label>
    </q-item-section>
  </q-item>
</template>
