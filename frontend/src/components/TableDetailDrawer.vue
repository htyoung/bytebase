<template>
  <Drawer :show="show" @close="$emit('dismiss')">
    <DrawerContent
      :title="$t('database.table-detail')"
      :content-props="{
        bodyContentClass: 'relative',
      }"
    >
      <div class="focus:outline-none w-[calc(100vw-256px)]" tabindex="0">
        <MaskSpinner v-if="isFetchingTableMetadata" />
        <main v-if="table" class="flex-1 relative pb-8 overflow-y-auto">
          <!-- Highlight Panel -->
          <div
            class="px-4 pb-4 border-b border-block-border md:flex md:items-center md:justify-between"
          >
            <div class="flex-1 min-w-0">
              <!-- Summary -->
              <div class="flex items-center">
                <div>
                  <div class="flex items-center">
                    <h1
                      class="pt-2 pb-2.5 text-xl font-bold leading-6 text-main truncate flex items-center gap-x-3"
                    >
                      {{ getTableName(table.name) }}
                      <BBBadge
                        v-if="isGhostTable(table)"
                        text="gh-ost"
                        :can-remove="false"
                        class="text-xs"
                      />
                    </h1>
                  </div>
                </div>
              </div>
              <dl
                class="flex flex-col space-y-1 md:space-y-0 md:flex-row md:flex-wrap"
              >
                <dt class="sr-only">{{ $t("common.environment") }}</dt>
                <dd class="flex items-center text-sm md:mr-4">
                  <span class="textlabel"
                    >{{ $t("common.environment") }}&nbsp;-&nbsp;</span
                  >
                  <EnvironmentV1Name
                    :environment="database.effectiveEnvironmentEntity"
                    icon-class="textinfolabel"
                  />
                </dd>
                <dt class="sr-only">{{ $t("common.instance") }}</dt>
                <dd class="flex items-center text-sm md:mr-4">
                  <span class="ml-1 textlabel"
                    >{{ $t("common.instance") }}&nbsp;-&nbsp;</span
                  >
                  <InstanceV1Name :instance="database.instanceEntity" />
                </dd>
                <dt class="sr-only">{{ $t("common.project") }}</dt>
                <dd class="flex items-center text-sm md:mr-4">
                  <span class="textlabel"
                    >{{ $t("common.project") }}&nbsp;-&nbsp;</span
                  >
                  <ProjectV1Name
                    :project="database.projectEntity"
                    hash="#databases"
                  />
                </dd>
                <dt class="sr-only">{{ $t("common.database") }}</dt>
                <dd class="flex items-center text-sm md:mr-4">
                  <span class="textlabel"
                    >{{ $t("common.database") }}&nbsp;-&nbsp;</span
                  >
                  <DatabaseV1Name :database="database" />
                </dd>
                <SQLEditorButtonV1
                  class="text-sm md:mr-4"
                  :database="database"
                  :label="true"
                  :disabled="!allowQuery"
                />
              </dl>
            </div>
          </div>

          <div class="mt-6">
            <div class="max-w-6xl px-6 space-y-6 divide-y divide-block-border">
              <!-- Description list -->
              <dl class="grid grid-cols-1 gap-x-4 gap-y-4 sm:grid-cols-3">
                <div v-if="hasTableEngineProperty" class="col-span-1">
                  <dt class="text-sm font-medium text-control-light">
                    {{ $t("database.engine") }}
                  </dt>
                  <dd class="mt-1 text-sm text-main">
                    {{ table.engine }}
                  </dd>
                </div>

                <div v-if="classificationConfig" class="col-span-1">
                  <dt class="text-sm text-control-light">
                    {{ $t("database.classification.self") }}
                  </dt>
                  <dd class="mt-1 text-lg sm:text-xl font-semibold">
                    <ClassificationLevelBadge
                      :classification="table.classification"
                      :classification-config="classificationConfig"
                      placeholder="-"
                    />
                  </dd>
                </div>

                <div class="col-span-1">
                  <dt class="text-sm font-medium text-control-light">
                    {{ $t("database.row-count-estimate") }}
                  </dt>
                  <dd class="mt-1 text-sm text-main">
                    {{ table.rowCount }}
                  </dd>
                </div>

                <div class="col-span-1">
                  <dt class="text-sm font-medium text-control-light">
                    {{ $t("database.data-size") }}
                  </dt>
                  <dd class="mt-1 text-sm text-main">
                    {{ bytesToString(table.dataSize.toNumber()) }}
                  </dd>
                </div>

                <div v-if="hasIndexSizeProperty" class="col-span-1">
                  <dt class="text-sm font-medium text-control-light">
                    {{ $t("database.index-size") }}
                  </dt>
                  <dd class="mt-1 text-sm text-main">
                    {{ bytesToString(table.indexSize.toNumber()) }}
                  </dd>
                </div>

                <template v-if="hasCollationProperty">
                  <div class="col-span-1">
                    <dt class="text-sm font-medium text-control-light">
                      {{ $t("db.collation") }}
                    </dt>
                    <dd class="mt-1 text-sm text-main">
                      {{ table.collation }}
                    </dd>
                  </div>
                </template>
              </dl>
            </div>
          </div>

          <div v-if="shouldShowPartitionTablesDataTable" class="mt-6 px-6">
            <div class="mb-4 w-full flex flex-row justify-between items-center">
              <div class="text-lg leading-6 font-medium text-main">
                {{ $t("database.partition-tables") }}
              </div>
              <div>
                <SearchBox
                  :value="state.partitionTableNameSearchKeyword"
                  :placeholder="$t('common.filter-by-name')"
                  @update:value="state.partitionTableNameSearchKeyword = $event"
                />
              </div>
            </div>
            <PartitionTablesDataTable
              :table="table"
              :search="state.partitionTableNameSearchKeyword"
            />
          </div>

          <div v-if="shouldShowColumnTable" class="mt-6 px-6">
            <div class="mb-4 w-full flex flex-row justify-between items-center">
              <div class="text-lg leading-6 font-medium text-main">
                {{ $t("database.columns") }}
              </div>
              <div>
                <SearchBox
                  :value="state.columnNameSearchKeyword"
                  :placeholder="$t('common.filter-by-name')"
                  @update:value="state.columnNameSearchKeyword = $event"
                />
              </div>
            </div>
            <ColumnDataTable
              :database="database"
              :schema="schemaName"
              :table="table"
              :column-list="table.columns"
              :mask-data-list="sensitiveDataList"
              :classification-config="classificationConfig"
              :search="state.columnNameSearchKeyword"
            />
          </div>

          <div v-if="instanceEngine !== Engine.SNOWFLAKE" class="mt-6 px-6">
            <div class="text-lg leading-6 font-medium text-main mb-4">
              {{ $t("database.indexes") }}
            </div>
            <IndexTable :database="database" :index-list="table.indexes" />
          </div>
        </main>
      </div>
    </DrawerContent>
  </Drawer>
