<script setup lang="ts">
  import { QTableColumn } from 'quasar';
  import { onMounted, ref } from 'vue';
  import { backend as models } from '../../wailsjs/go/models.js';
  import { GetMovieByID } from '../../wailsjs/go/backend/Movie.js';
  import { GetSubtitlesByMovieID } from '../../wailsjs/go/backend/Subtitle.js';
  import ImportSubtitle from '../components/subtitle/Import.vue';
  import { useRouter } from 'vue-router';

  const router = useRouter();
  const movieId = router.currentRoute.value.params.id;

  const loading = ref(true);
  const showImport = ref(false);

  const movie = ref<models.Movie>();
  const subtitles = ref<models.Subtitle[]>([]);
  const columns = ref<QTableColumn[]>([]);
  const pagination = ref<models.Pagination>({
    sortBy: 'sl_no',
    descending: false,
    page: 1,
    rowsPerPage: 20,
    rowsNumber: 0,
  });

  interface SubtitleRow {
    id: number;
    sl_no: number;
    time: string;
    [key: string]: string | number;
  }

  const rows = ref<SubtitleRow[]>([]);

  onMounted(async () => {
    getMovie();
    onRequest({ pagination: pagination.value });
  });

  const getMovie = async () => {
    const response = await GetMovieByID(Number(movieId));
    movie.value = response;
    setupColumns();
  };

  const getSubtitles = async (props: any) => {
    try {
      const response = await GetSubtitlesByMovieID(
        Number(movieId),
        props.pagination
      );
      subtitles.value = response.subtitles;
      setupRows();
      return response;
    } catch (error) {
      console.error(error);
    } finally {
      loading.value = false;
    }
  };

  const onRequest = async (props: any) => {
    const { page, rowsPerPage, sortBy, descending } = props.pagination;
    const response = await getSubtitles(props);
    if (response) {
      pagination.value.page = page;
      pagination.value.rowsPerPage = rowsPerPage;
      pagination.value.rowsNumber = response.pagination.rowsNumber;
      pagination.value.sortBy = sortBy;
      pagination.value.descending = descending;
    }
  };

  const setupColumns = () => {
    if (!movie.value?.default_language) return;

    let tempColumns: QTableColumn[] = [
      {
        name: 'sl_no',
        label: 'Sl No',
        field: 'sl_no',
        align: 'left' as const,
        sortable: false,
      },
      {
        name: 'time',
        label: 'Time',
        field: 'time',
        align: 'left' as const,
        sortable: false,
      },
      {
        name: movie.value?.default_language || '',
        label:
          movie.value?.languages[movie.value?.default_language] +
            ' (Default)' || '',
        field: movie.value?.default_language || '',
        align: 'left' as const,
        sortable: false,
      },
    ];

    // Add other languages
    Object.keys(movie.value?.languages || {}).forEach((code) => {
      if (code !== movie.value?.default_language) {
        tempColumns.push({
          name: code,
          label: movie.value?.languages[code] || '',
          field: code,
          align: 'left' as const,
          sortable: false,
        });
      }
    });

    columns.value.push(...tempColumns);
  };

  const setupRows = () => {
    if (!subtitles.value || !movie.value) return;

    rows.value = subtitles.value.map((subtitle) => {
      const row: SubtitleRow = {
        id: subtitle.id,
        sl_no: subtitle.sl_no,
        time: `${subtitle.start_time} - ${subtitle.end_time}`,
      };
      // Add content for each language
      Object.keys(movie.value?.languages || {}).forEach((code) => {
        row[code] = subtitle.content[code] || '';
      });
      return row;
    });
  };

  const onCellEdit = async (
    row: SubtitleRow,
    col: string,
    value: string | number | null
  ) => {
    if (!movie.value) return;

    const subtitle = subtitles.value.find((s) => s.id === row.id);
    if (!subtitle) return;

    // Update the content
    subtitle.content[col] = String(value || '');

    // TODO: Call backend to update subtitle
    // await UpdateSubtitle(subtitle);
  };
</script>

<template>
  <q-card
    flat
    class="full-width row justify-between items-center"
  >
    <q-card-section class="q-py-sm q-pl-none">
      <h6 class="text-h6">{{ movie?.title }}'s Subtitles</h6>
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
            showImport = true;
          }
        "
      >
        <q-tooltip> Import default language subtitle </q-tooltip>
      </q-btn>
    </q-card-section>
  </q-card>

  <q-table
    class="text-left "
    flat
    color="primary"
    bordered
    :columns="columns"
    :rows="rows"
    row-key="id"
    separator="cell"
    wrap-cells
    :loading="loading"
    v-model:pagination="pagination"
    :rows-per-page-options="[10, 20, 50, 100]"
    binary-state-sort
    rows-per-page-label="Records per page"
    @request="onRequest"
    :no-data-label="
      loading
        ? 'Loading...'
        : 'No subtitles found, Please import subtitle of default language'
    "
    :resizable-columns="true"
  >
    <template v-slot:body-cell-sl_no="props">
      <q-td
        :props="props"
        :style="{ width: '60px' }"
      >
        {{ props.value }}
      </q-td>
    </template>
    <template v-slot:body-cell-time="props">
      <q-td
        :props="props"
        :style="{ width: '200px' }"
      >
        {{ props.value }}
      </q-td>
    </template>
    <template v-slot:body-cell="props">
      <q-td :props="props">
        <q-input
          :readonly="props.col.name == movie?.default_language"
          v-model="props.row[props.col.name]"
          dense
          autogrow
          outlined
          @update:model-value="
            (val) => onCellEdit(props.row, props.col.name, val)
          "
        />
      </q-td>
    </template>
  </q-table>

  <q-dialog v-model="showImport">
    <ImportSubtitle
      :movie="movie as models.Movie"
      @onClose="showImport = false"
      @onImport="
        () => {
          showImport = false;
          onRequest({ pagination });
        }
      "
    />
  </q-dialog>
</template>
