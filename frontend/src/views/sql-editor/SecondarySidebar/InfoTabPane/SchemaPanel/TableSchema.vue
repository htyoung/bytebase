<template>
  <div class="overflow-hidden flex flex-col">
    <div class="flex items-center justify-between pl-4 pr-2 py-1 border-b">
      <div
        class="flex items-center flex-1 truncate cursor-pointer"
        @click="emit('close')"
      >
        <heroicons-outline:table class="h-4 w-4 mr-1 flex-shrink-0" />
        <span v-if="schema.name" class="text-sm">{{ schema.name }}.</span>
        <span class="text-sm">{{ table.name }}</span>
      </div>

      <div class="flex justify-end gap-x-0.5">
        <StringifyMetadataButton
          v-if="showStringifyMetadataButton"
          :database="db"
          :schema="schema"
          :table="table"
        />
        <HideInStandaloneMode>
          <ExternalLinkButton
            v-if="db.uid !== String(UNKNOWN_ID)"
            :link="tableDetailLink"
            :tooltip="$t('common.detail')"
          />
          <AlterSchemaButton
            :database="db"
            :schema="schema"
            :table="table"
            @click="
              editorEvents.emit('alter-schema', {
                databaseUID: db.uid,
                schema: schema.name,
                table: table.name,
              })
            "
          />
        </HideInStandaloneMode>
      </div>
    </div>

    <ColumnList
      :db="db"
      :database="database"
      :schema="schema"
      :columns="table.columns"
      class="w-full flex-1 py-1"
    />
  </div>
</template>

<script lang="ts" setup>
import { stringify } from "qs";
import { computed } from "vue";
import { UNKNOWN_ID, type ComposedDatabase } from "@/types";
import { Engine } from "@/types/proto/v1/common";
import type {
  DatabaseMetadata,
  SchemaMetadata,
  TableMetadata,
} from "@/types/proto/v1/database_service";
import { databaseV1Url } from "@/utils";
import { useSQLEditorContext } from "@/views/sql-editor/context";
import AlterSchemaButton from "./AlterSchemaButton.vue";
import ColumnList from "./ColumnList.vue";
import ExternalLinkButton from "./ExternalLinkButton.vue";
import StringifyMetadataButton from "./StringifyMetadataButton.vue";

const props = defineProps<{
  db: ComposedDatabase;
  database: DatabaseMetadata;
  schema: SchemaMetadata;
  table: TableMetadata;
}>();

const emit = defineEmits<{
  (e: "close"): void;
}>();

const { events: editorEvents } = useSQLEditorContext();

const tableDetailLink = computed((): string => {
  const { db: database, schema, table } = props;
  const query: Record<string, string> = {
    table: table.name,
  };
  if (schema.name) {
    query.schema = schema.name;
  }
  const url = `${databaseV1Url(database)}?${stringify(query)}`;

  return url;
});

const showStringifyMetadataButton = computed(() => {
  const engine = props.db.instanceEntity.engine;
  return [Engine.MYSQL, Engine.POSTGRES, Engine.TIDB].includes(engine);
});
</script>