</template>

<script lang="ts" setup>
import { computedAsync } from "@vueuse/core";
import { computed, reactive, ref } from "vue";
import {
  DatabaseV1Name,
  InstanceV1Name,
  Drawer,
  DrawerContent,
} from "@/components/v2";
import {
  useCurrentUserV1,
  useDatabaseV1Store,
  useDBSchemaV1Store,
} from "@/store";
import { usePolicyByParentAndType } from "@/store/modules/v1/policy";
import { DEFAULT_PROJECT_V1_NAME, defaultProject } from "@/types";
import { Engine } from "@/types/proto/v1/common";
import { PolicyType, MaskData } from "@/types/proto/v1/org_policy_service";
import { DataClassificationSetting_DataClassificationConfig } from "@/types/proto/v1/setting_service";
import {
  bytesToString,
  hasProjectPermissionV2,
  isDatabaseV1Queryable,
  isGhostTable,
} from "@/utils";
import ColumnDataTable from "./ColumnDataTable/index.vue";
import { SQLEditorButtonV1 } from "./DatabaseDetail";
import IndexTable from "./IndexTable.vue";
import PartitionTablesDataTable from "./PartitionTablesDataTable.vue";
import MaskSpinner from "./misc/MaskSpinner.vue";

interface LocalState {
  columnNameSearchKeyword: string;
  partitionTableNameSearchKeyword: string;
}

