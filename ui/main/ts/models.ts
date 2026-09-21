import type {EnvVar} from "../types/models";

export interface ScriptPayload {
    name: string
    workDir: string
    command: string[]
    environmentsid: string[]
}

export interface ExecuteScriptPayload {
    id: string
    command?: string[]
    environmentsid?: string[]
}

export interface ExecuteScriptsPayload {
    scripts: ExecuteScriptPayload[]
}

export interface EnvironmentPayload {
    name: string
    paths: string[]
    env: EnvVar[]
}

export interface FileChangeEventPayload {
    path: string
    scripts: ExecuteScriptPayload[]
}

export interface TimeEventPayload {
    interval: number
    repeat: boolean
    scripts: ExecuteScriptPayload[]
}

export interface ShortcutScriptPayload {
    id: string
    command: string[]
    environmentsid: string[]
}

export interface ShortcutSlotPayload {
    key: string
    scripts: ShortcutScriptPayload[]
}

export interface ConfigPayload {
    font_size: number
    shortcuts: ShortcutSlotPayload[] | null
}
