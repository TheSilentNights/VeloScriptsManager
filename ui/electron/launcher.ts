import { app } from "electron";
import {ChildProcess, spawn} from "node:child_process";
import net from "node:net";
import path from 'node:path';

let serverProcess: ChildProcess | null = null
let serverPort: number | null = null

function checkPortAvailable(port: number): Promise<boolean> {
    return new Promise((resolve) => {
        const server = net.createServer()
        server.once('error', () => resolve(false))
        server.once('listening', () => {
            server.close(() => resolve(true))
        })
        server.listen(port, '127.0.0.1')
    })
}

export function getServerPort(): number {
    return serverPort ? serverPort: 19278
}

async function findAvailablePort(startPort: number): Promise<number> {
    for (let port = startPort; port < startPort + 100; port++) {
        if (await checkPortAvailable(port)) {
            return port
        }
    }
    throw new Error(`no available port in range ${startPort}-${startPort + 99}`)
}

function resolveServerCommand(): { command: string; args: string[]; cwd?: string } {
    return {
        command: path.join(process.resourcesPath, 'server.exe'),
        args: ['--port', String(serverPort), '--release'],
        cwd: app.getPath('userData'),
    }
}

export async function waitForServer(port: number, timeoutMs = 15000): Promise<void> {
    const deadline = Date.now() + timeoutMs
    while (Date.now() < deadline) {
        if (serverProcess?.exitCode !== null && serverProcess?.exitCode !== undefined) {
            throw new Error(`server exited with code ${serverProcess.exitCode}`)
        }
        const ok = await new Promise<boolean>((resolve) => {
            const socket = net.connect(port, '127.0.0.1')
            socket.once('connect', () => {
                socket.destroy()
                resolve(true)
            })
            socket.once('error', () => resolve(false))
        })
        if (ok) return
        await new Promise((resolve) => setTimeout(resolve, 200))
    }
    throw new Error('server startup timeout')
}

export async function startServer(): Promise<void> {
    serverPort = await findAvailablePort(19278)
    const { command, args, cwd } = resolveServerCommand()

    serverProcess = spawn(command, args, {
        cwd,
        stdio: 'inherit',
        windowsHide: true,
    })

    serverProcess.once('exit', (code) => {
        if (code !== 0 && code !== null) {
            console.error(`backend server exited with code ${code}`)
        }
    })

    await waitForServer(serverPort)
}

export function stopServer(): void {
    if (!serverProcess || serverProcess.killed) return
    serverProcess.kill()
    serverProcess = null
}