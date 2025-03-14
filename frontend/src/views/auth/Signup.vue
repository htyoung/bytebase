<template>
  <div class="mx-auto w-full max-w-sm">
    <div>
      <BytebaseLogo component="span" class="mx-auto mb-8" />

      <h2 class="text-2xl leading-9 font-medium text-main">
        <template v-if="needAdminSetup">
          <i18n-t
            keypath="auth.sign-up.admin-title"
            tag="p"
            class="text-accent font-semnibold text-center"
          >
            <template #account>
              <span class="text-accent font-semnibold">{{
                $t("auth.sign-up.admin-account")
              }}</span>
            </template>
          </i18n-t>
        </template>
        <template v-else> {{ $t("auth.sign-up.title") }}</template>
      </h2>
    </div>

    <div class="mt-8">
      <div class="mt-6">
        <form class="space-y-6" @submit.prevent="trySignup">
          <div>
            <label
              for="email"
              class="block text-sm font-medium leading-5 text-control"
            >
              {{ $t("common.email") }} <span class="text-red-600">*</span>
            </label>
            <div class="mt-1 rounded-md shadow-sm">
              <BBTextField
                v-model:value="state.email"
                required
                placeholder="jim@example.com"
                :input-props="{ id: 'email' }"
                @input="onTextEmail"
              />
            </div>
          </div>

          <div>
            <label
              for="password"
              class="block text-sm font-medium leading-5 text-control"
            >
              {{ $t("common.password") }}
              <span class="text-red-600">*</span>
            </label>
            <div
              class="relative flex flex-row items-center mt-1 rounded-md shadow-sm"
            >
              <BBTextField
                v-model:value="state.password"
                :type="state.showPassword ? 'text' : 'password'"
                :input-props="{ id: 'password', autocomplete: 'off' }"
                required
                @input="refreshPasswordValidation"
              />
              <div
                class="hover:cursor-pointer absolute right-3"
                @click="
                  () => {
                    state.showPassword = !state.showPassword;
                  }
                "
              >
                <heroicons-outline:eye
                  v-if="state.showPassword"
                  class="w-4 h-4"
                />
                <heroicons-outline:eye-slash v-else class="w-4 h-4" />
              </div>
            </div>
          </div>

          <div>
            <label
              for="password-confirm"
              class="block text-sm font-medium leading-5 text-control"
            >
              {{ $t("auth.sign-up.confirm-password") }}
              <span class="text-red-600">
                *
                {{
                  state.showPasswordMismatchError
                    ? $t("auth.sign-up.password-mismatch")
                    : ""
                }}
              </span>
            </label>
            <div
              class="relative flex flex-row items-center mt-1 rounded-md shadow-sm"
            >
              <BBTextField
                v-model:value="state.passwordConfirm"
                :type="state.showPassword ? 'text' : 'password'"
                :input-props="{ id: 'password-confirm', autocomplete: 'off' }"
                :placeholder="$t('auth.sign-up.confirm-password-placeholder')"
                required
                @input="refreshPasswordValidation"
              />
              <div
                class="hover:cursor-pointer absolute right-3"
                @click="
                  () => {
                    state.showPassword = !state.showPassword;
                  }
                "
              >
                <heroicons-outline:eye
                  v-if="state.showPassword"
                  class="w-4 h-4"
                />
                <heroicons-outline:eye-slash v-else class="w-4 h-4" />
              </div>
            </div>
          </div>

          <div>
            <label
              for="email"
              class="block text-sm font-medium leading-5 text-control"
            >
              {{ $t("common.username") }}
              <span class="text-red-600"> * </span>
            </label>
            <div class="mt-1 rounded-md shadow-sm">
              <BBTextField
                id="name"
                v-model:value="state.name"
                required
                placeholder="Jim Gray"
                @input="onTextName"
              />
            </div>
          </div>

          <div
            v-if="needAdminSetup"
            class="w-full flex flex-row justify-start items-start"
          >
            <NCheckbox
              v-model:checked="state.acceptTermsAndPolicy"
              class="mt-0.5"
            />
            <i18n-t
              tag="span"
              keypath="auth.sign-up.accept-terms-and-policy"
              class="ml-1 select-none"
              @click="
                () => (state.acceptTermsAndPolicy = !state.acceptTermsAndPolicy)
              "
            >
              <template #terms>
                <a
                  href="https://www.bytebase.com/terms?source=console"
                  class="text-accent"
                  >{{ $t("auth.sign-up.terms-of-service") }}</a
                >
              </template>
              <template #policy>
                <a
                  href="https://www.bytebase.com/privacy?source=console"
                  class="text-accent"
                  >{{ $t("auth.sign-up.privacy-policy") }}</a
                >
              </template>
            </i18n-t>
          </div>

          <div class="w-full">
            <NButton
              attr-type="submit"
              type="primary"
              size="large"
              :disabled="!allowSignup"
              :loading="state.isLoading"
              style="width: 100%"
            >
              {{
                needAdminSetup
                  ? $t("auth.sign-up.create-admin-account")
                  : $t("common.sign-up")
              }}
            </NButton>
          </div>
        </form>
      </div>
    </div>

    <div v-if="!needAdminSetup" class="mt-6 relative">
      <div class="absolute inset-0 flex items-center" aria-hidden="true">
        <div class="w-full border-t border-control-border"></div>
      </div>
      <div class="relative flex justify-center text-sm">
        <span class="pl-2 bg-white text-control">
          {{ $t("auth.sign-up.existing-user") }}
        </span>
        <router-link to="/auth/signin" class="accent-link px-2 bg-white">
          {{ $t("common.sign-in") }}
        </router-link>
      </div>
    </div>
  </div>

  <AuthFooter />
