<script setup lang="ts">
  import { ref, computed } from 'vue';
  import { useQuasar } from 'quasar';
  import { UploadSRTFile } from '../../../wailsjs/go/backend/Subtitle.js';
  import { backend as models } from '../../../wailsjs/go/models.js';
  import Error from '../Error.vue';

  const $q = useQuasar();
  const props = defineProps<{
    movie: models.Movie;
  }>();
  const emit = defineEmits(['onClose', 'onImport']);

  const saving = ref(false);
  const selectedFile = ref<File | null>(null);
  const errors = ref<{ error?: string }>({});

  const isImportDisabled = computed(() => {
    return !selectedFile.value || !selectedFile.value.name.endsWith('.srt');
  });

  const handleFileSelect = (file: File | null) => {
    if (file) {
      if (file.name.endsWith('.srt')) {
        selectedFile.value = file;
        errors.value = {};
      } else {
        selectedFile.value = null;
        errors.value = { 
          error: 'Invalid file type. Please select a .srt file' 
        };
      }
    } else {
      selectedFile.value = null;
      errors.value = {};
    }
  };

  async function onSubmit() {
    if (!selectedFile.value) {
      errors.value = { 
        error: 'Please select a .srt file' 
      };
      return;
    }

    if (!selectedFile.value.name.endsWith('.srt')) {
      errors.value = { 
        error: 'Invalid file type. Please select a .srt file' 
      };
      return;
    }

    errors.value = {};
    saving.value = true;
    try {
      // Read file content
      const content = await selectedFile.value.text();
      
      // Call backend function
      const response = await UploadSRTFile(props.movie, content);
      console.log(response);

      emit('onImport');
      emit('onClose');
    } catch (err: any) {
      errors.value = { 
        error: err.message || 'Failed to upload subtitle file' 
      };
    } finally {
      saving.value = false;
    }
  }
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
      <span class="text-body2">Import Subtitle</span>
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

    <q-card-section class="q-pb-none">
      <div class="text-subtitle2 q-mb-sm">Subtitle File (.srt)</div>
      <q-file
        v-model="selectedFile"
        label="Select .srt file"
        accept=".srt"
        outlined
        :error="Object.keys(errors).length > 0"
        :error-message="errors.error"
        @update:model-value="handleFileSelect"
        clearable
      >
        <template v-slot:prepend>
          <q-icon name="fas fa-file" />
        </template>
      </q-file>
    </q-card-section>

    <q-card-section class="text-right q-mt-md">
      <q-btn
        flat
        color="negative"
        class="q-px-md"
        @click="emit('onClose')"
        :disable="saving"
        >Cancel</q-btn
      >
      <q-btn
        color="primary"
        class="q-px-md q-ml-md"
        @click="onSubmit"
        :disable="saving || isImportDisabled"
        >Import</q-btn
      >
    </q-card-section>
  </q-card>
</template>
