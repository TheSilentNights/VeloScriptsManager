import axios, {type AxiosRequestConfig} from "axios";
import type {Environment, EnvVar, ExecutionInfo, Script} from "../types/models";

const http = axios.create({
    headers: {"Content-Type": "application/json"},
});

let baseReady: Promise<void> | null = null;

function initBase(): Promise<void> {
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

export function fetchScripts(): Promise<Script[]> {
    return sendGet<Script[]>("/api/v1/getStoredScripts");
}

export interface ScriptPayload {
    name: string
    workDir: string
    command: string[]
    environmentsId: string[]
}

export function addScript(payload: ScriptPayload): Promise<ScriptPayload> {
    return sendPost<ScriptPayload>("/api/v1/addScript", payload);
}

export function updateScript(id: string, payload: ScriptPayload): Promise<ScriptPayload> {
    return sendPost("/api/v1/updateScript", {
        id:id,
        ...payload
    });
}

export function deleteScript(id: string): Promise<unknown> {
    return sendPost("/api/v1/deleteScript",{id:id});
}

export function executeScript(
    id: string,
    command: string[],
    environments: string[],
): Promise<ExecutionInfo> {
    return sendPost<ExecutionInfo>("/api/v1/executeScript", {
        id,
        command,
        environmentsid: environments,
    });
}

export function fetchEnvironments(): Promise<Environment[]> {
    return sendGet<Environment[]>("/api/v1/getEnvironments",{method: "GET"});
}

export interface EnvironmentPayload {
    name: string
    paths: string[]
    env: EnvVar[]
}

export function addEnvironment(payload: EnvironmentPayload): Promise<unknown> {
    return sendPost("/api/v1/addEnvironment", payload);
}

export function updateEnvironment(id: string, payload: EnvironmentPayload): Promise<unknown> {
    return sendPost("/api/v1/updateEnvironment", {
        id,
        ...payload,
    });
}

export function deleteEnvironment(id: string): Promise<unknown> {
    return sendPost("/api/v1/deleteEnvironment", {id: id});
}

export function getExecutions(): Promise<ExecutionInfo[]> {
    return sendGet<ExecutionInfo[]>("/api/v1/getExecutions");
}

export function deleteExecution(id: string): Promise<unknown> {
    return sendPost("/api/v1/deleteExecution", {id: id});
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
