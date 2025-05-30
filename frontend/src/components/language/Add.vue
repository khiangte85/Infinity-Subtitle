<script setup lang="ts">
  import { ref } from 'vue';
  import { useQuasar } from 'quasar';
  import { backend as models } from '../../../wailsjs/go/models.js';
  import { CreateLanguage } from '../../../wailsjs/go/backend/Language.js';
  import Error from '../Error.vue';

  const $q = useQuasar();
  const emit = defineEmits(['onClose', 'onAdded']);

  const saving = ref(false);

  const model = ref(new models.Language({
    id: 0,
    code: '',
    name: '',
    created_at: '',
  }));

  const errors = ref({});

  async function onSubmit() {
    errors.value = {};
    saving.value = true;
    try {
      await CreateLanguage(model.value.name, model.value.code);

      emit('onAdded');

      $q.notify({
        html: true,
        position: 'bottom',
        type: 'positive',
        icon: 'fas fa-circle-check',
        message: 'created successfully',
      });
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
      <span class="text-body2">{{ $t('Add Language') }}</span>
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
        v-model="model.name"
        :label="$t('Name')"
        dense
        outlined
        maxlength="100"
        lazy-rules
        :rules="[(val) => !!val || 'Field is required']"
      />
    </q-card-section>

    <q-card-section class="q-pt-lg q-pb-none">
      <q-input
        :autofocus="true"
        v-model="model.code"
        :label="$t('Code')"
        dense
        outlined
        maxlength="10"
        lazy-rules
        :rules="[(val) => !!val || 'Field is required']"
      />
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
