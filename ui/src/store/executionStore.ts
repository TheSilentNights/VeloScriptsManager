import {create} from "zustand";
import type {ExecutionInfo} from "../types/models";
import {getExecutions} from "../ts/api";

const POLL_INTERVAL = 3000;

interface ExecutionState {
    executions: ExecutionInfo[]
    loading: boolean
    error: string | null
    load: () => Promise<void>
    startPolling: () => () => void
}

let timer: ReturnType<typeof setInterval> | null = null;

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

    startPolling() {
        if (timer) {
            return () => {
            };
        }
        void get().load();
        timer = setInterval(() => {
            void get().load();
        }, POLL_INTERVAL);
        return () => {
            if (timer) {
                clearInterval(timer);
                timer = null;
            }
        };
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
