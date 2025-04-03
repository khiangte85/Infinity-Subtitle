<script setup lang="ts">
  import { QTableColumn } from 'quasar';
  import { ref } from 'vue';
  import { backend as models } from '../../wailsjs/go/models.js';
  import {
    ListMovies,
  } from '../../wailsjs/go/backend/Movie.js';
  import AddMovie from '../components/movie/Add.vue';
  import { GetAllLanguages } from '../../wailsjs/go/backend/Language.js';
  const loading = ref(true);
  const pagination = ref({
    rowsPerPage: 0,
  });

  const showDialog = ref(false);
  const movies = ref<models.Movie[]>([]);
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
      name: 'title',
      label: 'Title',
      field: 'title',
      sortable: true,
      align: 'left',
    },
    {
      name: 'default_language',
      label: 'Default Language',
      field: 'default_language',
      sortable: true,
      align: 'left',
      format: (val: string) => {
        const language = languages.value.find((language) => language.code === val);
        return language ? language.name : '';
      },
    },
    {
      name: 'languages',
      label: 'Subtitle Languages',
      field: 'languages',
      sortable: true,
      align: 'left',
      format: (val: string) => {
        return Object.entries(val).map(([key, value]) => `${value}`).join(', ');
      },
    },
    {
      name: 'created_at',
      label: 'Created At',
      field: 'created_at',
      sortable: true,
      align: 'left',
      format: (val: string) => {
        return new Date(val).toLocaleString();
      },
    },
  ];

  const paginateMovies = async () => {
    try {
      const response = await ListMovies('', '', false, 0, 10) as Record<string, any>;
      movies.value = response.movies;
      pagination.value.rowsPerPage = response.last_id;
    } catch (error) {
      console.error(error);
    } finally {
      loading.value = false;
    }
  };

  const getLanguages = async () => {
    try {
      const response = await GetAllLanguages();
      languages.value = response;

      paginateMovies();
    } catch (error) {
      console.error(error);
    }
  };

  getLanguages();
</script>

<template>
  <div class="q-pa-md row justify-between items-center">
    <div>
      <h5 class="text-h5">Movies</h5>
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
        <q-tooltip> Add Movie </q-tooltip>
      </q-btn>
    </div>
  </div>
  <q-table
    class="text-left"
    flat
    dense
    color="primary"
    bordered
    :columns="columns"
    :rows="movies"
    :loading="loading"
    separator="cell"
    wrap-cells
    row-key="id"
    :pagination="pagination"
    :rows-per-page-options="[0]"
  />

  <q-dialog v-model="showDialog">
    <AddMovie
      @onClose="showDialog = false"
      @onAdded="() => {
        showDialog = false;
        paginateMovies();
      }"
    />
  </q-dialog>
</template>
