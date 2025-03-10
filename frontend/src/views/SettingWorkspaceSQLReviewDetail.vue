<template>
  <FeatureAttention custom-class="mb-4" feature="bb.feature.sql-review" />
  <SQLReviewCreation
    v-if="state.editMode"
    key="sql-review-creation"
    :policy="reviewPolicy"
    :name="reviewPolicy.name"
    :selected-environment="reviewPolicy.environment"
    :selected-rule-list="ruleListOfPolicy"
    @cancel="state.editMode = false"
  />
  <div v-else>
    <div
      class="flex flex-col gap-y-2 items-start md:items-center gap-x-2 justify-center md:flex-row"
    >
      <div class="flex-1 flex space-x-2 items-center justify-start">
        <BBBadge
          v-if="reviewPolicy.environment"
          :can-remove="false"
          :link="`/${reviewPolicy.environment.name}`"
        >
          <EnvironmentV1Name
            :environment="reviewPolicy.environment"
            :link="false"
          />
        </BBBadge>
        <div v-if="!reviewPolicy.enforce" class="whitespace-nowrap">
          <BBBadge
            :text="$t('sql-review.disabled')"
            :can-remove="false"
            :badge-style="'DISABLED'"
            ::badge-style="'DISABLED'"
          />
        </div>
        <BBTextField
          class="flex-1 text-xl md:text-3xl py-0.5 px-0.5 font-bold truncate"
          :disabled="!hasPermission"
          :required="true"
          :focus-on-mount="false"
          :ends-on-enter="true"
          :bordered="false"
          :value="reviewPolicy.name"
          @end-editing="changeName"
        />
      </div>
      <div v-if="hasPermission" class="flex gap-x-2">
        <NButton
          v-if="reviewPolicy.enforce"
          @click.prevent="state.showDisableModal = true"
        >
          {{ $t("common.disable") }}
        </NButton>
        <NButton v-else @click.prevent="state.showEnableModal = true">
          {{ $t("common.enable") }}
        </NButton>
        <NButton type="primary" @click="onEdit">
          {{ $t("sql-review.create.configure-rule.change-template") }}
        </NButton>
      </div>
    </div>
    <BBAttention
      v-if="!reviewPolicy.environment"
      class="my-4"
      type="warning"
      :title="$t('sql-review.create.basic-info.no-linked-environments')"
    />
    <SQLRuleFilter
      :rule-list="state.ruleList"
      :params="filterParams"
      v-on="filterEvents"
    />

    <SQLRuleTable
      :class="[state.updating && 'pointer-events-none']"
      :rule-list="filteredRuleList"
      :editable="hasPermission"
      @level-change="onLevelChange"
      @payload-change="onPayloadChange"
      @comment-change="onCommentChange"
    />
    <BBButtonConfirm
      class="!my-4"
      :disabled="!hasPermission"
      :style="'DELETE'"
      :button-text="$t('sql-review.delete')"
      :ok-text="$t('common.delete')"
      :confirm-title="$t('common.delete') + ` '${reviewPolicy.name}'?`"
      :require-confirm="true"
      @confirm="onRemove"
    />
    <div
      v-if="state.rulesUpdated"
      class="w-full mt-4 py-4 border-t border-block-border flex justify-between bg-white sticky bottom-0 z-10"
    >
      <NButton @click.prevent="onCancelChanges">
        <span> {{ $t("common.cancel") }}</span>
      </NButton>
      <NButton type="primary" @click.prevent="onApplyChanges">
        {{ $t("common.confirm-and-update") }}
      </NButton>
    </div>
  </div>
  <BBAlert
    v-model:show="state.showDisableModal"
    type="warning"
    :ok-text="$t('common.disable')"
    :title="$t('common.disable') + ` '${reviewPolicy.name}'?`"
    description=""
    @ok="
      () => {
        state.showDisableModal = false;
        onArchive();
      }
    "
    @cancel="state.showDisableModal = false"
  />
  <BBAlert
    v-model:show="state.showEnableModal"
    type="info"
    :ok-text="$t('common.enable')"
    :title="$t('common.enable') + ` '${reviewPolicy.name}'?`"
    description=""
    @ok="
      () => {
        state.showEnableModal = false;
        onRestore();
      }
    "
    @cancel="state.showEnableModal = false"
  />
</template>

<script lang="ts" setup>
import { cloneDeep, groupBy } from "lodash-es";
import { computed, reactive, toRef, watch, watchEffect } from "vue";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";
import { BBTextField } from "@/bbkit";
import {
  payloadValueListToComponentList,
  SQLRuleFilter,
  useSQLRuleFilter,
  SQLRuleTable,
} from "@/components/SQLReview/components";
import { PayloadForEngine } from "@/components/SQLReview/components/RuleConfigComponents";
import { EnvironmentV1Name } from "@/components/v2";
import { SETTING_ROUTE_WORKSPACE_SQL_REVIEW } from "@/router/dashboard/workspaceSetting";
import {
  pushNotification,
  useCurrentUserV1,
  useSQLReviewStore,
  useSubscriptionV1Store,
} from "@/store";
import {
  unknown,
  RuleTemplate,
  SQLReviewPolicy,
  RuleType,
  SchemaPolicyRule,
  TEMPLATE_LIST,
  convertPolicyRuleToRuleTemplate,
  ruleIsAvailableInSubscription,
  convertRuleTemplateToPolicyRule,
} from "@/types";
import { Engine } from "@/types/proto/v1/common";
import { SQLReviewRuleLevel } from "@/types/proto/v1/org_policy_service";
import { idFromSlug, hasWorkspacePermissionV2 } from "@/utils";

const props = defineProps<{
  sqlReviewPolicySlug: string;
}>();

