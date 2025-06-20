<script setup lang="ts">
  import { ref, onMounted } from 'vue';
  import { useQuasar } from 'quasar';
  import { useI18n } from 'vue-i18n';
  import * as movieQueueAPI from '../../wailsjs/go/backend/MovieQueue';
  import * as languageAPI from '../../wailsjs/go/backend/Language';
  import { backend } from '../../wailsjs/go/models';
  import BatchUpload from '../components/movie-queue/BatchUpload.vue';
  import VideoToAudio from '../components/movie-queue/VideoToAudio.vue';
  import { EventsOn } from '../../wailsjs/runtime';

  const { t } = useI18n();
  const $q = useQuasar();
  const loading = ref(false);
  const languagesCodeMap = ref<Record<string, string>>({});
  const movies = ref<backend.MovieQueue[]>([]);
  const showBatchDialog = ref(false);
  const showVideoToAudioDialog = ref(false);
  const filter = ref({
    title: '',
  });

  const columns = [
    {
      name: 'id',
      label: t('#'),
      field: 'id',
      align: 'left' as const,
      sortable: true,
    },
    {
      name: 'name',
      label: t('Name'),
      field: 'name',
      align: 'left' as const,
      sortable: true,
    },
    {
      name: 'type',
      label: t('Type'),
      field: 'type',
      align: 'left' as const,
      sortable: true,
      format: (val: string) => {
        return val
      },
    },
    {
      name: 'file_type',
      label: t('File Type'),
      field: 'file_type',
      align: 'left' as const,
      sortable: true,
      format: (val: string) => {
        return val
      },
    },
    {
      name: 'source_language',
      label: t('Source Language'),
      field: 'source_language',
      align: 'left' as const,
      sortable: true,
    },
    {
      name: 'target_languages',
      label: t('Target Languages'),
      field: 'target_languages',
      align: 'left' as const,
      sortable: true,
    },
    {
      name: 'status',
      label: t('Status'),
      field: 'status',
      align: 'left' as const,
      sortable: true,
    },
    {
      name: 'created_at',
      label: t('Created At'),
      field: 'created_at',
      align: 'left' as const,
      sortable: true,
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
      label: t('Actions'),
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

  EventsOn('audio-transcribed', (id: number, status: number) => {
    movies.value = movies.value.map((movie) => {
      if (movie.id === id) {
        movie.status = status;
        return movie;
      }
      return movie;
    });
  });

  EventsOn('movie-created', (id: number, status: number) => {
    movies.value = movies.value.map((movie) => {
      if (movie.id === id) {
        movie.status = status;
        return movie;
      }
      return movie;
    });
  });

  EventsOn('subtitle-created', (id: number, status: number) => {
    movies.value = movies.value.map((movie) => {
      if (movie.id === id) {
        movie.status = status;
        return movie;
      }
      return movie;
    });
  });

  EventsOn('subtitle-translated', (id: number, status: number) => {
    movies.value = movies.value.map((movie) => {
      if (movie.id === id) {
        movie.status = status;
        return movie;
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
        return 'green';
      case 5:
        return 'negative';
      default:
        return 'grey';
    }
  };

  const getStatusText = (status: number) => {
    switch (status) {
      case 0:
        return t('Pending');
      case 1:
        return t('Audio transcribed');
      case 2:
        return t('Movie created');
      case 3:
        return t('Subtitle Created');
      case 4:
        return t('Subtitle Translated');
      default:
        return t('Unknown');
    }
  };

  const deleteMovie = async (id: number) => {
    try {
      await movieQueueAPI.DeleteFromQueue(id);
      $q.notify({
        color: 'positive',
        message: t('Movie deleted from queue'),
      });
      onRequest({
        pagination: pagination.value,
        filter: filter.value,
      });
    } catch (error) {
      $q.notify({
        color: 'negative',
        message: t('Failed to delete movie from queue'),
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
        message: t('Failed to load movies queue'),
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
      <h5 class="text-h5">{{ $t('Queues') }}</h5>
    </q-card-section>
    <q-space />
    <q-card-section class="q-py-sm">
      <!-- <q-btn
        round
        unelevated
        color="primary"
        icon="fas fa-file-audio"
        size="sm"
        class="q-mr-sm"
        @click="() => {
          showVideoToAudioDialog = true;
        }"
      >
        <q-tooltip>{{ $t('Convert Video to Audio') }}</q-tooltip>
      </q-btn> -->
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
        <q-tooltip>{{ $t('Create Batch') }}</q-tooltip>
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
        :placeholder="$t('Search')"
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
        <q-tooltip>{{ $t('Clear') }}</q-tooltip>
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
    :rows-per-page-label="$t('Records per page')"
    @request="onRequest"
  >
    <template v-slot:body-cell-status="props">
      <q-td :props="props">
        <q-chip
          size="10px"
          class="q-px-md q-py-md text-weight-bold"
          :color="getStatusColor(props.row.status)"
          text-color="white"
          :label="getStatusText(props.row.status).toUpperCase()"
        />
      </q-td>
    </template>

    <template v-slot:body-cell-source_language="props">
      <q-td :props="props">
        {{ languagesCodeMap[props.row.source_language] }}
      </q-td>
    </template>

    <template v-slot:body-cell-target_languages="props">
      <q-td :props="props">
        {{
          Object.keys(props.row.target_languages)
            .map((lang: string) => languagesCodeMap[lang])
            .join(', ')
        }}
      </q-td>
    </template>

    <template v-slot:body-cell-actions="props">
      <q-td :props="props">
        <q-btn
          flat
          round
          color="primary"
          icon="fas fa-trash"
          @click="deleteMovie(props.row.id)"
        >
          <q-tooltip>{{ $t('Delete from Queue') }}</q-tooltip>
        </q-btn>
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

  <q-dialog
    v-model="showVideoToAudioDialog"
    persistent
  >
    <VideoToAudio
      @onClose="() => {
        showVideoToAudioDialog = false;
      }"
      @onConvert="() => {
        showVideoToAudioDialog = false;
        onRequest({
          pagination: pagination,
          filter: filter,
        });
      }"
    />
  </q-dialog>
</template>
