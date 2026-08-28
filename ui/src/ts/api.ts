import axios, {type AxiosRequestConfig} from "axios";
import type {Environment, ExecutionInfo, Script} from "../types/models";

const API_BASE = "http://127.0.0.1:8080";

const http = axios.create({
    baseURL: API_BASE,
    headers: {"Content-Type": "application/json"},
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
    runner: string
    params: string[]
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
    params: string[],
    environments: string[],
): Promise<ExecutionInfo> {
    return sendPost<ExecutionInfo>("/api/v1/executeScript", {
        id,
        params,
        environmentsid: environments,
    });
}

export function fetchEnvironments(): Promise<Environment[]> {
    return sendGet<Environment[]>("/api/v1/getEnvironments",{method: "GET"});
}

export function getExecutions(): Promise<ExecutionInfo[]> {
    return sendGet<ExecutionInfo[]>("/api/v1/getExecutions");
}
