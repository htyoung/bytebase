<template>
  <div class="space-y-4">
    <div v-if="allowEdit" class="flex items-center justify-end">
      <NButton
        type="primary"
        class="capitalize"
        @click.prevent="addProjectWebhook"
      >
        {{ $t("project.webhook.add-a-webhook") }}
      </NButton>
    </div>
    <div v-if="projectWebhookList.length > 0" class="space-y-6">
      <template
        v-for="(projectWebhook, index) in projectWebhookList"
        :key="index"
      >
        <ProjectWebhookCard :project-webhook="projectWebhook" />
      </template>
    </div>
    <NoDataPlaceholder v-else>
      <div class="text-center">
        <h3 class="mt-2 text-sm font-medium text-main">
          {{ $t("project.webhook.no-webhook.title") }}
        </h3>
        <p class="mt-1 text-sm text-control-light">
          {{ $t("project.webhook.no-webhook.content") }}
        </p>
        <div v-if="allowEdit" class="mt-4">
          <NButton
            size="small"
            type="primary"
            @click.prevent="addProjectWebhook"
          >
            {{ $t("project.webhook.add-a-webhook") }}
          </NButton>
        </div>
      </div>
    </NoDataPlaceholder>
  </div>
</template>

<script lang="ts" setup>
import { computed, PropType } from "vue";
import { useRouter } from "vue-router";
import { PROJECT_V1_ROUTE_WEBHOOK_CREATE } from "@/router/dashboard/projectV1";
import { Project } from "@/types/proto/v1/project_service";
import ProjectWebhookCard from "./ProjectWebhookCard.vue";

const props = defineProps({
  project: {
    required: true,
    type: Object as PropType<Project>,
  },
  allowEdit: {
    default: true,
    type: Boolean,
  },
});
const router = useRouter();

const projectWebhookList = computed(() => {
  return props.project.webhooks;
});

const addProjectWebhook = () => {
  router.push({
    name: PROJECT_V1_ROUTE_WEBHOOK_CREATE,
  });
};
</script>
