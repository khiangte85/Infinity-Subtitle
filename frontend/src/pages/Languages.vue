<script setup lang="ts">
  import { QTableColumn } from 'quasar';
  import { ref } from 'vue';
  import { backend as models } from '../../wailsjs/go/models.js';
  import {
    GetAllLanguages,
  } from '../../wailsjs/go/backend/Language.js';
  import AddLanguage from '../components/language/Add.vue';

  const loading = ref(true);
  const pagination = ref({
    rowsPerPage: 0,
  });

  const showDialog = ref(false);
  const languages = ref<models.Language[]>([]);

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
</script>

<template>
  <div class="q-py-md row justify-between items-center">
    <div>
      <h5 class="text-h5">Languages</h5>
    </div>
    <div>
      <q-btn
        round
        unelevated
        color="primary"
        icon="fas fa-plus"
        size="sm"
        @click="
          () => {
            showDialog = true;
          }
        "
      >
        <q-tooltip> Add Language </q-tooltip>
      </q-btn>
    </div>
  </div>
  <q-table
    class="text-left"
    color="primary"
    flat
    dense
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

  <q-dialog v-model="showDialog">
    <AddLanguage
      @onClose="showDialog = false"
      @onAdded="() => {
        showDialog = false;
        getLanguages();
      }"
    />
  </q-dialog>
</template>
