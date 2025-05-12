<script setup lang="ts">
  import { QTableColumn } from 'quasar';
  import { onMounted, ref, watch } from 'vue';
  import { backend as models } from '../../wailsjs/go/models.js';
  import { ListMovies, DeleteMovie } from '../../wailsjs/go/backend/Movie.js';
  import AddMovie from '../components/movie/Add.vue';
  import EditMovie from '../components/movie/Edit.vue';
  import { GetAllLanguages } from '../../wailsjs/go/backend/Language.js';
  import { useRouter } from 'vue-router';
  import { useQuasar } from 'quasar';

  const $q = useQuasar();
  const router = useRouter();
  const loading = ref(true);
  const showEdit = ref(false);
  const showDelete = ref(false);
  const selectedMovie = ref<models.Movie>();
  const pagination = ref<models.Pagination>({
    sortBy: 'created_at',
    descending: true,
    page: 1,
    rowsPerPage: 10,
    rowsNumber: 0,
  });
  const filter = ref({
    title: '',
  });

  const showAdd = ref(false);
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
        const language = languages.value.find(
          (language) => language.code === val
        );
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
        return Object.entries(val)
          .map(([key, value]) => `${value}`)
          .join(', ');
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
    {
      name: 'actions',
      label: 'Actions',
      field: 'actions',
      align: 'left',
      sortable: false,
    },
  ];

  onMounted(async () => {
    getLanguages();
    onRequest({ pagination: pagination.value, filter: filter.value });
  });

  const paginateMovies = async (props: any) => {
    try {
      const response = await ListMovies(props.filter.title, props.pagination);
      movies.value = response.movies ?? [];
      return response;
    } catch (error) {
      console.error(error);
    } finally {
      loading.value = false;
    }
  };

  const onRequest = async (props: any) => {
    const { page, rowsPerPage, sortBy, descending } = props.pagination;
    props.filter = filter.value;
    const response = await paginateMovies(props);
    pagination.value.page = page;
    pagination.value.rowsPerPage = rowsPerPage;
    pagination.value.rowsNumber = response?.pagination.rowsNumber ?? 0;
    pagination.value.sortBy = sortBy;
    pagination.value.descending = descending;
  };

  const getLanguages = async () => {
    try {
      const response = await GetAllLanguages();
      languages.value = response;
    } catch (error) {
      console.error(error);
    } finally {
      loading.value = false;
    }
  };

  const deleteMovie = async (id: number) => {
    try {
      await DeleteMovie(id);
      onRequest({ pagination: pagination.value, filter: filter.value });

      $q.notify({
        color: 'positive',
        message: 'Movie deleted successfully',
      });
    } catch (error) {
      $q.notify({
        color: 'negative',
        message: 'Failed to delete movie',
      });
      console.error(error);
    } finally {
      showDelete.value = false;
    }
  };
</script>

<template>
  <q-card
    flat
    class="full-width row justify-between items-center"
  >
    <q-card-section class="q-py-sm q-pl-none">
      <h5 class="text-h5">Movies</h5>
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
            showAdd = true;
          }
        "
      >
        <q-tooltip> Add Movie </q-tooltip>
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
    <template v-slot:body-cell-actions="props">
      <q-td
        :props="props"
        style="min-width: 120px"
      >
        <q-btn
          round
          unelevated
          color="primary"
          icon="fas fa-edit"
          size="sm"
          @click="
            () => {
              selectedMovie = props.row;
              showEdit = true;
            }
          "
        >
          <q-tooltip>Edit</q-tooltip>
        </q-btn>

        <q-btn
          class="q-ml-sm"
          round
          unelevated
          color="primary"
          icon="fas fa-closed-captioning"
          size="sm"
          :to="`/movies/${props.row.id}/subtitles`"
        >
          <q-tooltip>Subtitles</q-tooltip>
        </q-btn>

        <q-btn
          flat
          round
          color="negative"
          icon="fas fa-trash"
          @click="
            () => {
              selectedMovie = props.row;
              showDelete = true;
            }
          "
        >
          <q-tooltip>Delete</q-tooltip>
        </q-btn>
      </q-td>
    </template>
  </q-table>

  <q-dialog v-model="showAdd">
    <AddMovie
      @onClose="showAdd = false"
      @onAdded="
        () => {
          showAdd = false;
          onRequest({ pagination: { ...pagination }, filter: { ...filter } });
        }
      "
    />
  </q-dialog>

  <q-dialog v-model="showEdit">
    <EditMovie
      :movie="selectedMovie as models.Movie"
      @onClose="showEdit = false"
      @onUpdated="
        () => {
          showEdit = false;
          onRequest({ pagination: { ...pagination }, filter: { ...filter } });
        }
      "
    />
  </q-dialog>

  <q-dialog v-model="showDelete">
    <q-card class="q-pa-sm">
      <q-card-section>
        <div class="text-h6">Delete Movie</div>
        <div class="text-body">Are you sure you want to delete this movie?</div>
      </q-card-section>
      <q-card-actions align="right">
        <q-btn
          label="Cancel"
          flat
          size="md"
          color="negative"
          @click="showDelete = false"
        />
        <q-btn
          color="negative"
          label="Delete"
          unelevated
          size="md"
          @click="deleteMovie(selectedMovie?.id ?? 0)"
        />
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>
