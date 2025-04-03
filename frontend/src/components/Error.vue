<script setup lang="ts">

const props = defineProps<{
  messages: {
    [key: string]: string;
  };
  translate?: boolean;
}>();
</script>
<template>
  <div
    class="bg-negative text-white q-pb-none"
    :style="{ borderRadius: '3px' }"
  >
    <ol class="q-my-none q-pt-sm">
      <template v-for="(value, key, i) in props.messages" :key="i">
        <li class="q-pb-sm">{{ value }}</li>
      </template>
    </ol>
  </div>
</template>
