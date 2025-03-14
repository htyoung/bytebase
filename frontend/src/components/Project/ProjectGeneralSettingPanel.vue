<template>
  <form class="w-full space-y-4 mx-auto">
    <p class="text-lg font-medium leading-7 text-main">
      {{ $t("common.general") }}
    </p>
    <div class="flex justify-start items-start gap-12">
      <dl class="">
        <dt class="text-sm font-medium text-control-light">
          {{ $t("common.name") }} <span class="text-red-600">*</span>
        </dt>
        <dd class="mt-1 text-sm text-main">
          <NInput
            id="projectName"
            v-model:value="state.title"
            :disabled="!allowEdit"
            required
          />
        </dd>
        <div class="mt-1">
          <ResourceIdField
            resource-type="project"
            :value="extractProjectResourceName(project.name)"
            :readonly="true"
          />
        </div>
      </dl>

      <dl class="">
        <dt class="flex text-sm font-medium text-control-light">
          {{ $t("common.key") }}
          <NTooltip>
            <template #trigger>
              <heroicons-outline:information-circle class="ml-1 w-4 h-auto" />
            </template>
            {{ $t("project.key-hint") }}
          </NTooltip>
          <span class="text-red-600">*</span>
        </dt>
        <dd class="mt-1 text-sm text-main">
          <NInput
            id="projectKey"
            v-model:value="state.key"
            :disabled="!allowEdit"
            required
            @update:value="(val: string) => state.key = val.toUpperCase()"
          />
        </dd>
      </dl>
    </div>

    <div class="flex flex-col">
      <div for="name" class="text-sm font-medium text-control-light">
        {{ $t("common.mode") }}
        <span class="text-red-600">*</span>
      </div>
      <div class="mt-2 textlabel">
        <ProjectModeRadioGroup
          v-model:value="state.tenantMode"
          :disabled="!allowEdit"
        />
      </div>
    </div>

    <div v-if="allowEdit" class="flex justify-end">
      <NButton type="primary" :disabled="!allowSave" @click.prevent="save">
        {{ $t("common.update") }}
      </NButton>
    </div>

    <FeatureModal
      :open="!!state.requiredFeature"
      :feature="state.requiredFeature"
      @cancel="state.requiredFeature = undefined"
    />
  </form>
</template>

<script lang="ts" setup>
import { cloneDeep, isEmpty } from "lodash-es";
import { NTooltip } from "naive-ui";
import { computed, reactive } from "vue";
import { useI18n } from "vue-i18n";
import ResourceIdField from "@/components/v2/Form/ResourceIdField.vue";
import { hasFeature, pushNotification, useProjectV1Store } from "@/store";
import { DEFAULT_PROJECT_ID, FeatureType } from "@/types";
import { Project, TenantMode } from "@/types/proto/v1/project_service";
import { extractProjectResourceName } from "@/utils";

interface LocalState {
  title: string;
  key: string;
  tenantMode: TenantMode;
  requiredFeature: FeatureType | undefined;
}

const props = defineProps<{
  project: Project;
  allowEdit: boolean;
}>();

const { t } = useI18n();
const projectV1Store = useProjectV1Store();

const state = reactive<LocalState>({
  title: props.project.title,
  key: props.project.key,
  tenantMode: props.project.tenantMode,
  requiredFeature: undefined,
});

const allowSave = computed((): boolean => {
  return (
    parseInt(props.project.uid, 10) !== DEFAULT_PROJECT_ID &&
    !isEmpty(state.title) &&
    (state.title !== props.project.title ||
      state.key !== props.project.key ||
      state.tenantMode !== props.project.tenantMode)
  );
});

const save = () => {
  const projectPatch = cloneDeep(props.project);
  const updateMask: string[] = [];
  if (state.title !== props.project.title) {
    projectPatch.title = state.title;
    updateMask.push("title");
  }
  if (state.key !== props.project.key) {
    projectPatch.key = state.key;
    updateMask.push("key");
  }
  if (state.tenantMode !== props.project.tenantMode) {
    if (state.tenantMode === TenantMode.TENANT_MODE_ENABLED) {
      if (!hasFeature("bb.feature.multi-tenancy")) {
        state.tenantMode = TenantMode.TENANT_MODE_DISABLED;
        state.requiredFeature = "bb.feature.multi-tenancy";
        return;
      }
    }
    projectPatch.tenantMode = state.tenantMode;
    updateMask.push("tenant_mode");
  }
  projectV1Store.updateProject(projectPatch, updateMask).then((updated) => {
    pushNotification({
      module: "bytebase",
      style: "SUCCESS",
      title: t("project.settings.success-updated"),
    });
    state.title = updated.title;
    state.key = updated.key;
    state.tenantMode = updated.tenantMode;
  });
};
</script>
