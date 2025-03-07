<template>
  <BBGrid
    :column-list="columnList"
    :row-clickable="false"
    :data-source="memberList"
    :custom-header="true"
    :ready="ready"
    class="border"
    row-key="email"
  >
    <template #header>
      <div role="table-row" class="bb-grid-row bb-grid-header-row group">
        <div
          v-for="(column, index) in columnList"
          :key="index"
          role="table-cell"
          class="bb-grid-header-cell capitalize"
          :class="[column.class]"
        >
          <template v-if="allowEdit && index === 0">
            <slot name="selection-all" :member-list="memberList" />
          </template>
          <template v-else>{{ column.title }}</template>
        </div>
      </div>
    </template>
    <template #item="{ item: projectMember }: ProjectMemberRow">
      <div v-if="allowEdit" class="bb-grid-cell items-start !px-2">
        <div class="h-10 flex flex-col items-center justify-center">
          <slot name="selection" :member="projectMember" />
        </div>
      </div>
      <div class="bb-grid-cell items-start gap-x-2">
        <UserAvatar class="mt-1" :user="projectMember.user" />
        <div class="flex flex-col">
          <div class="flex flex-row items-center space-x-2">
            <template v-if="projectMember.user.email !== ALL_USERS_USER_EMAIL">
              <router-link
                :to="`/users/${projectMember.user.email}`"
                class="normal-link"
                >{{ projectMember.user.title }}</router-link
              >
              <YouTag v-if="currentUserV1.name === projectMember.user.name" />
            </template>
            <template v-else>
              <span class="text-gray-800 pt-2.5">{{
                projectMember.user.title
              }}</span>
            </template>
          </div>
          <span
            v-if="projectMember.user.email !== ALL_USERS_USER_EMAIL"
            class="textinfolabel"
            >{{ projectMember.user.email }}</span
          >
        </div>
      </div>
      <div class="bb-grid-cell flex-wrap gap-x-2 gap-y-1">
        <div class="w-full flex flex-col justify-start items-start">
          <p
            v-for="binding in getSortedBindingList(projectMember.bindingList)"
            :key="binding.role"
            class="w-auto py-0.5 flex flex-row justify-start items-center flex-nowrap gap-x-2"
            :class="isExpired(binding) ? 'line-through' : ''"
          >
            <span class="block truncate">{{
              displayRoleTitle(binding.role)
            }}</span>
            <span
              v-if="getBindingConditionTitle(binding)"
              class="block truncate text-blue-600 cursor-pointer hover:text-blue-800"
              @click="editingBinding = binding"
            >
              {{
                getBindingConditionTitle(binding) ||
                displayRoleTitle(binding.role)
              }}
            </span>
            <span v-if="isExpired(binding)"
              >({{ $t("project.members.expired") }})</span
            >
          </p>
        </div>
      </div>
      <div class="bb-grid-cell flex-wrap gap-x-2 gap-y-1">
        <div class="w-full flex flex-col justify-start items-start">
          <p
            v-for="binding in getSortedBindingList(projectMember.bindingList)"
            :key="binding.role"
            class="w-full py-0.5 truncate"
            :class="isExpired(binding) ? 'line-through' : ''"
          >
            <span>{{ getExpiredTimeString(binding) || "*" }}</span>
          </p>
        </div>
      </div>
      <div class="bb-grid-cell items-start gap-x-2 justify-end">
        <div class="h-10 flex flex-col items-center justify-center">
          <NTooltip v-if="allowEdit" trigger="hover">
            <template #trigger>
              <button
                class="cursor-pointer opacity-60 hover:opacity-100"
                @click="editingMember = projectMember"
              >
                <heroicons-outline:pencil class="w-4 h-4" />
              </button>
            </template>
            {{ $t("common.edit") }}
          </NTooltip>
          <NButton
            v-else-if="allowView(projectMember)"
            size="tiny"
            @click="editingMember = projectMember"
          >
            {{ $t("common.view") }}
          </NButton>
        </div>
      </div>
    </template>

    <template #placeholder-content>
      <div class="p-2">-</div>
    </template>
  </BBGrid>

  <ProjectMemberRolePanel
    v-if="editingMember !== null"
    :project="project"
    :member="editingMember"
    @close="editingMember = null"
  />

  <EditProjectRolePanel
    v-if="editingBinding"
    :project="project"
    :binding="editingBinding"
    @close="editingBinding = null"
  />
