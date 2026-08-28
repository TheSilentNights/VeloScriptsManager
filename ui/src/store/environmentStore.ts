import {create} from "zustand";
import type {Environment} from "../types/models";
import {fetchEnvironments} from "../ts/api";

interface EnvironmentState {
    environments: Environment[]
    loading: boolean
    error: string | null
    load: () => Promise<void>
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

    nameOf(id) {
        return get().environments.find((e) => e.id === id)?.name ?? id;
    },
}));
