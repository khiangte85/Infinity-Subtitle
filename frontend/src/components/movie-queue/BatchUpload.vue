<script setup lang="ts">
  import { ref, computed, onMounted } from 'vue';
  import { useQuasar } from 'quasar';
  import Error from '../Error.vue';
  import { backend } from '../../../wailsjs/go/models';
  import { GetAllLanguages } from '../../../wailsjs/go/backend/Language';

  interface SelectedFile {
    file: File;
    name: string;
    sourceLanguage: string;
    targetLanguage: string;
  }

  const emit = defineEmits<{
    (e: 'onQueue'): void;
    (e: 'onClose'): void;
  }>();

  const $q = useQuasar();
  const saving = ref(false);
  const errors = ref<Record<string, string>>({});
  const files = ref<File[]>([]);
  const selectedFiles = ref<SelectedFile[]>([]);
  const languages = ref<backend.Language[]>([]);

  const validateSourceLanguage = (src: string) => {
    const index = selectedFiles.value.findIndex(
      (file) => file.sourceLanguage === src
    );
    if (index !== -1) {
      const tgt = selectedFiles.value[index].targetLanguage;
      if (src && tgt && src === tgt) {
        selectedFiles.value[index].sourceLanguage = '';
        errors.value[`source_${index}`] =
          'Source and target languages cannot be the same';
      } else {
        delete errors.value[`source_${index}`];
      }
    }
  };

  const validateTargetLanguage = (tgt: string) => {
    const index = selectedFiles.value.findIndex(
      (file) => file.targetLanguage === tgt
    );
    if (index !== -1) {
      const src = selectedFiles.value[index].sourceLanguage;
      if (src && tgt && src === tgt) {
        selectedFiles.value[index].targetLanguage = '';
        errors.value[`target_${index}`] =
          'Source and target languages cannot be the same';
      } else {
        delete errors.value[`target_${index}`];
      }
    }
  };

  const getLanguages = async () => {
    try {
      const response = await GetAllLanguages();
      languages.value = response;
    } catch (error) {
      console.error('Failed to get languages:', error);
    }
  };

  onMounted(() => {
    getLanguages();
  });

  const onFilesSelected = () => {
    if (!files.value || files.value.length === 0) {
      selectedFiles.value = [];
      return;
    }

    selectedFiles.value = files.value.map((file) => ({
      file,
      name: file.name.replace('.srt', ''),
      sourceLanguage: '',
      targetLanguage: '',
    }));
  };

  const canSave = computed(() => {
    if (selectedFiles.value.length === 0) {
      return false;
    }

    return (
      selectedFiles.value.every(
        (file) =>
          file.name &&
          file.sourceLanguage &&
          file.targetLanguage &&
          file.sourceLanguage !== file.targetLanguage
      ) && Object.keys(errors.value).length === 0
    );
  });

  const saveToQueue = async () => {
    if (Object.keys(errors.value).length > 0) {
      $q.notify({
        color: 'negative',
        message: 'Please fix the validation errors before saving',
      });
      return;
    }

    try {
      saving.value = true;
      for (const file of selectedFiles.value) {
        const content = await file.file.text();
        await window.backend.MovieQueue.AddToQueue({
          name: file.name,
          content,
          source_language: file.sourceLanguage,
          target_language: file.targetLanguage,
        });
      }

      $q.notify({
        color: 'positive',
        message: 'Movies added to queue successfully',
      });

      emit('onQueue');
    } catch (error) {
      $q.notify({
        color: 'negative',
        message: 'Failed to add movies to queue',
      });
    } finally {
      saving.value = false;
    }
  };
</script>

<template>
  <q-card
    :style="{
      width: $q.platform.is.mobile ? '100%' : '1000px',
      maxWidth: '100%',
    }"
  >
    <q-bar
      dark
      class="bg-primary text-white q-py-lg"
    >
      <span class="text-body2">Create Batch</span>
      <q-space />
      <q-btn
        dense
        flat
        icon="fas fa-times"
        @click="emit('onClose')"
        :disable="saving"
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

    <q-card-section class="q-mt-lg">
      <q-file
        outlined
        use-chips
        clearable
        v-model="files"
        label="Select SRT files"
        multiple
        append
        accept=".srt"
        @update:model-value="onFilesSelected"
        @clear="onFilesSelected"
      >
        <template v-slot:prepend>
          <q-icon name="attach_file" />
        </template>
      </q-file>

      <q-card
        v-if="selectedFiles.length > 0"
        class="q-mt-md"
      >
        <q-card-section
          v-for="(file, index) in selectedFiles"
          :key="index"
          class="q-mb-md"
        >
          <div class="row q-col-gutter-md">
            <div class="col-6">
              <q-input
                v-model="file.name"
                outlined
                label="Movie Name"
                :rules="[(val) => !!val || 'Movie name is required']"
              />
            </div>
            <div class="col-3">
              <q-select
                outlined
                v-model="file.sourceLanguage"
                :options="languages"
                option-value="code"
                option-label="name"
                emit-value
                map-options
                label="Source Language"
                :rules="[(val) => !!val || 'Source language is required']"
                @update:model-value="validateSourceLanguage"
              />
            </div>
            <div class="col-3">
              <q-select
                outlined
                v-model="file.targetLanguage"
                :options="languages"
                option-value="code"
                option-label="name"
                emit-value
                map-options
                label="Target Language"
                :rules="[(val) => !!val || 'Target language is required']"
                @update:model-value="validateTargetLanguage"
              />
            </div>
          </div>
        </q-card-section>
      </q-card>
    </q-card-section>

    <q-card-actions
      class="q-my-md q-px-md"
      align="right"
    >
      <q-btn
        flat
        label="Close"
        color="negative"
        @click="emit('onClose')"
        :disable="saving"
      />
      <q-btn
        size="md"
        label="Save to Queue"
        color="primary"
        @click="saveToQueue"
        :disable="!canSave"
      />
    </q-card-actions>
  </q-card>
</template>
