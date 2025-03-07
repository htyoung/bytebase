<template>
  <div v-if="shouldShowAssignee" class="flex flex-col gap-y-1">
    <div class="flex items-center gap-1">
      <NTooltip>
        <template #trigger>
          <div class="flex items-center gap-x-1 textlabel">
            <span>{{ $t("common.assignee") }}</span>
          </div>
        </template>
        <template #default>
          <div class="max-w-[14rem]">
            {{ $t("issue.assignee-tooltip") }}
          </div>
        </template>
      </NTooltip>
    </div>

    <NTooltip :disabled="errors.length === 0">
      <template #trigger>
        <UserSelect
          :multiple="false"
          :user="assigneeUID"
          :disabled="errors.length > 0 || isUpdating"
          :loading="isUpdating"
          :filter="filterAssignee"
          :include-system-bot="false"
          :map-options="mapUserOptions"
          :fallback-option="fallbackUser"
          :clearable="true"
          :auto-reset="false"
          style="width: 100%"
          @update:user="changeAssigneeUID"
        />
      </template>
      <template #default>
        <ErrorList :errors="errors" class="max-w-[24rem] !whitespace-normal" />
      </template>
    </NTooltip>
  </div>
</template>

<script setup lang="ts">
import { asyncComputed } from "@vueuse/core";
import { NTooltip, SelectGroupOption } from "naive-ui";
import { computed, ref } from "vue";
import { useI18n } from "vue-i18n";
import {
  allowUserToChangeAssignee,
  useIssueContext,
  useWrappedReviewStepsV1,
} from "@/components/IssueV1/logic";
import ErrorList, { ErrorItem } from "@/components/misc/ErrorList.vue";
import { UserSelect } from "@/components/v2";
import { issueServiceClient } from "@/grpcweb";
import { emitWindowEvent } from "@/plugins";
import {
  pushNotification,
  useCurrentUserV1,
  usePageMode,
  useUserStore,
} from "@/store";
import {
  SYSTEM_BOT_EMAIL,
  SYSTEM_BOT_ID,
  SYSTEM_BOT_USER_NAME,
  UNKNOWN_ID,
  unknownUser,
} from "@/types";
import { User } from "@/types/proto/v1/auth_service";
import { Issue } from "@/types/proto/v1/issue_service";
import {
  extractUserResourceName,
  extractUserUID,
  hasProjectPermissionV2,
  isDatabaseRelatedIssue,
} from "@/utils";

const { t } = useI18n();
const userStore = useUserStore();
const pageMode = usePageMode();
const { isCreating, issue, reviewContext, releaserCandidates } =
  useIssueContext();
const currentUser = useCurrentUserV1();
const isUpdating = ref(false);

const shouldShowAssignee = computed(() => {
  return (
    !isCreating.value &&
    pageMode.value === "BUNDLED" &&
    isDatabaseRelatedIssue(issue.value)
  );
});

const assigneeEmail = computed(() => {
  const assignee = issue.value.assignee;
  if (!assignee) return undefined;
  return extractUserResourceName(assignee);
});

const assigneeUID = computed(() => {
  if (!assigneeEmail.value) return undefined;
  if (assigneeEmail.value === SYSTEM_BOT_EMAIL) return String(SYSTEM_BOT_ID);
  const user = userStore.getUserByEmail(assigneeEmail.value);
  if (!user) return undefined;
  return extractUserUID(user.name);
});

const allowChangeAssignee = computed(() => {
  if (isCreating.value) {
    return true;
  }
  return allowUserToChangeAssignee(currentUser.value, issue.value);
});

const errors = asyncComputed(async () => {
  const errors: ErrorItem[] = [];
  if (!allowChangeAssignee.value) {
    errors.push(t("issue.you-are-not-allowed-to-change-this-value"));
  }
  return errors;
}, []);

const changeAssigneeUID = async (uid: string | undefined) => {
  if (!uid || uid === String(UNKNOWN_ID)) {
    issue.value.assignee = "";
  } else {
    const assignee = userStore.getUserById(uid);
    if (!assignee) {
      issue.value.assignee = "";
    } else {
      issue.value.assignee = `users/${assignee.email}`;
    }
  }

  if (!isCreating.value) {
    const issuePatch = Issue.fromJSON(issue.value);
    isUpdating.value = true;
    try {
      const updated = await issueServiceClient.updateIssue({
        issue: issuePatch,
        updateMask: ["assignee"],
      });
      Object.assign(issue.value, updated);
      pushNotification({
        module: "bytebase",
        style: "SUCCESS",
        title: t("common.updated"),
      });
      emitWindowEvent("bb.issue-field-update");
    } finally {
      isUpdating.value = false;
    }
  }
};

const filterAssignee = (user: User): boolean => {
  // Users can be assignee candidates only if they have permission to "See"
  // the issue
  return hasProjectPermissionV2(
    issue.value.projectEntity,
    user,
    "bb.issues.get"
  );
};

const mapUserOptions = (users: User[]) => {
  const project = issue.value.projectEntity;
  const phase = isCreating.value
    ? "PREVIEW"
    : reviewContext.done.value
    ? "CD"
    : "CI";
  const added = new Set<string>(); // by user.name
  const groups: SelectGroupOption[] = [];
  const mapUserOption = (user: User) => ({
    user,
    value: extractUserUID(user.name),
    label: user.title,
  });
  // `addGroup` ensures that
  // 1. any users will be added at most once
  // 2. empty groups will not be added
  const addGroup = (group: SelectGroupOption, users: User[]) => {
    const filteredUsers = users.filter((u) => !added.has(u.name));
    if (filteredUsers.length > 0) {
      filteredUsers.forEach((u) => added.add(u.name));
      group.children = filteredUsers.map(mapUserOption);
      groups.push(group);
    }
  };

  // Groups in order
  // - current assignee
  // - approvers of current step (CI phase only)
  // - releasers of current stage (CD phase only)
  // - project or workspace owners
  if (assigneeEmail.value) {
    const assignee = userStore.getUserByEmail(assigneeEmail.value);
    if (assignee) {
      addGroup(
        {
          type: "group",
          label: t("issue.assignee.current-assignee"),
          key: "current-assignee",
        },
        [assignee]
      );
    }
  }

  if (phase === "CI") {
    const steps = useWrappedReviewStepsV1(issue, reviewContext);
    const currentStep = steps.value?.find((step) => step.status === "CURRENT");
    addGroup(
      {
        type: "group",
        label: t("issue.assignee.approvers-of-current-step"),
        key: "approvers",
      },
      currentStep?.candidates ?? []
    );
  }
  if (phase === "CD") {
    const candidates = new Set(releaserCandidates.value.map((c) => c.name));
    const releasers = users.filter((user) => candidates.has(user.name));
    addGroup(
      {
        type: "group",
        label: t("issue.assignee.releasers-of-current-stage"),
        key: "releasers",
      },
      releasers
    );
  }
  const owners = users.filter((user) =>
    hasProjectPermissionV2(project, user, "bb.issues.update")
  );
  addGroup(
    {
      type: "group",
      label: t("issue.assignee.project-owners"),
      key: "project-owners",
    },
    owners
  );

  return groups;
};

const fallbackUser = (uid: string) => {
  if (uid === String(SYSTEM_BOT_ID)) {
    return {
      user: userStore.getUserByName(SYSTEM_BOT_USER_NAME)!,
      value: uid,
    };
  }
  return {
    user: unknownUser(),
    value: uid,
  };
};

defineExpose({
  shown: computed(() => {
    return shouldShowAssignee.value;
  }),
});
</script>
