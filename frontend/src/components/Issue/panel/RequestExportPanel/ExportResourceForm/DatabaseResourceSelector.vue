<template>
  <div class="w-[60rem] space-y-2">
    <NTransfer
      v-if="!loading"
      v-model:value="selectedValueList"
      class="select-table-transfer"
      :options="sourceTransferOptions"
      :render-source-list="renderSourceList"
      :render-target-list="renderTargetList"
      :source-filterable="true"
      :source-filter-placeholder="$t('common.filter-by-name')"
    />
    <div
      v-else
      style="height: 512px"
      class="border flex items-center justify-center"
    >
      <BBSpin size="large" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { orderBy } from "lodash-es";
import {
  NTransfer,
  TransferRenderSourceList,
  NTree,
  TreeOption,
} from "naive-ui";
import { computed, h, ref, watch } from "vue";
import {
  useDatabaseV1Store,
  useDBSchemaV1Store,
  useProjectV1Store,
} from "@/store";
import { DatabaseResource } from "@/types";
import Label from "./Label.vue";
import {
  flattenTreeOptions,
  mapTreeOptions,
  DatabaseTreeOption,
} from "./common";

const props = defineProps<{
  projectId: string;
  databaseId: string;
  databaseResources: DatabaseResource[];
}>();

const emit = defineEmits<{
  (e: "update", databaseResourceList: DatabaseResource[]): void;
}>();

const databaseStore = useDatabaseV1Store();
const dbSchemaStore = useDBSchemaV1Store();
const selectedValueList = ref<string[]>(
  props.databaseResources
    .map((databaseResource) => {
      const database = databaseStore.getDatabaseByName(
        databaseResource.databaseName
      );
      if (databaseResource.table !== undefined) {
        return `t-${database.uid}-${databaseResource.schema}-${databaseResource.table}`;
      } else {
        return "";
      }
    })
    .filter((item) => item !== "")
);
const databaseResourceMap = ref<Map<string, DatabaseResource>>(new Map());
const loading = ref(true);

watch(
  () => props.projectId,
  async (projectId) => {
    loading.value = true;

    const project = await useProjectV1Store().getOrFetchProjectByUID(projectId);
    const databaseList = await databaseStore.fetchDatabaseList({
      parent: "instances/-",
      filter: `project == "${project.name}"`,
    });

    for (const database of databaseList) {
      const databaseMetadata = await dbSchemaStore.getOrFetchDatabaseMetadata({
        database: database.name,
      });
      databaseResourceMap.value.set(`d-${database.uid}`, {
        databaseName: database.name,
      });
      for (const schema of databaseMetadata.schemas) {
        databaseResourceMap.value.set(`s-${database.uid}-${schema.name}`, {
          databaseName: database.name,
          schema: schema.name,
        });
        for (const table of schema.tables) {
          databaseResourceMap.value.set(
            `t-${database.uid}-${schema.name}-${table.name}`,
            {
              databaseName: database.name,
              schema: schema.name,
              table: table.name,
            }
          );
        }
      }
    }
    loading.value = false;
  },
  {
    immediate: true,
  }
);

const databaseList = computed(() => {
  const project = useProjectV1Store().getProjectByUID(props.projectId);
  const list = orderBy(
    databaseStore.databaseListByProject(project.name),
    [
      (db) => db.effectiveEnvironmentEntity.order,
      (db) => db.effectiveEnvironmentEntity.title,
      (db) => db.databaseName,
      (db) => db.instanceEntity.title,
    ],
    ["desc", "asc", "asc", "asc"]
  );
  return props.databaseId
    ? list.filter((item) => item.uid === props.databaseId)
    : list;
});

const sourceTreeOptions = computed(() => {
  return mapTreeOptions(databaseList.value);
});

const sourceTransferOptions = computed(() => {
  const options = flattenTreeOptions(sourceTreeOptions.value);
  return options;
});

const renderSourceList: TransferRenderSourceList = ({ onCheck, pattern }) => {
  return h(NTree, {
    keyField: "value",
    checkable: true,
    selectable: false,
    data: sourceTreeOptions.value,
    defaultExpandAll: true,
    blockLine: true,
    style: "height: 428px", // since <NTransfer> height is 512
    renderLabel: ({ option }: { option: TreeOption }) => {
      return h(Label, {
        option: option as DatabaseTreeOption,
        keyword: pattern,
      });
    },
    pattern,
    checkedKeys: selectedValueList.value,
    showIrrelevantNodes: false,
    onUpdateCheckedKeys: (checkedKeys: string[]) => {
      onCheck(checkedKeys);
    },
  });
};

const targetTreeOptions = computed(() => {
  return mapTreeOptions(databaseList.value, selectedValueList.value);
});

const renderTargetList: TransferRenderSourceList = ({ onCheck }) => {
  return h(NTree, {
    keyField: "value",
    checkable: true,
    selectable: false,
    checkOnClick: true,
    defaultExpandAll: true,
    data: targetTreeOptions.value,
    blockLine: true,
    style: "height: 468px", // since <NTransfer> height is 512
    renderLabel: ({ option }: { option: TreeOption }) => {
      return h(Label, {
        option: option as DatabaseTreeOption,
      });
    },
    checkedKeys: selectedValueList.value,
    showIrrelevantNodes: false,
    onUpdateCheckedKeys: (checkedKeys: string[]) => {
      onCheck(checkedKeys);
    },
  });
};

// Clear selectedValueList when projectId changed.
watch(
  () => props.projectId,
  () => {
    selectedValueList.value = [];
  }
);

watch(
  () => [props.projectId, props.databaseId],
  () => {
    if (!props.databaseId) {
      selectedValueList.value = [];
      return;
    }
    const database = databaseStore.getDatabaseByUID(props.databaseId);
    selectedValueList.value = selectedValueList.value.filter((key) =>
      key.startsWith(`t-${database.databaseName}`)
    );
  }
);

watch(selectedValueList, () => {
  emit(
    "update",
    selectedValueList.value.map((key) => databaseResourceMap.value.get(key)!)
  );
});
</script>

<style>
/* Hide the select all button in NTransfer */
.select-table-transfer
  > .n-transfer-list--source
  > .n-transfer-list-header
  > .n-transfer-list-header__button {
  @apply hidden;
}
</style>