</template>

<script setup lang="ts">
import { NButton } from "naive-ui";
import { computed, ref } from "vue";
import { useI18n } from "vue-i18n";
import { type BBGridColumn, type BBGridRow, BBGrid } from "@/bbkit";
import UserAvatar from "@/components/User/UserAvatar.vue";
import YouTag from "@/components/misc/YouTag.vue";
import { featureToRef, useCurrentUserV1 } from "@/store";
import { ALL_USERS_USER_EMAIL, ComposedProject, PRESET_ROLES } from "@/types";
import { Binding } from "@/types/proto/v1/iam_policy";
import { displayRoleTitle } from "@/utils";
import {
  getExpiredTimeString,
  isExpired,
  getExpiredDateTime,
} from "../ProjectRoleTable/utils";
import { getBindingConditionTitle } from "../common/util";
import ProjectMemberRolePanel from "./ProjectMemberRolePanel.vue";
import { ComposedProjectMember } from "./types";

export type ProjectMemberRow = BBGridRow<ComposedProjectMember>;

const props = defineProps<{
  project: ComposedProject;
  allowEdit: boolean;
  memberList: ComposedProjectMember[];
  ready?: boolean;
}>();

const { t } = useI18n();
const hasRBACFeature = featureToRef("bb.feature.rbac");
const currentUserV1 = useCurrentUserV1();
const editingMember = ref<ComposedProjectMember | null>(null);
const editingBinding = ref<Binding | null>(null);

const columnList = computed(() => {
  const ACCOUNT: BBGridColumn = {
    title: t("common.user"),
    width: hasRBACFeature.value ? "minmax(min-content, auto)" : "1fr",
  };
  const ROLE: BBGridColumn = {
    title: t("common.role.self"),
    width: "minmax(min-content, auto)",
  };
  const EXPIRATION: BBGridColumn = {
    title: t("common.expiration"),
    width: "minmax(min-content, auto)",
  };
  const OPERATIONS: BBGridColumn = {
    title: "",
    width: "minmax(min-content, auto)",
  };
  const list = hasRBACFeature.value
    ? [ACCOUNT, ROLE, EXPIRATION, OPERATIONS]
    : [ACCOUNT, OPERATIONS];
  if (props.allowEdit) {
    list.unshift({
      title: "",
      width: "minmax(auto, 2rem)",
      class: "items-center !px-2",
    });
  }
  return list;
});

const allowView = (item: ComposedProjectMember) => {
  return item.user.name === currentUserV1.value.name;
};

const getSortedBindingList = (bindingList: Binding[]) => {
  let roleMap = new Map<string, Binding[]>();
  for (const binding of bindingList) {
    const role = binding.role;
    if (!roleMap.has(role)) {
      roleMap.set(role, []);
    }
    roleMap.get(role)?.push(binding);
  }
  // Sort by role type.
  roleMap = new Map(
    [...roleMap].sort((a, b) => {
      if (!PRESET_ROLES.includes(a[0])) return -1;
      if (!PRESET_ROLES.includes(b[0])) return 1;
      return PRESET_ROLES.indexOf(a[0]) - PRESET_ROLES.indexOf(b[0]);
    })
  );
  // Sort by expiration time.
  for (const role of roleMap.keys()) {
    roleMap.set(
      role,
      roleMap.get(role)?.sort((a, b) => {
        return (
          (getExpiredDateTime(b)?.getTime() ?? -1) -
          (getExpiredDateTime(a)?.getTime() ?? -1)
        );
      }) || []
    );
  }
  return Array.from(roleMap.values()).flat();
};
</script>
