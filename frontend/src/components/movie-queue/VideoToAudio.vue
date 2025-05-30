<script setup lang="ts">
  import { ref } from 'vue';
  import { useQuasar } from 'quasar';
  import { Dialogs } from '@wailsio/runtime';

  const emit = defineEmits<{
    (e: 'onClose'): void;
    (e: 'onConvert', filePath: string): void;
  }>();

  const $q = useQuasar();
  const videoFile = ref<File | null>(null);
  const converting = ref(false);

  const filePath = ref('');

  async function selectFile() {
    const path = await Dialogs.OpenFile({
      AllowsMultipleSelection: false,
      Filters: [
        {
          Pattern: '*.mp4;*.avi;*.mkv;*.mov',
        },
      ],
    });

    if (path) {
      filePath.value = path;
      console.log('Selected file path:', path);
    }
  }

  const convertVideoToAudio = async () => {
    if (!filePath.value) return;

    converting.value = true;
    try {
      emit('onConvert', filePath.value);
      emit('onClose');
      videoFile.value = null;
    } catch (error) {
      console.error('Error converting video:', error);
      $q.notify({
        color: 'negative',
        message: 'Failed to convert video to audio',
      });
    } finally {
      converting.value = false;
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
      <span class="text-body2">{{ $t('Convert Video to Audio') }}</span>
      <q-space />
      <q-btn
        dense
        flat
        icon="fas fa-times"
        @click="emit('onClose')"
        :disable="converting"
      >
        <q-tooltip>{{ $t('Close') }}</q-tooltip>
      </q-btn>
    </q-bar>

    <q-card-section>
      <q-btn
        :label="$t('Select Video File')"
        color="primary"
        @click="selectFile"
      />
    </q-card-section>

    <q-card-actions
      align="right"
      class="q-pa-md"
    >
      <q-btn
        flat
        :label="$t('Cancel')"
        color="negative"
        v-close-popup
        @click="emit('onClose')"
      />
      <q-btn
        :label="$t('Convert')"
        color="primary"
        @click="convertVideoToAudio"
        :loading="converting"
        :disable="!filePath"
      />
    </q-card-actions>
  </q-card>
</template>
