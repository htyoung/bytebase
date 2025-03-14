<template>
  <NDropdown
    v-if="allowInputValue"
    ref="dropdownRef"
    trigger="click"
    placement="bottom-start"
    :options="dropdownOptions"
    :width="consistentMenuWidth ? 'trigger' : undefined"
    style="max-height: 20rem; overflow-y: auto; overflow-x: hidden"
    v-bind="dropdownProps"
    @select="$emit('update:value', $event as string)"
  >
    <NInput
      :value="value"
      v-bind="$attrs"
      @update:value="$emit('update:value', $event)"
    >
      <template #suffix>
        <!-- use the same icon and style with NSelect -->
        <NElement
          tag="button"
          style="color: var(--placeholder-color)"
          class="absolute top-1/2 right-[9px] -translate-y-1/2 cursor-pointer"
          :class="suffixClass"
          :style="suffixStyle"
        >
          <svg
            viewBox="0 0 16 16"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
            class="w-4 h-4"
          >
            <path
              d="M3.14645 5.64645C3.34171 5.45118 3.65829 5.45118 3.85355 5.64645L8 9.79289L12.1464 5.64645C12.3417 5.45118 12.6583 5.45118 12.8536 5.64645C13.0488 5.84171 13.0488 6.15829 12.8536 6.35355L8.35355 10.8536C8.15829 11.0488 7.84171 11.0488 7.64645 10.8536L3.14645 6.35355C2.95118 6.15829 2.95118 5.84171 3.14645 5.64645Z"
              fill="currentColor"
            ></path>
          </svg>
        </NElement>
      </template>
    </NInput>
  </NDropdown>
  <NSelect
    v-else
    :value="value"
    :options="options"
    :consistent-menu-width="consistentMenuWidth"
    v-bind="$attrs"
    @update:value="$emit('update:value', $event as string)"
  />
</template>

<script setup lang="ts">
import {
  DropdownOption,
  DropdownProps,
  NDropdown,
  NElement,
  NInput,
  NSelect,
  SelectOption,
} from "naive-ui";
import { computed, h, ref } from "vue";
import { VueClass, VueStyle } from "@/utils";

const props = withDefaults(
  defineProps<{
    value?: string | undefined | null;
    allowInputValue?: boolean;
    allowFilter?: boolean;
    options: SelectOption[];
    dropdownProps?: DropdownProps;
    suffixClass?: VueClass;
    suffixStyle?: VueStyle;
    consistentMenuWidth?: boolean;
  }>(),
  {
    value: undefined,
    allowInputValue: true,
    allowFilter: false,
    dropdownProps: undefined,
    suffixClass: undefined,
    suffixStyle: undefined,
    consistentMenuWidth: true,
  }
);
const emit = defineEmits<{
  (event: "update:value", value: string): void;
}>();

const dropdownRef = ref();

const dropdownOptions = computed(() => {
  return props.options
    .map<DropdownOption>((opt) => ({
      key: opt.value,
      ...opt,
      type: "render",
      render: () =>
        h(
          "div",
          {
            class:
              "px-3 py-2 textinfo text-sm hover:bg-gray-100 cursor-pointer",
          },
          opt.value
        ),
      props: {
        onClick: () => {
          emit("update:value", opt.value as string);
          if (dropdownRef.value) {
            dropdownRef.value.doUpdateShow(false);
          }
        },
      },
    }))
    .filter((opt: DropdownOption) => {
      if (!props.value || !props.allowFilter) {
        return true;
      }
      return (opt.key as string).startsWith(props.value);
    });
});
</script>
