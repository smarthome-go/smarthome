/*
 * This file contains data types used in a Homescript context
 * Only types need to be imported from this file
* /

/* Homescript data type and container */
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
}

/* Homescript run request response */
// Is returned as a response to a Homescript run request
export interface homescriptResponse {
    success: boolean
    exitcode: number
    message: string
    output: string
    errors: homescriptError[]
}

export interface homescriptError {
    errorType: string
    location: location
    message: string
}

export interface location {
    filename: string
    line: number
    column: number
    index: number
}

/* Homescript arguments */
// Container for homescript argument
export interface homecriptArg {
    id: number
    data: homescriptArgData
}

// Main data of a Homescript argument
export interface homescriptArgData {
    argKey: string
    homescriptId: string
    prompt: string
    inputType: "string" | "number" | "boolean"
    display: "type_default" | "string_switches" | "boolean_yes_no" | "boolean_on_off" | "number_hour" | "number_minute"
}
