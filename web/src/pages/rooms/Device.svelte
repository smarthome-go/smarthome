<script lang="ts">
    import IconButton from '@smui/icon-button'
    import Switch from '@smui/switch'
    import { createEventDispatcher, onMount } from 'svelte'
    import Progress from '../../components/Progress.svelte'
    import { createSnackbar, hasPermission, sleep } from '../../global'
    import EditDevice from './dialogs/device/EditDevice.svelte'
    import DeviceInfo from './dialogs/device/DeviceInfo.svelte'
    import Ripple from '@smui/ripple'
    import type { DeviceResponse } from '../../device';
    import Slider from '@smui/slider';
    import FormField from '@smui/form-field';
    import Button, { Label, Icon } from '@smui/button';
    // import Terminal from '../../components/Homescript/ExecutionResultPopup/Terminal.svelte'
    import ExecutionResultPopup from '../../components/Homescript/ExecutionResultPopup/ExecutionResultPopup.svelte'
    import GenericDevice from './GenericDevice.svelte';

    import type { DeviceCapability, ValidationError } from '../../driver';
    import type { homescriptError } from '../../homescript';

    // Event dispatcher
    const dispatch = createEventDispatcher()

    let deviceInfoOpen = false
    let deviceEditOpen = false

    export let data: DeviceResponse = {
        type: 'INPUT',
        id: '',
        name: '',
        vendorId: '',
        modelId: '',
        roomId: '',
        singletonJson: {},
        hmsErrors: [],
        config: {
            capabilities: [],
            info: null
        },
        powerInformation: {
            state: false,
            powerDrawWatts: 0
        },
        dimmables: [],
    }

    let requests = 0
    let loading = false

    // Is bound to the `editSwitch` in order to pass an event to a child
    let showEditDevice: () => void

    let showDeviceInfo = () => deviceInfoOpen = true

    // Determines if edit button should be shown
    let hasEditPermission: boolean
    onMount(async () => {
        hasEditPermission = await hasPermission('modifyRooms')
    })

    let isWide = hasEditPermission

    $: loading = requests !== 0
    async function toggle(event: CustomEvent<{ selected: boolean }>) {
        // Send a event in order to signal that the cameras should be reloaded
        dispatch('powerChange', null)
        requests++
        try {
            const res = await (
                await fetch('/api/devices/action/power', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        deviceId: data.id,
                        power: {
                            state: event.detail.selected,
                        },
                    }),
                })
            ).json()

            if (!res.success) {
                for (let error of (res.hmsErrors as homescriptError[])) {
                    pushUserError(error)
                }
            }
        } catch (err) {
            $createSnackbar(
                `Failed to set device power '${data.name}' to ${event.detail.selected ? 'on' : 'off'}: ${err}`,
            )
        }
        await sleep(500)
        requests--
        dispatch('powerChangeDone', null)
    }

    // TODO: introduce timer to only update if the user has finished their input.
    async function dim(value: number, label: string) {
        // Send a event in order to signal that the cameras should be reloaded
        dispatch('dim', null)
        requests++
        try {
            const res = await (
                await fetch('/api/devices/action/dim', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        deviceId: data.id,
                        dim: {
                            percent: value,
                            label,
                        },
                    }),
                })
            ).json()

            if (!res.success) {
                for (let error of (res.hmsErrors as homescriptError[])) {
                    pushUserError(error)
                }
            }
        } catch (err) {
            $createSnackbar(
                `Failed to set device '${data.name}' dimmable '${label}' to ${value}: ${err}`,
            )
        }
        await sleep(500)
        requests--
        dispatch('dimDone', null)
    }

    let homescriptCode: Map<string, string> = new Map()
    let sourcesUpToDate = false

    interface ErrorWrapper {
        userCaused: boolean
        error: homescriptError
    }

    // Error handling and recovery
    let errors: ErrorWrapper[] = []
    $: if (data.hmsErrors !== null && data.hmsErrors.length > 0) {
        errors = data.hmsErrors.map((error) => Object.create({userCaused: false, error}))
    }
    $: if(errors.length) loadHmsSources(errors.map(e => e.error.span.filename))

    // TODO: optimize this!
    async function loadHmsSources(ids: string[]) {
        sourcesUpToDate = false

        let res = await (await fetch("/api/homescript/sources", {
            method: 'PUT',
            body: JSON.stringify({
                ids: [...new Set(ids)],
            })
        })).json()

        for (let item of Object.keys(res)) {
            homescriptCode.set(item, res[item])
        }

        sourcesUpToDate = true
    }

    function pushUserError(error: homescriptError) {
        errors = [...errors, {
            userCaused: true,
            error,
        }]
    }

    let errorsOpen = false
    $: errorsOpen = errors.find((e) => e.userCaused) !== undefined

    function hasCapability(capability: DeviceCapability): boolean { return data.config.capabilities.includes(capability) }
