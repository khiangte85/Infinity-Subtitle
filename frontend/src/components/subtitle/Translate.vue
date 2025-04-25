<script setup lang="ts">
  import { ref, computed, onMounted, watch } from 'vue';
  import { backend as models } from '../../../wailsjs/go/models.js';
  import { GetSubtitlesByMovieID, TranslateSubtitles } from '../../../wailsjs/go/backend/Subtitle.js';
  import Error from '../Error.vue';
  import { QTableColumn } from 'quasar';

  const props = defineProps<{
    movie: models.Movie;
  }>();
  const emit = defineEmits(['onClose']);

  const sourceLanguage = ref<string>('');
  const targetLanguage = ref<string>('');
  const errors = ref<{ error?: string }>({});
  const subtitles = ref<models.Subtitle[]>([]);
  const loading = ref(true);
  const columns = ref<QTableColumn[]>([]);
  const rows = ref<any[]>([]);
  const pagination = ref({
    sortBy: 'sl_no',
    descending: false,
    page: 1,
    rowsPerPage: 20,
    rowsNumber: 0,
  });

  const computedPagination = computed({
    get: () => pagination.value,
    set: (val) => {
      pagination.value = val;
    },
  });

  const languageOptions = computed(() => {
    return Object.entries(props.movie.languages).map(([code, name]) => ({
      label: name,
      value: code,
    }));
  });

  const setupColumns = () => {
    if (!sourceLanguage.value) return;

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
        name: sourceLanguage.value,
        label: props.movie.languages[sourceLanguage.value],
        field: sourceLanguage.value,
        align: 'left' as const,
        sortable: false,
      },
    ];

    if (targetLanguage.value) {
      tempColumns.push({
        name: targetLanguage.value,
        label: props.movie.languages[targetLanguage.value],
        field: targetLanguage.value,
        align: 'left' as const,
        sortable: false,
      });
    }

    columns.value = tempColumns;
  };

  const setupRows = () => {
    if (!subtitles.value || !sourceLanguage.value) return;

    rows.value = subtitles.value.map((subtitle) => {
      const row: any = {
        id: subtitle.id,
        sl_no: subtitle.sl_no,
        time: `${subtitle.start_time} - ${subtitle.end_time}`,
      };

      // Add source language content
      row[sourceLanguage.value] = subtitle.content[sourceLanguage.value] || '';

      // Add target language content if selected
      if (targetLanguage.value) {
        row[targetLanguage.value] =
          subtitle.content[targetLanguage.value] || '';
      }

      return row;
    });
  };

  const getSubtitles = async (paginationState: any) => {
    try {
      loading.value = true;
      const response = await GetSubtitlesByMovieID(
        Number(props.movie.id),
        paginationState
      );
      subtitles.value = response.subtitles;
      setupRows();
      return response;
    } catch (error) {
      console.error(error);
      errors.value = { error: 'Failed to load subtitles' };
    } finally {
      loading.value = false;
    }
  };

  const onRequest = async (props: any) => {
    loading.value = true;
    try {
      const { page, rowsPerPage, sortBy, descending } = props.pagination;
      const response = await getSubtitles({
        page,
        rowsPerPage,
        sortBy,
        descending,
        rowsNumber: pagination.value.rowsNumber
      });
      if (response) {
        pagination.value = {
          ...props.pagination,
          rowsNumber: response.pagination.rowsNumber
        };
      }
    } catch (error) {
      console.error(error);
      errors.value = { error: 'Failed to load subtitles' };
    } finally {
      loading.value = false;
    }
  };

  onMounted(async () => {
    try {
      const response = await getSubtitles(pagination.value);
      if (response) {
        pagination.value.rowsNumber = response.pagination.rowsNumber;
      }
      // Set default language as source language
      sourceLanguage.value = props.movie.default_language;
      setupColumns();
      setupRows();
    } catch (error) {
      console.error(error);
      errors.value = { error: 'Failed to load subtitles' };
    } finally {
      loading.value = false;
    }
  });

  const validateSourceLanguage = () => {
    errors.value = {};
    if (!sourceLanguage.value) return false;

    // Check if source language has content in first x subtitles
    const hasContent = subtitles.value.some(
      (subtitle) =>
        subtitle.content[sourceLanguage.value] &&
        subtitle.content[sourceLanguage.value].trim() !== ''
    );

    if (!hasContent) {
      errors.value = {
        error: `No subtitles available in ${
          props.movie.languages[sourceLanguage.value]
        }`,
      };
      sourceLanguage.value = '';
      return false;
    }

    setupColumns();
    setupRows();
    return true;
  };

  const validateTargetLanguage = () => {
    errors.value = {};
    if (sourceLanguage.value === targetLanguage.value) {
      errors.value = {
        error: 'Source and target languages cannot be the same',
      };
      targetLanguage.value = '';
      return false;
    }
    setupColumns();
    setupRows();
    return true;
  };

  const validate = () => {
    return validateSourceLanguage() && validateTargetLanguage();
  };

  const onCellEdit = (row: any, col: string, value: string | number | null) => {
    if (!targetLanguage.value || col !== targetLanguage.value) return;

    // Update the content
    row[col] = String(value || '');
  };

  const onSubmit = async () => {
    try {
      loading.value = true;
      if (!validate()) return;

    // TODO: Implement translation logic
    const response = await TranslateSubtitles(
      Number(props.movie.id),
      sourceLanguage.value,
      targetLanguage.value
    );


    console.log(
      'Translating from',
      sourceLanguage.value,
      'to',
      targetLanguage.value
      );
      console.log(response);
      // emit('onClose');
    } catch (error) {
      console.error(error);
      errors.value = { error: 'Failed to translate subtitles' };
    } finally {
      loading.value = false;
    }
  };
