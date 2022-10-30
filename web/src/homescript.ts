/*
 * This file contains data types used in a Homescript context
 * Only types need to be imported from this file
* /

/* Homescript data type and container */

import type { GenericResponse } from './global'

export interface homescriptJob {
    id: number
    initiator: string
    homescriptId: string
}

// A Homescript with its arguments
export interface homescriptWithArgs {
    data: homescript
    arguments: homescriptArg[]
}

// Homescript container
export interface homescript {
    owner: string
    data: homescriptData
}

// Includes the main data of a Homescript
export interface homescriptData {
    id: string
    name: string
    description: string
    mdIcon: string
    code: string
    quickActionsEnabled: boolean
    schedulerEnabled: boolean
    workspace: string
}

/* Homescript run request response */
// Is returned as a response to a Homescript run request
export interface homescriptResponseWrapper {
    response: homescriptResponse
    code: string
    modeRun: boolean
}

export interface homescriptResponse {
    success: boolean
    id: string
    exitCode: number
    message: string
    output: string
    errors: homescriptError[]
}

export interface homescriptError {
    kind: string
    message: string
    span: span
}

export interface span {
    start: location
    end: location
}

export interface location {
    line: number
    column: number
    index: number
}

/* Homescript arguments */
// Is used when requesting the execution of a Homescript
export interface homescriptArgSubmit {
    key: string
    value: string
}

// Container for homescript arguments
export interface homescriptArg {
    id: number
    data: homescriptArgData
}

// Main data of a Homescript argument
export interface homescriptArgData {
    argKey: string
    homescriptId: string
    prompt: string
    mdIcon: string
    inputType: 'string' | 'number' | 'boolean'
    display: 'type_default' | 'string_switches' | 'boolean_yes_no' | 'boolean_on_off' | 'number_hour' | 'number_minute'
}

// Is used for visual applications which require labels and a logical connection between type and display
export interface DisplayOpt {
    identifier:
        | 'type_default'
        | 'string_switches'
        | 'boolean_yes_no'
        | 'boolean_on_off'
        | 'number_hour'
        | 'number_minute'
    label: string
    type: 'string' | 'number' | 'boolean'
}

// Used for displaying the options for `inputType` and `display`
export const inputTypeOpts = ['string', 'number', 'boolean']
export const displayOpts: DisplayOpt[] = [
    // Default display
    { identifier: 'type_default', label: 'Type default', type: 'string' },
    { identifier: 'type_default', label: 'Type default', type: 'number' },
    { identifier: 'type_default', label: 'Type default', type: 'boolean' },
    // Switch listing as string
    {
        identifier: 'string_switches',
        label: 'Select switch',
        type: 'string',
    },
    // Yes / No prompt as boolean
    {
        identifier: 'boolean_yes_no',
        label: 'Yes / No (bool)',
        type: 'boolean',
    },
    // On / Off prompt as boolean
    {
        identifier: 'boolean_on_off',
        label: 'On / Off (bool)',
        type: 'boolean',
    },
    // Time prompts as either hour or minute
    { identifier: 'number_hour', label: 'Hour', type: 'number' },
    { identifier: 'number_minute', label: 'Minute', type: 'number' },
]

// Sends an execution request to the server
// Returns the Homescript Response
// Can throw an error if non-Homescript errors occur
export async function runHomescriptById(id: string, args: homescriptArgSubmit[]): Promise<homescriptResponse> {
    const res = await fetch(`/api/homescript/run`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id, args: args }),
    })
    if (res.status !== 200 && res.status !== 500) throw Error(await (res.json()))
    return await (res.json())
}

// Sends an execution request to the server
// Returns the Homescript Response
// Can throw an error if non-Homescript errors occur
export async function runHomescriptCode(code: string, args: homescriptArgSubmit[]): Promise<homescriptResponse> {
    const res = await fetch(`/api/homescript/run/live`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ code, args: args }),
    })
    if (res.status !== 200 && res.status !== 500) throw Error(await (res.json()))
    return await (res.json())
}

// Sends a lint request to the server
// Returns the Homescript Response
// Can throw an error if non-Homescript errors occur
export async function lintHomescriptById(id: string, args: homescriptArgSubmit[]): Promise<homescriptResponse> {
    const res = await fetch(`/api/homescript/lint`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id, args: args }),
    })
    if (res.status !== 200 && res.status !== 500) throw Error(await (res.json()))
    return await (res.json())
}

// Sends a lint request to the server
// Returns the Homescript Response
// Can throw an error if non-Homescript errors occur
export async function lintHomescriptCode(code: string, args: homescriptArgSubmit[]): Promise<homescriptResponse> {
    const res = await fetch(`/api/homescript/lint/live`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ code, args: args }),
    })
    if (res.status !== 200 && res.status !== 500) throw Error(await (res.json()))
    return await (res.json())
}

// Returns all currently active Homescript jobs
export async function getRunningJobs(): Promise<homescriptJob[]> {
    const res = await (await fetch('/api/homescript/jobs')).json()
    if (res.success != undefined && !res.success) throw Error(res.error)
    return res
}

// Sends a request to kill all running executions of a given script (by id)
export async function killAllJobsById(id: string): Promise<GenericResponse> {
    const res = await fetch(`/api/homescript/kill/script/${encodeURIComponent(id)}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
    })
    if (res.status !== 200 && res.status !== 500) throw Error(await (res.json()))
    return await (res.json())
}
