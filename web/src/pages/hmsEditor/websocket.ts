import type {  homescriptError } from '../../homescript'

export interface hmsOutMessage {
    kind: 'out' | 'res'
    payload: string
}

export interface hmsResMessage {
    kind: 'out' | 'res'
    exitCode: number
    errors: homescriptError[]
}

export interface hmsResWrapper {
    modeRun: boolean
    fileContents: Map<string, string>
    errors: homescriptError[]
    success: boolean
}
