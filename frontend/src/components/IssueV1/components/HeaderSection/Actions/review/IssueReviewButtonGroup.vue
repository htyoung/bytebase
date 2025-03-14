<template>
  <div class="flex items-stretch gap-x-3">
    <template v-if="showIssueReviewActions">
      <ReviewActionButton
        v-for="action in issueReviewActionList"
        :key="action"
        :action="action"
        @perform-action="
          (action) => events.emit('perform-issue-review-action', { action })
        "
      />
    </template>

    <IssueStatusActionButtonGroup
      :display-mode="displayMode"
      :issue-status-action-list="issueStatusActionList"
      :extra-action-list="forceRolloutActionList"
    />
  </div>
</template>

<script lang="ts" setup>
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import {
  IssueReviewAction,
  getApplicableIssueStatusActionList,
  getApplicableStageRolloutActionList,
  getApplicableTaskRolloutActionList,
  taskRolloutActionDisplayName,
  useIssueContext,
} from "@/components/IssueV1";
import { useActuatorV1Store, useCurrentUserV1, usePageMode } from "@/store";
import { PresetRoleType } from "@/types";
import {
  IssueStatus,
  Issue_Approver_Status,
} from "@/types/proto/v1/issue_service";
import { extractUserResourceName } from "@/utils";
import { ExtraActionOption, IssueStatusActionButtonGroup } from "../common";
import ReviewActionButton from "./ReviewActionButton.vue";

const { t } = useI18n();
const actuatorStore = useActuatorV1Store();
const currentUser = useCurrentUserV1();
const pageMode = usePageMode();
const { issue, phase, reviewContext, events, activeTask, activeStage } =
  useIssueContext();
const { ready, status, done } = reviewContext;

const shouldShowApproveOrReject = computed(() => {
  if (
    issue.value.status === IssueStatus.CANCELED ||
    issue.value.status === IssueStatus.DONE
  ) {
    return false;
  }

  if (phase.value !== "REVIEW") {
    return false;
  }

  if (!ready.value) return false;
  if (done.value) return false;

  return true;
});
const shouldShowApprove = computed(() => {
  if (!shouldShowApproveOrReject.value) return false;

  return status.value === Issue_Approver_Status.PENDING;
});
const shouldShowReject = computed(() => {
  if (!shouldShowApproveOrReject.value) return false;
  return status.value === Issue_Approver_Status.PENDING;
});
const shouldShowReRequestReview = computed(() => {
  return (
    extractUserResourceName(issue.value.creator) === currentUser.value.email &&
    status.value === Issue_Approver_Status.REJECTED
  );
});
const issueReviewActionList = computed(() => {
  const actionList: IssueReviewAction[] = [];
  if (shouldShowReject.value) {
    actionList.push("SEND_BACK");
  }
  if (shouldShowApprove.value) {
    actionList.push("APPROVE");
  }
  if (shouldShowReRequestReview.value) {
    actionList.push("RE_REQUEST");
  }

  return actionList;
});

const issueStatusActionList = computed(() => {
  return getApplicableIssueStatusActionList(issue.value);
});
const forceRolloutActionList = computed((): ExtraActionOption[] => {
  // Still using role based permission checks
  if (
    !currentUser.value.roles.includes(PresetRoleType.WORKSPACE_ADMIN) &&
    !currentUser.value.roles.includes(PresetRoleType.WORKSPACE_DBA)
  ) {
    // Only for workspace admins and DBAs.
    return [];
  }

  const taskExtraActionOptions = getApplicableTaskRolloutActionList(
    issue.value,
    activeTask.value
  ).map<ExtraActionOption>((action) => ({
    type: "TASK",
    action,
    target: activeTask.value,
    label: `${t("common.force-verb", {
      verb: taskRolloutActionDisplayName(action).toLowerCase(),
    })}`,
    key: `force-${action}-task`,
  }));
  const stageExtraActionOptions = getApplicableStageRolloutActionList(
    issue.value,
    activeStage.value
  ).map<ExtraActionOption>(({ action, tasks }) => {
    const taskActionName = taskRolloutActionDisplayName(action);
    const stageActionName = t("issue.action-to-current-stage", {
      action: taskActionName,
    });
    return {
      type: "TASK-BATCH",
      action,
      target: tasks,
      label: t("common.force-verb", { verb: stageActionName.toLowerCase() }),
      key: `force-${action}-stage`,
    };
  });

  return [...taskExtraActionOptions, ...stageExtraActionOptions];
});

const showIssueReviewActions = computed(() => {
  return !(
    pageMode.value === "STANDALONE" && actuatorStore.customTheme === "lixiang"
  );
});

const displayMode = computed(() => {
  if (forceRolloutActionList.value.length > 0) return "DROPDOWN";
  return issueReviewActionList.value.length > 0 ? "DROPDOWN" : "BUTTON";
});
</script>
