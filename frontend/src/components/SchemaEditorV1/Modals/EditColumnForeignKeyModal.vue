<template>
  <BBModal
    :title="$t('schema-editor.edit-foreign-key')"
    class="shadow-inner outline outline-gray-200"
    @close="dismissModal"
  >
    <div v-if="shouldShowSchemaSelector" class="w-72">
      <p class="mb-2">{{ $t("schema-editor.select-reference-schema") }}</p>
      <NSelect
        v-model:value="state.referencedSchemaId"
        :placeholder="$t('schema-editor.schema.select')"
        :options="schemas.map((item) => ({ value: item.id, label: item.name }))"
      />
    </div>
    <div class="w-72">
      <p class="mt-4 mb-2">{{ $t("schema-editor.select-reference-table") }}</p>
      <NSelect
        v-model:value="state.referencedTableId"
        :placeholder="$t('schema-editor.table.select')"
        :options="
          tableList.map((item) => ({ value: item.id, label: item.name }))
        "
      />
    </div>
    <div class="w-72">
      <p class="mt-4 mb-2">{{ $t("schema-editor.select-reference-column") }}</p>
      <NSelect
        v-model:value="state.referencedColumnId"
        :placeholder="$t('schema-editor.column.select')"
        :options="
          columnList.map((item) => ({ value: item.id, label: item.name }))
        "
      />
    </div>
    <div
      class="w-full flex items-center justify-between mt-6 space-x-2 pr-1 pb-1"
    >
      <div class="flex items-center justify-start">
        <button
          v-if="foreignKey !== undefined"
          type="button"
          class="btn-danger"
          @click="handleRemoveFKButtonClick"
        >
          {{ $t("common.delete") }}
        </button>
      </div>
      <div class="flex items-center justify-end space-x-2">
        <button type="button" class="btn-normal" @click="dismissModal">
          {{ $t("common.cancel") }}
        </button>
        <button class="btn-primary" @click="handleConfirmButtonClick">
          {{ $t("common.save") }}
        </button>
      </div>
    </div>
  </BBModal>
</template>

<script lang="ts" setup>
import { isUndefined } from "lodash-es";
import { NSelect } from "naive-ui";
import { computed, onMounted, reactive, watch } from "vue";
import { BBModal } from "@/bbkit";
import { useSchemaEditorV1Store } from "@/store";
import { Engine } from "@/types/proto/v1/common";
import { Column, ForeignKey } from "@/types/v1/schemaEditor";

interface LocalState {
  referencedSchemaId?: string;
  referencedTableId?: string;
  referencedColumnId?: string;
}

const props = defineProps<{
  parentName: string;
  schemaId: string;
  tableId: string;
  columnId: string;
}>();

const emit = defineEmits<{
  (event: "close"): void;
}>();

const schemaEditorV1Store = useSchemaEditorV1Store();
const state = reactive<LocalState>({
  referencedSchemaId: props.schemaId,
});

const engine = computed(() => {
  return schemaEditorV1Store.getCurrentEngine(props.parentName);
});

const parentResource = computed(() => {
  return schemaEditorV1Store.resourceMap[schemaEditorV1Store.resourceType].get(
    props.parentName
  );
});

const schemas = computed(() => {
  return parentResource.value?.schemaList || [];
});

const schema = computed(() => {
  return schemaEditorV1Store.getSchema(props.parentName, props.schemaId);
});

const table = computed(() => {
  return schema.value?.tableList.find((item) => item.id === props.tableId);
});

const propsColumn = computed(() => {
  return table.value?.columnList.find((item) => item.id === props.columnId);
});

const foreignKeyList = computed(() => {
  return table.value?.foreignKeyList || [];
});

const shouldShowSchemaSelector = computed(() => {
  return engine.value === Engine.POSTGRES;
});

const selectedSchema = computed(() => {
  return parentResource.value?.schemaList.find(
    (schema) => schema.id === state.referencedSchemaId
  );
});

const tableList = computed(() => {
  return selectedSchema.value?.tableList || [];
});

const selectedTable = computed(() => {
  return tableList.value.find((table) => table.id === state.referencedTableId);
});

const columnList = computed(() => {
  if (!selectedTable.value) {
    return [];
  }

  return selectedTable.value.columnList.filter(
    (column) =>
      column.id !== props.columnId &&
      column.type.toUpperCase() === propsColumn.value?.type.toUpperCase()
  );
});

const foreignKey = computed(() => {
  for (const fk of foreignKeyList.value) {
    const foundIndex = fk.columnIdList.findIndex((id) => id === props.columnId);
    if (foundIndex > -1) {
      return fk;
    }
  }
  return undefined;
});

onMounted(() => {
  if (foreignKey.value) {
    const foundIndex = foreignKey.value.columnIdList.findIndex(
      (id) => id === props.columnId
    );
    if (foundIndex > -1) {
      state.referencedSchemaId = foreignKey.value.referencedSchemaId;
      state.referencedTableId = foreignKey.value.referencedTableId;
      state.referencedColumnId =
        foreignKey.value.referencedColumnIdList[foundIndex];
    }
  }
});

watch(
  () => state.referencedTableId,
  () => {
    const found = columnList.value.find(
      (column) => column.id === state.referencedColumnId
    );
    if (!found) {
      state.referencedColumnId = undefined;
    }
  }
);

const handleRemoveFKButtonClick = async () => {
  if (isUndefined(foreignKey.value)) {
    return;
  }

  const index = foreignKey.value.columnIdList.findIndex(
    (id) => id === propsColumn.value?.id
  );
  if (index > -1) {
    foreignKey.value.referencedColumnIdList.splice(index, 1);
    foreignKey.value.columnIdList.splice(index, 1);
    dismissModal();
  }
};

const handleConfirmButtonClick = async () => {
  if (
    isUndefined(state.referencedSchemaId) ||
    isUndefined(state.referencedTableId) ||
    isUndefined(state.referencedColumnId)
  ) {
    return;
  }

  const column = propsColumn.value as Column;
  if (isUndefined(foreignKey.value)) {
    const fk: ForeignKey = {
      // NOTE: keep it empty, we will generate a formated name in `mergeSchemaEditToMetadata`.
      name: "",
      tableId: props.tableId,
      columnIdList: [column.id],
      referencedSchemaId: state.referencedSchemaId,
      referencedTableId: state.referencedTableId,
      referencedColumnIdList: [state.referencedColumnId],
    };
    table.value?.foreignKeyList.push(fk);
  } else {
    const index = foreignKey.value.columnIdList.findIndex(
      (id) => id === column.id
    );
    if (index >= 0) {
      foreignKey.value.referencedColumnIdList[index] = state.referencedColumnId;
      foreignKey.value.columnIdList[index] = column.id;
    } else {
      foreignKey.value.referencedColumnIdList.push(state.referencedColumnId);
      foreignKey.value.columnIdList.push(column.id);
    }
  }
  dismissModal();
};

const dismissModal = () => {
  emit("close");
};
</script>
