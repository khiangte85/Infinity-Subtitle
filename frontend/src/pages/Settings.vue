<script setup lang="ts">
  import { ref, onMounted } from 'vue';
  import { useQuasar } from 'quasar';
  import { GetOpenAIKey, SaveOpenAIKey } from '../../wailsjs/go/backend/Setting';

  const $q = useQuasar();
  const loading = ref(false);
  const apiKey = ref('');
  const showKey = ref(false);

  onMounted(async () => {
    try {
      loading.value = true;
      const key = await GetOpenAIKey();
      apiKey.value = key;
    } catch (error) {
      console.error(error);
      $q.notify({
        message: 'Failed to load API key',
        color: 'negative',
        icon: 'fas fa-times',
      });
    } finally {
      loading.value = false;
    }
  });

  const saveKey = async () => {
    try {
      loading.value = true;
      await SaveOpenAIKey(apiKey.value);
      $q.notify({
        message: 'API key saved successfully',
        color: 'primary',
        icon: 'fas fa-check',
      });
    } catch (error) {
      console.error(error);
      $q.notify({
        message: 'Failed to save API key',
        color: 'negative',
        icon: 'fas fa-times',
      });
    } finally {
      loading.value = false;
    }
  };
</script>

<template>
  <q-card flat>
    <q-card-section>
      <h6 class="text-h6">Settings</h6>
    </q-card-section>

    <q-card-section>
      <div class="text-subtitle2 q-mb-sm">OpenAI API Key</div>
      <div class="row q-col-gutter-md">
        <div class="col-12 col-md-6">
          <q-input
            v-model="apiKey"
            label="API Key"
            outlined
            :type="showKey ? 'text' : 'password'"
            :loading="loading"
          >
            <template v-slot:append>
              <q-icon
                :name="showKey ? 'fas fa-eye-slash' : 'fas fa-eye'"
                class="cursor-pointer"
                @click="showKey = !showKey"
              >
                <q-tooltip>{{ showKey ? 'Hide' : 'Show' }} API Key</q-tooltip>
              </q-icon>
            </template>
          </q-input>
          <div class="text-caption text-grey q-mt-sm">
            Your OpenAI API key will be stored locally in the .env file.
          </div>
        </div>
      </div>
    </q-card-section>

    <q-card-section>
      <q-btn
        color="primary"
        :loading="loading"
        @click="saveKey"
        :disable="!apiKey"
      >
        Save Changes
      </q-btn>
    </q-card-section>
  </q-card>
</template>