</template>

<script lang="ts" setup>
import { NCheckbox } from "naive-ui";
import { storeToRefs } from "pinia";
import { computed, onMounted, onUnmounted, reactive } from "vue";
import { useRouter } from "vue-router";
import {
  useActuatorV1Store,
  useAuthStore,
  useOnboardingStateStore,
} from "@/store";
import { SignupInfo, TEXT_VALIDATION_DELAY } from "@/types";
import { isValidEmail } from "@/utils";
import AuthFooter from "./AuthFooter.vue";

interface LocalState {
  email: string;
  password: string;
  passwordConfirm: string;
  passwordValidationTimer?: ReturnType<typeof setTimeout>;
  showPasswordMismatchError: boolean;
  name: string;
  nameManuallyEdited: boolean;
  acceptTermsAndPolicy: boolean;
  showPassword: boolean;
  isLoading: boolean;
}

const actuatorStore = useActuatorV1Store();
const router = useRouter();

const state = reactive<LocalState>({
  email: "",
  password: "",
  passwordConfirm: "",
  showPasswordMismatchError: false,
  name: "",
  nameManuallyEdited: false,
  acceptTermsAndPolicy: true,
  showPassword: false,
  isLoading: false,
});

onUnmounted(() => {
  if (state.passwordValidationTimer) {
    clearInterval(state.passwordValidationTimer);
  }
});

const { needAdminSetup, disallowSignup } = storeToRefs(actuatorStore);

const allowSignup = computed(() => {
  return (
    isValidEmail(state.email) &&
    state.password &&
    !state.showPasswordMismatchError &&
    state.acceptTermsAndPolicy &&
    !disallowSignup.value
  );
});

const passwordMatch = computed(() => {
  return state.password == state.passwordConfirm;
});

onMounted(() => {
  if (needAdminSetup.value) {
    state.acceptTermsAndPolicy = false;
  }
});

const refreshPasswordValidation = () => {
  if (state.passwordValidationTimer) {
    clearInterval(state.passwordValidationTimer);
  }

  if (passwordMatch.value) {
    state.showPasswordMismatchError = false;
  } else {
    state.passwordValidationTimer = setTimeout(() => {
      // If error is already displayed, we hide the error only if there is valid input.
      // Otherwise, we hide the error if input is either empty or valid.
      if (state.showPasswordMismatchError) {
        state.showPasswordMismatchError = !passwordMatch.value;
      } else {
        state.showPasswordMismatchError =
          state.password != "" &&
          state.passwordConfirm != "" &&
          !passwordMatch.value;
      }
    }, TEXT_VALIDATION_DELAY);
  }
};

const onTextEmail = () => {
  const email = state.email.trim().toLowerCase();
  state.email = email;
  if (!state.nameManuallyEdited) {
    const emailParts = email.split("@");
    if (emailParts.length > 0) {
      if (emailParts[0].length > 0) {
        const name = emailParts[0].replace("_", ".");
        const nameParts = name.split(".");
        if (nameParts.length >= 2) {
          state.name = [
            nameParts[0].charAt(0).toUpperCase() + nameParts[0].slice(1),
            nameParts[1].charAt(0).toUpperCase() + nameParts[1].slice(1),
          ].join(" ");
        } else {
          state.name = name.charAt(0).toUpperCase() + name.slice(1);
        }
      }
    }
  }
};

const onTextName = () => {
  const name = state.name.trim();
  state.nameManuallyEdited = name.length > 0;
};

const trySignup = async () => {
  if (state.isLoading) return;

  if (!passwordMatch.value) {
    state.showPasswordMismatchError = true;
  } else {
    state.isLoading = true;

    try {
      const signupInfo: SignupInfo = {
        email: state.email,
        password: state.password,
        name: state.name,
      };
      await useAuthStore().signup(signupInfo);
      if (needAdminSetup.value) {
        await actuatorStore.fetchServerInfo();
        // When the first time we created an end user, the server-side will
        // generate onboarding data.
        // We write a flag here to indicate that the workspace is just created
        // and we can consume this flag somewhere else if needed.
        useOnboardingStateStore().initialize();
      }
      router.replace("/");
    } finally {
      state.isLoading = false;
    }
  }
};
</script>
