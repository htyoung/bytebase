<template>
  <div
    v-if="state.isLoaded"
    class="flex-1 overflow-auto focus:outline-none"
    tabindex="0"
    v-bind="$attrs"
  >
    <main class="flex-1 relative overflow-y-auto space-y-4">
      <div
        class="space-y-2 lg:space-y-0 lg:flex lg:items-center lg:justify-between"
      >
        <div class="flex-1 min-w-0 shrink-0">
          <!-- Summary -->
          <div class="flex items-center">
            <div>
              <div class="flex items-center">
                <h1
                  class="pt-2 pb-2.5 text-xl font-bold leading-6 text-main truncate flex items-center gap-x-3"
                >
                  {{ databaseGroup.databasePlaceholder }}
                  <BBBadge
                    text="Database Group"
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
                :environment="environment"
                icon-class="textinfolabel"
              />
            </dd>
            <dd class="flex items-center text-sm md:mr-4">
              <span class="textlabel"
                >{{ $t("common.project") }}&nbsp;-&nbsp;</span
              >
              <ProjectV1Name :project="project" hash="#database-groups" />
            </dd>
          </dl>
        </div>

        <div
          class="flex flex-row justify-end items-center flex-wrap shrink gap-x-2 gap-y-2"
        >
          <NButton v-if="allowEdit" @click="handleEditDatabaseGroup">
            {{ $t("common.configure") }}
          </NButton>
          <NButton
            v-if="hasPermissionToCreateIssue"
            @click="createMigration('bb.issue.database.schema.update')"
          >
            {{ $t("database.edit-schema") }}
          </NButton>
          <NButton
            v-if="hasPermissionToCreateIssue"
            @click="createMigration('bb.issue.database.data.update')"
          >
            {{ $t("database.change-data") }}
          </NButton>
        </div>
      </div>

      <hr />

      <FeatureAttentionForInstanceLicense
        v-if="existMatchedUnactivateInstance"
        type="warning"
        feature="bb.feature.database-grouping"
      />

      <div class="w-full max-w-5xl grid grid-cols-5 gap-x-6">
        <div class="col-span-3">
          <p class="pl-1 text-lg mb-2">
            {{ $t("database-group.condition.self") }}
          </p>
          <ExprEditor
            :expr="state.expr!"
            :allow-admin="false"
            :factor-list="FactorList.get('DATABASE_GROUP') ?? []"
          />
        </div>
        <div class="col-span-2">
          <MatchedDatabaseView
            :loading="false"
            :matched-database-list="matchedDatabaseList"
            :unmatched-database-list="unmatchedDatabaseList"
          />
        </div>
      </div>

      <hr />
      <div class="w-full max-w-5xl">
        <div class="w-full flex flex-row justify-between items-center">
          <p class="my-4">{{ $t("database-group.table-group.self") }}</p>
          <NButton v-if="allowEdit" @click.prevent="handleCreateSchemaGroup">
            {{ $t("database-group.table-group.create") }}
          </NButton>
        </div>
        <SchemaGroupTable
          :schema-group-list="schemaGroupList"
          @edit="handleEditSchemaGroup"
        />
      </div>
    </main>
  </div>

  <DatabaseGroupPanel
    :show="editState.showConfigurePanel"
    :project="project"
    :resource-type="editState.type"
    :database-group="editState.databaseGroup"
    :parent-database-group="editState.parentDatabaseGroup"
    @close="editState.showConfigurePanel = false"
  />

  <DatabaseGroupPrevEditorModal
    v-if="issueType"
    :issue-type="issueType"
    :database-group="databaseGroup"
    @close="issueType = undefined"
  />
</template>

<script lang="ts" setup>
import { useDebounceFn } from "@vueuse/core";
import { NButton } from "naive-ui";
import { onMounted, reactive, computed, watch, ref } from "vue";
import DatabaseGroupPrevEditorModal from "@/components/AlterSchemaPrepForm/DatabaseGroupPrevEditorModal.vue";
import DatabaseGroupPanel from "@/components/DatabaseGroup/DatabaseGroupPanel.vue";
import MatchedDatabaseView from "@/components/DatabaseGroup/MatchedDatabaseView.vue";
import SchemaGroupTable from "@/components/DatabaseGroup/SchemaGroupTable.vue";
import { FactorList, ResourceType } from "@/components/DatabaseGroup/utils";
import ExprEditor from "@/components/ExprEditor";
import { ConditionGroupExpr } from "@/plugins/cel";
import {
  useCurrentUserV1,
  useDBGroupStore,
  useProjectV1Store,
  useSubscriptionV1Store,
} from "@/store";
import { databaseGroupNamePrefix } from "@/store/modules/v1/common";
import { projectNamePrefix } from "@/store/modules/v1/common";
import { ComposedDatabase, ComposedDatabaseGroup } from "@/types";
import { DatabaseGroup, SchemaGroup } from "@/types/proto/v1/project_service";
import { hasProjectPermissionV2 } from "@/utils";

