<template>
  <div class="flex flex-row items-center justify-between gap-x-2 group">
    <div class="flex flex-row items-center gap-x-2">
      <NTag>
        <span class="inline-block w-[30px] text-center">
          {{ displaySemanticType(changeHistory.type) }}
        </span>
      </NTag>
      <div
        class="flex flex-row items-center gap-x-1 border-b border-transparent group-hover:border-control-border cursor-pointer"
        @click="$emit('click-item', change)"
      >
        <RichDatabaseName
          :database="database"
          :show-instance="false"
          :show-arrow="false"
          :show-production-environment-icon="false"
          tooltip="instance"
        />
        <span>@</span>
        <span class="text-sm">{{ changeHistory.version }}</span>
        <router-link
          :to="{
            name: PROJECT_V1_ROUTE_ISSUE_DETAIL,
            params: {
              projectId: extractProjectResourceName(changeHistory.issue),
              issueSlug: extractIssueUID(changeHistory.issue),
            },
          }"
          class="normal-link text-sm hover:!no-underline"
          target="_blank"
          @click.stop
        >
          #{{ extractIssueUID(changeHistory.issue) }}
        </router-link>
      </div>
    </div>
    <div
      class="flex flex-row items-center justify-end gap-x-1 invisible group-hover:visible"
    >
      <NButton
        size="small"
        quaternary
        style="--n-padding: 0 6px"
        @click.stop="$emit('remove-item', change)"
      >
        <template #icon>
          <heroicons:x-mark />
        </template>
      </NButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import { NButton, NTag } from "naive-ui";
import { computed } from "vue";
import { RichDatabaseName } from "@/components/v2";
import { PROJECT_V1_ROUTE_ISSUE_DETAIL } from "@/router/dashboard/projectV1";
import { useChangeHistoryStore, useDatabaseV1Store } from "@/store";
import { Changelist_Change as Change } from "@/types/proto/v1/changelist_service";
import { ChangeHistory } from "@/types/proto/v1/database_service";
import {
  extractDatabaseResourceName,
  extractIssueUID,
  extractProjectResourceName,
} from "@/utils";
import { displaySemanticType } from "./utils";

const props = defineProps<{
  change: Change;
}>();

defineEmits<{
  (event: "click-item", change: Change): void;
  (event: "remove-item", change: Change): void;
}>();

const changeHistory = computed(() => {
  const name = props.change.source;
  return (
    useChangeHistoryStore().getChangeHistoryByName(name) ??
    ChangeHistory.fromPartial({
      name,
      version: "<<Unknown Change History>>",
    })
  );
});

const database = computed(() => {
  const { full: databaseResourceName } = extractDatabaseResourceName(
    changeHistory.value.name
  );
  return useDatabaseV1Store().getDatabaseByName(databaseResourceName);
});
</script>
