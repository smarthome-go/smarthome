<script lang="ts">
    import Button from "@smui/button";
    import type { DriverData, FetchedDriver } from "../driver";
    import DynamicConfigurator from "../../../components/Homescript/DynamicConfigurator.svelte";
    import { createEventDispatcher, onMount } from "svelte";
    import EditDriver from "./EditDriver.svelte";
    import IconButton from "@smui/icon-button";
    import DriverInfo from "./DriverInfo.svelte";
    import Ripple from '@smui/ripple'
    import { createSnackbar } from "../../../global";

    export let driver: FetchedDriver = null

    const dispatch = createEventDispatcher()

    // NOTE: commented out code is only used for development purposes.

    // let textareaContent = ""
    // let preventReacttoOutput = false
    // let textarea: HTMLTextAreaElement = null

    let dirty = false
    let lastOutput = null
    // Determines whether this is the first time that the output hook fired.
    // Only if this is `false`, does the `dirty` flag change.
    let initialOutput = false

    // function reactToInput() {
    //     textareaContent = textarea.value
    //     try {
    //         driver.configuration = JSON.parse(textareaContent)
    //         lastOutput = JSON.parse(textareaContent)
    //
    //         if (!preventReacttoOutput) {
    //             preventReacttoOutput = true
    //             setTimeout(() => preventReacttoOutput = false, 10)
    //         }
    //     } catch (err) {
    //         console.error(`JSON parse error: `, err)
    //     }
    // }

    function reactToOutput(data: any) {
        if (!initialOutput) { initialOutput = true } else {
            dirty = true
        }
        lastOutput = structuredClone(data)
    }


    function commitConfig() {
        dirty = false
        dispatch('save', lastOutput)
    }

    // If dynamic data is not `null`, it was modified and should be updated.
    async function modifyDriver(dataIn: DriverData, dynamicData: {} | null) {
        try {
            const res = await fetch('/api/system/hardware/driver/modify', {
                method: 'PUT',
                body: JSON.stringify(dataIn)
            });

            if (res.status !== 200)  {
                let msg  = await res.json()
                throw `${msg.message}: ${msg.error}`
            }

        } catch (error) {
            $createSnackbar(`Could not modify driver: ${error}`)
        }

        if (dynamicData !== null) {
            await saveDriverConfig(dynamicData)
        }

        driver.driver = dataIn
    }

    async function saveDriverConfig(input: {}) {
        try {
            let res = await fetch(
                '/api/system/hardware/driver/configure', {
                    method: "PUT",
                    body: JSON.stringify({
                        driver: {
                            vendorId: driver.driver.vendorId,
                            modelId: driver.driver.modelId
                        },
                        data: input
                    })
                }
            )

            if (res.status !== 200) {
                let msg  = await res.json()
                throw `${msg.message}: ${msg.error}`
            }
        } catch (err) {
             $createSnackbar(`Saving driver configuration failed: ${err}`)
        }
    }

    async function deleteDriver() {
        dispatch('delete')
    }

    let editDriverShow = null

    // function editCodeUrl(vendorId: string, modelId: string): string {
    //     const hmsEditorUrl = "/hmsEditor"
    //     return `${hmsEditorUrl}?id=@driver:${vendorId}:${modelId}`
    // }

    let infoOpen = false

    function openInfo() {
        infoOpen = true
    }
</script>

{#if driver !== null}
    <EditDriver
        data={driver.driver}
        configSchema={driver.info.driver.info.config}
        dynamicConfig={driver.configuration}
        on:modify={(e) => modifyDriver(e.detail.data, e.detail.dynamic)}
        on:delete={deleteDriver}
        bind:show={editDriverShow}
    />
    <DriverInfo bind:data={driver} bind:open={infoOpen}/>
{/if}

<div class="driver">
    <div class="driver__top">
        <span class="driver__top__name">{driver.driver.name}</span>
        <div class="driver__top__meta">
            <span class="driver__top__meta__type">
                Input
                <i class="material-icons">power</i>
            </span>
            <div class="driver__top__meta__capabilities">
                {#if driver.info.driver.capabilities !== null}
                    {#each driver.info.driver.capabilities as capability}
                        <div class="driver__top__meta__capabilities__chip capability">
                            <span>{capability}</span>
                        </div>
                    {/each}
                {/if}
            </div>
        </div>
    </div>

    <div class="bottom">
        <div class="driver__health">
            <span class='text-hint'>Status:</span>
            <!-- Homescript Status -->
            <div
                class="driver__health__chip"
                use:Ripple={{ surface: true }}
                class:ok={driver.validationErrors.length === 0}
                on:click={openInfo}
                on:keydown={openInfo}
            >
                <span>Homescript</span>
            </div>
            <!-- Driver Integrity -->
            <div
                class="driver__health__chip"
                use:Ripple={{ surface: true }}
                class:ok={true}
                on:click={openInfo}
                on:keydown={openInfo}
            >
                <span>Integrity</span>
            </div>
        </div>

        <div class="driver__bottom">
            <span class="driver__bottom__id text-hint">
                <span><code>{driver.driver.vendorId}:{driver.driver.modelId}</code></span>
                <i class="material-icons driver__bottom__id__icon">
                    power
                </i>
            </span>
            <div class="bottom__buttons">
                <IconButton class="material-icons" on:click={editDriverShow}>edit</IconButton>
                <IconButton class="material-icons" on:click={openInfo}>info</IconButton>
            </div>
        </div>
    </div>
</div>

<style lang="scss">
    @use './drivers.scss' as *;

    .driver {
        height: $driver-height;
        width: $driver-width;

        border-radius: 0.3rem;
        padding: 1rem;
        display: flex;
        flex-direction: column;
        justify-content: space-between;

        background-color: var(--clr-height-1-3);

        &__top {
            display: flex;
            flex-direction: column;
            padding-bottom: .7rem;

            &__name {
                font-weight: bold;
            }

            &__meta {
                display: flex;
                align-items: center;

                &__type {
                    display: flex;
                    align-items: center;
                    gap: 0.1rem;
                    font-size: 0.85rem;
                    color: var(--clr-text-hint);

                    i {
                        font-size: 1rem;
                    }
                }
            }
        }

        &__health, &__top__meta__capabilities {
            display: flex;
            gap: 0.2rem;
            flex-wrap: nowrap;
            overflow-x: hidden;
            align-items: center;

            &__chip {
                user-select: none;
                border-radius: 0.6rem;
                background-color: var(--clr-height-3-4);
                opacity: 70%;
                padding: 0 0.5rem;
                font-size: 0.8rem;
                cursor: default;
                display: flex;
                align-items: center;
                gap: 0.4rem;
                max-width: 5rem;
                color: var(--clr-error);

                &.capability {
                    color: var(--clr-priority-medium);
                }

                &.ok {
                    color: var(--clr-success);
                }

                span {
                    overflow: hidden;
                    white-space: nowrap;
                    text-overflow: ellipsis;
                }
            }
        }


        &__bottom {
            display: flex;
            gap: 0.5rem;
            align-items: center;
            justify-content: space-between;

            &__buttons {
                display: flex;
            }

            &__id {
                display: flex;
                align-items: center;
                gap: 0.5rem;
                font-size: 0.9rem;

                &__icon {
                    font-size: 1.2rem;
                }
            }
        }
    }
</style>
