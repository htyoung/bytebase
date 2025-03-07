<template>
  <div class="space-y-4">
    <div class="flex justify-end">
      <div v-if="vcsWithUIType" class="flex flex-row items-center space-x-2">
        <div class="textlabel whitespace-nowrap">
          {{ vcsWithUIType.title }}
        </div>
        <VCSIcon custom-class="h-6" :type="vcsWithUIType.type" />
      </div>
    </div>

    <div>
      <label for="instanceurl" class="textlabel">
        {{ $t("common.instance") }} URL
      </label>
      <BBTextField
        id="instanceurl"
        name="instanceurl"
        class="mt-1 w-full"
        :disabled="true"
        :value="vcs?.url"
      />
    </div>

    <div>
      <label for="name" class="textlabel">
        {{ $t("gitops.setting.add-git-provider.basic-info.display-name") }}
      </label>
      <p class="mt-1 textinfolabel">
        {{
          $t("gitops.setting.add-git-provider.basic-info.display-name-label")
        }}
      </p>
      <BBTextField
        id="name"
        v-model:value="state.title"
        name="name"
        class="mt-1 w-full"
      />
    </div>

    <div>
      <label for="applicationid" class="textlabel">
        {{ $t("common.application") }} ID
      </label>
      <p class="mt-1 textinfolabel">
        <template v-if="vcsUIType == 'GITLAB_SELF_HOST'">
          {{
            $t(
              "gitops.setting.git-provider.gitlab-self-host-application-id-label"
            )
          }}
          <a :href="adminApplicationUrl" target="_blank" class="normal-link">{{
            $t("gitops.setting.git-provider.view-in-gitlab")
          }}</a>
        </template>
        <template v-else-if="vcsUIType == 'GITLAB_COM'">
          {{
            $t("gitops.setting.git-provider.gitlab-com-application-id-label")
          }}
          <a :href="adminApplicationUrl" target="_blank" class="normal-link">{{
            $t("gitops.setting.git-provider.view-in-gitlab")
          }}</a>
        </template>
        <template v-else-if="vcsUIType == 'GITHUB_COM'">
          {{ $t("gitops.setting.git-provider.github-application-id-label") }}
        </template>
        <template v-else-if="vcsUIType == 'AZURE_DEVOPS'">
          {{ $t("gitops.setting.git-provider.azure-application-id-label") }}
          <a :href="adminApplicationUrl" target="_blank" class="normal-link">{{
            $t("gitops.setting.git-provider.view-in-azure")
          }}</a>
        </template>
      </p>
      <BBTextField
        id="applicationid"
        v-model:value="state.applicationId"
        name="applicationid"
        class="mt-1 w-full"
      />
    </div>

    <div>
      <label for="secret" class="textlabel"> Secret </label>
      <p class="mt-1 textinfolabel">
        <template v-if="vcsUIType == 'GITLAB_SELF_HOST'">
          {{ $t("gitops.setting.git-provider.gitlab-self-host-secret-label") }}
        </template>
        <template v-else-if="vcsUIType == 'GITLAB_COM'">
          {{ $t("gitops.setting.git-provider.gitlab-com-secret-label") }}
        </template>
        <template v-else-if="vcsUIType == 'GITHUB_COM'">
          {{ $t("gitops.setting.git-provider.secret-label-github") }}
        </template>
        <template v-else-if="vcsUIType == 'AZURE_DEVOPS'">
          {{ $t("gitops.setting.git-provider.azure-secret-label") }}
        </template>
      </p>
      <BBTextField
        id="secret"
        v-model:value="state.secret"
        name="secret"
        class="mt-1 w-full"
        :placeholder="$t('common.sensitive-placeholder')"
      />
    </div>

    <div class="pt-4 flex border-t justify-between">
      <template v-if="repositoryList.length == 0">
        <BBButtonConfirm
          :style="'DELETE'"
          :button-text="$t('gitops.setting.git-provider.delete')"
          :ok-text="$t('common.delete')"
          :confirm-title="
            $t('gitops.setting.git-provider.delete-confirm', {
              name: vcs?.title,
            })
          "
          :disabled="!hasDeleteVCSPermission"
          :require-confirm="true"
          @confirm="deleteVCS"
        />
      </template>
      <template v-else>
        <div class="mt-1 textinfolabel">
          {{ $t("gitops.setting.git-provider.delete-forbidden") }}
        </div>
      </template>
      <div class="space-x-3">
        <NButton @click.prevent="cancel">
          {{ $t("common.cancel") }}
        </NButton>
        <NButton
          type="primary"
          :disabled="!allowUpdate"
          @click.prevent="doUpdate"
        >
          {{ $t("common.update") }}
        </NButton>
      </div>
    </div>
  </div>

  <div class="py-6">
    <div class="text-lg leading-6 font-medium text-main">
      {{ $t("repository.linked") + ` (${repositoryList.length})` }}
    </div>
    <div class="mt-4">
      <RepositoryTable :repository-list="repositoryList" />
    </div>
  </div>
