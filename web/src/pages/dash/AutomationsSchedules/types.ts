
export interface automation {
    id: number;
    name: string;
    description: string;
    cronExpression: string;
    cronDescription: string;
    homescriptId: string;
    owner: string;
    enabled: boolean;
    timingMode: "normal" | "sunrise" | "sunset";
}

export interface Schedule {
    id: number;
    owner: string;
    data: ScheduleData;
}

export interface ScheduleData {
    name: string;
    hour: number;
    minute: number;
    targetMode: "code" | "hms" | "switches";
    homescriptCode: string;
    homescriptTargetId: string;
    switchJobs: SwitchJob[];
}

export interface SwitchJob {
    switchId: string;
    powerOn: boolean;
}
