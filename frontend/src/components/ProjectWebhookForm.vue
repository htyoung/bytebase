<template>
  <div class="space-y-4">
    <div v-if="create">
      <div>
        <label for="name" class="textlabel">
          {{ $t("project.webhook.destination") }}
          <span class="text-red-600">*</span>
        </label>
      </div>
      <NRadioGroup class="w-full mt-1" :value="state.webhook.type">
        <div class="grid grid-cols-1 gap-4 sm:grid-cols-7">
          <template
            v-for="(item, index) in projectWebhookV1TypeItemList()"
            :key="index"
          >
            <div
              class="flex justify-center px-2 py-4 rounded border border-control-border hover:bg-control-bg-hover cursor-pointer"
              @click.capture="state.webhook.type = item.type"
            >
              <div class="flex flex-col items-center">
                <WebhookTypeIcon :type="item.type" class="h-10 w-10" />
                <p class="mt-1 text-center textlabel">
                  {{ item.name }}
                </p>
                <div class="mt-3 radio text-sm">
                  <NRadio :value="item.type" />
                </div>
              </div>
            </div>
          </template>
        </div>
      </NRadioGroup>
    </div>
    <div>
      <label for="name" class="textlabel">
        {{ $t("common.name") }} <span class="text-red-600">*</span>
      </label>
      <NInput
        id="name"
        v-model:value="state.webhook.title"
        name="name"
        class="mt-1 w-full"
        :placeholder="namePlaceholder"
        :disabled="!allowEdit"
      />
    </div>
    <div>
      <label for="url" class="textlabel">
        {{ $t("project.webhook.webhook-url") }}
        <span class="text-red-600">*</span>
      </label>
      <div class="mt-1 textinfolabel">
        <template v-if="state.webhook.type === Webhook_Type.TYPE_SLACK">
          {{
            $t("project.webhook.creation.desc", {
              destination: $t("common.slack"),
            })
          }}
          <a
            href="https://api.slack.com/messaging/webhooks"
            target="__blank"
            class="normal-link"
            >{{
              $t("project.webhook.creation.view-doc", {
                destination: $t("common.slack"),
              })
            }}</a
          >.
        </template>
        <template v-else-if="state.webhook.type === Webhook_Type.TYPE_DISCORD">
          {{
            $t("project.webhook.creation.desc", {
              destination: $t("common.discord"),
            })
          }}
          <a
            href="https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks"
            target="__blank"
            class="normal-link"
            >{{
              $t("project.webhook.creation.view-doc", {
                destination: $t("common.discord"),
              })
            }}</a
          >.
        </template>
        <template v-else-if="state.webhook.type === Webhook_Type.TYPE_TEAMS">
          {{
            $t("project.webhook.creation.desc", {
              destination: $t("common.teams"),
            })
          }}
          <a
            href="https://docs.microsoft.com/en-us/microsoftteams/platform/webhooks-and-connectors/how-to/add-incoming-webhook"
            target="__blank"
            class="normal-link"
            >{{
              $t("project.webhook.creation.view-doc", {
                destination: $t("common.teams"),
              })
            }}</a
          >.
        </template>
        <template v-else-if="state.webhook.type === Webhook_Type.TYPE_DINGTALK">
          {{
            $t("project.webhook.creation.desc", {
              destination: $t("common.dingtalk"),
            }) +
            ". " +
            $t("project.webhook.creation.how-to-protect")
          }}
          <a
            href="https://developers.dingtalk.com/document/robots/custom-robot-access"
            target="__blank"
            class="normal-link"
            >{{
              $t("project.webhook.creation.view-doc", {
                destination: $t("common.dingtalk"),
              })
            }}</a
          >.
        </template>
        <template v-else-if="state.webhook.type === Webhook_Type.TYPE_FEISHU">
          {{
            $t("project.webhook.creation.desc", {
              destination: $t("common.feishu"),
            }) +
            ". " +
            $t("project.webhook.creation.how-to-protect")
          }}

          <a
            href="https://open.feishu.cn/document/client-docs/bot-v3/add-custom-bot"
            target="__blank"
            class="normal-link"
            >{{
              $t("project.webhook.creation.view-doc", {
                destination: $t("common.feishu"),
              })
            }}</a
          >.
        </template>
        <!-- WeCom doesn't seem to provide official webhook setup guide for the enduser -->
        <template v-else-if="state.webhook.type === Webhook_Type.TYPE_WECOM">
          {{
            $t("project.webhook.creation.desc", {
              destination: $t("common.wecom"),
            })
          }}
          <a
            href="https://open.work.weixin.qq.com/help2/pc/14931"
            target="__blank"
            class="normal-link"
            >{{
              $t("project.webhook.creation.view-doc", {
                destination: $t("common.wecom"),
              })
            }}</a
          >.
        </template>
        <template v-else-if="state.webhook.type === Webhook_Type.TYPE_CUSTOM">
          {{
            $t("project.webhook.creation.desc", {
              destination: $t("common.custom"),
            })
          }}
          <a
            href="https://www.bytebase.com/docs/change-database/webhook#custom?source=console"
            target="__blank"
            class="normal-link"
            >{{
              $t("project.webhook.creation.view-doc", {
                destination: $t("common.custom"),
              })
            }}</a
          >.
        </template>
      </div>
      <NInput
        id="url"
        v-model:value="state.webhook.url"
        name="url"
        class="mt-1 w-full"
        :placeholder="urlPlaceholder"
        :disabled="!allowEdit"
      />
    </div>
    <div>
      <div class="text-md leading-6 font-medium text-main">
        {{ $t("project.webhook.triggering-activity") }}
      </div>
      <div
        v-for="(item, index) in projectWebhookV1ActivityItemList()"
        :key="index"
        class="mt-4 space-y-4"
      >
        <div>
          <NCheckbox
            :label="item.title"
            :checked="isEventOn(item.activity)"
            @update:checked="
            (on: boolean) => {
              toggleEvent(item.activity, on);
            }
          "
          />
          <div class="textinfolabel">{{ item.label }}</div>
        </div>
      </div>
    </div>
    <NButton @click.prevent="testWebhook">
      {{ $t("project.webhook.test-webhook") }}
    </NButton>
    <div
      class="flex pt-5"
      :class="!create && allowEdit ? 'justify-between' : 'justify-end'"
    >
      <BBButtonConfirm
        v-if="!create && allowEdit"
        :style="'DELETE'"
        :button-text="$t('project.webhook.deletion.btn-text')"
        :ok-text="$t('common.delete')"
        :confirm-title="
          $t('project.webhook.deletion.confirm-title', { title: webhook.title })
        "
        :confirm-description="$t('common.cannot-undo-this-action')"
        :require-confirm="true"
        @confirm="deleteWebhook"
      />
      <div class="space-x-3">
        <NButton @click.prevent="cancel">
          {{ allowEdit ? $t("common.cancel") : $t("common.back") }}
        </NButton>
        <template v-if="allowEdit">
          <NButton
            v-if="create"
            type="primary"
            :disabled="!allowCreate"
            @click.prevent="createWebhook"
          >
            {{ $t("common.create") }}
          </NButton>
          <NButton
            v-else
            type="primary"
            :disabled="
              !valueChanged || state.webhook.notificationTypes.length === 0
            "
            @click.prevent="updateWebhook"
          >
            {{ $t("common.update") }}
          </NButton>
        </template>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { cloneDeep, isEmpty, isEqual } from "lodash-es";
