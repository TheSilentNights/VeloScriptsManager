export interface EnvVar {
    key: string
    value: string
}

export interface Script {
    id: string
    name: string
    workDir: string
    runner: string
    params: string[]
    environments: string[]
}

export interface Environment {
    id: string
    name: string
    type: string
    path: string
    env: EnvVar[]
    children: string[]
}

export interface ExecutionInfo {
    executionId: string
    scriptId: string
    name: string
}

export interface ApiEnvelope<T> {
    code: number
    message: string
    data: T
}
