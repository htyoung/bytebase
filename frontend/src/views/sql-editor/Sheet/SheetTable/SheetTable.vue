<template>
  <div class="w-full grid grid-cols-1">
    <div class="sheet-list-container text-sm text-gray-400" :class="view">
      <span v-for="header in headers" :key="header.key" class="uppercase">{{
        header.label
      }}</span>
    </div>
    <div
      v-if="loading"
      class="w-full flex flex-col py-6 justify-start items-center"
    >
      <span class="text-sm leading-6 text-gray-500">{{
        $t("sql-editor.loading-data")
      }}</span>
    </div>
    <div
      v-for="sheet in sheetList"
      :key="sheet.name"
      class="sheet-list-container text-sm cursor-pointer hover:bg-gray-100"
      :class="view"
      @click="handleSheetClick(sheet)"
    >
      <span
        v-for="value in getValueList(sheet)"
        :key="value.key"
        class="truncate w-5/6"
        >{{ value.value }}</span
      >
      <div class="flex flex-row justify-end items-center" @click.stop>
        <Dropdown :sheet="sheet" :view="view" @refresh="$emit('refresh')" />
      </div>
    </div>

    <div
      v-show="!loading && sheetList.length === 0"
      class="w-full flex flex-col py-6 justify-start items-center"
    >
      <heroicons-outline:inbox class="w-12 h-auto text-gray-500" />
      <span class="text-sm leading-6 text-gray-500">{{
        $t("common.no-data")
      }}</span>
    </div>
  </div>
</template>

<script lang="ts" setup>
import dayjs from "dayjs";
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";
import { SQL_EDITOR_SHARE_MODULE } from "@/router/sqlEditor";
import { useUserStore, useProjectV1Store } from "@/store";
import {
  Worksheet,
  Worksheet_Visibility,
} from "@/types/proto/v1/worksheet_service";
import { projectV1Name, worksheetSlugV1 } from "@/utils";
import { SheetViewMode } from "../types";
import Dropdown from "./Dropdown.vue";

const props = withDefaults(
  defineProps<{
    view: SheetViewMode;
    sheetList: Worksheet[];
    loading?: boolean;
  }>(),
  {
    view: "my",
    loading: false,
  }
);

defineEmits<{
  (event: "refresh"): void;
}>();

const { t } = useI18n();
const router = useRouter();
const projectV1Store = useProjectV1Store();

const showCreator = computed(() => {
  return props.view === "shared" || props.view === "starred";
});

const handleSheetClick = (sheet: Worksheet) => {
  router.push({
    name: SQL_EDITOR_SHARE_MODULE,
    params: {
      sheetSlug: worksheetSlugV1(sheet),
    },
  });
};

const headers = computed(() => {
  const labelList = [
    {
      key: "name",
      label: t("common.name"),
    },
    {
      key: "project",
      label: t("common.project"),
    },
    {
      key: "visibility",
      label: t("common.visibility"),
    },
  ];

  if (showCreator.value) {
    labelList.push({
      key: "creator",
      label: t("common.creator"),
    });
  }

  labelList.push({
    key: "updated",
    label: t("common.updated-at"),
  });

  return labelList;
});

const getValueList = (sheet: Worksheet) => {
  const project = projectV1Store.getProjectByName(sheet.project);
  let visibility = t("sql-editor.private");
  switch (sheet.visibility) {
    case Worksheet_Visibility.VISIBILITY_PROJECT:
      visibility = t("common.project");
      break;
    case Worksheet_Visibility.VISIBILITY_PUBLIC:
      visibility = t("sql-editor.public");
  }
  const valueList = [
    {
      key: "name",
      value: sheet.title,
    },
    {
      key: "project",
      value: projectV1Name(project),
    },
    {
      key: "visibility",
      value: visibility,
    },
  ];

  if (showCreator.value) {
    valueList.push({
      key: "creator",
      value:
        useUserStore().getUserByIdentifier(sheet.creator)?.title ??
        sheet.creator,
    });
  }

  valueList.push({
    key: "updated",
    value: dayjs
      .duration((sheet.updateTime ?? new Date()).getTime() - Date.now())
      .humanize(true),
  });

  return valueList;
};
</script>

<style scoped lang="postcss">
.sheet-list-container {
  @apply w-full grid py-3 px-4 border-b text-sm leading-6 select-none;
}
.sheet-list-container.my {
  grid-template-columns: 2fr repeat(4, 1fr) 32px;
}
.sheet-list-container.shared,
.sheet-list-container.starred {
  grid-template-columns: 2fr repeat(5, 1fr) 32px;
}
</style>
