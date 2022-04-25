<script lang="ts">
    import IconButton from '@smui/icon-button'
    import Switch from '@smui/switch'
    import { onMount } from 'svelte/internal'
    import Progress from '../../components/Progress.svelte'
    import { createSnackbar,hasPermission,sleep } from '../../global'

    export let id: string
    export let label: string
    export let checked: boolean

    let requests = 0
    let loading = false

    // Determines if edit button should be shown
    let hasEditPermission: boolean
    onMount(async () => {
        hasEditPermission = await hasPermission('modifyRooms')
    })

    $: loading = requests !== 0
    async function toggle(event: CustomEvent<{ selected: boolean }>) {
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
                `Failed to set switch '${label}' to ${
                    event.detail.selected ? 'on' : 'off'
                }`
            )
        }
        await sleep(500)
        requests--
    }
</script>

<div class="switch mdc-elevation--z3">
    <div>
        <Switch icons={false} bind:checked on:SMUISwitch:change={toggle} />
        <span>{label}</span>
    </div>
    <div class="right">
        {#if hasEditPermission}
            <IconButton class="material-icons" title="Edit Switch"
                >edit</IconButton
            >
        {/if}
        <div>
            <Progress type="circular" bind:loading />
        </div>
    </div>
</div>

<style lang="scss">
    .switch {
        background-color: var(--clr-height-1-3);
        border-radius: 0.3rem;
        min-width: 15rem;
        height: 3.3rem;
        padding: 0.5rem;
        display: flex;
        justify-content: space-between;
        align-items: center;

        & > * {
            display: flex;
            align-items: center;
        }
    }
    .right {
        div {
            margin-right: 14px;
        }
    }
</style>
