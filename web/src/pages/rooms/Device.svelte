<script lang="ts">
    // import IconButton from '@smui/icon-button'
    import Switch from '@smui/switch'
    import { createEventDispatcher, onMount } from 'svelte'
    // import Progress from '../../components/Progress.svelte'
    import { createSnackbar, hasPermission, sleep } from '../../global'
    import EditDevice from './dialogs/device/EditDevice.svelte'
    import DeviceInfo from './dialogs/device/DeviceInfo.svelte'
    import Ripple from '@smui/ripple'
    import type { DeviceExtractions, HydratedDeviceResponse, ShallowDeviceResponse } from '../../device';
    import Slider from '@smui/slider';
    import FormField from '@smui/form-field';
    // import Button, { Label, Icon } from '@smui/button';
    // import Terminal from '../../components/Homescript/ExecutionResultPopup/Terminal.svelte'
    import ExecutionResultPopup from '../../components/Homescript/ExecutionResultPopup/ExecutionResultPopup.svelte'
    import GenericDevice from './GenericDevice.svelte';

    import type { DeviceCapability, ValidationError } from '../../driver';
    import type { homescriptError } from '../../homescript';

    // Event dispatcher
    const dispatch = createEventDispatcher()

    let deviceInfoOpen = false
    let deviceEditOpen = false

    export let style = ''

    export let shallow: ShallowDeviceResponse = {
        type: 'INPUT', id: '',
        name: '',
        vendorId: '',
        modelId: '',
        roomId: '',
        singletonJson: {},
    }

    let extractionsLoaded = false
    let extractions: DeviceExtractions = {
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
        sensors: [],
    }

    export let capabilities: DeviceCapability[] = []
    $: capabilities = extractions.config.capabilities


    let localStorageKey = `amount_of_sliders_for_device_id__${shallow.id}`

    async function loadExtractions() {
        // TODO: bug
        requests++

        try {
            let res = await fetch(`/api/devices/extract/${shallow.id}`)
            let responseJson = await res.json()
            if (responseJson.error !== undefined) {
                throw(responseJson.error)
            }
            if (res.status !== 200) {
                throw(responseJson)
            }

            const extractionsTemp = (responseJson as HydratedDeviceResponse).extractions
            shallow = (responseJson as HydratedDeviceResponse).shallow

            console.dir(extractions)

            // Write the amount of sliders into the cache for smoother loading.
            if (extractionsTemp.dimmables != null) {
                writeNumSliders(extractionsTemp.dimmables.length)
            }

            extractionsLoaded = true
            extractions = extractionsTemp
        } catch (err) {
            $createSnackbar(`Failed to hydrate device: ${err}`)
        }

        requests--
    }

    // export let data: DeviceResponse = {
    //     shallow: {
    //         type: 'INPUT', id: '',
    //         name: '',
    //         vendorId: '',
    //         modelId: '',
    //         roomId: '',
    //         singletonJson: {},
    //     },
    //     hmsErrors: [],
    //     config: {
    //         capabilities: [],
    //         info: null
    //     },
    //     powerInformation: {
    //         state: false,
    //         powerDrawWatts: 0
    //     },
    //     dimmables: [],
    //     sensors: [],
    // }

    let requests = 0
    let loading = false
    let isInitialLoad = false

    // Is bound to the `editSwitch` in order to pass an event to a child
    let showEditDevice: () => void

    let showDeviceInfo = () => deviceInfoOpen = true

    // Determines if edit button should be shown
    let hasEditPermission: boolean

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
                        deviceId: shallow.id,
                        power: {
                            state: event.detail.selected,
                        },
                    }),
                })
            ).json()

            if (!res.success) {
                errors = []
                for (let error of (res.hmsErrors as homescriptError[])) {
                    pushUserError(error)
                }
            }
        } catch (err) {
            $createSnackbar(
                `Failed to set device power '${shallow.name}' to ${event.detail.selected ? 'on' : 'off'}: ${err}`,
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
                        deviceId: shallow.id,
                        dim: {
                            percent: value,
                            label,
                        },
                    }),
                })
            ).json()

            if (!res.success) {
                errors = []
                for (let error of (res.hmsErrors as homescriptError[])) {
                    pushUserError(error)
                }
            }
        } catch (err) {
            $createSnackbar(
                `Failed to set device '${shallow.name}' dimmable '${label}' to ${value}: ${err}`,
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
    $: if (extractionsLoaded && extractions.hmsErrors !== null && extractions.hmsErrors.length > 0) {
        errors = extractions.hmsErrors.map((error) => Object.create({userCaused: false, error}))
    }
    $: if(errors.length && canFetchSources !== undefined) loadHmsSources(errors.map(e => e.error.span.filename))

    let canFetchSources = undefined

    // TODO: optimize this!
    async function loadHmsSources(ids: string[]) {
        // They would not see any code as there would be a 403.
        if (!canFetchSources) {
            console.log("not fetching sources...")
            return
        }
        // TODO: what to do?

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

        errorsOpen = errors.find((e) => e.userCaused) !== undefined
    }

    let errorsOpen = false

    function hasCapability(self: DeviceCapability[], capability: DeviceCapability): boolean {
        return self !== null && self.includes(capability)
    }


    function deviceHeight(ex: DeviceExtractions): number {
        let height = 1;

        for (let c of ex.config.capabilities) {
            switch (c) {
                case 'base':
                    break
                case 'dimmable':
                    height += ex.dimmables.length
                    break
                case 'power':
                    break
                case 'sensor':
                    height += ex.sensors.length
                    break
            }
        }

        return height
    }

    let hasErrors = false
    $: hasErrors = errors !== null && errors.length > 0

    function writeNumSliders(num: number) {
        window.localStorage.setItem(localStorageKey, num.toString())
    }

    let height = 0;

    async function mount() {
        isInitialLoad = true

        let numSlidersRaw = window.localStorage.getItem(localStorageKey)
        if (numSlidersRaw == null) {
            writeNumSliders(0)
            numSlidersRaw = '0';
        }

        const numSliders = parseInt(numSlidersRaw)
        for (let i = 0; i < numSliders; i++) {
            extractions.dimmables.push({
                value: 0,
                label: "loading...",
                range: {
                    lower: 0,
                    upper: 100,
                }
            })
        }


        hasEditPermission = await hasPermission('modifyRooms')
        await loadExtractions()

        canFetchSources = (await hasPermission('modifyServerConfig')) && (await hasPermission('homescript'))
        console.log(`Configured error display: user can fetch sources: ${canFetchSources}`)

        let height = deviceHeight(extractions)
        style = `grid-row-end: span ${height};`

        isInitialLoad = false
    }

    onMount(mount)
</script>

{#if extractionsLoaded}
    <EditDevice
        on:delete={() => dispatch('delete', null)}
        on:modify={e => dispatch('modify', e.detail)}
        bind:show={showEditDevice}
        data={ { shallow: shallow, extractions } }
    />

    <DeviceInfo bind:open={deviceInfoOpen} data={{shallow, extractions}} />
{/if}

<GenericDevice
    {style}
    name={shallow.name}
    {loading}
    {isInitialLoad}
    {hasEditPermission}
    isTall={hasCapability(capabilities, 'dimmable') || hasCapability(capabilities, 'sensor')}
    on:info_show={() => deviceInfoOpen = true}
    on:edit_show={showEditDevice}
    {hasErrors}
>
    <div slot='top'>
        {#if hasCapability(capabilities, 'power')}
            <div class="device__power">
                <Switch icons={false} bind:checked={extractions.powerInformation.state} on:SMUISwitch:change={toggle} />
            </div>
        {/if}
    </div>

    <div slot='extend'>
        {#if hasCapability(capabilities, 'dimmable')}
            <div class="device__dim">
                {#each extractions.dimmables as dimmable}
                    <div class="device__dim__sep"/>
                    <div class="device__dim__item">
                        <span class="device__dim__item__name text-hint">{dimmable.label}</span>
                        <div class="device__dim__item__body">
                            <div class="device__dim__item__body__left">
                                <FormField align="start" style="display: flex;">
                                    <!-- TODO: does this also update the value??? -->
                                    <Slider
                                        min={dimmable.range.lower}
                                        max={dimmable.range.upper}
                                        step={1}
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

        {#if hasCapability(capabilities, 'sensor')}
            <div class="device__sensor">
                {#if extractions.sensors !== null}
                    {#each extractions.sensors as sensor}
                        <div class="device__sensor__sep"/>
                        <div class="device__sensor__reading">
                            <span class='text-disabled'>
                                {sensor.label}
                            </span>
                            <span class="text-hint">
                                {sensor.value}
                                {sensor.unit}
                            </span>
                        </div>
                    {/each}
                {/if}
            </div>
        {/if}
    </div>

    <div slot="bottom">
        {#if hasErrors}
            <div class="device__errors">
                {#if !canFetchSources}
                    <ExecutionResultPopup
                        bind:open={errorsOpen}
                        data={{
                            modeRun: true,
                            response: {
                                title: `Driver invocation '${shallow.name}'`,
                                success: false,
                                output: "",
                                fileContents: new Map(),
                                errors: errors.map(w => w.error),
                            },
                        }}
                        on:close={() => {
                            // This hack is required so that the window still remains scrollable after removal
                        }}
                    />
                {:else if sourcesUpToDate}
                    <ExecutionResultPopup
                        bind:open={errorsOpen}
                        data={{
                            modeRun: true,
                            response: {
                                title: `Driver invocation '${shallow.name}'`,
                                success: false,
                                output: "",
                                fileContents: homescriptCode,
                                errors: errors.map(w => w.error),
                            },
                        }}
                        on:close={() => {
                            // This hack is required so that the window still remains scrollable after removal
                        }}
                    />
                {/if}

                <span
                    class="device__errors__banner"
                    use:Ripple={{ surface: true }}
                    on:click={() => errorsOpen=true}
                    on:keydown={() => errorsOpen=true}
                >
                    <i class="material-icons">cancel</i>
                    {errors.length} Error {errors.length != 1 ? 's' : ''}
                </span>
            </div>
        {/if}
   </div>
</GenericDevice>

<style lang='scss'>
    .device {
        &__errors {
            display: flex;
            justify-content: space-between;
            align-items: center;
            user-select: none;

            &__banner {
                font-weight: bold;
                color: var(--clr-error);
                font-size: .95rem;
                padding: .1rem .4rem;
                border-radius: .3rem;
                cursor: pointer;
                display: flex;
                align-items: center;
                gap: .3rem;

                i {
                    font-size: 1rem;
                }
            }
        }

        @mixin separator {
            width: 100%;
            background-color: var(--clr-height-3-6);
            border-radius: .3rem;
            height: .125rem;
        }

        &__sensor {
            display: flex;
            flex-direction: column;
            flex-grow: 0;
            margin-top: -.75rem;
            padding: 0 1.5rem;
            padding-right: 1rem;

            &__reading {
                display: flex;
                gap: .4rem;
            }
        }

        &__dim {
            display: flex;
            flex-direction: column;
            flex-grow: 0;

            &__sep {
                @include separator;

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
