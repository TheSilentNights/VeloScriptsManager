import {create} from "zustand";
import {
    fetchConfig,
    updateConfig as apiUpdateConfig,
    type ConfigPayload,
    type ShortcutSlotPayload,
} from "../ts/api";

const slotCount = 10;

function normalizeShortcuts(shortcuts: ShortcutSlotPayload[] | null | undefined): ShortcutSlotPayload[] {
    const normalized: ShortcutSlotPayload[] = Array.from({length: slotCount}, () => ({
        key: "",
        scripts: [],
    }));
    if (!shortcuts) {
        return normalized;
    }
    for (let i = 0; i < shortcuts.length && i < slotCount; i++) {
        const slot = shortcuts[i];
        normalized[i] = {
            key: slot.key ?? "",
            scripts: (slot.scripts ?? []).map((script) => ({
                id: script.id ?? "",
                command: script.command ?? [],
                environmentsid: script.environmentsid ?? [],
            })),
        };
    }
    return normalized;
}

interface ConfigState {
    font_size: number | null
    shortcuts: ShortcutSlotPayload[]
    loading: boolean
    error: string | null
    load: () => Promise<void>
    update: (payload: ConfigPayload) => Promise<void>
}

export const useConfigStore = create<ConfigState>((set) => ({
    font_size: null,
    shortcuts: normalizeShortcuts(null),
    loading: false,
    error: null,

    async load() {
        set({loading: true, error: null});
        try {
            const config = await fetchConfig();
            set({
                font_size: config.font_size,
                shortcuts: normalizeShortcuts(config.shortcuts),
                loading: false,
            });
        } catch (e) {
            set({loading: false, error: (e as Error).message});
        }
    },

    async update(payload) {
        await apiUpdateConfig(payload);
        set({
            font_size: payload.font_size,
            shortcuts: normalizeShortcuts(payload.shortcuts),
        });
    },
}));