</template>

<script lang="ts" setup>
import isEmpty from "lodash-es/isEmpty";
import { reactive, computed, watchEffect, onMounted, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import RepositoryTable from "@/components/RepositoryTable.vue";
import { vcsListByUIType } from "@/components/VCS/utils";
import { SETTING_ROUTE_WORKSPACE_GITOPS } from "@/router/dashboard/workspaceSetting";
import {
  pushNotification,
  useCurrentUserV1,
  useRepositoryV1Store,
  useVCSV1Store,
} from "@/store";
import {
  openWindowForOAuth,
  OAuthWindowEventPayload,
  VCSUIType,
} from "@/types";
import {
  OAuthToken,
  ExternalVersionControl,
  ExternalVersionControl_Type,
} from "@/types/proto/v1/externalvs_service";
import {
  idFromSlug,
  uidFromSlug,
  getVCSUIType,
  hasWorkspacePermissionV2,
} from "@/utils";

interface LocalState {
  title: string;
  applicationId: string;
  secret: string;
  oAuthResultCallback?: (token: OAuthToken | undefined) => void;
}

const props = defineProps<{
  vcsSlug: string;
}>();

const router = useRouter();
const currentUser = useCurrentUserV1();
const vcsV1Store = useVCSV1Store();
const repositoryV1Store = useRepositoryV1Store();

const vcs = computed((): ExternalVersionControl | undefined => {
  return vcsV1Store.getVCSByUid(idFromSlug(props.vcsSlug));
});

const vcsUIType = computed((): VCSUIType => {
  if (vcs.value) {
    return getVCSUIType(vcs.value);
  }
  return "GITLAB_SELF_HOST";
});

const vcsWithUIType = computed(() => {
  return vcsListByUIType.value.find((data) => data.uiType === vcsUIType.value);
});

const state = reactive<LocalState>({
  title: vcs.value?.title ?? "",
  applicationId: vcs.value?.applicationId ?? "",
  secret: "",
});

const hasUpdateVCSPermission = computed(() => {
  return hasWorkspacePermissionV2(
    currentUser.value,
    "bb.externalVersionControls.update"
  );
});

const hasDeleteVCSPermission = computed(() => {
  return hasWorkspacePermissionV2(
    currentUser.value,
    "bb.externalVersionControls.delete"
  );
});

onMounted(() => {
  window.addEventListener("bb.oauth.register-vcs", eventListener, false);
});

onUnmounted(() => {
  window.removeEventListener("bb.oauth.register-vcs", eventListener);
});

const eventListener = (event: Event) => {
  const payload = (event as CustomEvent).detail as OAuthWindowEventPayload;
  if (isEmpty(payload.error)) {
    if (
      vcs.value?.type === ExternalVersionControl_Type.GITLAB ||
      vcs.value?.type === ExternalVersionControl_Type.GITHUB ||
      vcs.value?.type === ExternalVersionControl_Type.BITBUCKET
    ) {
      vcsV1Store
        .exchangeToken({
          code: payload.code,
          vcsName: vcs.value.name,
          clientId: state.applicationId,
          clientSecret: state.secret,
        })
        .then((token: OAuthToken) => {
          state.oAuthResultCallback!(token);
        })
        .catch(() => {
          state.oAuthResultCallback!(undefined);
        });
    }
  } else {
    state.oAuthResultCallback!(undefined);
  }
};

watchEffect(async () => {
  const vcs = await vcsV1Store.fetchVCSByUid(uidFromSlug(props.vcsSlug));
  if (vcs) {
    if (
      !hasWorkspacePermissionV2(
        currentUser.value,
        "bb.externalVersionControls.listProjects"
      )
    ) {
      pushNotification({
        module: "bytebase",
        style: "CRITICAL",
        title: `You don't have permission to list projects in '${vcs.title}'`,
      });
      return;
    }
    await repositoryV1Store.fetchRepositoryListByVCS(vcs.name);
  }
});

const adminApplicationUrl = computed(() => {
  switch (vcsUIType.value) {
    case "AZURE_DEVOPS":
      return `https://app.vsaex.visualstudio.com/app/view?clientId=${vcs.value?.applicationId}`;
    case "GITLAB_COM":
      return "https://gitlab.com/-/profile/applications";
    case "GITLAB_SELF_HOST":
      return `${vcs.value?.url}/admin/applications`;
    default:
      return "";
  }
});

const repositoryList = computed(() =>
  repositoryV1Store.getRepositoryListByVCS(vcs.value?.name ?? "")
);

const allowUpdate = computed(() => {
  return (
    (state.title != vcs.value?.title ||
      state.applicationId != vcs.value?.applicationId ||
      !isEmpty(state.secret)) &&
    hasUpdateVCSPermission.value
  );
});

const doUpdate = () => {
  if (!vcs.value) {
    return;
  }

  if (
    state.applicationId != vcs.value.applicationId ||
    !isEmpty(state.secret)
  ) {
    let authorizeUrl = `${vcs.value.url}/oauth/authorize`;
    if (vcs.value.type === ExternalVersionControl_Type.GITHUB) {
      authorizeUrl = `https://github.com/login/oauth/authorize`;
    } else if (vcs.value.type === ExternalVersionControl_Type.BITBUCKET) {
      authorizeUrl = `https://bitbucket.org/site/oauth2/authorize`;
    } else if (vcs.value.type === ExternalVersionControl_Type.AZURE_DEVOPS) {
      authorizeUrl = "https://app.vssps.visualstudio.com/oauth2/authorize";
    }
    const newWindow = openWindowForOAuth(
      authorizeUrl,
      state.applicationId,
      "bb.oauth.register-vcs",
      vcs.value.type
    );
    if (newWindow) {
      state.oAuthResultCallback = (token: OAuthToken | undefined) => {
        if (!vcs.value) {
          return;
        }
        if (token) {
          const vcsPatch: Partial<ExternalVersionControl> = {
            name: vcs.value.name,
          };
          if (state.title != vcs.value.title) {
            vcsPatch.title = state.title;
          }
          if (state.applicationId != vcs.value.applicationId) {
            vcsPatch.applicationId = state.applicationId;
          }
          if (!isEmpty(state.secret)) {
            vcsPatch.secret = state.secret;
          }
          vcsV1Store
            .updateVCS(vcsPatch)
            .then((vcs: ExternalVersionControl | undefined) => {
              if (!vcs) {
                return;
              }
              pushNotification({
                module: "bytebase",
                style: "SUCCESS",
                title: `Successfully updated '${vcs.title}'`,
              });
            });
        } else {
          // If the application ID mismatches, the OAuth workflow will stop early.
          // So the only possibility to reach here is we have a matching application ID, while
          // we failed to exchange a token, and it's likely we are requesting with a wrong secret.
          let description = "";
          if (vcs.value.type == ExternalVersionControl_Type.GITLAB) {
            description =
              "Please make sure Secret matches the one from your GitLab instance Application.";
          } else if (vcs.value.type == ExternalVersionControl_Type.GITHUB) {
            description =
              "Please make sure Client secret matches the one from your GitHub.com Application.";
          } else if (vcs.value.type == ExternalVersionControl_Type.BITBUCKET) {
            description =
              "Please make sure Secret matches the one from your Bitbucket.org consumer.";
          }
          pushNotification({
            module: "bytebase",
            style: "CRITICAL",
            title: `Failed to update '${vcs.value.title}'`,
            description: description,
          });
        }
      };
    }
  } else if (state.title != vcs.value.title) {
    const vcsPatch: Partial<ExternalVersionControl> = {
      name: vcs.value.name,
      title: state.title,
    };
    vcsV1Store
      .updateVCS(vcsPatch)
      .then((updatedVCS: ExternalVersionControl | undefined) => {
        if (!updatedVCS) {
          return;
        }
        pushNotification({
          module: "bytebase",
          style: "SUCCESS",
          title: `Successfully updated '${updatedVCS.title}'`,
        });
      });
  }
};

const cancel = () => {
  router.push({
    name: SETTING_ROUTE_WORKSPACE_GITOPS,
  });
};

const deleteVCS = () => {
  if (!vcs.value) {
    return;
  }
  const title = vcs.value.title;
  vcsV1Store.deleteVCS(vcs.value.name).then(() => {
    pushNotification({
      module: "bytebase",
      style: "SUCCESS",
      title: `Successfully deleted '${title}'`,
    });
    router.push({
      name: SETTING_ROUTE_WORKSPACE_GITOPS,
    });
  });
};
</script>
