// Homescript data types and utilities
export interface homescript {
    owner: string
    data: homescriptData
}

export interface homescriptData {
    id: string
    name: string
    description: string
    mdIcon: string
    code: string
    quickActionsEnabled: boolean
    schedulerEnabled: boolean
}

export interface location {
    filename: string
    line: number
    column: number
    index: number
}

export interface homescriptError {
    errorType: string
    location: location
    message: string
}

export interface homescriptResponse {
    success: boolean
    exitcode: number
    message: string
    output: string
    errors: homescriptError[]
}
