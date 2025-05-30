<script setup lang="ts">
  import { ref, onMounted } from 'vue';
  import { useQuasar } from 'quasar';
  import { useI18n } from 'vue-i18n';
  import { backend as models } from '../../../wailsjs/go/models.js';
  import { CreateMovie, UpdateMovie } from '../../../wailsjs/go/backend/Movie.js';
  import { GetAllLanguages } from '../../../wailsjs/go/backend/Language.js';
  import Error from '../Error.vue';

  const { t } = useI18n();
  const $q = useQuasar();
  const emit = defineEmits(['onClose', 'onUpdated']);
  const props = defineProps<{
    movie: models.Movie;
  }>();

  console.log(Object.keys(props.movie.languages));

  const saving = ref(false);

  const model = ref<models.Movie>(new models.Movie({
    ...props.movie,
    languages: Object.keys(props.movie.languages),
  }));

  const selectedLanguages = ref<string[]>([...Object.keys(props.movie.languages)]);

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

      await UpdateMovie(model.value);

      emit('onUpdated');
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
      <span class="text-body2">{{ $t('Edit Movie') }}</span>
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
      <q-input
        :autofocus="true"
        v-model="model.title"
        :label="$t('Title')"
        dense
        outlined
        maxlength="100"
        lazy-rules
        :rules="[(val) => !!val || $t('Field is required')]"
      />
    </q-card-section>

    <q-card-section class="q-pb-none">
      <q-select
        :disabled="true"
        :readonly="true"
        :autofocus="true"
        v-model="model.default_language"
        :options="languages"
        :label="$t('Default Language')"
        emit-value
        map-options
        option-label="name"
        option-value="code"
        dense
        outlined
        lazy-rules
        :rules="[(val) => !!val || $t('Field is required')]"
      />
    </q-card-section>

    <q-card-section class="q-pb-none">
      <div class="text-subtitle2 q-mb-sm">{{ $t('Subtitle Languages') }}</div>
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
        >{{ $t('Cancel') }}</q-btn
      >
      <q-btn
        color="primary"
        class="q-px-md q-ml-md"
        @click="onSubmit"
        :disable="saving"
        >{{ $t('Save') }}</q-btn
      >
    </q-card-section>
  </q-card>
</template>
