import {create} from "zustand";
import type {ExecutionInfo} from "../types/models";
import {
    deleteExecution as apiDeleteExecution,
    getExecutions,
} from "../ts/api";

interface ExecutionState {
    executions: ExecutionInfo[]
    loading: boolean
    error: string | null
    load: () => Promise<void>
    remove: (id: string) => Promise<void>
}

export const useExecutionStore = create<ExecutionState>((set, get) => ({
    executions: [],
    loading: false,
    error: null,

    async load() {
        set({loading: true, error: null});
        try {
            const executions = await getExecutions();
            set({executions, loading: false});
        } catch (e) {
            set({loading: false, error: (e as Error).message});
        }
    },

    async remove(id) {
        await apiDeleteExecution(id);
        set({executions: get().executions.filter((e) => e.executionId !== id)});
    },
}));

// runningCount returns how many executions of the given script are still
// running (status === "running"). Derived from the cached executions list.
export function selectRunningCount(
    executions: ExecutionInfo[],
    scriptId: string,
): number {
    let count = 0;
    for (const e of executions) {
        if (e.scriptId === scriptId && e.status === "running") {
            count++;
        }
    }
    return count;
}
