<script setup lang="ts">
  import { QTableColumn } from 'quasar';
  import { ref, computed, nextTick } from 'vue';

  const loading = ref(false);
  const pagination = ref({
    rowsPerPage: 0,
  });

  const columns: QTableColumn[] = [
    {
      name: 'index',
      label: '#',
      field: 'index',
      sortable: true,
      align: 'left',
    },
    {
      name: 'name',
      label: 'Name',
      field: 'name',
      sortable: true,
      align: 'left',
    },
    {
      name: 'code',
      label: 'Code',
      field: 'code',
      sortable: true,
      align: 'left',
    },
  ];

  const rows = [
    {
      index: 1,
      name: 'English',
      code: 'en',
    },
    {
      index: 2,
      name: 'Spanish',
      code: 'es',
    },
    {
      index: 3,
      name: 'French',
      code: 'fr',
    },
    {
      index: 4,
      name: 'German',
      code: 'de',
    },
    {
      index: 5,
      name: 'Italian',
      code: 'it',
    },
    {
      index: 6,
      name: 'Portuguese',
      code: 'pt',
    },
    {
      index: 7,
      name: 'Russian',
      code: 'ru',
    },
  ];

  const onScroll = (event: any) => {
    console.log(event);
  };
</script>

<template>
  <h5 class="text-h5">Languages</h5>
  <q-table
    class="my-sticky-dynamic text-left"
    flat
    color="primary"
    bordered
    :columns="columns"
    :rows="rows"
    :loading="loading"
    separator="cell"
    wrap-cells
    row-key="index"
    virtual-scroll
    :virtual-scroll-item-size="5"
    :virtual-scroll-sticky-size-start="5"
    :pagination="pagination"
    :rows-per-page-options="[0]"
    @virtual-scroll="onScroll"
  />
</template>
