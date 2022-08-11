<script lang="ts">
    import IconButton from "@smui/icon-button";

    import Page from "../../Page.svelte";
    import Logs from "./Logs.svelte";
    import Progress from "../../components/Progress.svelte";
    import Button, { Icon, Label } from "@smui/button";
    import Textfield from "@smui/textfield";
    import type { config } from "./main";
    import { onMount } from "svelte";
    import { createSnackbar } from "../../global";
    import HelperText from "@smui/textfield/helper-text";
    import GeoHelp from "./dialogs/GeoHelp.svelte";
    import FormField from "@smui/form-field";
    import Switch from "@smui/switch";
    import ExportImport from "./ExportImport.svelte";

    let loading = false;

    let automationEnabledLoading = false;
    let lockDownModeEnabledLoading = false;

    // Specifies whether the log event dialog should be visible or not
    let logsOpen = false;

    // Specifies whether the geolocation help dialog should be open
    let geoHelpOpen = false;

    let config: config = {
        automationEnabled: false,
        lockDownMode: false,
        latitude: 0.0,
        longitude: 0.0,
    };

    let latitudeInput = 0.0;
    let longitudeInput = 0.0;

    async function fetchConfig() {
        loading = true;
        try {
            const res = await (await fetch("/api/system/config")).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            config = res;
            latitudeInput = res.latitude;
            longitudeInput = res.longitude;
        } catch (err) {
            $createSnackbar(`Failed to load system configuration: ${err}`);
        }
        loading = false;
    }

    async function updateGeolocation() {
        loading = true;
        try {
            const res = await (
                await fetch("/api/system/location/modify", {
                    method: "PUT",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        latitude: latitudeInput,
                        longitude: longitudeInput,
                    }),
                })
            ).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            config.latitude = latitudeInput;
            config.longitude = longitudeInput;
        } catch (err) {
            $createSnackbar(`Failed to update geolocation: ${err}`);
        }
        loading = false;
    }

    async function setAutomationsEnabled(enabled: boolean) {
        automationEnabledLoading = true;
        try {
            const res = await (
                await fetch("/api/automation/state/global", {
                    method: "PUT",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ enabled }),
                })
            ).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            config.automationEnabled = enabled;
            setTimeout(() => {
                automationEnabledLoading = false;
            }, 750);
        } catch (err) {
            $createSnackbar(`Failed to update automation system state: ${err}`);
            automationEnabledLoading = false;
        }
    }

    async function setLockDownModeEnabled(enabled: boolean) {
        lockDownModeEnabledLoading = true;
        try {
            const res = await (
                await fetch("/api/system/lockdown/modify", {
                    method: "PUT",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ enabled }),
                })
            ).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            config.lockDownMode = enabled;
            setTimeout(() => {
                lockDownModeEnabledLoading = false;
            }, 750);
        } catch (err) {
            $createSnackbar(`Failed to update lockdown mode: ${err}`);
            lockDownModeEnabledLoading = false;
        }
    }

    // As soon as the component is mounted, fetch the configuration
    onMount(fetchConfig);
</script>

<Logs />

<!-->Log record dialog</-->
<Logs bind:open={logsOpen} />

<!-->Geolocation help dialog</-->
<GeoHelp bind:open={geoHelpOpen} />

