<script setup lang="ts">
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { backend as models } from '../../../wailsjs/go/models.js';
  import { ExportSubtitle as ExportSubtitleAPI } from '../../../wailsjs/go/backend/Subtitle.js';
  import { useQuasar } from 'quasar';
  import Error from '../Error.vue';

  const { t } = useI18n();

  declare global {
    interface Window {
      runtime: {
        BrowserOpenFile: (path: string) => void;
        BrowserOpenURL: (url: string) => void;
      };
    }
  }

  interface ExportResponse {
    file_path: string;
  }

  const props = defineProps<{
    movie: models.Movie;
  }>();

  const emit = defineEmits<{
    (e: 'onClose'): void;
    (e: 'onExport'): void;
  }>();

  const $q = useQuasar();
  const loading = ref(false);
  const language = ref<string>('');
  const filePath = ref<string>('');
  const showSuccess = ref(false);
  const errors = ref<{ error?: string }>({});

  const onExport = async () => {
    if (!language.value) {
      errors.value = {
        error: t('Please select a language'),
      };
      return;
    }
    try {
      loading.value = true;
      const response = await ExportSubtitleAPI(
        Number(props.movie.id),
        language.value
      ) as ExportResponse;
      filePath.value = response.file_path;
      $q.notify({
        message: t('Subtitles exported successfully'),
        color: 'primary',
        icon: 'fas fa-check',
        timeout: 3000,
      });
    } catch (error) {
      console.error(error);
      errors.value = {
        error: t('Failed to export subtitles'),
      };
    } finally {
      loading.value = false;
    }
  };

  const openFileLocation = () => {
    if (filePath.value) {
      const directory = filePath.value.substring(0, filePath.value.lastIndexOf('/'));
      window.runtime.BrowserOpenURL(`file://${directory}`);
    }
  };
</script>

<template>
  <q-card
    :style="{
      width: $q.platform.is.mobile ? '100%' : '600px',
      maxWidth: '100%',
    }"
  >
    <q-bar
      dark
      class="bg-primary text-white q-py-lg"
    >
      <span class="text-body2">{{ $t('Export Subtitle') }}</span>
      <q-space />
      <q-btn
        dense
        flat
        icon="fas fa-times"
        @click="emit('onClose')"
      >
        <q-tooltip>{{ $t('Close') }}</q-tooltip>
      </q-btn>
    </q-bar>

    <q-card-section
      v-if="Object.keys(errors).length"
      class="q-pb-none"
    >
      <Error :messages="errors" />
    </q-card-section>

    <q-card-section class="q-pb-none">
      <div class="text-subtitle2 q-mb-sm">{{ $t('Select Language') }}</div>
      <q-select
        v-model="language"
        :options="
          Object.entries(movie.languages || {}).map(([code, name]) => ({
            label: name,
            value: code,
          }))
        "
        :label="$t('Language')"
        outlined
        emit-value
        map-options
        :error="Object.keys(errors).length > 0"
        :error-message="errors.error"
      >
        <template v-slot:prepend>
          <q-icon name="fas fa-language" />
        </template>
      </q-select>
    </q-card-section>

    <q-card-section
      v-if="filePath"
      class="q-pb-none"
    >
      <div class="text-subtitle2 q-mb-sm">{{ $t('Export Location') }}</div>
      <div class="row items-center">
        <div class="col text-caption text-grey text-break">
          <div class="text-weight-medium">{{ filePath }}</div>
        </div>
        <div class="col-auto">
          <q-btn
            flat
            dense
            icon="fas fa-folder-open"
            @click="openFileLocation"
          >
            <q-tooltip>{{ $t('Open File Location') }}</q-tooltip>
          </q-btn>
        </div>
      </div>
    </q-card-section>

    <q-card-section class="text-right q-mt-md">
      <q-btn
        flat
        color="negative"
        class="q-px-md"
        @click="emit('onClose')"
        :disable="loading"
        >{{ $t('Close') }}</q-btn
      >
      <q-btn
        color="primary"
        class="q-px-md q-ml-md"
        @click="onExport"
        :disable="!language || loading"
        :loading="loading"
        >{{ $t('Export') }}</q-btn
      >
    </q-card-section>
  </q-card>

  <q-dialog v-model="showSuccess">
    <q-card
      :style="{
        width: $q.platform.is.mobile ? '100%' : '600px',
        maxWidth: '100%',
      }"
    >
      <q-bar
        dark
        class="bg-primary text-white q-py-lg"
      >
        <span class="text-body2">{{ $t('Export Complete') }}</span>
        <q-space />
        <q-btn
          dense
          flat
          icon="fas fa-times"
          v-close-popup
        >
          <q-tooltip>{{ $t('Close') }}</q-tooltip>
        </q-btn>
      </q-bar>

      <q-card-section>
        <div class="text-body2">
          {{ $t('Subtitles have been exported successfully.') }}
        </div>
        <div class="text-caption text-grey q-mt-sm">
          {{ filePath }}
        </div>
      </q-card-section>

      <q-card-section class="text-right">
        <q-btn
          flat
          color="negative"
          class="q-px-md"
          v-close-popup
          >{{ $t('Close') }}</q-btn
        >
        <q-btn
          color="primary"
          class="q-px-md q-ml-md"
          @click="openFileLocation"
          >{{ $t('Open File Location') }}</q-btn
        >
      </q-card-section>
    </q-card>
  </q-dialog>
</template> 