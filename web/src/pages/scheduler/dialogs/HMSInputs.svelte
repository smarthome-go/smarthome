<script lang="ts">
    import Tab, { Icon, Label } from "@smui/tab";
    import TabBar from "@smui/tab-bar";
    import { createSnackbar } from "../../../global";
    import HmsEditor from "../../../components/Homescript/HmsEditor/HmsEditor.svelte";
    import HmsInputsReset from "./HMSInputsReset.svelte";
    import HmsSelector from "../../../components/Homescript/HmsSelector.svelte";
    import Progress from "../../../components/Progress.svelte";
    import IconButton from "@smui/icon-button";
    import FormField from "@smui/form-field";
    import Switch from "@smui/switch";
    import Select, { Option } from "@smui/select";
    import { devices, devicesLoaded, homescripts } from "../main";
    import type { ScheduleData, SwitchResponse, ScheduleTargetMode } from "../main";
    import Button from "@smui/button/src/Button.svelte";

    export let data: ScheduleData = {
        name: "",
        hour: 0,
        minute: 0,
        targetMode: "hms",
        homescriptCode: "",
        homescriptTargetId: "",
        deviceJobs: [],
    };

    /*
        //// Tabs (active mode selection) ////
    */

    // Specifies which mode is currently being used for editing
    let active  = data.targetMode;

    // Saves the tab data for the editor type selection
    const tabs: ScheduleTargetMode[] = ["hms", "devices", "code"];
    const tabData: { label: ScheduleTargetMode; icon: string }[] = [
        {
            label: "hms",
            icon: "list",
        },
        {
            label: "devices",
            icon: "power",
        },
        {
            label: "code",
            icon: "code",
        },
    ];

    /*
        //// Devices ////
        Used for when the active mode is set to `devices`
    */
    // Saves the devices which are available to the current user
    // Used for displaying the device selector
    let deviceToBeInserted: string;
    let devicesAvailable: SwitchResponse[] = [];

    $: if (
        $devicesLoaded &&
        $devices.length > 0 &&
        data.deviceJobs !== undefined
    )
        updateSwitchesAvailable();

    function updateSwitchesAvailable() {
        devicesAvailable = $devices.filter((device) => {
            return (
                data.deviceJobs.filter((job) => job.deviceId === device.id).length === 0
            );
        });
        // Causes an update in the selection element
        if (devicesAvailable.length === 1)
            deviceToBeInserted = devicesAvailable[0].id;
        else deviceToBeInserted = undefined;
    }
</script>