<Page>
    <div id="header" class="mdc-elevation--z4">
        <h6>System Configuration</h6>
        <div id="header__buttons">
            <IconButton title="Refresh" class="material-icons"
                >refresh</IconButton
            >
            <Button on:click={() => (logsOpen = true)}>Logs</Button>
        </div>
    </div>
    <Progress id="loader" loading={false} />
    <div id="content">
        <div id="left" class="mdc-elevation--z1">
            <div class="geo">
                <div class="geo__title">
                    <h6>Geolocation</h6>
                    <IconButton
                        class="geo__title__help"
                        on:click={() => (geoHelpOpen = true)}
                        size="button"
                    >
                        <Icon class="material-icons">help</Icon>
                    </IconButton>
                    <Button
                        on:click={updateGeolocation}
                        disabled={(latitudeInput === config.latitude &&
                            longitudeInput === config.longitude) ||
                            latitudeInput < -90 ||
                            latitudeInput > 90 ||
                            longitudeInput < -180 ||
                            longitudeInput > 180}
                    >
                        <Label>Save</Label>
                        <Icon class="material-icons">save</Icon>
                    </Button>
                </div>
                <div class="geo__inputs">
                    <div class="geo__inputs__lat">
                        <Textfield
                            bind:value={latitudeInput}
                            label="Latitude °"
                            type="number"
                            invalid={config.latitude < -90.0 ||
                                config.latitude > 90.0}
                        >
                            <HelperText slot="helper"
                                >Latitude° of your geolocation</HelperText
                            >
                        </Textfield>
                    </div>
                    <div class="geo_inputs__long">
                        <Textfield
                            bind:value={longitudeInput}
                            label="Longitude °"
                            type="number"
                            invalid={config.longitude < -180.0 ||
                                config.longitude > 180}
                        >
                            <HelperText slot="helper"
                                >Longitude° of your geolocation</HelperText
                            >
                        </Textfield>
                    </div>
                </div>
            </div>
            <div class="automation">
                <h6>Automation</h6>
                <FormField>
                    <Switch
                        disabled={automationEnabledLoading}
                        checked={config.automationEnabled}
                        on:SMUISwitch:change={(e) =>
                            setAutomationsEnabled(e.detail.selected)}
                    />
                    <div slot="label" class="automation__label">
                        <span
                            >Automations & Schedules {config.automationEnabled
                                ? "enabled"
                                : "disabled"}</span
                        >
                        <Progress
                            type={"circular"}
                            loading={automationEnabledLoading}
                        />
                    </div>
                </FormField>
            </div>
            <div class="lockdown">
                <h6>Lockdown Mode</h6>
                <FormField>
                    <Switch
                        disabled={lockDownModeEnabledLoading}
                        checked={config.lockDownMode}
                        on:SMUISwitch:change={(e) =>
                            setLockDownModeEnabled(e.detail.selected)}
                    />
                    <div slot="label" class="lockdown__label">
                        <span
                            >Power requests {config.lockDownMode
                                ? "blocked"
                                : "allowed"}</span
                        >
                        <Progress
                            type={"circular"}
                            loading={lockDownModeEnabledLoading}
                        />
                    </div>
                </FormField>
            </div>
            <ExportImport />
        </div>
        <div id="logs" class="mdc-elevation--z1" />
    </div></Page
>

<style lang="scss">
    @use "../../mixins" as *;

    #header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0.1rem 1.3rem;
        box-sizing: border-box;
        background-color: var(--clr-height-1-4);
        min-height: 3.5rem;

        &__buttons {
            display: flex;
            align-items: center;
        }

        h6 {
            margin: 0.5em 0;
            @include mobile {
                // Hide title on mobile due to space limitations
                display: none;
            }
        }
    }

    #content {
        padding: 1rem 1.5rem;
        box-sizing: border-box;
        flex-direction: column;
        display: flex;
        gap: 1rem;

        @include widescreen {
            height: calc(100vh - 60px);
            flex-direction: row;
        }

        @include mobile {
            min-height: calc(100vh - 48px - 3.5rem);
            padding: 1rem;
        }

        #left {
            background-color: var(--clr-height-0-1);
            border-radius: 0.4rem;
            height: 75%;
            width: 100%;
            box-sizing: border-box;
            padding: 1rem 1.5rem;

            h6 {
                margin-bottom: 0.5rem;
                margin-top: 1rem;
                font-size: 1.1rem;
                color: var(--clr-text-hint);
            }

            .geo {
                &__title {
                    display: flex;
                    align-items: center;

                    h6 {
                        margin: 0;
                    }

                    :global &__help {
                        color: var(--clr-text-disabled);
                    }
                }
                &__inputs {
                    display: flex;
                    gap: 1rem;
                }
            }

            .automation {
                &__label {
                    display: flex;
                    gap: 1rem;
                    align-items: center;
                }
            }

            .lockdown {
                &__label {
                    display: flex;
                    gap: 1rem;
                    align-items: center;
                }
            }

            @include widescreen {
                height: 100%;
                width: 80%;
            }
        }

        #logs {
            background-color: var(--clr-height-0-1);
            border-radius: 0.4rem;
            height: 25%;
            width: 100%;

            @include widescreen {
                height: 100%;
                width: 20%;
            }
        }
    }
</style>