<script setup lang="ts">
  import { ref, onMounted } from 'vue';
  import { useQuasar } from 'quasar';
  import { backend as models } from '../../../wailsjs/go/models.js';
  import { CreateMovie } from '../../../wailsjs/go/backend/Movie.js';
  import { GetAllLanguages } from '../../../wailsjs/go/backend/Language.js';
  import Error from '../Error.vue';

  const $q = useQuasar();
  const emit = defineEmits(['onClose', 'onAdded']);

  const saving = ref(false);

  const model = ref(
    new models.Movie({
      id: 0,
      title: '',
      default_language: 'en',
      languages: {},
      created_at: '',
    })
  );

  const selectedLanguages = ref<string[]>([]);

  const languages = ref<models.Language[]>([]);

  const errors = ref({});

  onMounted(async () => {
    await getLanguages();
  });

  const getLanguages = async () => {
    try {
      languages.value = await GetAllLanguages();
    } catch (error) {
      console.error(error);
    }
  };

  async function onSubmit() {
    errors.value = {};
    saving.value = true;
    try {
      const subtitleLanguages = languages.value.filter((lang: models.Language) =>{ 
        const code = lang.code;
        return model.value.default_language == code || selectedLanguages.value.includes(code);
      });

      const languagesKV  = subtitleLanguages.reduce((acc: any, lang) => {
        const code = lang.code;
        const value = lang.name;
        acc[code] = value;
        return acc;
      }, {});

      model.value.languages = languagesKV;

      await CreateMovie(
        model.value.title,
        model.value.default_language,
        model.value.languages
      );

      emit('onAdded');
    } catch (err: any) {
      errors.value = { error: err };
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
      <span class="text-body2">Add Movie</span>
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
      <q-input
        :autofocus="true"
        v-model="model.title"
        label="Title"
        dense
        outlined
        maxlength="100"
        lazy-rules
        :rules="[(val) => !!val || 'Field is required']"
      />
    </q-card-section>

    <q-card-section class="q-pb-none">
      <q-select
        :autofocus="true"
        v-model="model.default_language"
        :options="languages"
        label="Default Subtitle Language"
        emit-value
        map-options
        option-label="name"
        option-value="code"
        dense
        outlined
        lazy-rules
        :rules="[(val) => !!val || 'Field is required']"
      />
    </q-card-section>

    <q-card-section class="q-pb-none">
      <div class="text-subtitle2 q-mb-sm">Subtitle Languages</div>
      <div class="row">
        <template
          v-for="(lang, index) in languages"
          :key="lang.code"
        >
          <div
            class="q-mb-xs col-4"
            v-if="lang.code !== model.default_language"
          >
            <q-checkbox
              v-model="selectedLanguages"
              :label="lang.name"
              :val="lang.code"
            />
          </div>
        </template>
      </div>
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
        :disable="saving"
        >Save</q-btn
      >
    </q-card-section>
  </q-card>
</template>
