<script setup lang="ts">
  import { ref, onMounted } from 'vue';
  import { useQuasar } from 'quasar';
  import * as movieQueueAPI from '../../wailsjs/go/backend/MovieQueue';
  import * as languageAPI from '../../wailsjs/go/backend/Language';
  import { backend } from '../../wailsjs/go/models';
  import BatchUpload from '../components/movie-queue/BatchUpload.vue';
  import { EventsOn } from '../../wailsjs/runtime';
  const $q = useQuasar();
  const loading = ref(false);
  const languagesCodeMap = ref<Record<string, string>>({});
  const movies = ref<backend.MovieQueue[]>([]);
  const showBatchDialog = ref(false);
  const filter = ref({
    title: '',
  });

  const columns = [
    {
      name: 'name',
      label: 'Name',
      field: 'name',
      align: 'left' as const,
    },
    {
      name: 'source_language',
      label: 'Source Language',
      field: 'source_language',
      align: 'left' as const,
    },
    {
      name: 'target_language',
      label: 'Target Language',
      field: 'target_language',
      align: 'left' as const,
    },
    {
      name: 'status',
      label: 'Status',
      field: 'status',
      align: 'left' as const,
    },
    {
      name: 'created_at',
      label: 'Created At',
      field: 'created_at',
      align: 'left' as const,
      format: (val: string) => {
        return new Date(val)
          .toLocaleDateString('en-US', {
            day: '2-digit',
            month: '2-digit',
            year: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
            hour12: true,
          })
          .replace(/\//g, '-');
      },
    },
    {
      name: 'actions',
      label: 'Actions',
      field: 'actions',
      align: 'right' as const,
    },
  ];

  const pagination = ref<backend.Pagination>({
    sortBy: 'created_at',
    descending: true,
    page: 1,
    rowsPerPage: 10,
    rowsNumber: 0,
  });

  EventsOn('subtitle-created', (data: backend.MovieQueue) => {
    movies.value = movies.value.map((movie) => {
      if (movie.id === data.id) {
        return data;
      }
      return movie;
    });
  });

  EventsOn('movie-created', (data: backend.MovieQueue) => {
    movies.value = movies.value.map((movie) => {
      if (movie.id === data.id) {
        return data;
      }
      return movie;
    });
  });

  const getStatusColor = (status: number) => {
    switch (status) {
      case 0:
        return 'grey';
      case 1:
        return 'primary';
      case 2:
        return 'primary';
      case 3:
        return 'primary';
      case 4:
        return 'primary';
      default:
        return 'grey';
    }
  };

  const getStatusText = (status: number) => {
    switch (status) {
      case 0:
        return 'Pending';
      case 1:
        return 'Movie created';
      case 2:
        return 'Subtitle created';
      case 3:
        return 'Subtitle translated';
      case 4:
        return 'Failed';
      default:
        return 'Unknown';
    }
  };

  const deleteMovie = async (movie: backend.MovieQueue) => {
    try {
      await movieQueueAPI.DeleteFromQueue(movie.id);
      $q.notify({
        color: 'positive',
        message: 'Movie deleted from queue',
      });
      onRequest({
        pagination: pagination.value,
        filter: filter.value,
      });
    } catch (error) {
      $q.notify({
        color: 'negative',
        message: 'Failed to delete movie from queue',
      });
    }
  };

  const fetchMovies = async (props: any) => {
    loading.value = true;
    try {
      const response = await movieQueueAPI.ListQueue(
        filter.value.title,
        props.pagination
      );
      movies.value = response.movies ?? [];
      return response;
    } catch (error) {
      console.error('Error fetching movies:', error);
      $q.notify({
        color: 'negative',
        message: 'Failed to load movies queue',
      });
      return null;
    } finally {
      loading.value = false;
    }
  };

  const getLanguages = async () => {
    try {
      const response = await languageAPI.GetAllLanguages();
      response.forEach((language) => {
        languagesCodeMap.value[language.code] = language.name;
      });
    } catch (error) {
      console.error('Error fetching languages:', error);
      return [];
    }
  };

  const onRequest = async (props: any) => {
    const { page, rowsPerPage, sortBy, descending } = props.pagination;
    props.filter = filter.value;
    const response = await fetchMovies(props);
    pagination.value.page = page;
    pagination.value.rowsPerPage = rowsPerPage;
    pagination.value.rowsNumber = response?.pagination.rowsNumber ?? 0;
    pagination.value.sortBy = sortBy;
    pagination.value.descending = descending;
  };

  getLanguages();

  onMounted(async () => {
    onRequest({
      pagination: pagination.value,
      filter: filter.value,
    });
  });
</script>

<template>
  <q-card
    flat
    class="full-width row justify-between items-center"
  >
    <q-card-section class="q-py-sm q-pl-none">
      <h5 class="text-h5">Queues</h5>
    </q-card-section>
    <q-space />
    <q-card-section class="q-py-sm">
      <q-btn
        round
        unelevated
        color="primary"
        icon="fas fa-plus"
        size="sm"
        @click="
          () => {
            showBatchDialog = true;
          }
        "
      >
        <q-tooltip> Create Batch </q-tooltip>
      </q-btn>
    </q-card-section>
  </q-card>

  <q-card
    flat
    bordered
    class="full-width q-mb-md"
  >
    <q-card-section class="row q-py-lg">
      <q-input
        class="q-ml-md"
        dense
        outlined
        debounce="300"
        v-model="filter.title"
        autocomplete="off"
        clearable
        placeholder="Search"
        :style="{ minWidth: '400px', maxWidth: '600px' }"
      />

      <q-btn
        class="q-mx-md"
        :round="true"
        unelevated
        icon="fas fa-filter-circle-xmark"
        color="primary"
        @click="
          () => {
            filter.title = '';
            onRequest({
              pagination: { ...pagination },
              filter: { ...filter },
            });
          }
        "
      >
        <q-tooltip>Clear</q-tooltip>
      </q-btn>
    </q-card-section>
  </q-card>

  <q-table
    class="text-left"
    flat
    color="primary"
    bordered
    :columns="columns"
    :rows="movies"
    :filter="filter"
    :loading="loading"
    separator="cell"
    wrap-cells
    row-key="id"
    v-model:pagination="pagination"
    :rows-per-page-options="[10, 20, 50]"
    binary-state-sort
    rows-per-page-label="Records per page"
    @request="onRequest"
  >
    <template v-slot:body-cell-status="props">
      <q-td :props="props">
        <q-chip
          :color="getStatusColor(props.row.status)"
          text-color="white"
        >
          {{ getStatusText(props.row.status) }}
        </q-chip>
      </q-td>
    </template>

    <template v-slot:body-cell-source_language="props">
      <q-td :props="props">
        {{ languagesCodeMap[props.row.source_language] }}
      </q-td>
    </template>

    <template v-slot:body-cell-target_language="props">
      <q-td :props="props">
        {{ languagesCodeMap[props.row.target_language] }}
      </q-td>
    </template>

    <template v-slot:body-cell-actions="props">
      <q-td :props="props">
        <q-btn
          flat
          round
          color="primary"
          icon="delete"
          @click="deleteMovie(props.row)"
          :disable="props.row.status !== 0"
        />
      </q-td>
    </template>
  </q-table>

  <q-dialog
    v-model="showBatchDialog"
    persistent
  >
    <BatchUpload
      @onClose="
        () => {
          showBatchDialog = false;
        }
      "
      @on-queue="
        () => {
          showBatchDialog = false;
          onRequest({
            pagination: pagination,
            filter: filter,
          });
        }
      "
    />
  </q-dialog>
</template>
