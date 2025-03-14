<template>
  <div class="w-[32rem] h-[20rem] flex flex-col gap-y-2">
    <div class="flex flex-row content-justify items-center gap-x-2">
      <div class="textlabel flex-1 truncate">
        {{ name }}
      </div>
      <div class="flex flex-row justify-end items-center">
        <NCheckbox v-if="!formatted.error" v-model:checked="format">
          {{ $t("sql-editor.format") }}
        </NCheckbox>
      </div>
    </div>
    <div class="w-full flex-1 relative">
      <MonacoEditor
        :content="format ? formatted.data : view.definition"
        :readonly="true"
        class="border w-full h-full"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computedAsync, useLocalStorage } from "@vueuse/core";
import { NCheckbox } from "naive-ui";
import { computed } from "vue";
import { MonacoEditor } from "@/components/MonacoEditor";
import formatSQL from "@/components/MonacoEditor/sqlFormatter";
import { ComposedDatabase, dialectOfEngineV1 } from "@/types";
import { Engine } from "@/types/proto/v1/common";
import {
  DatabaseMetadata,
  SchemaMetadata,
  ViewMetadata,
} from "@/types/proto/v1/database_service";

const props = defineProps<{
  db: ComposedDatabase;
  database: DatabaseMetadata;
  schema: SchemaMetadata;
  view: ViewMetadata;
}>();
const format = useLocalStorage<boolean>(
  "bb.sql-editor.schema-hover-panel.view-info.format",
  false
);
const instanceEngine = computed(() => props.db.instanceEntity.engine);

const hasSchemaProperty = computed(() => {
  return [Engine.POSTGRES, Engine.RISINGWAVE].includes(instanceEngine.value);
});

const name = computed(() => {
  const { schema, view } = props;
  if (hasSchemaProperty.value) {
    return `${schema.name}.${view.name}`;
  }
  return view.name;
});

const formatted = computedAsync(
  async () => {
    const sql = props.view.definition;
    try {
      const result = await formatSQL(
        sql,
        dialectOfEngineV1(instanceEngine.value)
      );
      return result;
    } catch (err) {
      return {
        error: err,
        data: sql,
      };
    }
  },
  {
    error: null,
    data: props.view.definition,
  }
);
</script>
