<script setup lang="ts">
  import { QTableColumn } from 'quasar';
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { backend as models } from '../../wailsjs/go/models.js';
  import {
    GetAllLanguages,
  } from '../../wailsjs/go/backend/Language.js';
  import AddLanguage from '../components/language/Add.vue';

  const { t } = useI18n();
  const loading = ref(true);
  const pagination = ref({
    rowsPerPage: 0,
  });

  const showDialog = ref(false);
  const languages = ref<models.Language[]>([]);

  const columns: QTableColumn[] = [
    {
      name: 'id',
      label: t('#'),
      field: 'id',
      sortable: true,
      align: 'left',
    },
    {
      name: 'name',
      label: t('Name'),
      field: 'name',
      sortable: true,
      align: 'left',
    },
    {
      name: 'code',
      label: t('Code'),
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
      <h5 class="text-h5">{{ $t('Languages') }}</h5>
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
        <q-tooltip>{{ $t('Add Language') }}</q-tooltip>
      </q-btn>
    </div>
  </div>
  <q-table
    class="text-left"
    color="primary"
    flat
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
