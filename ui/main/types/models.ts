export interface EnvVar {
    key: string
    value: string
}

export interface Script {
    id: string
    name: string
    workDir: string
    command: string[]
    environments: string[]
}

export interface Environment {
    id: string
    name: string
    paths: string[]
    env: EnvVar[]
}

export interface ExecutionInfo {
    executionId: string
    scriptId: string
    name: string
    status: string // running | finished | failed
    command: string[]
    environments: string[]
    startedAt?: string
    exitCode?: number
    error?: string
}

export interface FileChangeEventDetail {
    ID: string
    Path: string
}

export interface TimeEventDetail {
    ID: string
    Interval: number
    Repeat: boolean
}

export interface FileChangeEventInfo {
    eventId: string
    type: "fileChangeEvent"
    event: FileChangeEventDetail
}

export interface TimeEventInfo {
    eventId: string
    type: "timeEvent"
    event: TimeEventDetail
}

export type EventInfo = FileChangeEventInfo | TimeEventInfo

export interface ApiEnvelope<T> {
    code: number
    message: string
    data: T
}