interface LocalState {
  showDisableModal: boolean;
  showEnableModal: boolean;
  selectedCategory?: string;
  editMode: boolean;
  checkedEngine: Set<Engine>;
  checkedLevel: Set<SQLReviewRuleLevel>;
  ruleList: RuleTemplate[];
  rulesUpdated: boolean;
  updating: boolean;
}

const { t } = useI18n();
const store = useSQLReviewStore();
const router = useRouter();
const currentUserV1 = useCurrentUserV1();
const subscriptionStore = useSubscriptionV1Store();

const state = reactive<LocalState>({
  showDisableModal: false,
  showEnableModal: false,
  editMode: false,
  checkedEngine: new Set<Engine>(),
  checkedLevel: new Set<SQLReviewRuleLevel>(),
  ruleList: [],
  rulesUpdated: false,
  updating: false,
});

watchEffect(async () => {
  await store.getOrFetchReviewPolicyByEnvironmentId(
    idFromSlug(props.sqlReviewPolicySlug)
  );
});

const hasPermission = computed(() => {
  return hasWorkspacePermissionV2(currentUserV1.value, "bb.policies.update");
});

const reviewPolicy = computed((): SQLReviewPolicy => {
  return (
    store.getReviewPolicyByEnvironmentId(
      String(idFromSlug(props.sqlReviewPolicySlug))
    ) || (unknown("SQL_REVIEW") as SQLReviewPolicy)
  );
});

const ruleListOfPolicy = computed((): RuleTemplate[] => {
  if (!reviewPolicy.value) {
    return [];
  }

  const ruleTemplateList: RuleTemplate[] = [];
  const ruleTemplateMap: Map<RuleType, RuleTemplate> = TEMPLATE_LIST.reduce(
    (map, template) => {
      for (const rule of template.ruleList) {
        map.set(rule.type, rule);
      }
      return map;
    },
    new Map<RuleType, RuleTemplate>()
  );

  const groupByRule = groupBy(reviewPolicy.value.ruleList, (rule) => rule.type);

  for (const [type, ruleList] of Object.entries(groupByRule)) {
    const rule = ruleTemplateMap.get(type as RuleType);
    if (!rule) {
      continue;
    }

    const data = convertPolicyRuleToRuleTemplate(ruleList, rule);
    if (data) {
      ruleTemplateList.push(data);
      ruleTemplateMap.delete(type as RuleType);
    }
  }

  for (const rule of ruleTemplateMap.values()) {
    ruleTemplateList.push({
      ...rule,
      level: SQLReviewRuleLevel.DISABLED,
    });
  }

  return ruleTemplateList;
});

watch(
  ruleListOfPolicy,
  (ruleList) => {
    state.ruleList = cloneDeep(ruleList);
  },
  { immediate: true }
);

const {
  params: filterParams,
  events: filterEvents,
  filteredRuleList,
} = useSQLRuleFilter(toRef(state, "ruleList"));

const changeName = async (name: string) => {
  const policy = reviewPolicy.value;
  if (name === policy.name) {
    return;
  }
  const upsert = {
    name,
    ruleList: policy.ruleList,
  };

  await store.updateReviewPolicy({
    id: policy.id,
    ...upsert,
  });
  pushNotification({
    module: "bytebase",
    style: "SUCCESS",
    title: t("sql-review.policy-updated"),
  });
};

const markChange = (rule: RuleTemplate, overrides: Partial<RuleTemplate>) => {
  if (
    !ruleIsAvailableInSubscription(rule.type, subscriptionStore.currentPlan)
  ) {
    return;
  }

  const index = state.ruleList.findIndex((r) => r.type === rule.type);
  if (index < 0) {
    return;
  }
  const newRule = {
    ...state.ruleList[index],
    ...overrides,
  };
  state.ruleList[index] = newRule;
  state.rulesUpdated = true;
};

const onPayloadChange = (rule: RuleTemplate, data: PayloadForEngine) => {
  markChange(rule, payloadValueListToComponentList(rule, data));
};

const onLevelChange = (rule: RuleTemplate, level: SQLReviewRuleLevel) => {
  markChange(rule, { level });
};

const onCommentChange = (rule: RuleTemplate, comment: string) => {
  markChange(rule, { comment });
};

const onCancelChanges = () => {
  state.ruleList = cloneDeep(ruleListOfPolicy.value);
  state.rulesUpdated = false;
};

const onApplyChanges = async () => {
  const policy = reviewPolicy.value;

  const ruleList: SchemaPolicyRule[] = [];
  for (const rule of state.ruleList) {
    ruleList.push(...convertRuleTemplateToPolicyRule(rule));
  }

  state.updating = true;
  try {
    await store.updateReviewPolicy({
      id: policy.id,
      name: policy.name,
      ruleList,
    });
    state.rulesUpdated = false;
    pushNotification({
      module: "bytebase",
      style: "SUCCESS",
      title: t("sql-review.policy-updated"),
    });
  } finally {
    state.updating = false;
  }
};

const onEdit = () => {
  state.editMode = true;
  filterParams.searchText = "";
  filterParams.selectedCategory = undefined;
};

const onArchive = () => {
  store.updateReviewPolicy({
    id: reviewPolicy.value.id,
    enforce: false,
  });
};

const onRestore = () => {
  store.updateReviewPolicy({
    id: reviewPolicy.value.id,
    enforce: true,
  });
};

const onRemove = () => {
  store.removeReviewPolicy(reviewPolicy.value.id).then(() => {
    pushNotification({
      module: "bytebase",
      style: "SUCCESS",
      title: t("sql-review.policy-removed"),
    });
  });
  router.replace({
    name: SETTING_ROUTE_WORKSPACE_SQL_REVIEW,
  });
};
</script>
