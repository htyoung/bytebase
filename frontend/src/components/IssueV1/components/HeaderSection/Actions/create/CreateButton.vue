<template>
  <NTooltip :disabled="issueCreateErrorList.length === 0" placement="top">
    <template #trigger>
      <NButton
        type="primary"
        size="medium"
        tag="div"
        :disabled="issueCreateErrorList.length > 0 || loading"
        :loading="loading"
        @click="doCreateIssue"
      >
        {{ loading ? $t("common.creating") : $t("common.create") }}
      </NButton>
    </template>

    <template #default>
      <ErrorList :errors="issueCreateErrorList" />
    </template>
  </NTooltip>

  <!-- prevent clicking the page when creating in progress -->
  <div
    v-if="loading"
    v-zindexable="{ enabled: true }"
    class="fixed inset-0 pointer-events-auto flex flex-col items-center justify-center"
    @click.stop.prevent
  />
</template>

<script setup lang="ts">
import { NTooltip, NButton } from "naive-ui";
import { zindexable as vZindexable } from "vdirs";
import { computed, nextTick, ref, toRaw } from "vue";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";
import { ErrorList } from "@/components/IssueV1/components/common";
import {
  databaseForTask,
  getLocalSheetByName,
  isValidStage,
  specForTask,
  useIssueContext,
} from "@/components/IssueV1/logic";
import formatSQL from "@/components/MonacoEditor/sqlFormatter";
import { useSQLCheckContext } from "@/components/SQLCheck";
import { issueServiceClient, rolloutServiceClient } from "@/grpcweb";
import { emitWindowEvent } from "@/plugins";
import { PROJECT_V1_ROUTE_ISSUE_DETAIL } from "@/router/dashboard/projectV1";
import { useCurrentUserV1, useDatabaseV1Store, useSheetV1Store } from "@/store";
import { ComposedIssue, dialectOfEngineV1, languageOfEngineV1 } from "@/types";
import { Issue } from "@/types/proto/v1/issue_service";
import { Plan_ChangeDatabaseConfig } from "@/types/proto/v1/rollout_service";
import { Sheet } from "@/types/proto/v1/sheet_service";
import {
  extractDeploymentConfigName,
  extractProjectResourceName,
  extractSheetUID,
  flattenTaskV1List,
  getSheetStatement,
  hasPermissionToCreateChangeDatabaseIssueInProject,
  issueSlug,
  setSheetStatement,
  sheetNameOfTaskV1,
} from "@/utils";

const MAX_FORMATTABLE_STATEMENT_SIZE = 10000; // 10K characters

const { t } = useI18n();
const router = useRouter();
const { issue, formatOnSave } = useIssueContext();
const { runSQLCheck } = useSQLCheckContext();
const sheetStore = useSheetV1Store();
const me = useCurrentUserV1();
const loading = ref(false);

const issueCreateErrorList = computed(() => {
  const errorList: string[] = [];
  if (
    !hasPermissionToCreateChangeDatabaseIssueInProject(
      issue.value.projectEntity,
      me.value
    )
  ) {
    errorList.push(t("common.missing-permission"));
  }
  if (!issue.value.title.trim()) {
    errorList.push("Missing issue title");
  }
  if (!issue.value.rolloutEntity.stages.every((stage) => isValidStage(stage))) {
    errorList.push("Missing SQL statement in some stages");
  }
  return errorList;
});

const doCreateIssue = async () => {
  loading.value = true;
  const check = runSQLCheck.value;
  if (check && !(await check())) {
    loading.value = false;
    return;
  }

  try {
    await createSheets();
    const createdPlan = await createPlan();
    if (!createdPlan) return;

    issue.value.plan = createdPlan.name;
    issue.value.planEntity = createdPlan;

    const issueCreate = {
      ...Issue.fromPartial(issue.value),
      rollout: "",
    };
    const createdIssue = await issueServiceClient.createIssue({
      parent: issue.value.project,
      issue: issueCreate,
    });

    const createdRollout = await rolloutServiceClient.createRollout({
      parent: issue.value.project,
      plan: createdPlan.name,
    });

    createdIssue.rollout = createdRollout.name;

    const composedIssue: ComposedIssue = {
      ...issue.value,
      ...createdIssue,
      planEntity: createdPlan,
      rolloutEntity: createdRollout,
    };

    await emitIssueCreateWindowEvent(composedIssue);
    nextTick(() => {
      router.push({
        name: PROJECT_V1_ROUTE_ISSUE_DETAIL,
        params: {
          projectId: extractProjectResourceName(composedIssue.project),
          issueSlug: issueSlug(composedIssue.title, composedIssue.uid),
        },
      });
    });

    return composedIssue;
  } catch {
    loading.value = false;
  }
};

