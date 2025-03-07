<template>
  <div
    v-if="shouldShowTips"
    class="mt-1 text-sm text-control-light flex flex-row items-center space-x-1"
  >
    <heroicons-outline:arrow-uturn-left class="w-4 h-4" />
    <i18n-t keypath="issue.will-rollback" tag="span">
      <template #link>
        <router-link :to="rollbackIssueLink" class="normal-link">
          <i18n-t keypath="issue.issue-link-with-task">
            <template #issue>#{{ rollbackFromIssue.uid }}</template>
            <template #task>[{{ rollbackFromTask.title }}]</template>
          </i18n-t>
        </router-link>
      </template>
    </i18n-t>
  </div>
</template>

<script lang="ts" setup>
import { computedAsync } from "@vueuse/core";
import { computed } from "vue";
import { useRoute } from "vue-router";
import { experimentalFetchIssueByUID } from "@/store";
import { RollbackDetail, UNKNOWN_ID, unknownIssue, unknownTask } from "@/types";
import { Task_Type } from "@/types/proto/v1/rollout_service";
import {
  buildIssueV1LinkWithTask,
  extractIssueUID,
  extractProjectResourceName,
  extractTaskUID,
  flattenTaskV1List,
} from "@/utils";
import {
  databaseForTask,
  getRollbackTaskMappingFromQuery,
  useIssueContext,
} from "../../logic";

const { isCreating, issue, selectedTask } = useIssueContext();
const route = useRoute();

const taskConfig = computed(() => {
  if (isCreating.value) return undefined;

  const task = selectedTask.value;
  if (task.type === Task_Type.DATABASE_DATA_UPDATE) {
    return task.databaseDataUpdate;
  }
  return undefined;
});

const rollbackDetail = computed((): RollbackDetail | undefined => {
  if (!isCreating.value) return undefined;
  // Try to find out the relationship between databaseId and rollback issue/task
  // Id from the URL query.
  const databaseUID = databaseForTask(issue.value, selectedTask.value).uid;

  const mapping = getRollbackTaskMappingFromQuery(route);
  return mapping.get(databaseUID);
});

const rollbackFromIssueUID = computed(() => {
  if (isCreating.value) {
    return String(rollbackDetail.value?.issueId ?? UNKNOWN_ID);
  }
  if (taskConfig.value) {
    return (
      extractIssueUID(taskConfig.value.rollbackFromIssue) || String(UNKNOWN_ID)
    );
  }
  return String(UNKNOWN_ID);
});

const rollbackFromTaskUID = computed(() => {
  if (isCreating.value) {
    return String(rollbackDetail.value?.taskId ?? UNKNOWN_ID);
  }
  if (taskConfig.value) {
    return (
      extractTaskUID(taskConfig.value.rollbackFromTask) || String(UNKNOWN_ID)
    );
  }
  return String(UNKNOWN_ID);
});

const rollbackFromIssue = computedAsync(async () => {
  const uid = rollbackFromIssueUID.value;
  // Here we assume that the rollback issue and the original issue are always
  // in the same project
  return await experimentalFetchIssueByUID(
    uid,
    extractProjectResourceName(issue.value.project)
  );
}, unknownIssue());

const rollbackFromTask = computed(() => {
  const issue = rollbackFromIssue.value;
  if (issue.uid === String(UNKNOWN_ID)) return unknownTask();

  const taskUID = rollbackFromTaskUID.value;
  if (taskUID === String(UNKNOWN_ID)) return unknownTask();

  const task = flattenTaskV1List(issue.rolloutEntity).find(
    (t) => t.uid === taskUID
  );
  if (task) {
    return task;
  }

  return unknownTask();
});

const shouldShowTips = computed(() => {
  // Show the tips when rollbackFromIssue and rollbackFromTask are both ready.
  return (
    rollbackFromIssue.value.uid !== String(UNKNOWN_ID) &&
    rollbackFromTask.value.uid !== String(UNKNOWN_ID)
  );
});

const rollbackIssueLink = computed(() => {
  if (!shouldShowTips.value) return "";
  return buildIssueV1LinkWithTask(
    rollbackFromIssue.value,
    rollbackFromTask.value
  );
});

defineExpose({
  shown: computed(() => {
    return shouldShowTips.value;
  }),
});
</script>
