<script setup lang="ts">
  import { ref, onMounted } from 'vue';
  import { useQuasar } from 'quasar';
  import { backend } from '../../wailsjs/go/models';
  import BatchUpload from '../components/movie-queue/BatchUpload.vue';

  interface AddToQueueRequest {
    name: string;
    content: string;
    source_language: string;
    target_language: string;
  }

  interface MovieQueue {
    id: number;
    name: string;
    content: string;
    source_language: string;
    target_language: string;
    status: number;
    created_at: string;
    processed_at: string | null;
  }

  interface MovieQueueResponse {
    movies: MovieQueue[];
    pagination: {
      rowsNumber: number;
    };
  }

  interface GetQueueRequest {
    page: number;
    rowsPerPage: number;
    sortBy: string;
    descending: boolean;
  }

  declare global {
    interface Window {
      backend: {
        MovieQueue: {
          AddToQueue: (request: AddToQueueRequest) => Promise<void>;
          DeleteFromQueue: (id: number) => Promise<void>;
          GetQueue: (request: GetQueueRequest) => Promise<MovieQueueResponse>;
        };
      };
    }
  }

  const $q = useQuasar();
  const loading = ref(false);
  const movies = ref<MovieQueue[]>([]);
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
    },
    {
      name: 'processed_at',
      label: 'Processed At',
      field: 'processed_at',
      align: 'left' as const,
    },
    {
      name: 'actions',
      label: 'Actions',
      field: 'actions',
      align: 'right' as const,
    },
  ];

  const pagination = ref({
    sortBy: 'created_at',
    descending: true,
    page: 1,
    rowsPerPage: 10,
    rowsNumber: 0,
  });

  const getStatusColor = (status: number) => {
    switch (status) {
      case 0:
        return 'primary';
      case 1:
        return 'warning';
      case 2:
        return 'positive';
      case 3:
        return 'negative';
      default:
        return 'grey';
    }
  };

  const getStatusText = (status: number) => {
    switch (status) {
      case 0:
        return 'Pending';
      case 1:
        return 'Processing';
      case 2:
        return 'Completed';
      case 3:
        return 'Failed';
      default:
        return 'Unknown';
    }
  };

  const deleteMovie = async (movie: MovieQueue) => {
    try {
      await window.backend.MovieQueue.DeleteFromQueue(movie.id);
      $q.notify({
        color: 'positive',
        message: 'Movie deleted from queue',
      });
      loadMovies();
    } catch (error) {
      $q.notify({
        color: 'negative',
        message: 'Failed to delete movie from queue',
      });
    }
  };

  const loadMovies = async () => {
    loading.value = true;
    try {
      const response = await window.backend.MovieQueue.GetQueue({
        page: pagination.value.page,
        rowsPerPage: pagination.value.rowsPerPage,
        sortBy: pagination.value.sortBy,
        descending: pagination.value.descending,
      });
      movies.value = response.movies;
      pagination.value.rowsNumber = response.pagination.rowsNumber;
    } catch (error) {
      $q.notify({
        color: 'negative',
        message: 'Failed to load movies queue',
      });
    } finally {
      loading.value = false;
    }
  };

  const onRequest = (props: any) => {
    pagination.value = props.pagination;
    loadMovies();
  };

  onMounted(() => {
    loadMovies();
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

  <q-dialog        v-model="showBatchDialog" persistent>
    <BatchUpload
      @onClose="() => {
        showBatchDialog = false;
      }"
      @on-queue="() => {
        showBatchDialog = false;
        loadMovies();
      }"
    />
  </q-dialog>

</template>
