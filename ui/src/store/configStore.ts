import {create} from "zustand";
import {fetchConfig, updateConfig as apiUpdateConfig, type ConfigPayload} from "../ts/api";

interface ConfigState {
    font_size: number | null
    loading: boolean
    error: string | null
    load: () => Promise<void>
    update: (payload: ConfigPayload) => Promise<void>
}

export const useConfigStore = create<ConfigState>((set) => ({
    font_size: null,
    loading: false,
    error: null,

    async load() {
        set({loading: true, error: null});
        try {
            const config = await fetchConfig();
            set({font_size: config.font_size, loading: false});
        } catch (e) {
            set({loading: false, error: (e as Error).message});
        }
    },

    async update(payload) {
        await apiUpdateConfig(payload);
        set({font_size: payload.font_size});
    },
}));
