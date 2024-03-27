import type { homescript, homescriptWithArgs } from "../../homescript"

export interface EditorHms {
    data: homescriptWithArgs,
    unsaved: boolean,
    errors: boolean,
}
