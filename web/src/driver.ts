export type DeviceCapability = 'base' | 'power' | 'dimmable' | 'sensor'

export interface ConfigSpecWrapper {
    capabilities: DeviceCapability[]
    info: ConfigSpecWrapperInner
    // NOTE: hms type from backend ignored here, would add unnecessary bloat and is not needed here.
}

export interface ConfigSpecWrapperInner {
    config: ConfigSpec
}

export type ConfigSpec = ConfigSpecAtom | ConfigSpecInner | ConfigSpecStruct | null;

export type ConfigSpecType = 'INT' | 'FLOAT' | 'BOOL' | 'STRING' | 'LIST' | 'STRUCT' | 'OPTION'

export interface ConfigSpecAtom {
    type:   ConfigSpecType;
}

export interface ConfigSpecInner {
    type:   ConfigSpecType;
    inner: ConfigSpec;
}

export interface ConfigSpecStruct {
    type: ConfigSpecType;
    fields: ConfigSpecStructField[];
}

export interface ConfigSpecStructField {
    name: ConfigSpecType;
    type: ConfigSpec;
}

export interface ValidationError {
    level:   number;
    message: string;
    notes:   string[];
    span:    Span;
}

export interface Span {
    start:    Location;
    end:      Location;
    filename: string;
}

export interface Location {
    line:   number;
    column: number;
    index:  number;
}


