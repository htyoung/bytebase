<template>
  <div v-if="databaseMetadata" class="h-full overflow-hidden flex flex-col">
    <div class="flex items-center justify-between px-2 py-1 border-b gap-x-1">
      <div
        class="flex items-center flex-1 truncate"
        :class="[headerClickable && 'cursor-pointer']"
        @click="handleClickHeader"
      >
        <heroicons-outline:database class="h-4 w-4 mr-1 flex-shrink-0" />
        <span class="text-sm">{{ database.databaseName }}</span>
      </div>
      <div class="flex justify-end gap-x-0.5">
        <HideInStandaloneMode>
          <ExternalLinkButton
            v-if="database.uid !== String(UNKNOWN_ID)"
            :link="databaseV1Url(database)"
            :tooltip="$t('common.detail')"
          />
        </HideInStandaloneMode>
        <HideInStandaloneMode>
          <AlterSchemaButton
            :database="database"
            @click="
              editorEvents.emit('alter-schema', {
                databaseUID: database.uid,
                schema: '',
                table: '',
              })
            "
          />
        </HideInStandaloneMode>
      </div>
    </div>

    <TableList
      class="flex-1 w-full py-1"
      :db="database"
      :database="databaseMetadata"
      :schema-list="availableSchemas"
      :row-clickable="rowClickable"
      @select-table="(schema, table) => $emit('select-table', schema, table)"
      @select-external-table="
        (schema, externalTable) =>
          $emit('select-external-table', schema, externalTable)
      "
    />
  </div>
</template>

<script lang="ts" setup>
import { computed } from "vue";
import { useCurrentUserV1 } from "@/store";
import { UNKNOWN_ID, type ComposedDatabase } from "@/types";
import { Engine } from "@/types/proto/v1/common";
import {
  DatabaseMetadata,
  ExternalTableMetadata,
  SchemaMetadata,
  TableMetadata,
} from "@/types/proto/v1/database_service";
import { databaseV1Url, isTableQueryable } from "@/utils";
import { useSQLEditorContext } from "@/views/sql-editor/context";
import AlterSchemaButton from "./AlterSchemaButton.vue";
import ExternalLinkButton from "./ExternalLinkButton.vue";
import TableList from "./TableList.vue";

const props = defineProps<{
  database: ComposedDatabase;
  databaseMetadata: DatabaseMetadata;
  headerClickable: boolean;
}>();

const emit = defineEmits<{
  (e: "click-header"): void;
  (e: "select-table", schema: SchemaMetadata, table: TableMetadata): void;
  (
    e: "select-external-table",
    schema: SchemaMetadata,
    externalTable: ExternalTableMetadata
  ): void;
}>();

const { events: editorEvents } = useSQLEditorContext();
const currentUser = useCurrentUserV1();

const engine = computed(() => props.database.instanceEntity.engine);

const rowClickable = computed(() => engine.value !== Engine.MONGODB);

const availableSchemas = computed(() => {
  const schemas = props.databaseMetadata.schemas
    .map((schema) => {
      return {
        ...schema,
        tables: schema.tables.filter((table) => {
          return isTableQueryable(
            props.database,
            schema.name,
            table.name,
            currentUser.value
          );
        }),
      };
    })
    .filter((schema) => {
      return schema.tables.length !== 0;
    });
  return schemas;
});

const handleClickHeader = () => {
  if (!props.headerClickable) return;
  emit("click-header");
};
</script>
