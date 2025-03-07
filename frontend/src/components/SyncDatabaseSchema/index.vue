<template>
  <div class="w-full h-full overflow-hidden flex flex-col">
    <p class="text-sm text-gray-500">
      {{ $t("database.sync-schema.description") }}
      <LearnMoreLink
        url="https://www.bytebase.com/docs/change-database/synchronize-schema?source=console"
      />
    </p>
    <StepTab
      class="pt-4 flex-1 overflow-hidden flex flex-col"
      :step-list="stepTabList"
      :current-index="state.currentStep"
      :show-cancel="false"
      :allow-next="allowNext"
      :finish-title="$t('database.sync-schema.preview-issue')"
      pane-class="flex-1 overflow-y-auto"
      :next-button-props="nextButtonProps"
      @cancel="cancelSetup"
      @update:current-index="tryChangeStep"
      @finish="tryFinishSetup"
    >
      <template #0>
        <div class="mb-4">
          <NRadioGroup v-model:value="state.sourceSchemaType" class="space-x-4">
            <NRadio
              :value="'SCHEMA_HISTORY_VERSION'"
              :label="$t('database.sync-schema.schema-history-version')"
            />
            <NRadio
              :value="'RAW_SQL'"
              :label="$t('database.sync-schema.copy-schema')"
            />
          </NRadioGroup>
        </div>
        <DatabaseSchemaSelector
          v-if="state.sourceSchemaType === 'SCHEMA_HISTORY_VERSION'"
          :select-state="changeHistorySourceSchemaState"
          :disable-project-select="!!project"
          @update="handleChangeHistorySchemaVersionChanges"
        />
        <RawSQLEditor
          v-if="state.sourceSchemaType === 'RAW_SQL'"
          :project-id="rawSQLState.projectId"
          :engine="rawSQLState.engine"
          :statement="rawSQLState.statement"
          :sheet-id="rawSQLState.sheetId"
          :disable-project-select="!!project"
          @update="handleRawSQLStateChange"
        />
      </template>
      <template #1>
        <SelectTargetDatabasesView
          ref="targetDatabaseViewRef"
          :project-id="projectId!"
          :source-schema-type="state.sourceSchemaType"
          :database-source-schema="(changeHistorySourceSchemaState as any)"
          :raw-sql-state="rawSQLState"
        />
      </template>
    </StepTab>
  </div>
</template>

<script lang="ts" setup>
import dayjs from "dayjs";
import { isNull, isUndefined } from "lodash-es";
import { NRadioGroup, NRadio, useDialog, ButtonProps } from "naive-ui";
import { computed, reactive, ref } from "vue";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";
import { StepTab } from "@/components/v2";
import { PROJECT_V1_ROUTE_ISSUE_DETAIL } from "@/router/dashboard/projectV1";
import { WORKSPACE_HOME_MODULE } from "@/router/dashboard/workspaceRoutes";
import { useProjectV1Store } from "@/store";
import { UNKNOWN_ID, ComposedProject } from "@/types";
import { Engine } from "@/types/proto/v1/common";
import { extractProjectResourceName } from "@/utils";
import DatabaseSchemaSelector from "./DatabaseSchemaSelector.vue";
import RawSQLEditor from "./RawSQLEditor.vue";
import SelectTargetDatabasesView from "./SelectTargetDatabasesView.vue";
import {
  ChangeHistorySourceSchema,
  RawSQLState,
  SourceSchemaType,
} from "./types";

defineOptions({
  name: "SyncDatabaseSchema",
});

const SELECT_SOURCE_SCHEMA = 0;
const SELECT_TARGET_DATABASE_LIST = 1;

type Step = typeof SELECT_SOURCE_SCHEMA | typeof SELECT_TARGET_DATABASE_LIST;

interface LocalState {
  sourceSchemaType: SourceSchemaType;
  currentStep: Step;
}

const props = defineProps<{
  project?: ComposedProject;
  sourceSchemaType?: SourceSchemaType;
}>();

const { t } = useI18n();
const router = useRouter();
const dialog = useDialog();
const projectStore = useProjectV1Store();
const targetDatabaseViewRef =
  ref<InstanceType<typeof SelectTargetDatabasesView>>();
const state = reactive<LocalState>({
  sourceSchemaType: props.sourceSchemaType ?? "SCHEMA_HISTORY_VERSION",
  currentStep: SELECT_SOURCE_SCHEMA,
});
const changeHistorySourceSchemaState = reactive<ChangeHistorySourceSchema>({
  projectId: props.project?.uid,
});
const rawSQLState = reactive<RawSQLState>({
  projectId: props.project?.uid,
  engine: Engine.MYSQL,
  statement: "",
});

const projectId = computed(() => {
  if (props.project) {
    return props.project.uid;
  }
  if (state.sourceSchemaType === "SCHEMA_HISTORY_VERSION") {
    return changeHistorySourceSchemaState.projectId;
  } else {
    return rawSQLState.projectId;
  }
});

