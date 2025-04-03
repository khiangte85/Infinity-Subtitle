<script setup lang="ts">
  import { QTableColumn } from 'quasar';
  import { reactive, ref } from 'vue';
  import { backend as models} from '../../wailsjs/go/models.js';
  import { GetAllLanguages } from '../../wailsjs/go/backend/Language.js';

  const loading = ref(true);
  const pagination = ref({
    rowsPerPage: 0,
  });

  const languages = ref<models.Language[]>([]);

  const getLanguages = async () => {
    try {
      languages.value = await GetAllLanguages();
    } catch (error) {
      console.error(error);
    } finally {
      loading.value = false;
    }
  };

  getLanguages();

  const columns: QTableColumn[] = [
    {
      name: 'id',
      label: '#',
      field: 'id',
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
</script>

<template>
  <h5 class="text-h5">Languages</h5>
  <q-table
    class="text-left"
    flat
    color="primary"
    bordered
    :columns="columns"
    :rows="languages"
    :loading="loading"
    separator="cell"
    wrap-cells
    row-key="id"
    :pagination="pagination"
    :rows-per-page-options="[0]"
  />
</template>
