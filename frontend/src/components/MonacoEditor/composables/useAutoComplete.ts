import type monaco from "monaco-editor";
import { Ref, computed, watch } from "vue";
import { UNKNOWN_ID } from "@/types";
import {
  extractDatabaseResourceName,
  extractInstanceResourceName,
} from "@/utils";
import type { MonacoModule } from "../types";

export type AutoCompleteContext = {
  instance: string; // instances/{instance}
  database?: string; // instances/{instance}/databases/{database_name}
};

type SetMetadataParams = {
  instanceId: string; // instances/{instance}
  databaseName: string;
};

export const useAutoComplete = async (
  monaco: MonacoModule,
  editor: monaco.editor.IStandaloneCodeEditor,
  context: Ref<AutoCompleteContext | undefined>
) => {
  const { executeCommand, useLSPClient } = await import("../lsp-client");

  const client = useLSPClient();
  const params = computed(() => {
    const p: SetMetadataParams = {
      instanceId: "",
      databaseName: "",
    };
    const ctx = context.value;
    if (ctx) {
      const instance = extractInstanceResourceName(ctx.instance);
      if (instance && instance !== String(UNKNOWN_ID)) {
        p.instanceId = ctx.instance;
      }
      const databaseName = extractDatabaseResourceName(
        ctx.database ?? ""
      ).database;
      if (databaseName && databaseName !== String(UNKNOWN_ID)) {
        p.databaseName = databaseName;
      }
    }
    return p;
  });
  watch(
    () => JSON.stringify(params.value),
    async () => {
      const result = executeCommand(client, "setMetadata", [params.value]);
      console.debug(
        `setMetadata(${JSON.stringify(params.value)}): ${JSON.stringify(
          result
        )}`
      );
    },
    { immediate: true }
  );
};
