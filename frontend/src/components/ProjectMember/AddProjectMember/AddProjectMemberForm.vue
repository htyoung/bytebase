<template>
  <div class="w-full flex flex-col justify-start items-start gap-y-4">
    <div class="w-full">
      <div class="flex items-center justify-between">
        {{ $t("project.members.select-users") }}

        <NButton v-if="allowRemove" text @click="$emit('remove')">
          <template #icon>
            <heroicons:trash class="w-4 h-4" />
          </template>
        </NButton>
      </div>
      <UserSelect
        v-model:users="state.userUidList"
        class="mt-2"
        style="width: 100%"
        :multiple="true"
        :include-all-users="true"
      />
    </div>
    <div class="w-full">
      <span>{{ $t("project.members.assign-role") }}</span>
      <ProjectRoleSelect v-model:role="state.role" class="mt-2" />
    </div>
    <div class="w-full">
      <span>{{ $t("common.reason") }}</span>
      <NInput
        v-model:value="state.reason"
        class="mt-2"
        type="textarea"
        rows="2"
        :placeholder="$t('project.members.assign-reason')"
      />
    </div>
    <div
      v-if="
        state.role === PresetRoleType.PROJECT_QUERIER ||
        state.role === PresetRoleType.PROJECT_EXPORTER
      "
      class="w-full"
    >
      <span class="block mb-2">{{ $t("common.databases") }}</span>
      <QuerierDatabaseResourceForm
        :project-id="project.uid"
        :database-resources="state.databaseResources"
        @update:condition="state.databaseResourceCondition = $event"
        @update:database-resources="state.databaseResources = $event"
      />
    </div>
    <template v-if="state.role === PresetRoleType.PROJECT_EXPORTER">
      <div class="w-full flex flex-col justify-start items-start">
        <span class="mb-2">
          {{ $t("issue.grant-request.export-rows") }}
        </span>
        <NInputNumber
          v-model:value="state.maxRowCount"
          required
          :placeholder="$t('issue.grant-request.export-rows')"
        />
      </div>
    </template>

    <div class="w-full">
      <span>{{ $t("common.expiration") }}</span>
      <ExpirationSelector
        class="mt-2"
        :options="expireDaysOptions"
        :value="state.expireDays"
        @update="state.expireDays = $event"
      />
    </div>
  </div>
</template>

<script lang="ts" setup>
/* eslint-disable vue/no-mutating-props */
import dayjs from "dayjs";
import { head } from "lodash-es";
import { NInputNumber, NInput } from "naive-ui";
import { computed, onMounted, reactive, watch } from "vue";
import { useI18n } from "vue-i18n";
import ExpirationSelector from "@/components/ExpirationSelector.vue";
import QuerierDatabaseResourceForm from "@/components/Issue/panel/RequestQueryPanel/DatabaseResourceForm/index.vue";
import ProjectRoleSelect from "@/components/v2/Select/ProjectRoleSelect.vue";
import { useUserStore } from "@/store";
import { getUserId } from "@/store/modules/v1/common";
import {
  ComposedProject,
  DatabaseResource,
  getUserEmailInBinding,
  PresetRoleType,
} from "@/types";
import { Expr } from "@/types/proto/google/type/expr";
import { Binding } from "@/types/proto/v1/iam_policy";
import { displayRoleTitle, extractDatabaseResourceName } from "@/utils";

const props = defineProps<{
  project: ComposedProject;
  binding: Binding;
  allowRemove: boolean;
}>();

defineEmits<{
  (event: "remove"): void;
}>();

interface LocalState {
  userUidList: string[];
  role?: string;
  reason: string;
  expireDays: number;
  // Querier and exporter options.
  databaseResourceCondition?: string;
  databaseResources?: DatabaseResource[];
  // Exporter options.
  maxRowCount: number;
  databaseId?: string;
}

const { t } = useI18n();
const userStore = useUserStore();
const state = reactive<LocalState>({
  userUidList: [],
  reason: "",
  // Default is never expires.
  expireDays: 0,
  // Exporter options.
  maxRowCount: 1000,
});

onMounted(() => {
  if (props.binding) {
    const userUidList = [];
    for (const member of props.binding.members) {
      // Member format: user:{email}
      const user = userStore.getUserByEmail(member.slice(5));
      if (user) {
        const userUid = getUserId(user.name);
        userUidList.push(String(userUid));
      }
    }
    state.userUidList = userUidList;
  }
});