<div class="main">
    <div class="main__header">
        <TabBar {tabs} let:tab bind:active>
            <Tab {tab}>
                <Icon class="material-icons"
                    >{tabData.find((t) => t.label === tab).icon}</Icon
                >
                <Label>{tab}</Label>
            </Tab>
        </TabBar>
    </div>
    <div class="main__editor" class:hms={active === "hms"}>
        {#if active === "hms"}
            {#if data.targetMode !== active}
                <HmsInputsReset
                    activeInCode={data.targetMode}
                    icon="auto_fix_off"
                    on:reset={() => (data.targetMode = active)}
                />
            {:else}
                <div class="main__editor__homescript">
                    {#if $homescripts.length > 0}
                        <HmsSelector
                            homescripts={$homescripts}
                            bind:selection={data.homescriptTargetId}
                        />
                    {:else}
                        <div class="main__editor__homescript__empty">
                            <i class="material-icons">code_off</i>
                            <div class="main__editor__homescript__empty__text">
                                <h6>No Homescripts available</h6>
                                <span class="text-hint"
                                    >Make sure the <span
                                        style="color: var(--clr-primary)"
                                        >'Show Selection'</span
                                    > setting is enabled for Homescripts which should
                                    appear up here.</span
                                >
                                <br>
                                <span class="text-disabled"
                                    >You can find this setting under 'Selection
                                    and visibility'</span
                                >
                            </div>
                            <Button href="/homescript" variant="outlined"
                                >To Homescript</Button
                            >
                        </div>
                    {/if}
                </div>
            {/if}
        {:else if active === "devices"}
            {#if data.targetMode !== active}
                <HmsInputsReset
                    {active}
                    activeInCode={data.targetMode}
                    icon="auto_fix_off"
                    on:reset={() => (data.targetMode = active)}
                />
            {:else if devicesLoaded}
                <div class="main__editor__switches__header mdc-elevation--z1">
                    <Select
                        bind:value={deviceToBeInserted}
                        label="Select Switch"
                        disabled={devicesAvailable.length <= 1}
                    >
                        {#each devicesAvailable as swOpt}
                            <Option value={swOpt.id}>{swOpt.name}</Option>
                        {/each}
                    </Select>
                    <IconButton
                        class="material-icons"
                        disabled={deviceToBeInserted === undefined}
                        on:click={() => {
                            if (devicesAvailable.length === 0) {
                                $createSnackbar(
                                    "Only one action per device is allowed."
                                );
                                return;
                            }

                            data.deviceJobs = [
                                ...data.deviceJobs,
                                {
                                    deviceId: deviceToBeInserted,
                                    powerOn: false,
                                },
                            ];
                        }}
                    >
                        add
                    </IconButton>
                </div>
                <div class="main__editor__switches__wizard">
                    {#if data.deviceJobs.length === 0}
                        <div class="main__editor__switches__no-selection">
                            <i
                                class="main__editor__switches__no-selection__icon material-icons"
                                >power_off</i
                            >
                            <div
                                class="main__editor__switches__no-selection__text"
                            >
                                {#if $devices.length === 0 && $devices}
                                    <h6>No Devices Available</h6>
                                    You need to have access to at least one device.
                                    <br />
                                    <span class="text-disabled">
                                        If this is unintentional, contact your
                                        administrator.
                                    </span>
                                {:else}
                                    <h6>Empty Procedure</h6>
                                    Your current procedure is empty, use the menu
                                    above in order to create a new switch action.
                                {/if}
                            </div>
                        </div>
                    {/if}
                    {#each data.deviceJobs as job (job.deviceId)}
                        <div
                            class="main__editor__switches__wizard__item mdc-elevation--z1"
                        >
                            <FormField>
                                <Switch
                                    bind:checked={job.powerOn}
                                    icons={false}
                                />
                                <span slot="label">{$devices.find((device) => device.id === job.deviceId).name}</span>
                            </FormField>
                            <IconButton
                                class="material-icons"
                                on:click={() => {
                                    data.deviceJobs = data.deviceJobs.filter(
                                        (j) => j.deviceId !== job.deviceId
                                    );
                                }}
                            >
                                delete
                            </IconButton>
                        </div>
                    {/each}
                </div>
            {:else}
                <Progress type="circular" loading={true} />
            {/if}
        {:else if data.targetMode !== active}
            <HmsInputsReset
                {active}
                activeInCode={data.targetMode}
                icon="code_off"
                on:reset={() => (data.targetMode = active)}
            />
        {:else}
            <HmsEditor registerCtrlSCatcher bind:code={data.homescriptCode} />
        {/if}
    </div>
</div>

<style lang="scss">
    @use "../../../mixins" as *;
    .main {
        &__editor {
            height: 25rem;

            @include mobile {
                height: auto;
                min-height: 20rem;
            }

            &__homescript {
                &__empty {
                    margin-top: 4rem;
                    display: flex;
                    flex-direction: column;
                    align-items: center;
                    gap: 1rem;

                    @include mobile {
                        margin-top: 2rem;
                    }

                    i {
                        font-size: 5rem;
                        color: var(--clr-text-disabled);
                    }

                    &__text {
                        max-width: 50%;

                        @include widescreen {
                            max-width: 60%;
                        }

                        @include mobile {
                            max-width: 100%;
                        }

                        h6 {
                            margin: 0.5rem 0;
                        }

                        span {
                            // Placeholder
                        }
                    }
                }
            }

            &__switches {
                &__header {
                    background-color: var(--clr-height-0-1);
                    padding: 0.5rem;
                    overflow: visible;
                }

                &__no-selection {
                    display: flex;
                    flex-direction: column;
                    align-items: center;
                    gap: 1rem;
                    margin-top: 2rem;

                    &__icon {
                        display: block;
                        color: var(--clr-text-disabled);
                        font-size: 5rem;
                    }

                    &__text {
                        max-width: 50%;

                        @include widescreen {
                            max-width: 60%;
                        }

                        h6 {
                            margin: 0.1rem 0;
                        }
                    }
                }

                &__wizard {
                    display: flex;
                    flex-direction: column;
                    gap: 0.5rem;
                    margin-top: 1rem;

                    &__item {
                        height: 3rem;
                        border-radius: 0.3rem;
                        background-color: var(--clr-height-0-1);
                        padding: 0.5rem;
                        display: flex;
                        align-items: center;
                        justify-content: space-between;
                    }
                }
            }

            margin-top: 1rem;
            overflow: auto;

            @include not-widescreen {
                overflow: visible;
            }
        }
    }
</style>