</script>

<EditDevice
    on:delete={() => dispatch('delete', null)}
    on:modify={event => {
        // TODO: implement copy
        // name = event.detail.name
        // watts = event.detail.watts
        // targetNode = event.detail.targetNode
        // event.detail.id = id
        dispatch('modify', event.detail)
    }}
    {data}
    bind:show={showEditDevice}
/>

<DeviceInfo bind:open={deviceInfoOpen} {data} />

<GenericDevice
    name={data.name}
    {hasEditPermission}
    isTall={hasCapability('dimmable')}
    on:info_show={() => deviceInfoOpen = true}
    on:edit_show={showEditDevice}
>
    <div slot='top'>
        {#if hasCapability('power')}
            <div class="device__power">
                <Switch icons={false} bind:checked={data.powerInformation.state} on:SMUISwitch:change={toggle} />
            </div>
        {/if}
    </div>

    <div slot='extend'>
        {#if (errors !== null) && errors.length > 0}
                {#if sourcesUpToDate}
                    <ExecutionResultPopup
                        open={errorsOpen && sourcesUpToDate}
                        data={{
                            modeRun: true,
                            response: {
                                title: `Driver invocation '${data.name}'`,
                                success: false,
                                output: "",
                                fileContents: homescriptCode, // TODO
                                errors: errors.map(w => w.error),
                            },
                        }}
                        on:close={() => {
                            // This hack is required so that the window still remains scrollable after removal
                        }}
                    />
                {/if}

                <div class="device__error">
                    {data.hmsErrors.length} Error {data.hmsErrors.length != 1 ? 's' : ''}
                    <Button on:click={() => errorsOpen=true}>
                        <Label>Inspect</Label>
                        <Icon class="material-icons">bug_report</Icon>
                    </Button>
                </div>
        {/if}

        {#if hasCapability('dimmable')}
            <div class="device__dim">
                {#each data.dimmables as dimmable}
                    <div class="device__dim__sep"/>
                    <div class="device__dim__item">
                        <span class="device__dim__item__name text-hint">{dimmable.label}</span>
                        <div class="device__dim__item__body">
                            <div class="device__dim__item__body__left">
                                <FormField align="start" style="display: flex;">
                                    <!-- TODO: does this also update the value??? -->
                                    <Slider
                                        style="flex-grow: 1;"
                                        bind:value={dimmable.value}
                                        on:SMUISlider:change={(e) => dim(e.detail.value, dimmable.label)}
                                    />
                                </FormField>
                            </div>
                            <div class="device__dim__item__body__right">
                                <span class="status text-hint">{dimmable.value}</span>
                            </div>
                        </div>
                    </div>
                {/each}
            </div>
        {/if}
    </div>
</GenericDevice>

<style lang='scss'>
    .device {
        &__dim {
            display: flex;
            flex-direction: column;
            flex-grow: 0;

            &__sep {
                width: 100%;
                background-color: var(--clr-height-3-6);
                border-radius: .3rem;
                height: .125rem;

                // TODO: decide whether to include this.
                //&:first-of-type {
                    //display: none;
                //}
            }

            &__item {
                background-color: var(--clr-height-1-3);
                border-radius: 0.3rem;
                padding: 0.8rem;
                padding-left: 0;
                display: flex;
                flex-direction: column;

                &__name {
                    font-size: .65rem;
                    margin-bottom: -.5rem;
                    padding-left: .85rem;
                }

                &__body {
                    display: flex;

                    &__left {
                        width: 85%;
                    }

                    &__right {
                        display: flex;
                        flex-direction: column;
                        justify-content: center;
                        width: 15%;
                        font-size: .8rem;
                    }
                }
            }
        }
    }
</style>
