<script lang="ts">
    import IconButton from '@smui/icon-button'
    import Switch from '@smui/switch'
    import { createEventDispatcher, onMount } from 'svelte/internal'
    import Progress from '../../components/Progress.svelte'
    import { createSnackbar, hasPermission, sleep } from '../../global'
    import EditDevice from './dialogs/device/EditDevice.svelte'
    import DeviceInfo from './dialogs/device/DeviceInfo.svelte'
    import Ripple from '@smui/ripple'
    import type { DeviceResponse } from './main';
    import Slider from '@smui/slider';
    import FormField from '@smui/form-field';
    import Button from '@smui/button';

    // Event dispatcher
    const dispatch = createEventDispatcher()

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
        dimmableInformation: {
            percent: 0,
        }
    }

    let requests = 0
    let loading = false

    // Is bound to the `editSwitch` in order to pass an event to a child
    let showEditSwitch: () => void

    // Is bound to the `switchInfo` in order to pass an event to a child
    let showSwitchInfo: () => void

    // Determines if edit button should be shown
    let hasEditPermission: boolean
    onMount(async () => {
        hasEditPermission = await hasPermission('modifyRooms')
    })

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
            if (!res.success) throw Error(res.error)
        } catch (err) {
            $createSnackbar(
                `Failed to set device power '${data.name}' to ${event.detail.selected ? 'on' : 'off'}: ${err}`,
            )
        }
        await sleep(500)
        requests--
        dispatch('powerChangeDone', null)
    }

    $: dim(data.dimmableInformation.percent)

    async function dim(percent: number) {
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
                            percent,
                        },
                    }),
                })
            ).json()
            if (!res.success) throw Error(res.error)
        } catch (err) {
            $createSnackbar(
                `Failed to set device '${data.name}' brightness to ${data.dimmableInformation}: ${err}`,
            )
        }
        await sleep(500)
        requests--
        dispatch('dimDone', null)
    }
</script>

<!-- <EditSwitch -->
<!--     on:delete={() => dispatch('delete', null)} -->
<!--     on:modify={event => { -->
<!--         name = event.detail.name -->
<!--         watts = event.detail.watts -->
<!--         targetNode = event.detail.targetNode -->
<!--         event.detail.id = id -->
<!--         dispatch('modify', event.detail) -->
<!--     }} -->
<!--     {id} -->
<!--     {name} -->
<!--     {watts} -->
<!--     {targetNode} -->
<!--     bind:show={showEditSwitch} -->
<!-- /> -->

<!-- <SwitchInfo bind:show={showSwitchInfo} {id} {name} {watts} {targetNode} /> -->

<div class="switch mdc-elevation--z3" class:wide={hasEditPermission}>
    {#if data.config.capabilities.includes('power')}
        <div class="switch__power">
            <div class="switch__power__left">
                <Switch icons={false} bind:checked={data.powerInformation.state} on:SMUISwitch:change={toggle} />
                <div
                    class="switch__power__name__box"
                    use:Ripple={{ surface: true }}
                    on:click={() => showSwitchInfo()}
                    on:keydown={() => showSwitchInfo()}
                >
                    <span class="switch__power__name"> {data.name}</span>
                </div>
            </div>
            <div class="switch__power__right">
                <div>
                    <Progress type="circular" bind:loading />
                </div>
                {#if hasEditPermission}
                    <IconButton class="material-icons" title="Edit Switch" on:click={showEditSwitch}
                        >edit</IconButton
                    >
                {/if}
            </div>
        </div>
    {/if}

    {#if data.config.capabilities.includes('dimmable')}
        <div class="switch__dim">
                <div class="switch__dim__left">
                    <FormField align="start" style="display: flex;">
                        <Slider style="flex-grow: 1;" bind:value={data.dimmableInformation.percent} />
                        <span
                            slot="label"
                            style="padding-right: 12px; width: max-content; display: block;"
                        >
                        </span>
                    </FormField>
                </div>
                <div class="switch__dim__right">
                    <span class="status text-hint">{data.dimmableInformation.percent}%</span>
                </div>
        </div>
    {/if}
</div>

<style lang="scss">
    @use '../../mixins' as *;
    .switch {
        display: flex;
        flex-direction: column;
        gap: .3rem;

        &__dim {
            background-color: var(--clr-height-1-3);
            border-radius: 0.3rem;
            padding: 0.5rem;
            padding-left: 0;
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

        &__power {
            background-color: var(--clr-height-1-3);
            border-radius: 0.3rem;
            width: 15rem;
            height: 3.3rem;
            padding: 0.5rem;
            display: flex;
            align-items: center;
            justify-content: space-between;

            &.wide {
                width: 17rem;

                @include mobile {
                    width: 90%;
                }
            }

            & > * {
                display: flex;
                align-items: center;
            }
            &__left {
                max-width: 70%;
                gap: 0.2rem;
            }
            &__right {
                div {
                    margin-right: 14px;
                    display: flex;
                    align-items: center;
                }
            }

            &__name {
                overflow: hidden;
                text-overflow: ellipsis;

                &__box {
                    padding: 2px 5px;
                    border-radius: 5px;
                    cursor: pointer;
                }
            }

            @include mobile {
                width: 90%;
                height: auto;
                flex-wrap: wrap;
            }
        }
    }
</style>