interface LocalState {
  isLoaded: boolean;
  expr?: ConditionGroupExpr;
}

interface EditDatabaseGroupState {
  showConfigurePanel: boolean;
  type: ResourceType;
  databaseGroup?: DatabaseGroup | SchemaGroup;
  parentDatabaseGroup?: ComposedDatabaseGroup;
}

const props = defineProps<{
  projectId: string;
  databaseGroupName: string;
  allowEdit: boolean;
}>();

const projectStore = useProjectV1Store();
const dbGroupStore = useDBGroupStore();
const subscriptionV1Store = useSubscriptionV1Store();

const state = reactive<LocalState>({
  isLoaded: false,
});
const editState = reactive<EditDatabaseGroupState>({
  showConfigurePanel: false,
  type: "DATABASE_GROUP",
});
const issueType = ref<
  | "bb.issue.database.schema.update"
  | "bb.issue.database.data.update"
  | undefined
>();
const project = computed(() => {
  return projectStore.getProjectByName(
    `${projectNamePrefix}${props.projectId}`
  );
});
const databaseGroupResourceName = computed(() => {
  return `${project.value.name}/${databaseGroupNamePrefix}${props.databaseGroupName}`;
});
const databaseGroup = computed(() => {
  return dbGroupStore.getDBGroupByName(
    databaseGroupResourceName.value
  ) as ComposedDatabaseGroup;
});
const schemaGroupList = computed(() => {
  return dbGroupStore.getSchemaGroupListByDBGroupName(
    databaseGroupResourceName.value
  );
});
const environment = computed(() => databaseGroup.value?.environment);
const hasPermissionToCreateIssue = computed(() =>
  hasProjectPermissionV2(
    project.value,
    useCurrentUserV1().value,
    "bb.issues.create"
  )
);

onMounted(async () => {
  await dbGroupStore.getOrFetchDBGroupByName(databaseGroupResourceName.value);
});

const handleEditDatabaseGroup = () => {
  editState.type = "DATABASE_GROUP";
  editState.databaseGroup = databaseGroup.value;
  editState.showConfigurePanel = true;
};

const handleCreateSchemaGroup = () => {
  editState.type = "SCHEMA_GROUP";
  editState.databaseGroup = undefined;
  editState.showConfigurePanel = true;
  editState.parentDatabaseGroup = databaseGroup.value;
};

const handleEditSchemaGroup = (schemaGroup: SchemaGroup) => {
  editState.type = "SCHEMA_GROUP";
  editState.databaseGroup = schemaGroup;
  editState.showConfigurePanel = true;
};

const createMigration = (
  type: "bb.issue.database.schema.update" | "bb.issue.database.data.update"
) => {
  issueType.value = type;
};

watch(
  () => [databaseGroup.value],
  async () => {
    if (!databaseGroup.value) {
      return;
    }

    state.expr = databaseGroup.value.simpleExpr;
    await dbGroupStore.getOrFetchSchemaGroupListByDBGroupName(
      databaseGroup.value.name
    );

    state.isLoaded = true;
  },
  {
    immediate: true,
  }
);

const matchedDatabaseList = ref<ComposedDatabase[]>([]);
const unmatchedDatabaseList = ref<ComposedDatabase[]>([]);

const updateDatabaseMatchingState = useDebounceFn(async () => {
  if (!state.isLoaded) {
    return;
  }
  if (!environment.value) {
    return;
  }
  if (!project.value) {
    return;
  }
  if (!state.expr) {
    return;
  }

  const result = await dbGroupStore.fetchDatabaseGroupMatchList({
    projectName: project.value.name,
    environmentId: environment.value.uid,
    expr: state.expr,
  });

  matchedDatabaseList.value = result.matchedDatabaseList;
  unmatchedDatabaseList.value = result.unmatchedDatabaseList;
}, 500);

watch(
  [
    () => state.isLoaded,
    () => project.value,
    () => environment.value,
    () => state.expr,
  ],
  updateDatabaseMatchingState,
  {
    immediate: true,
    deep: true,
  }
);

const existMatchedUnactivateInstance = computed(() => {
  return matchedDatabaseList.value.some(
    (database) =>
      !subscriptionV1Store.hasInstanceFeature(
        "bb.feature.database-grouping",
        database.instanceEntity
      )
  );
});
</script>