const handleChangeHistorySchemaVersionChanges = (
  schemaVersion: ChangeHistorySourceSchema
) => {
  Object.assign(changeHistorySourceSchemaState, schemaVersion);
};

const isValidId = (id: any): id is string => {
  if (isNull(id) || isUndefined(id) || String(id) === String(UNKNOWN_ID)) {
    return false;
  }
  return true;
};

const stepTabList = computed(() => {
  return [
    { title: t("database.sync-schema.select-source-schema") },
    { title: t("database.sync-schema.select-target-databases") },
  ];
});

const allowNext = computed(() => {
  if (state.currentStep === SELECT_SOURCE_SCHEMA) {
    if (state.sourceSchemaType === "SCHEMA_HISTORY_VERSION") {
      return (
        !changeHistorySourceSchemaState.isFetching &&
        isValidId(changeHistorySourceSchemaState.environmentId) &&
        isValidId(changeHistorySourceSchemaState.databaseId) &&
        !isUndefined(changeHistorySourceSchemaState.changeHistory)
      );
    } else {
      return (
        !isUndefined(rawSQLState.projectId) &&
        (rawSQLState.statement !== "" || !isUndefined(rawSQLState.sheetId))
      );
    }
  } else {
    if (!targetDatabaseViewRef.value) {
      return false;
    }
    const targetDatabaseList = targetDatabaseViewRef.value?.targetDatabaseList;
    const targetDatabaseDiffList = targetDatabaseList
      .map((db) => {
        const diff = targetDatabaseViewRef.value!.databaseDiffCache[db.uid];
        return {
          id: db.uid,
          diff: diff?.edited || "",
        };
      })
      .filter((item) => item.diff !== "");
    return targetDatabaseDiffList.length > 0;
  }
});

const nextButtonProps = computed((): ButtonProps | undefined => {
  if (state.currentStep === SELECT_SOURCE_SCHEMA) {
    if (state.sourceSchemaType === "SCHEMA_HISTORY_VERSION") {
      if (changeHistorySourceSchemaState.isFetching) {
        return {
          loading: true,
        };
      }
    }
  }
  return undefined;
});

const handleRawSQLStateChange = (state: RawSQLState) => {
  Object.assign(rawSQLState, state);
};

const tryChangeStep = async (nextStepIndex: number) => {
  if (state.currentStep === 1 && nextStepIndex === 0) {
    const targetDatabaseList =
      targetDatabaseViewRef.value?.targetDatabaseList || [];
    if (targetDatabaseList.length > 0) {
      dialog.create({
        positiveText: t("common.confirm"),
        negativeText: t("common.cancel"),
        title: t("deployment-config.confirm-to-revert"),
        autoFocus: false,
        closable: false,
        maskClosable: false,
        closeOnEsc: false,
        onNegativeClick: () => {
          // nothing to do
        },
        onPositiveClick: () => {
          state.currentStep = nextStepIndex as Step;
        },
      });
      return;
    }
  }
  state.currentStep = nextStepIndex as Step;
};

const tryFinishSetup = async () => {
  if (!targetDatabaseViewRef.value) {
    return;
  }

  const targetDatabaseList = targetDatabaseViewRef.value.targetDatabaseList;
  const targetDatabaseDiffList = targetDatabaseList
    .map((db) => {
      const diff = targetDatabaseViewRef.value!.databaseDiffCache[db.uid];
      return {
        id: db.uid,
        diff: diff.edited,
      };
    })
    .filter((item) => item.diff !== "");
  const databaseIdList = targetDatabaseDiffList.map((item) => item.id);
  const statementList = targetDatabaseDiffList.map((item) => item.diff);
  const project = await projectStore.getOrFetchProjectByUID(projectId.value!);

  const query: Record<string, any> = {
    template: "bb.issue.database.schema.update",
    project: project.uid,
    mode: "normal",
    ghost: undefined,
  };
  query.databaseList = databaseIdList.join(",");
  query.sqlList = JSON.stringify(statementList);
  query.name = generateIssueName(
    targetDatabaseList.map((db) => db.databaseName)
  );

  const routeInfo = {
    name: PROJECT_V1_ROUTE_ISSUE_DETAIL,
    params: {
      projectId: extractProjectResourceName(project.name),
      issueSlug: "create",
    },
    query,
  };
  router.push(routeInfo);
};

const generateIssueName = (databaseNameList: string[]) => {
  const issueNameParts: string[] = [];
  if (databaseNameList.length === 1) {
    issueNameParts.push(`[${databaseNameList[0]}]`);
  } else {
    issueNameParts.push(`[${databaseNameList.length} databases]`);
  }
  issueNameParts.push(`Alter schema`);
  const datetime = dayjs().format("@MM-DD HH:mm");
  const tz = "UTC" + dayjs().format("ZZ");
  issueNameParts.push(`${datetime} ${tz}`);
  return issueNameParts.join(" ");
};

const cancelSetup = () => {
  router.replace({
    name: WORKSPACE_HOME_MODULE,
  });
};
</script>
