<script lang="ts">
    import IconButton from '@smui/icon-button'
    import Switch from '@smui/switch'
    import { createEventDispatcher,onMount } from 'svelte/internal'
    import Progress from '../../components/Progress.svelte'
    import { createSnackbar,hasPermission,sleep } from '../../global'
    import EditSwitch from './dialogs/switch/EditSwitch.svelte'

    // Event dispatcher
    const dispatch = createEventDispatcher()

    export let id: string
    export let name: string
    export let watts: number
    export let checked: boolean

    let requests = 0
    let loading = false

    // Is binded to the `editSwitch` in order to pass an event to a child
    let showEditSwitch: () => void

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
                await fetch('/api/power/set', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        switch: id,
                        powerOn: event.detail.selected,
                    }),
                })
            ).json()
            if (!res.success) throw Error()
        } catch {
            $createSnackbar(
                `Failed to set switch '${name}' to ${
                    event.detail.selected ? 'on' : 'off'
                }`
            )
        }
        await sleep(500)
        requests--
        dispatch('powerChangeDone', null)
    }
</script>

<EditSwitch
    on:delete={() => dispatch('delete', null)}
    {id}
    bind:name
    bind:watts
    bind:show={showEditSwitch}
/>

<div class="switch mdc-elevation--z3">
    <div class="switch__left">
        <Switch icons={false} bind:checked on:SMUISwitch:change={toggle} />
        <span class="switch__name">{name}</span>
    </div>
    <div class="switch__right">
        <div>
            <Progress type="circular" bind:loading />
        </div>
        {#if hasEditPermission}
            <IconButton
                class="material-icons"
                title="Edit Switch"
                on:click={showEditSwitch}>edit</IconButton
            >
        {/if}
    </div>
</div>

<style lang="scss">
    @use '../../mixins' as *;
    .switch {
        background-color: var(--clr-height-1-3);
        border-radius: 0.3rem;
        width: 15rem;
        height: 3.3rem;
        padding: 0.5rem;
        display: flex;
        align-items: center;
        justify-content: space-between;

        & > * {
            display: flex;
            align-items: center;
        }
        &__left {
            max-width: 70%;
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
        }
        @include mobile {
            width: 90%;
            height: auto;
            flex-wrap: wrap;
        }
    }
</style>
