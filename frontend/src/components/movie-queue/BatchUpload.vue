<script setup lang="ts">
  import { ref, computed, onMounted } from 'vue';
  import { useQuasar } from 'quasar';
  import Error from '../Error.vue';
  import { backend } from '../../../wailsjs/go/models';
  import { AddToQueue } from '../../../wailsjs/go/backend/MovieQueue';
  import { GetAllLanguages } from '../../../wailsjs/go/backend/Language';
  import { EventsEmit } from '../../../wailsjs/runtime';

  interface SelectedFile {
    file: File;
    name: string;
    sourceLanguage: string;
    targetLanguages: string[];
  }

  interface SelectedAudioFile {
    file: File;
    name: string;
    sourceLanguage: string;
    targetLanguages: string[];
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
  const activeTab = ref('subtitle');
  const audioFiles = ref<File[]>([]);
  const selectedAudioFiles = ref<SelectedAudioFile[]>([]);

  onMounted(() => {
    getLanguages();
  });

  const canSave = computed(() => {
    if (selectedFiles.value.length === 0) {
      return false;
    }

    // Check each file's validation
    const allFilesValid = selectedFiles.value.map((file) => {
      if (
        !file.name ||
        !file.sourceLanguage ||
        file.targetLanguages.length === 0
      ) {
        return false;
      }

      let targetLanguages = file.targetLanguages.filter(
        (tgt) => tgt !== file.sourceLanguage
      );
      if (targetLanguages.length === 0) {
        return false;
      }

      return true;
    });

    return (
      allFilesValid.every((valid) => valid) &&
      Object.keys(errors.value).length === 0
    );
  });

  const canSaveAudio = computed(() => {
    if (selectedAudioFiles.value.length === 0) {
      return false;
    }

    return selectedAudioFiles.value.every(
      (file) => 
        file.name && 
        file.sourceLanguage && 
        file.targetLanguages.length > 0 &&
        !file.targetLanguages.includes(file.sourceLanguage)
    );
  });

  const onFilesSelected = () => {
    if (!files.value || files.value.length === 0) {
      selectedFiles.value = [];
      return;
    }

    // Create a map of existing files and their selections
    const existingSelections = new Map(
      selectedFiles.value
        .filter((existingFile) =>
          files.value.some((newFile) => newFile.name === existingFile.file.name)
        )
        .map((file) => [
          file.file.name,
          {
            sourceLanguage: file.sourceLanguage,
            targetLanguages: file.targetLanguages,
          },
        ])
    );

    selectedFiles.value = files.value.map((file) => {
      const existingSelection = existingSelections.get(file.name);
      return {
        file,
        name: file.name.replace('.srt', ''),
        sourceLanguage: existingSelection?.sourceLanguage || '',
        targetLanguages: existingSelection?.targetLanguages || [],
      };
    });
  };

  const onAudioFilesSelected = () => {
    if (!audioFiles.value || audioFiles.value.length === 0) {
      selectedAudioFiles.value = [];
      return;
    }

    selectedAudioFiles.value = audioFiles.value.map((file) => ({
      file,
      name: file.name.replace(/\.[^/.]+$/, ''),
      sourceLanguage: '',
      targetLanguages: [],
    }));
  };

  const validateSourceLanguage = (src: string, index: number, type: 'subtitle' | 'audio') => {
    const files = type === 'subtitle' ? selectedFiles.value : selectedAudioFiles.value;
    if (index !== -1) {
      const targetLangs = files[index].targetLanguages;
      if (src && targetLangs.includes(src)) {
        files[index].sourceLanguage = '';
        errors.value[`${type}_source_${index}`] =
          'Source language cannot be in target languages';
      } else {
        delete errors.value[`${type}_source_${index}`];
      }
    }
  };

  const validateTargetLanguages = (tgt: string[], index: number, type: 'subtitle' | 'audio') => {
    const files = type === 'subtitle' ? selectedFiles.value : selectedAudioFiles.value;
    if (index !== -1) {
      const src = files[index].sourceLanguage;
      if (src && tgt.includes(src)) {
        files[index].targetLanguages = [];
        errors.value[`${type}_target_${index}`] =
          'Target languages cannot include source language';
      } else {
        delete errors.value[`${type}_target_${index}`];
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
      let req: backend.AddToQueueRequest[] = [];

      if (activeTab.value === 'subtitle') {
        req = await Promise.all(
          selectedFiles.value.map(async (file) => ({
            name: file.name,
            type: 'subtitle',
            file_type: 'srt',
            content: await file.file.text(),
            source_language: file.sourceLanguage,
            target_languages: file.targetLanguages,
          }))
        );
      } else {
        // Handle audio files
        req = await Promise.all(
          selectedAudioFiles.value.map(async (file): Promise<backend.AddToQueueRequest> => {
            // Convert audio file to base64
            const arrayBuffer = await file.file.arrayBuffer();
            const base64 = btoa(
              Array.from(new Uint8Array(arrayBuffer))
                .map(byte => String.fromCharCode(byte))
                .join('')
            );
            
            return {
              name: file.name,
              type: 'audio',
              file_type: file.file.name.split('.').pop() || '',
              content: base64,
              source_language: file.sourceLanguage,
              target_languages: file.targetLanguages,
            };
          })
        );
      }

      await AddToQueue(req);
      EventsEmit('on-queue-added', req);

      $q.notify({
        color: 'positive',
        message: 'Files added to queue successfully',
      });

      emit('onQueue');
    } catch (error) {
      console.error('Failed to add files to queue:', error);
      $q.notify({
        color: 'negative',
        message: 'Failed to add files to queue',
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
      <q-tabs
        v-model="activeTab"
        class="text-primary"
        active-color="primary"
        indicator-color="primary"
        align="justify"
      >
        <q-tab name="subtitle" label="Subtitle Files" />
        <q-tab name="audio" label="Audio Files" />
      </q-tabs>

      <q-tab-panels v-model="activeTab" animated>
        <q-tab-panel name="subtitle">
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
              class="q-pb-none"
            >
              <div class="row q-col-gutter-md">
                <div class="col-6">
                  <q-input
                    dense
                    v-model="file.name"
                    outlined
                    label="Movie Name"
                    :rules="[(val) => !!val || 'Movie name is required']"
                  />
                </div>
                <div class="col-3">
                  <q-select
                    dense
                    outlined
                    v-model="file.sourceLanguage"
                    :options="languages"
                    option-value="code"
                    option-label="name"
                    emit-value
                    map-options
                    label="Subtitle Language"
                    :rules="[(val) => !!val || 'Subtitle language is required']"
                    @update:model-value="(val) => validateSourceLanguage(val, index, 'subtitle')"
                  />
                </div>
                <div class="col-3">
                  <q-select
                    dense
                    outlined
                    v-model="file.targetLanguages"
                    :options="languages"
                    option-value="code"
                    option-label="name"
                    emit-value
                    map-options
                    use-chips
                    multiple
                    label="Target Languages"
                    :rules="[
                      (val) =>
                        val.length > 0 ||
                        'At least one target language is required',
                    ]"
                    @update:model-value="(val) => validateTargetLanguages(val, index, 'subtitle')"
                  />
                </div>
              </div>
            </q-card-section>
          </q-card>
        </q-tab-panel>

        <q-tab-panel name="audio">
          <q-file
            outlined
            use-chips
            clearable
            v-model="audioFiles"
            label="Select Audio files"
            multiple
            append
            accept=".mp3,.wav,audio/mp3,audio/mpeg,audio/wav,audio/wave,audio/x-wav"
            @update:model-value="onAudioFilesSelected"
            @clear="onAudioFilesSelected"
          >
            <template v-slot:prepend>
              <q-icon name="audiotrack" />
            </template>
            <template v-slot:append>
              <q-badge class="q-pa-md" color="primary" text-color="white">
               Only mp3, wav, mpeg
              </q-badge>
            </template>
          </q-file>
          <q-card v-if="selectedAudioFiles.length > 0" class="q-mt-md">
            <q-card-section
              v-for="(file, index) in selectedAudioFiles"
              :key="index"
              class="q-pb-none"
            >
              <div class="row q-col-gutter-md">
                <div class="col-6">
                  <q-input
                    dense
                    v-model="file.name"
                    outlined
                    label="Movie Name"
                    :rules="[(val) => !!val || 'Movie name is required']"
                  />
                </div>
                <div class="col-3">
                  <q-select
                    dense
                    outlined
                    v-model="file.sourceLanguage"
                    :options="languages"
                    option-value="code"
                    option-label="name"
                    emit-value
                    map-options
                    label="Audio Language"
                    :rules="[(val) => !!val || 'Audio language is required']"
                    @update:model-value="(val) => validateSourceLanguage(val, index, 'audio')"
                  />
                </div>
                <div class="col-3">
                  <q-select
                    dense
                    outlined
                    v-model="file.targetLanguages"
                    :options="languages"
                    option-value="code"
                    option-label="name"
                    emit-value
                    map-options
                    use-chips
                    multiple
                    label="Target Languages"
                    :rules="[
                      (val) =>
                        val.length > 0 ||
                        'At least one target language is required',
                    ]"
                    @update:model-value="(val) => validateTargetLanguages(val, index, 'audio')"
                  />
                </div>
              </div>
            </q-card-section>
          </q-card>
        </q-tab-panel>
      </q-tab-panels>
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
        label="Add to Queue"
        color="primary"
        @click="saveToQueue"
        :disable="activeTab === 'subtitle' ? !canSave : !canSaveAudio || saving"
      />
    </q-card-actions>
  </q-card>
</template>
