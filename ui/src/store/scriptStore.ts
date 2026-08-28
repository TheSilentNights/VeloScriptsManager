import {create} from "zustand";
import type {Script} from "../types/models";
import {
    addScript as apiAddScript,
    deleteScript as apiDeleteScript,
    fetchScripts,
    updateScript as apiUpdateScript,
    type ScriptPayload,
} from "../ts/api";

interface ScriptState {
    scripts: Script[]
    loading: boolean
    error: string | null
    load: () => Promise<void>
    add: (payload: ScriptPayload) => Promise<void>
    update: (id: string, payload: ScriptPayload) => Promise<void>
    remove: (id: string) => Promise<void>
}

export const useScriptStore = create<ScriptState>((set, get) => ({
    scripts: [],
    loading: false,
    error: null,

    async load() {
        set({loading: true, error: null});
        try {
            const scripts = await fetchScripts();
            set({scripts, loading: false});
        } catch (e) {
            set({loading: false, error: (e as Error).message});
        }
    },

    async add(payload) {
        await apiAddScript(payload);
        await get().load();
    },

    async update(id, payload) {
        await apiUpdateScript(id, payload);
        await get().load();
    },

    async remove(id) {
        await apiDeleteScript(id);
        set({scripts: get().scripts.filter((s) => s.id !== id)});
    },
}));