import { NCheckbox, NRadio, NRadioGroup } from "naive-ui";
import { reactive, computed, PropType, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";
import {
  PROJECT_V1_ROUTE_WEBHOOKS,
  PROJECT_V1_ROUTE_WEBHOOK_DETAIL,
} from "@/router/dashboard/projectV1";
import {
  pushNotification,
  useProjectWebhookV1Store,
  useGracefulRequest,
} from "@/store";
import {
  projectWebhookV1ActivityItemList,
  projectWebhookV1TypeItemList,
} from "@/types";
import {
  Activity_Type,
  Project,
  Webhook,
  Webhook_Type,
} from "@/types/proto/v1/project_service";
import { projectWebhookV1Slug } from "../utils";

interface LocalState {
  webhook: Webhook;
}

const props = defineProps({
  allowEdit: {
    default: true,
    type: Boolean,
  },
  create: {
    type: Boolean,
    default: false,
  },
  project: {
    required: true,
    type: Object as PropType<Project>,
  },
  webhook: {
    required: true,
    type: Object as PropType<Webhook>,
  },
});

const router = useRouter();
const { t } = useI18n();

const projectWebhookV1Store = useProjectWebhookV1Store();
const state = reactive<LocalState>({
  webhook: cloneDeep(props.webhook),
});

watch(
  () => props.webhook,
  (webhook) => {
    state.webhook = cloneDeep(webhook);
  }
);

const namePlaceholder = computed(() => {
  if (state.webhook.type === Webhook_Type.TYPE_SLACK) {
    return `${t("common.slack")} Webhook`;
  } else if (state.webhook.type === Webhook_Type.TYPE_DISCORD) {
    return `${t("common.discord")} Webhook`;
  } else if (state.webhook.type === Webhook_Type.TYPE_TEAMS) {
    return `${t("common.teams")} Webhook`;
  } else if (state.webhook.type === Webhook_Type.TYPE_DINGTALK) {
    return `${t("common.dingtalk")} Webhook`;
  } else if (state.webhook.type === Webhook_Type.TYPE_FEISHU) {
    return `${t("common.feishu")} Webhook`;
  } else if (state.webhook.type === Webhook_Type.TYPE_WECOM) {
    return `${t("common.wecom")} Webhook`;
  } else if (state.webhook.type === Webhook_Type.TYPE_CUSTOM) {
    return `${t("common.custom")} Webhook`;
  }

  return "My Webhook";
});

const urlPlaceholder = computed(() => {
  if (state.webhook.type === Webhook_Type.TYPE_SLACK) {
    return "https://hooks.slack.com/services/...";
  } else if (state.webhook.type === Webhook_Type.TYPE_DISCORD) {
    return "https://discord.com/api/webhooks/...";
  } else if (state.webhook.type === Webhook_Type.TYPE_TEAMS) {
    return "https://acme123.webhook.office.com/webhookb2/...";
  } else if (state.webhook.type === Webhook_Type.TYPE_DINGTALK) {
    return "https://oapi.dingtalk.com/robot/...";
  } else if (state.webhook.type === Webhook_Type.TYPE_FEISHU) {
    return "https://open.feishu.cn/open-apis/bot/v2/hook/...";
  } else if (state.webhook.type === Webhook_Type.TYPE_WECOM) {
    return "https://qyapi.weixin.qq.com/cgi-bin/webhook/...";
  } else if (state.webhook.type === Webhook_Type.TYPE_CUSTOM) {
    return "https://example.com/api/webhook/...";
  }

  return "Webhook URL";
});

const valueChanged = computed(() => {
  return !isEqual(props.webhook, state.webhook);
});

const allowCreate = computed(() => {
  return (
    !isEmpty(state.webhook.title) &&
    !isEmpty(state.webhook.url) &&
    !isEmpty(state.webhook.notificationTypes)
  );
});

const isEventOn = (type: Activity_Type) => {
  return state.webhook.notificationTypes.includes(type);
};

const toggleEvent = (type: Activity_Type, on: boolean) => {
  if (on) {
    if (state.webhook.notificationTypes.includes(type)) {
      return;
    }
    state.webhook.notificationTypes.push(type);
  } else {
    const index = state.webhook.notificationTypes.indexOf(type);
    if (index >= 0) {
      state.webhook.notificationTypes.splice(index, 1);
    }
  }
  state.webhook.notificationTypes.sort();
};

const cancel = () => {
  router.push({
    name: PROJECT_V1_ROUTE_WEBHOOKS,
  });
};

const createWebhook = () => {
  useGracefulRequest(async () => {
    const { webhook } = state;
    const updatedProject = await projectWebhookV1Store.createProjectWebhook(
      props.project,
      webhook
    );

    pushNotification({
      module: "bytebase",
      style: "SUCCESS",
      title: t("project.webhook.success-created-prompt", {
        name: webhook.title,
      }),
    });
    const createdWebhook = updatedProject.webhooks.find((wh) => {
      return (
        wh.title === webhook.title &&
        wh.type == webhook.type &&
        wh.url === webhook.url
      );
    });
    if (createdWebhook) {
      router.push({
        name: PROJECT_V1_ROUTE_WEBHOOK_DETAIL,
        params: {
          projectWebhookSlug: projectWebhookV1Slug(createdWebhook),
        },
      });
    }
  });
};

const updateWebhook = () => {
  useGracefulRequest(async () => {
    const updateMask: string[] = [];
    if (state.webhook.title !== props.webhook.title) {
      updateMask.push("title");
    }
    if (state.webhook.url !== props.webhook.url) {
      updateMask.push("url");
    }
    if (
      !isEqual(state.webhook.notificationTypes, props.webhook.notificationTypes)
    ) {
      updateMask.push("notification_type");
    }
    await projectWebhookV1Store.updateProjectWebhook(state.webhook, updateMask);
    pushNotification({
      module: "bytebase",
      style: "SUCCESS",
      title: t("project.webhook.success-updated-prompt", {
        name: state.webhook.title,
      }),
    });
  });
};

const deleteWebhook = () => {
  useGracefulRequest(async () => {
    const name = state.webhook.title;
    await projectWebhookV1Store.deleteProjectWebhook(state.webhook);
    pushNotification({
      module: "bytebase",
      style: "SUCCESS",
      title: t("project.webhook.success-deleted-prompt", {
        name,
      }),
    });
    cancel();
  });
};

const testWebhook = () => {
  useGracefulRequest(async () => {
    const result = await useProjectWebhookV1Store().testProjectWebhook(
      props.project,
      state.webhook
    );

    if (result.error) {
      pushNotification({
        module: "bytebase",
        style: "CRITICAL",
        title: t("project.webhook.fail-tested-title"),
        description: result.error,
        manualHide: true,
      });
    } else {
      pushNotification({
        module: "bytebase",
        style: "SUCCESS",
        title: t("project.webhook.success-tested-prompt"),
      });
    }
  });
};
</script>