// Create sheets for spec configs and update their resource names.
const createSheets = async () => {
  const steps = issue.value.planEntity?.steps ?? [];
  const flattenSpecList = steps.flatMap((step) => {
    return step.specs;
  });

  const configWithSheetList: Plan_ChangeDatabaseConfig[] = [];
  const pendingCreateSheetMap = new Map<string, Sheet>();

  for (let i = 0; i < flattenSpecList.length; i++) {
    const spec = flattenSpecList[i];
    const config = spec.changeDatabaseConfig;
    if (!config) continue;
    configWithSheetList.push(config);
    if (pendingCreateSheetMap.has(config.sheet)) continue;
    const uid = extractSheetUID(config.sheet);
    if (uid.startsWith("-")) {
      // The sheet is pending create
      const sheet = getLocalSheetByName(config.sheet);
      sheet.database = config.target;
      pendingCreateSheetMap.set(sheet.name, sheet);

      await maybeFormatSQL(sheet, sheet.database);
    }
  }
  const pendingCreateSheetList = Array.from(pendingCreateSheetMap.values());
  const sheetNameMap = new Map<string, string>();
  for (let i = 0; i < pendingCreateSheetList.length; i++) {
    const sheet = pendingCreateSheetList[i];
    if (extractDeploymentConfigName(sheet.database)) {
      // If a sheet's target is a deploymentConfig, it should be unset
      // since it actually doesn't belongs to any exact database.
      sheet.database = "";
    }
    sheet.title = issue.value.title;
    const createdSheet = await sheetStore.createSheet(
      issue.value.project,
      sheet
    );
    sheetNameMap.set(sheet.name, createdSheet.name);
  }
  configWithSheetList.forEach((config) => {
    const uid = extractSheetUID(config.sheet);
    if (uid.startsWith("-")) {
      config.sheet = sheetNameMap.get(config.sheet) ?? "";
    }
  });
};

const createPlan = async () => {
  const plan = issue.value.planEntity;
  if (!plan) return;
  const createdPlan = await rolloutServiceClient.createPlan({
    parent: issue.value.project,
    plan,
  });
  return createdPlan;
};

const maybeFormatSQL = async (sheet: Sheet, target: string) => {
  if (!formatOnSave.value) {
    return;
  }
  const db = useDatabaseV1Store().getDatabaseByName(target);
  if (!db) {
    return;
  }
  const language = languageOfEngineV1(db.instanceEntity.engine);
  if (language !== "sql") {
    return;
  }
  const dialect = dialectOfEngineV1(db.instanceEntity.engine);

  const statement = getSheetStatement(sheet);
  if (statement.length > MAX_FORMATTABLE_STATEMENT_SIZE) {
    return;
  }
  const { error, data: formatted } = await formatSQL(statement, dialect);
  if (error) {
    return;
  }

  setSheetStatement(sheet, formatted);
};

const emitIssueCreateWindowEvent = async (issue: ComposedIssue) => {
  const eventParams = {
    uid: issue.uid,
    description: issue.description,
    project: toRaw(issue.projectEntity),
    tasks: [],
  };
  const tasks = flattenTaskV1List(issue.rolloutEntity);
  for (const task of tasks) {
    const spec = specForTask(issue.planEntity, task);
    const database = databaseForTask(issue, task);
    const sheetName = sheetNameOfTaskV1(task);
    const sheet = await sheetStore.getOrFetchSheetByName(sheetName, "FULL");
    const statement = sheet ? getSheetStatement(sheet) : "";
    eventParams.tasks.push({
      database: toRaw(database),
      earliestAllowedTime: spec?.earliestAllowedTime,
      statement,
    } as never);
  }
  emitWindowEvent("bb.issue-create", eventParams);
};
</script>
