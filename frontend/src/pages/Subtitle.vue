<script setup lang="ts">
  import { QTableColumn } from 'quasar';
  import { onMounted, ref, watch } from 'vue';
  import { backend as models } from '../../wailsjs/go/models.js';
  import { GetMovieByID } from '../../wailsjs/go/backend/Movie.js';
  import { GetSubtitlesByMovieID } from '../../wailsjs/go/backend/Subtitle.js';
  import AddMovie from '../components/movie/Add.vue';
  import { GetAllLanguages } from '../../wailsjs/go/backend/Language.js';
  import { useRouter } from 'vue-router';

  const router = useRouter();
  const movieId = router.currentRoute.value.params.id;
  const loading = ref(true);
  const showEdit = ref(false);

  const showAdd = ref(false);
  const movie = ref<models.Movie>();
  const subtitles = ref<models.Subtitle[]>([]);
  const languages = ref<models.Language[]>([]);
  const columns = ref<QTableColumn[]>([]);

  interface SubtitleRow {
    id: number;
    sl_no: number;
    time: string;
    [key: string]: string | number;
  }

  const rows = ref<SubtitleRow[]>([]);

  onMounted(async () => {
    getLanguages();
    getMovie();
    getSubtitles();
  });

  const getMovie = async () => {
    const response = await GetMovieByID(Number(movieId));
    movie.value = response;
    setupColumns();
  };

  const getSubtitles = async () => {
    try {
      const response = await GetSubtitlesByMovieID(Number(movieId));
      subtitles.value = response;
      setupRows();
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
    } catch (error) {
      console.error(error);
    }
  };

  const setupColumns = () => {
    const currentMovie = movie.value;
    if (!currentMovie?.default_language) return;

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
    Object.keys(currentMovie.languages).forEach((code) => {
      if (code !== currentMovie.default_language) {
        tempColumns.push({
          name: code,
          label: currentMovie.languages[code],
          field: code,
          align: 'left' as const,
          sortable: false,
        });
      }
    });

    columns.value.push(...tempColumns);
  };

  const setupRows = () => {
    // if (!subtitles.value || !movie.value) return;
    // rows.value = subtitles.value.map((subtitle) => {
    //   const row: SubtitleRow = {
    //     id: subtitle.id,
    //     sl_no: subtitle.sl_no,
    //     time: `${subtitle.start_time} - ${subtitle.end_time}`,
    //   };
    //   // Add content for each language
    //   Object.keys(movie.value.languages).forEach((code) => {
    //     row[code] = subtitle.content[code] || '';
    //   });
    //   return row;
    // });
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
            showAdd = true;
          }
        "
      >
        <q-tooltip> Import default language subtitle </q-tooltip>
      </q-btn>
    </q-card-section>
  </q-card>

  <q-table
    class="text-left"
    flat
    color="primary"
    bordered
    :columns="columns"
    :rows="rows"
    row-key="id"
    separator="cell"
    wrap-cells
    :pagination="{ rowsPerPage: 50 }"
    :no-data-label="
      loading
        ? 'Loading...'
        : 'No subtitles found, Please import subtitle of default language'
    "
  >
    <template v-slot:body-cell="props">
      <q-td :props="props">
        <template
          v-if="props.col.name === 'sl_no' || props.col.name === 'time'"
        >
          {{ props.value }}
        </template>
        <template v-else>
          <q-input
            v-model="props.row[props.col.name]"
            dense
            borderless
            @update:model-value="
              (val) => onCellEdit(props.row, props.col.name, val)
            "
          />
        </template>
      </q-td>
    </template>
  </q-table>

  <q-dialog v-model="showAdd">
    <AddMovie
      @onClose="showAdd = false"
      @onAdded="
        () => {
          showAdd = false;
        }
      "
    />
  </q-dialog>
</template>
