import {create} from "zustand";
import type {EventInfo} from "../types/models";
import {
    fetchEvents,
    registerFileChangeEvent as apiRegisterFileChangeEvent,
    registerTimeEvent as apiRegisterTimeEvent,

} from "../ts/api";
import type {FileChangeEventPayload, TimeEventPayload} from "../ts/models";



interface EventState {
    events: EventInfo[]
    loading: boolean
    error: string | null
    load: () => Promise<void>
    registerFileChange: (payload: FileChangeEventPayload) => Promise<void>
    registerTime: (payload: TimeEventPayload) => Promise<void>
}

export const useEventStore = create<EventState>((set, get) => ({
    events: [],
    loading: false,
    error: null,

    async load() {
        set({loading: true, error: null});
        try {
            const events = await fetchEvents();
            set({events, loading: false});
        } catch (e) {
            set({loading: false, error: (e as Error).message});
        }
    },

    async registerFileChange(payload) {
        await apiRegisterFileChangeEvent(payload);
        await get().load();
    },

    async registerTime(payload) {
        await apiRegisterTimeEvent(payload);
        await get().load();
    },
}));
