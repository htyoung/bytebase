<template>
  <Drawer
    :show="true"
    :close-on-esc="!loading"
    :mask-closable="!loading"
    width="auto"
    @update:show="(show: boolean) => !show && $emit('close')"
  >
    <DrawerContent
      :title="$t('database.sync-schema.target-databases')"
      :closable="true"
      class="w-[30rem] max-w-[100vw] relative"
    >
      <div class="flex items-center justify-end mx-2 mb-2">
        <SearchBox
          v-model:value="state.searchText"
          :placeholder="$t('database.filter-database')"
        />
      </div>
      <NCollapse
        class="overflow-y-auto"
        arrow-placement="left"
        :default-expanded-names="
          databaseListGroupByEnvironment.map((group) => group.environment.uid)
        "
      >
        <NCollapseItem
          v-for="{
            environment,
            databaseList: databaseListInEnvironment,
          } in databaseListGroupByEnvironment"
          :key="environment.uid"
          :name="environment.uid"
        >
          <template #header>
            <label class="flex items-center gap-x-2" @click.stop.prevent>
              <NCheckbox
                v-bind="
                  getAllSelectionStateForEnvironment(
                    environment,
                    databaseListInEnvironment
                  )
                "
                @update:checked="
                  toggleAllDatabasesSelectionForEnvironment(
                    environment,
                    databaseListInEnvironment,
                    $event
                  )
                "
              />
              <EnvironmentV1Name :environment="environment" :link="false" />
            </label>
          </template>

          <template #header-extra>
            <div class="flex items-center text-xs text-gray-500 mr-2">
              {{
                $t(
                  "database.n-selected-m-in-total",
                  getSelectionStateSummaryForEnvironment(
                    environment,
                    databaseListInEnvironment
                  )
                )
              }}
            </div>
          </template>

          <div class="relative bg-white rounded-md -space-y-px px-2">
            <template
              v-for="database in databaseListInEnvironment"
              :key="database.uid"
            >
              <label
                class="border-control-border relative border p-3 flex flex-col gap-y-2 md:flex-row md:pl-4 md:pr-6"
                :class="
                  database.syncState === State.ACTIVE
                    ? 'cursor-pointer'
                    : 'cursor-not-allowed'
                "
              >
                <div class="radio text-sm flex justify-start md:flex-1">
                  <NCheckbox
                    :class="database.syncState !== State.ACTIVE && 'opacity-40'"
                    :checked="isDatabaseSelected(database.uid)"
                    :label="database.databaseName"
                    @update:checked="
                      (checked) => toggleDatabaseSelected(database.uid, checked)
                    "
                  />
                </div>
                <div
                  class="flex items-center gap-x-1 textinfolabel ml-6 pl-0 md:ml-0 md:pl-0 md:justify-end"
                >
                  <InstanceV1Name
                    :instance="database.instanceEntity"
                    :link="false"
                  />
                </div>
              </label>
            </template>
          </div>
        </NCollapseItem>
      </NCollapse>
      <MaskSpinner v-if="loading" />

      <template #footer>
        <div class="flex items-center justify-end gap-x-2">
          <NButton @click="$emit('close')">{{ $t("common.cancel") }}</NButton>
          <NButton
            :disabled="state.selectedDatabaseList.length === 0"
            type="primary"
            @click="handleConfirm"
          >
            {{ $t("common.select") }}
          </NButton>
        </div>
      </template>
    </DrawerContent>
  </Drawer>
</template>

<script setup lang="ts">
import { NCollapse, NCollapseItem, NButton, NCheckbox } from "naive-ui";
import { computed, reactive } from "vue";
import {
  Drawer,
  DrawerContent,
  EnvironmentV1Name,
  InstanceV1Name,
} from "@/components/v2";
import { useDatabaseV1Store, useEnvironmentV1Store } from "@/store";
import { ComposedDatabase } from "@/types";
import { Engine, State } from "@/types/proto/v1/common";
import { Environment } from "@/types/proto/v1/environment_service";
import MaskSpinner from "../misc/MaskSpinner.vue";

type LocalState = {
  searchText: string;
  selectedDatabaseList: ComposedDatabase[];
};

const props = defineProps<{
  projectId: string;
  engine: Engine;
  selectedDatabaseIdList: string[];
  loading?: boolean;
}>();

const emit = defineEmits<{
  (event: "close"): void;
  (event: "update", databaseIdList: string[]): void;
}>();

const environmentV1Store = useEnvironmentV1Store();
const databaseStore = useDatabaseV1Store();
const state = reactive<LocalState>({
  searchText: "",
  selectedDatabaseList: props.selectedDatabaseIdList.map((id) => {
    return databaseStore.getDatabaseByUID(id);
  }),
});

const databaseListGroupByEnvironment = computed(() => {
  const databaseList =
    databaseStore.databaseList
      .filter((db) => db.projectEntity.uid === props.projectId)
      .filter((db) => db.databaseName.includes(state.searchText))
      .filter((db) => db.instanceEntity.engine === props.engine) || [];
  const listByEnv = environmentV1Store.environmentList.map((environment) => {
    const list = databaseList.filter(
      (db) => db.effectiveEnvironment === environment.name
    );
    return {
      environment,
      databaseList: list,
    };
  });

  return listByEnv.filter((group) => group.databaseList.length > 0);
});

const isDatabaseSelected = (databaseId: string) => {
  const idList = state.selectedDatabaseList.map((db) => db.uid);
  return idList.includes(databaseId);
};

const toggleDatabaseSelected = (databaseId: string, selected: boolean) => {
  const index = state.selectedDatabaseList.findIndex(
    (db) => db.uid === databaseId
  );
  if (selected) {
    if (index < 0) {
      state.selectedDatabaseList.push(
        databaseStore.getDatabaseByUID(databaseId)
      );
    }
  } else {
    if (index >= 0) {
      state.selectedDatabaseList.splice(index, 1);
    }
  }
};

const toggleAllDatabasesSelectionForEnvironment = (
  environment: Environment,
  databaseList: ComposedDatabase[],
  on: boolean
) => {
  databaseList
    .filter((db) => db.effectiveEnvironment === environment.name)
    .forEach((db) => toggleDatabaseSelected(db.uid, on));
};

const getAllSelectionStateForEnvironment = (
  environment: Environment,
  databaseList: ComposedDatabase[]
): { checked: boolean; indeterminate: boolean } => {
  const set = new Set(
    state.selectedDatabaseList
      .filter((db) => db.effectiveEnvironment === environment.name)
      .map((db) => db.uid)
  );
  const checked = set.size > 0 && databaseList.every((db) => set.has(db.uid));
  const indeterminate = !checked && databaseList.some((db) => set.has(db.uid));

  return {
    checked,
    indeterminate,
  };
};

const getSelectionStateSummaryForEnvironment = (
  environment: Environment,
  databaseList: ComposedDatabase[]
) => {
  const set = new Set(
    state.selectedDatabaseList
      .filter((db) => db.effectiveEnvironment === environment.name)
      .map((db) => db.uid)
  );
  const selected = databaseList.filter((db) => set.has(db.uid)).length;
  const total = databaseList.length;

  return { selected, total };
};

const handleConfirm = async () => {
  const databaseIdList = state.selectedDatabaseList
    .filter((db) => db.databaseName.includes(state.searchText))
    .map((db) => db.uid);
  emit("update", databaseIdList);
};
</script>
