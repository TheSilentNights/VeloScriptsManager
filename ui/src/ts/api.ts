import type {Environment, ExecutionInfo, Script} from "../types/models";

const API_BASE = "http://localhost:8080";

async function request<T>(path: string, init?: RequestInit): Promise<T> {
    const res = await fetch(`${API_BASE}${path}`, {
        headers: {"Content-Type": "application/json"},
        ...init,
    });

    const text = await res.text();
    const body = text ? JSON.parse(text) : {};

    if (body.code !== undefined && body.code !== 200) {
        throw new Error(body.message || `request failed with ${res.status}`);
    }

    return body.data as T;
}

export function fetchScripts(): Promise<Script[]> {
    return request<Script[]>("/api/v1/getStoredScripts");
}

export interface ScriptPayload {
    name: string
    workDir: string
    runner: string
    params: string[]
    environments: string[]
}

export function addScript(payload: ScriptPayload): Promise<unknown> {
    return request("/api/v1/addScript", {
        method: "POST",
        body: JSON.stringify(payload),
    });
}

export function updateScript(id: string, payload: ScriptPayload): Promise<unknown> {
    return request("/api/v1/updateScript", {
        method: "POST",
        body: JSON.stringify({id, ...payload}),
    });
}

export function deleteScript(id: string): Promise<unknown> {
    return request("/api/v1/deleteScript", {
        method: "POST",
        body: JSON.stringify({id}),
    });
}

export function executeScript(
    id: string,
    params: string[],
    environments: string[],
): Promise<ExecutionInfo> {
    return request<ExecutionInfo>("/api/v1/executeScript", {
        method: "POST",
        body: JSON.stringify({id, params, environmentsid: environments}),
    });
}

export function fetchEnvironments(): Promise<Environment[]> {
    return request<Environment[]>("/api/v1/getEnvironments");
}
