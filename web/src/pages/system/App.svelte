<script lang="ts">
    import IconButton from "@smui/icon-button";
    import Tab, { Label } from '@smui/tab'
    import TabBar from '@smui/tab-bar'
    import Page from "../../Page.svelte";
    import Logs from "./Logs.svelte";
    import Progress from "../../components/Progress.svelte";
    import Button, { Icon } from "@smui/button";
    import Textfield from "@smui/textfield";
    import type { mqttSystemConfig, systemConfig } from "./main";
    import { onMount } from "svelte";
    import { createSnackbar } from "../../global";
    import HelperText from "@smui/textfield/helper-text";
    import GeoHelp from "./dialogs/GeoHelp.svelte";
    import FormField from "@smui/form-field";
    import Switch from "@smui/switch";
    import ExportImport from "./ExportImport.svelte";
    import Hardware from "./hardware/Hardware.svelte";
    import PurgeCache from "./dialogs/PurgeCache.svelte";
    import Drivers from "./hardware/Drivers.svelte";

    let loading = 0;

    let automationEnabledLoading = false;
    let lockDownModeEnabledLoading = false;
    let mqttLoading = 0;


    // Specifies whether the dialog for flushing cached data should be open or closed
    let purgeCacheOpen = false;

    // Specifies whether the log event dialog should be visible or not
    let logsOpen = false;

    // Specifies whether the geolocation help dialog should be open
    let geoHelpOpen = false;

    let config: systemConfig = {
        automationEnabled: false,
        lockDownMode: false,
        openWeatherMapApiKey: "",
        latitude: 0.0,
        longitude: 0.0,
        mqtt: {
            enabled: false,
            host: "",
            port: 0,
            username: "",
            password: "",
        }
    };

    let latitudeInput = 0.0;
    let longitudeInput = 0.0;
    let owmInput = "";

    let mqttInput: mqttSystemConfig = {
        enabled: false,
        host: "",
        port: 0,
        username: "",
        password: ""
    }

    let statusMQTT: mqttStatus = {
        working: false,
        error: null
    }

    async function fetchConfig() {
        loading++
        try {
            const res = await (await fetch("/api/system/config")).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            config = res;
            owmInput = res.openWeatherMapApiKey;
            latitudeInput = res.latitude;
            longitudeInput = res.longitude;

            mqttInput = structuredClone(config.mqtt)
        } catch (err) {
            $createSnackbar(`Failed to load system configuration: ${err}`);
        }
        loading--
    }

    async function updateOWMKey() {
        loading++
        try {
            const res = await (
                await fetch("/api/weather/key/modify", {
                    method: "PUT",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        key: owmInput,
                    }),
                })
            ).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            config.openWeatherMapApiKey = owmInput;
        } catch (err) {
            $createSnackbar(`Failed to update OpenWeatherMap API key: ${err}`);
        }
        setTimeout(() => {
            loading--
        }, 750);
    }

    async function fetchMQTTStatus() {
        mqttLoading++
        try {
            const res = await (await fetch("/api/system/mqtt/status")).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            statusMQTT = res
        } catch (err) {
            $createSnackbar(`Failed to load MQTT status: ${err}`);
        }

        setTimeout(() => {
            mqttLoading--
        }, 750);
    }

    async function updateMQTT() {
        loading++
        mqttLoading++

        try {
            let res = await fetch("/api/system/mqtt/config", {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(mqttInput),
            })

            let resJson: any = {}
            if (res.status === 200 || res.status === 500) {
                resJson = await res.json();
            }

            if (res.status === 500) {
                statusMQTT.workning = false
                statusMQTT.error = resJson.error
            } else if (resJson.success !== undefined && !resJson.success) {
                throw Error(resJson.error);
            }

            config.mqtt = structuredClone(mqttInput)
            await fetchMQTTStatus()
        } catch (err) {
            $createSnackbar(`Failed to update MQTT configuration: ${err}`);
        }

        setTimeout(() => {
            mqttLoading--
            loading--
        }, 750);
    }

    async function updateGeolocation() {
        loading++
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
        loading--
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


    type Activity = 'general' | 'drivers'
    let currentActivity: Activity = 'general'
    let activities: Activity[] = ['general', 'drivers']

    // As soon as the component is mounted, fetch the configuration
    onMount(() => {
        fetchConfig()
        fetchMQTTStatus()
    });
</script>


<!-->Purge cache dialog</-->
<PurgeCache bind:open={purgeCacheOpen} />

<!-->Log record dialog</-->
<Logs bind:open={logsOpen} />

<!-->Geolocation help dialog</-->
<GeoHelp bind:open={geoHelpOpen} />

<Page>
    <div id="header" class="mdc-elevation--z4">
        <TabBar tabs={activities} let:tab bind:active={currentActivity}>
            <Tab {tab} minWidth>
                <Label>{tab}</Label>
            </Tab>
        </TabBar>

        <div id="header__buttons">
            <IconButton
                title="Purge Cache"
                class="material-icons"
                on:click={() => purgeCacheOpen = true}>cleaning_services</IconButton
            >
            <IconButton
                title="Refresh"
                class="material-icons"
                on:click={fetchConfig}>refresh</IconButton
            >
            <Button on:click={() => (logsOpen = true)}>Logs</Button>
        </div>
    </div>
    <Progress loading={loading > 0} />
    <div id="content">
        {#if currentActivity == 'general'}
        <div id="general" class="mdc-elevation--z1">
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
                    <IconButton
                        class="material-icons"
                        title="Reset"
                        size="button"
                        on:click={() => {
                            latitudeInput = config.latitude;
                            longitudeInput = config.longitude;
                        }}
                        disabled={latitudeInput === config.latitude &&
                            longitudeInput === config.longitude}
                        >undo</IconButton
                    >
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
            <div class="owm">
                <div class="owm__title">
                    <h6>Open Weather Map</h6>
                    <Button
                        on:click={updateOWMKey}
                        disabled={config.openWeatherMapApiKey === owmInput}
                    >
                        <Label>Save</Label>
                        <Icon class="material-icons">save</Icon>
                    </Button>
                </div>
                <Textfield
                    bind:value={owmInput}
                    label="API Key"
                    type="password"
                >
                    <HelperText slot="helper"
                        >Your OWM API Key for weather data</HelperText
                    >
                </Textfield>
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

            <div class="mqtt-status">
                <div class="mqtt-status__title">
                    <h6>MQTT Status</h6>
                    <IconButton
                        title="Update"
                        class="material-icons"
                        on:click={fetchMQTTStatus}>refresh</IconButton
                    >
                </div>

                <div class="mqtt-status__status">
                    {#if statusMQTT.working}
                        Working <i class="material-icons">check</i>
                    {:else}
                        Failure: <code>{statusMQTT.error}</code> <i class="material-icons">error</i>
                    {/if}

                    <Progress
                        type={"circular"}
                        loading={mqttLoading > 0}
                    />
                </div>
            </div>

            <div class="mqtt">
                <div class="mwtt__title">
                    <h6>MQTT Configuration</h6>
                    <Button
                        on:click={updateMQTT}
                        disabled={JSON.stringify(mqttInput) == JSON.stringify(config.mqtt)}
                    >
                        <Label>Save</Label>
                        <Icon class="material-icons">save</Icon>
                    </Button>
                </div>

                <FormField>
                    <Switch
                        bind:checked={mqttInput.enabled}
                    />
                    <div slot="label" class="lockdown__label">
                        <span
                            >MQTT subsystem {config.mqtt.enabled
                                ? "online"
                                : "offline"}</span
                        >
                    </div>
                </FormField>

                <div class="mqtt__host">
                    <div>
                        <Textfield
                            bind:value={mqttInput.host}
                            label="Host"
                            type="text"
                        >
                            <HelperText slot="helper"
                                >Hostname without port.</HelperText
                            >
                        </Textfield>
                    </div>

                    <div>
                        <Textfield
                            bind:value={mqttInput.port}
                            label="Port"
                            type="number"
                        >
                            <HelperText slot="helper"
                                >MQTT Port.</HelperText
                            >
                        </Textfield>
                    </div>
                </div>

                <div class="mqtt__credentials">
                    <div>
                        <Textfield
                            bind:value={mqttInput.username}
                            label="Username"
                            type="text"
                        >
                            <HelperText slot="helper"
                                >MQTT Username</HelperText
                            >
                        </Textfield>
                    </div>

                    <div>
                        <Textfield
                            bind:value={mqttInput.password}
                            label="Password"
                            type="password"
                        >
                            <HelperText slot="helper"
                                >MQTT Password</HelperText
                            >
                        </Textfield>
                    </div>
                </div>
            </div>
            <ExportImport />
        </div>
        {:else if currentActivity == 'drivers'}
            <!-- TODO: is being replaced by generic drivers -->
            <!-- <div id="hardware-left" class="mdc-elevation--z1"> -->
            <!--     <Hardware /> -->
            <!-- </div> -->
            <div id="hardware-right" class="mdc-elevation--z1">
                <Drivers />
            </div>
        {:else}
            Unsupported activity
        {/if}
    </div></Page
>

<style lang="scss">
    @use "../../mixins" as *;

    #header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        box-sizing: border-box;
        background-color: var(--clr-height-1-4);
        min-height: 3.5rem;

        &__buttons {
            display: flex;
            align-items: center;
        }
    }

    #content {
        padding: 1rem 1.5rem;
        box-sizing: border-box;
        flex-direction: column;
        display: flex;
        gap: 1rem;
        height: calc(100vh - 60px);

        @include widescreen {
            flex-direction: row;
        }

        @include mobile {
            height: auto;
            padding: 1rem;
        }

        #general {
            background-color: var(--clr-height-0-1);
            border-radius: 0.4rem;
            height: 100%;
            width: 100%;
            box-sizing: border-box;
            padding: 1rem 1.5rem;
            overflow-y: auto;

            h6 {
                margin: 0;
                color: var(--clr-text-hint);
                font-size: 1rem;

                @include widescreen {
                    margin-bottom: 0.5rem;
                    margin-top: 1rem;
                    font-size: 1.1rem;
                }
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

            .owm {
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
            }

            .mqtt-status {
                &__title {
                    display: flex;
                    align-items: center;

                    h6 {
                        margin: 0;
                    }
                }

                &__status {
                    display: flex;
                    align-items: center;
                    gap: 1rem;
                }
            }

            .mqtt {
                &__host, &__credentials {
                    display: flex;
                    align-items: center;
                    gap: 1rem;
                }

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
            }

            @include mobile {
                height: 100%;
            }
        }

        #hardware-left, #hardware-right {
            background-color: var(--clr-height-0-1);
            border-radius: 0.4rem;
            width: 100%;
            overflow-y: auto;
            padding-bottom: 1rem;

            @include widescreen {
                padding-bottom: 0;
            }

            @include mobile {
                height: auto;
            }
        }
    }
</style>
