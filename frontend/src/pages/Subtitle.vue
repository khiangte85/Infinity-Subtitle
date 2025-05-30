<script setup lang="ts">
  import { QTableColumn } from 'quasar';
  import { onMounted, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { backend as models } from '../../wailsjs/go/models.js';
  import { GetMovieByID } from '../../wailsjs/go/backend/Movie.js';
  import {
    GetSubtitlesByMovieID,
    UpdateSubtitle,
    ExportSubtitle as ExportSubtitleAPI,
  } from '../../wailsjs/go/backend/Subtitle.js';
  import ImportSubtitle from '../components/subtitle/Import.vue';
  import TranslateSubtitle from '../components/subtitle/Translate.vue';
  import ExportSubtitle from '../components/subtitle/Export.vue';
  import { useRouter } from 'vue-router';
  import { useQuasar } from 'quasar';

  const { t } = useI18n();
  const $q = useQuasar();
  const router = useRouter();
  const movieId = router.currentRoute.value.params.id;

  const loading = ref(true);
  const showImport = ref(false);
  const showTranslate = ref(false);
  const showExport = ref(false);
  const visibleLanguages = ref<Record<string, boolean>>({});
  const selectedLanguages = ref<string[]>([]);

  const movie = ref<models.Movie>();
  const subtitles = ref<models.Subtitle[]>([]);
  const columns = ref<QTableColumn[]>([]);
  const pagination = ref<models.Pagination>({
    sortBy: 'sl_no',
    descending: false,
    page: 1,
    rowsPerPage: 10,
    rowsNumber: 0,
  });

  interface SubtitleRow {
    row_id: number;
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
    // Initialize visibility for all non-default languages
    if (movie.value?.languages) {
      Object.keys(movie.value.languages).forEach((code) => {
        if (code !== movie.value?.default_language) {
          visibleLanguages.value[code] = true;
          selectedLanguages.value.push(code);
        }
      });
    }
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
    loading.value = true;
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
        label: t('Sl No'),
        field: 'sl_no',
        align: 'left' as const,
        sortable: false,
      },
      {
        name: 'time',
        label: t('Time'),
        field: 'time',
        align: 'left' as const,
        sortable: false,
      },
      {
        name: movie.value?.default_language || '',
        label:
          movie.value?.languages[movie.value?.default_language] +
            ` (${t('Default')})` || '',
        field: movie.value?.default_language || '',
        align: 'left' as const,
        sortable: false,
      },
    ];

    // Add other languages based on visibility
    Object.keys(movie.value?.languages || {}).forEach((code) => {
      if (
        code !== movie.value?.default_language &&
        visibleLanguages.value[code]
      ) {
        tempColumns.push({
          name: code,
          label: movie.value?.languages[code] || '',
          field: code,
          align: 'left' as const,
          sortable: false,
        });
      }
    });

    columns.value = tempColumns;
  };

  const updateLanguageVisibility = (selected: string[]) => {
    // Update all languages to false first
    Object.keys(visibleLanguages.value).forEach((code) => {
      visibleLanguages.value[code] = false;
    });
    // Set selected languages to true
    selected.forEach((code) => {
      visibleLanguages.value[code] = true;
    });
    setupColumns();
  };

  const setupRows = () => {
    if (!subtitles.value || !movie.value) return;

    rows.value = [];

    rows.value = subtitles.value.map((subtitle) => {
      const row: SubtitleRow = {
        row_id: subtitle.id,
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

  const onSubtitleUpdate = async (
    row: SubtitleRow,
    col: string,
    value: string | number | null,
    event: KeyboardEvent
  ) => {
    try {
      if (!movie.value) return;
      console.log(event);
      if (
        !(
          (event.ctrlKey || event.metaKey) &&
          (event.key === 's' || event.key === 'S')
        )
      )
        return;

      const subtitle = subtitles.value.find((s) => s.id === row.row_id);
      if (!subtitle) return;

      if (value === null || value === undefined || value === '') {
        $q.notify({
          message: t('Subtitle cannot be empty'),
          color: 'negative',
          icon: 'fas fa-times',
        });
        return;
      }
      // Update the content
      subtitle.content[col] = String(value || '').trim();
      await UpdateSubtitle(subtitle);

      $q.notify({
        message: t('Subtitle updated successfully'),
        color: 'primary',
        icon: 'fas fa-check',
      });
    } catch (error) {
      $q.notify({
        message: t('Failed to update subtitle'),
        color: 'negative',
        icon: 'fas fa-times',
      });
      console.error(error);
    }
  };
</script>

<template>
  <q-card
    flat
    class="full-width row justify-between items-center"
  >
    <q-card-section class="q-py-none q-pl-none">
      <h6 class="text-h6">{{ movie?.title }}'s {{ $t('subtitles') }}</h6>
    </q-card-section>
    <q-card-section class="q-py-sm">
      <div class="row q-gutter-md justify-end">
        <q-btn
          round
          unelevated
          color="primary"
          icon="fas fa-file-import"
          size="sm"
          @click="showImport = true"
        >
          <q-tooltip>{{ $t('Import default language subtitle') }}</q-tooltip>
        </q-btn>
        <q-btn
          round
          unelevated
          color="primary"
          icon="fas fa-language"
          size="sm"
          @click="showTranslate = true"
        >
          <q-tooltip>{{ $t('Translate subtitles') }}</q-tooltip>
        </q-btn>
        <q-btn
          round
          unelevated
          color="primary"
          icon="fas fa-download"
          size="sm"
          @click="showExport = true"
        >
          <q-tooltip>{{ $t('Export subtitles') }}</q-tooltip>
        </q-btn>
      </div>
    </q-card-section>
  </q-card>

  <q-card
    flat
    class="full-width row justify-between items-center q-pb-sm"
  >
    <div class="row q-gutter-md q-ml-none">
      <q-badge
        class="full-width text-body2 bg-primary text-white q-mb-sm q-pa-sm"
      >
        <q-icon name="fas fa-keyboard" />
        <span class="q-ml-sm">{{ $t('Focus on the subtitle and press Ctrl+S to save or update the subtitle') }}</span>
      </q-badge>
    </div>
    <q-space />
    <div class="q-gutter-md">
      <q-select
        v-if="movie?.languages"
        v-model="selectedLanguages"
        :options="
          Object.entries(movie.languages)
            .filter(([code]) => code !== movie?.default_language)
            .map(([code, name]) => ({ label: name, value: code }))
        "
        :label="$t('Select Languages')"
        emit-value
        map-options
        multiple
        outlined
        dense
        style="min-width: 200px"
        @update:model-value="updateLanguageVisibility"
      >
        <template v-slot:option="{ itemProps, opt, selected, toggleOption }">
          <q-item v-bind="itemProps">
            <q-item-section>
              <q-item-label>{{ opt.label }}</q-item-label>
            </q-item-section>
            <q-item-section side>
              <q-checkbox
                :model-value="selected"
                @update:model-value="toggleOption(opt)"
              />
            </q-item-section>
          </q-item>
        </template>
      </q-select>
    </div>
  </q-card>

  <q-table
    class="text-left table-sticky-header"
    flat
    color="primary"
    bordered
    :columns="columns"
    :rows="rows"
    row-key="row_id"
    separator="cell"
    wrap-cells
    :loading="loading"
    v-model:pagination="pagination"
    :rows-per-page-options="[10, 20, 50]"
    binary-state-sort
    :rows-per-page-label="$t('Records per page')"
    @request="onRequest"
    :no-data-label="
      loading
        ? $t('Loading...')
        : $t('No subtitles found, Please import subtitle of default language')
    "
    :resizable-columns="true"
    :style="{ height: 'calc(100vh - 300px)' }"
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
          type="textarea"
          v-model="props.row[props.col.name]"
          rows="2"
          dense
          outlined
          @keyup.ctrl.s="
            (event: KeyboardEvent) => onSubtitleUpdate(props.row, props.col.name, props.row[props.col.name], event)
          "
          @keyup.meta.s="
            (event: KeyboardEvent) => onSubtitleUpdate(props.row, props.col.name, props.row[props.col.name], event)
          "
        />
      </q-td>
    </template>
  </q-table>

  <q-dialog v-model="showImport">
    <ImportSubtitle
      :movie="movie as models.Movie"
      @onClose="
        () => {
          showImport = false;
          onRequest({ pagination });
        }
      "
      @onImport="
        () => {
          showImport = false;
          onRequest({ pagination });
        }
      "
    />
  </q-dialog>

  <q-dialog
    v-model="showTranslate"
    full-width
    full-height
    persistent
  >
    <TranslateSubtitle
      :movie="movie as models.Movie"
      @onClose="
        () => {
          rows = [];
          showTranslate = false;
          onRequest({ pagination });
        }
      "
      @onTranslate="
        () => {
          rows = [];
          showTranslate = false;
          onRequest({ pagination });
        }
      "
    />
  </q-dialog>

  <q-dialog
    v-model="showExport"
    persistent
  >
    <ExportSubtitle
      :movie="movie as models.Movie"
      @onClose="showExport = false"
      @onExport="() => onRequest({ pagination })"
    />
  </q-dialog>
</template>