const expireDaysOptions = computed(() => {
  if (state.role === PresetRoleType.PROJECT_EXPORTER) {
    return [
      {
        value: 1,
        label: t("common.date.days", { days: 1 }),
      },
      {
        value: 3,
        label: t("common.date.days", { days: 3 }),
      },
      {
        value: 30,
        label: t("common.date.days", { days: 30 }),
      },
      {
        value: 90,
        label: t("common.date.days", { days: 90 }),
      },
      {
        value: 0,
        label: t("project.members.never-expires"),
      },
    ];
  }
  return [
    {
      value: 7,
      label: t("common.date.days", { days: 7 }),
    },
    {
      value: 30,
      label: t("common.date.days", { days: 30 }),
    },
    {
      value: 90,
      label: t("common.date.days", { days: 90 }),
    },
    {
      value: 365,
      label: t("common.date.years", { years: 1 }),
    },
    {
      value: 0,
      label: t("project.members.never-expires"),
    },
  ];
});

watch(
  () => state.role,
  () => {
    state.databaseResourceCondition = undefined;
    state.databaseResources = undefined;
  },
  {
    immediate: true,
  }
);

watch(
  () => state,
  () => {
    const conditionName = generateConditionTitle();
    if (state.userUidList) {
      props.binding.members = state.userUidList.map((uid) => {
        const user = userStore.getUserById(uid);
        return getUserEmailInBinding(user!.email);
      });
    }
    if (state.role) {
      props.binding.role = state.role;
    }
    const expression: string[] = [];
    if (state.expireDays > 0) {
      const now = dayjs();
      const expiresAt = now.add(state.expireDays, "days");
      expression.push(`request.time < timestamp("${expiresAt.toISOString()}")`);
    }
    if (state.role === PresetRoleType.PROJECT_QUERIER) {
      if (state.databaseResourceCondition) {
        expression.push(state.databaseResourceCondition);
      }
    }
    if (state.role === PresetRoleType.PROJECT_EXPORTER) {
      if (state.databaseResourceCondition) {
        expression.push(state.databaseResourceCondition);
      }
      if (state.maxRowCount) {
        expression.push(`request.row_limit <= ${state.maxRowCount}`);
      }
    }
    if (expression.length > 0) {
      props.binding.condition = Expr.create({
        title: conditionName,
        description: state.reason,
        expression: expression.join(" && "),
      });
    } else {
      props.binding.condition = Expr.create({
        title: conditionName,
        description: state.reason,
        expression: undefined,
      });
    }
  },
  {
    deep: true,
  }
);

const generateConditionTitle = () => {
  if (!state.role) {
    return "";
  }

  let conditionSuffix = "";
  if (
    state.role === PresetRoleType.PROJECT_QUERIER ||
    state.role === PresetRoleType.PROJECT_QUERIER
  ) {
    if (!state.databaseResources || state.databaseResources.length === 0) {
      conditionSuffix = `${conditionSuffix} All`;
    } else if (state.databaseResources.length <= 3) {
      const databaseResourceNames = state.databaseResources.map((ds) =>
        getDatabaseResourceName(ds)
      );
      conditionSuffix = `${conditionSuffix} ${databaseResourceNames.join(
        ", "
      )}`;
    } else {
      const firstDatabaseResourceName = getDatabaseResourceName(
        head(state.databaseResources)!
      );
      conditionSuffix = `${conditionSuffix} ${firstDatabaseResourceName} and ${
        state.databaseResources.length - 1
      } more`;
    }
  }
  if (state.expireDays > 0) {
    const now = dayjs();
    const expiresAt = now.add(state.expireDays, "days");
    conditionSuffix = `${conditionSuffix} ${now.format("L")}-${expiresAt.format(
      "L"
    )}`;
  }

  if (conditionSuffix !== "") {
    return displayRoleTitle(state.role) + conditionSuffix;
  }
  return "";
};

const getDatabaseResourceName = (databaseResource: DatabaseResource) => {
  const { database } = extractDatabaseResourceName(
    databaseResource.databaseName
  );
  if (databaseResource.table) {
    if (databaseResource.schema) {
      return `${database}.${databaseResource.schema}.${databaseResource.table}`;
    } else {
      return `${database}.${databaseResource.table}`;
    }
  } else if (databaseResource.schema) {
    return `${database}.${databaseResource.schema}`;
  } else {
    return database;
  }
};

defineExpose({
  allowConfirm: computed(() => {
    if (state.userUidList.length <= 0) {
      return false;
    }
    if ((!state.expireDays && state.expireDays !== 0) || state.expireDays < 0) {
      return false;
    }
    // TODO: use parsed expression to check if the expression is valid.
    return true;
  }),
});
</script>
