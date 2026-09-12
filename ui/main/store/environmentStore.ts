import {create} from "zustand";
import type {Environment} from "../types/models";
import {
    addEnvironment as apiAddEnvironment,
    deleteEnvironment as apiDeleteEnvironment,
    fetchEnvironments,
    updateEnvironment as apiUpdateEnvironment,
    type EnvironmentPayload,
} from "../ts/api";

interface EnvironmentState {
    environments: Environment[]
    loading: boolean
    error: string | null
    load: () => Promise<void>
    add: (payload: EnvironmentPayload) => Promise<void>
    update: (id: string, payload: EnvironmentPayload) => Promise<void>
    remove: (id: string) => Promise<void>
    nameOf: (id: string) => string
}

export const useEnvironmentStore = create<EnvironmentState>((set, get) => ({
    environments: [],
    loading: false,
    error: null,

    async load() {
        set({loading: true, error: null});
        try {
            const environments = await fetchEnvironments();
            set({environments, loading: false});
        } catch (e) {
            set({loading: false, error: (e as Error).message});
        }
    },

    async add(payload) {
        await apiAddEnvironment(payload);
        await get().load();
    },

    async update(id, payload) {
        await apiUpdateEnvironment(id, payload);
        await get().load();
    },

    async remove(id) {
        await apiDeleteEnvironment(id);
        set({environments: get().environments.filter((e) => e.id !== id)});
    },

    nameOf(id) {
        return get().environments.find((e) => e.id === id)?.name ?? id;
    },
}));