const props = defineProps<{
  show: boolean;
  // Format: /databases/:databaseName
  databaseName: string;
  schemaName: string;
  tableName: string;
  classificationConfig?: DataClassificationSetting_DataClassificationConfig;
}>();

defineEmits(["dismiss"]);

const databaseV1Store = useDatabaseV1Store();
const dbSchemaStore = useDBSchemaV1Store();
const currentUserV1 = useCurrentUserV1();
const state = reactive<LocalState>({
  columnNameSearchKeyword: "",
  partitionTableNameSearchKeyword: "",
});
const isFetchingTableMetadata = ref(false);
const table = computedAsync(
  async () => {
    const { databaseName, tableName, schemaName } = props;
    if (!tableName) {
      return undefined;
    }
    return dbSchemaStore.getOrFetchTableMetadata({
      database: databaseName,
      schema: schemaName,
      table: tableName,
      skipCache: false,
      silent: false,
    });
  },
  undefined,
  {
    evaluating: isFetchingTableMetadata,
  }
);

const database = computed(() => {
  return databaseV1Store.getDatabaseByName(props.databaseName);
});

const instanceEngine = computed(() => {
  return database.value.instanceEntity.engine;
});

const allowQuery = computed(() => {
  if (database.value.project === DEFAULT_PROJECT_V1_NAME) {
    return hasProjectPermissionV2(
      defaultProject,
      currentUserV1.value,
      "bb.databases.query"
    );
  }
  return isDatabaseV1Queryable(database.value, currentUserV1.value);
});

const hasSchemaProperty = computed(
  () =>
    instanceEngine.value === Engine.POSTGRES ||
    instanceEngine.value === Engine.RISINGWAVE
);

const hasPartitionTables = computed(() => {
  return (
    // Only show partition tables for PostgreSQL.
    database.value.instanceEntity.engine === Engine.POSTGRES &&
    table.value &&
    table.value.partitions.length > 0
  );
});

const hasTableEngineProperty = computed(() => {
  return ![Engine.POSTGRES, Engine.SNOWFLAKE].includes(instanceEngine.value);
});
const hasIndexSizeProperty = computed(() => {
  return ![Engine.CLICKHOUSE, Engine.SNOWFLAKE].includes(instanceEngine.value);
});
const hasCollationProperty = computed(() => {
  return ![Engine.POSTGRES, Engine.CLICKHOUSE, Engine.SNOWFLAKE].includes(
    instanceEngine.value
  );
});

const shouldShowPartitionTablesDataTable = computed(() => {
  return hasPartitionTables.value;
});

const shouldShowColumnTable = computed(() => {
  return instanceEngine.value !== Engine.MONGODB;
});

const getTableName = (tableName: string) => {
  if (hasSchemaProperty.value) {
    return `"${props.schemaName}"."${tableName}"`;
  }
  return tableName;
};

const sensitiveDataPolicy = usePolicyByParentAndType(
  computed(() => ({
    parentPath: database.value.name,
    policyType: PolicyType.MASKING,
  }))
);

const sensitiveDataList = computed((): MaskData[] => {
  const policy = sensitiveDataPolicy.value;
  if (!policy || !policy.maskingPolicy) {
    return [];
  }

  return policy.maskingPolicy.maskData;
});
</script>