</script>

<template>
  <q-card
    :style="{
      width: '100%',
      height: '100%',
    }"
  >
    <q-bar
      dark
      class="bg-primary text-white q-py-lg"
    >
      <span class="text-body2">Translate Subtitles</span>
      <q-space />
      <q-btn
        dense
        flat
        icon="fas fa-times"
        @click="emit('onClose')"
      >
        <q-tooltip>Close</q-tooltip>
      </q-btn>
    </q-bar>

    <q-card-section
      v-if="Object.keys(errors).length"
      class="q-pb-none"
    >
      <Error :messages="errors" />
    </q-card-section>

    <q-card-section>
      <div class="row q-col-gutter-md">
        <div class="col-12 col-md-6">
          <q-select
            v-model="sourceLanguage"
            :options="languageOptions"
            label="Source Language"
            outlined
            emit-value
            map-options
            :loading="loading"
            @update:model-value="validateSourceLanguage"
          />
        </div>
        <div class="col-12 col-md-6">
          <q-select
            v-model="targetLanguage"
            :options="languageOptions"
            label="Target Language"
            outlined
            emit-value
            map-options
            :loading="loading"
            @update:model-value="validateTargetLanguage"
          />
        </div>
      </div>
    </q-card-section>

    <q-card-section v-if="sourceLanguage">
      <q-table
        class="text-left table-sticky-header"
        flat
        color="primary"
        bordered
        :columns="columns"
        :rows="rows"
        row-key="id"
        separator="cell"
        wrap-cells
        :loading="loading"
        :no-data-label="loading ? 'Loading...' : 'No subtitles found'"
        :resizable-columns="true"
        v-model:pagination="pagination"
        :rows-per-page-options="[20]"
        rows-per-page-label="Records per page"
        :style="{ height: 'calc(100vh - 300px)' }"
        @request="onRequest"
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
              :readonly="props.col.name === sourceLanguage"
              v-model="props.row[props.col.name]"
              dense
              autogrow
              outlined
              @update:model-value="
                (val) => onCellEdit(props.row, props.col.name, val)
              "
              :disable="loading"
            />
          </q-td>
        </template>
      </q-table>
    </q-card-section>

    <q-card-section class="text-right">
      <q-btn
        flat
        color="negative"
        class="q-px-md"
        @click="emit('onClose')"
        :disable="loading"
        >Close</q-btn
      >
      <q-btn
        color="primary"
        class="q-px-md q-ml-md"
        @click="onSubmit"
        :disable="!sourceLanguage || !targetLanguage || loading"
        >Translate</q-btn
      >
    </q-card-section>
  </q-card>
</template>
