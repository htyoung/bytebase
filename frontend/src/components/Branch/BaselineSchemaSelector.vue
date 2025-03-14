<template>
  <div class="contents">
    <EnvironmentSelect
      name="environment"
      :disabled="readonly || props.loading"
      :selected-id="state.environmentId"
      :select-default="false"
      :environment="state.environmentId"
      @update:environment="handleEnvironmentSelect"
    />
    <DatabaseSelect
      style="width: 100%"
      :placeholder="$t('schema-designer.select-database-placeholder')"
      :disabled="readonly || props.loading"
      :allowed-engine-type-list="allowedEngineTypeList"
      :environment="state.environmentId"
      :project="projectId"
      :database="state.databaseId ?? String(UNKNOWN_ID)"
      :fallback-option="false"
      @update:database="handleDatabaseSelect"
    />
  </div>
</template>

<script lang="ts" setup>
import { isNull, isUndefined } from "lodash-es";
import { reactive, watch } from "vue";
import { EnvironmentSelect, DatabaseSelect } from "@/components/v2";
import { useDatabaseV1Store, useEnvironmentV1Store } from "@/store";
import { UNKNOWN_ID } from "@/types";
import { Engine } from "@/types/proto/v1/common";

const props = defineProps<{
  projectId?: string;
  databaseId?: string;
  readonly?: boolean;
  loading?: boolean;
}>();

const emit = defineEmits<{
  (event: "update:database-id", databaseId: string | undefined): void;
}>();

interface LocalState {
  environmentId?: string;
  databaseId?: string;
}

const state = reactive<LocalState>({});
const databaseStore = useDatabaseV1Store();
const environmentStore = useEnvironmentV1Store();

watch(
  () => props.projectId,
  () => {
    const database = isValidId(state.databaseId)
      ? databaseStore.getDatabaseByUID(state.databaseId)
      : undefined;
    if (!database || !props.projectId) {
      return;
    }
    if (database.projectEntity.uid !== props.projectId) {
      state.environmentId = undefined;
      state.databaseId = undefined;
    }
  }
);

watch(
  () => state.databaseId,
  async (databaseId) => {
    const database = isValidId(state.databaseId)
      ? databaseStore.getDatabaseByUID(state.databaseId)
      : undefined;
    try {
      if (database) {
        state.databaseId = database.uid;
        state.environmentId = database.effectiveEnvironmentEntity.uid;
      }
    } catch (error) {
      // do nothing.
    }
    emit("update:database-id", databaseId);
  }
);

const allowedEngineTypeList: Engine[] = [
  Engine.MYSQL,
  Engine.TIDB,
  Engine.POSTGRES,
];

const isValidId = (id: any): id is string => {
  if (isNull(id) || isUndefined(id) || String(id) === String(UNKNOWN_ID)) {
    return false;
  }
  return true;
};

const handleEnvironmentSelect = (environmentId?: string) => {
  if (environmentId !== state.environmentId) {
    state.environmentId = environmentId;
    state.databaseId = undefined;
  }
};

const handleDatabaseSelect = (databaseId?: string) => {
  if (isValidId(databaseId)) {
    const database = databaseStore.getDatabaseByUID(databaseId);
    if (!database) {
      return;
    }

    const environment = environmentStore.getEnvironmentByName(
      database.effectiveEnvironment
    );
    state.environmentId = environment?.uid;
    state.databaseId = databaseId;
  }
};
</script>

<style lang="postcss">
.bb-baseline-schema-select-option .n-base-select-option__content {
  @apply w-full;
}
</style>
