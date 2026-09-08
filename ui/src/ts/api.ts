import axios, {type AxiosRequestConfig} from "axios";
import type {Environment, EnvVar, EventInfo, ExecutionInfo, Script} from "../types/models";

const http = axios.create({
    headers: {"Content-Type": "application/json"},
});

let baseReady: Promise<void> | null = null;

export function initBase(): Promise<void> {
    if (!baseReady) {
        baseReady = (async () => {
            const port = await window.electronAPI.getServerPort();
            if (!port) {
                throw new Error("failed to resolve server port");
            }
            http.defaults.baseURL = `http://127.0.0.1:${port}`;
        })();
    }
    return baseReady;
}

http.interceptors.request.use(async (config) => {
    await initBase();
    return config;
});


interface ApiEnvelope<T> {
    code: number
    message: string
    data: T
}

async function sendGet<T>(path: string, config?:AxiosRequestConfig): Promise<T> {
    const res = await http.get<ApiEnvelope<T>>(path,config);
    const body = res.data;

    console.log(body.message)

    if (body.data === undefined || body.data === null) {
        throw new Error(`unexpected response from ${path}: missing data field`);
    }

    return body.data;
}

async function sendPost<T>(path: string, data? : any, config?: AxiosRequestConfig): Promise<T> {
    const res = await http.post<ApiEnvelope<T>>(path,data,config).catch((err) => {
        console.error(err.response.data.message)
        throw err
    });
    const body = res.data;

    console.log(body.message)

    return body.data;
}

async function sendPut<T>(path: string, data? : any, config?: AxiosRequestConfig): Promise<T> {
    const res = await http.put<ApiEnvelope<T>>(path,data,config).catch((err) => {
        console.error(err.response.data.message)
        throw err
    });
    const body = res.data;

    console.log(body.message)

    return body.data;
}

async function sendDelete<T>(path: string, config?: AxiosRequestConfig): Promise<T> {
    const res = await http.delete<ApiEnvelope<T>>(path,config).catch((err) => {
        console.error(err.response.data.message)
        throw err
    });
    const body = res.data;

    console.log(body.message)

    return body.data;
}



export function fetchScripts(): Promise<Script[]> {
    return sendGet<Script[]>("/api/v1/scripts/");
}

export interface ScriptPayload {
    name: string
    workDir: string
    command: string[]
    environmentsid: string[]
}

export function addScript(payload: ScriptPayload): Promise<unknown> {
    return sendPost("/api/v1/scripts/add", payload);
}

export function updateScript(id: string, payload: ScriptPayload): Promise<unknown> {
    return sendPut("/api/v1/scripts/update", {
        id:id,
        ...payload
    });
}

export function deleteScript(id: string): Promise<unknown> {
    return sendDelete("/api/v1/scripts/delete",{params: {id:id}});
}

export function executeScript(
    id: string,
    command: string[],
    environments: string[],
): Promise<void> {
    return sendPost<void>("/api/v1/execution/execute", {
        id,
        command,
        environmentsid: environments,
    });
}

export function fetchEnvironments(): Promise<Environment[]> {
    return sendGet<Environment[]>("/api/v1/environments/",{method: "GET"});
}

export interface EnvironmentPayload {
    name: string
    paths: string[]
    env: EnvVar[]
}

export function addEnvironment(payload: EnvironmentPayload): Promise<unknown> {
    return sendPost("/api/v1/environments/add", payload);
}

export function updateEnvironment(id: string, payload: EnvironmentPayload): Promise<unknown> {
    return sendPut("/api/v1/environments/update", {
        id,
        ...payload,
    });
}

export function deleteEnvironment(id: string): Promise<unknown> {
    return sendDelete("/api/v1/environments/delete",{params: {id:id}});
}

export function getExecutions(): Promise<ExecutionInfo[]> {
    return sendGet<ExecutionInfo[]>("/api/v1/execution/");
}

export function deleteExecution(id: string): Promise<unknown> {
    return sendPost("/api/v1/execution/kill", {id: id});
}

export function fetchEvents(): Promise<EventInfo[]> {
    return sendGet<EventInfo[]>("/api/v1/event/");
}

export interface ExecuteScriptPayload {
    id: string
    command?: string[]
    environmentsid?: string[]
}

export interface FileChangeEventPayload {
    path: string
    execute_script: ExecuteScriptPayload
}

export interface TimeEventPayload {
    interval: number
    repeat: boolean
    execute_script: ExecuteScriptPayload
}

export function registerFileChangeEvent(payload: FileChangeEventPayload): Promise<void> {
    return sendPost<void>("/api/v1/event/registerFileChangeEvent", payload);
}

export function registerTimeEvent(payload: TimeEventPayload): Promise<void> {
    return sendPost<void>("/api/v1/event/registerTimeEvent", payload);
}

export interface ConfigPayload {
    fontSize: number
}

export function fetchConfig(): Promise<ConfigPayload> {
    return sendGet<ConfigPayload>("/api/v1/getConfig");
}

export function updateConfig(payload: ConfigPayload): Promise<unknown> {
    return sendPost("/api/v1/updateConfig", payload);
}
