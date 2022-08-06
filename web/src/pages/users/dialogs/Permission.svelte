<script lang="ts">
    import FormField from '@smui/form-field/'
    import Switch from '@smui/switch'
    import Progress from '../../../components/Progress.svelte'
    import { sleep } from '../../../global'

    export let name: string
    export let description: string
    export let permission: string

    export let active = false // Shows if the user currently has the permission
    let loading = false

    export let grantFunc: (_: string) => Promise<void>
    export let removeFunc: (_: string) => Promise<void>

    let loadingStart: number = Date.now()
    let loadingTime: number

    // Handle switch updates
    async function toggle(event: CustomEvent<{ selected: boolean }>) {
        loadingStart = Date.now()
        loading = true
        updateLoadingTime()
        try {
            if (event.detail.selected) {
                await grantFunc(permission)
            } else await removeFunc(permission)
        } catch (err) {
            await sleep(1000)
            active = !active
        }
        loading = false
    }

    // Calculate the time spent waiting for the serve'rs response
    // Is used in the switch to disabled it when the time is greater than 50ms
    // Prevents ugly flickering and stops the user from clicking multiple times causing interference
    async function updateLoadingTime() {
        if (loading) {
            loadingTime = (loadingStart - Date.now()) * -1
            await sleep(5)
            updateLoadingTime()
        }
    }
</script>

<div class="permission mdc-elevation--z3">
    <div class="top">
        <h6>{name}</h6>
        <div>
            <Progress type="circular" bind:loading />
            <pre>{permission}</pre>
        </div>
    </div>
    <span class="description">{description}</span>

    <FormField>
        <Switch
            on:SMUISwitch:change={toggle}
            checked={active}
            disabled={loading && loadingTime > 50}
        />
        <span slot="label">Permission {active ? 'granted' : 'denied'}</span>
    </FormField>
</div>

<style lang="scss">
    @use '../../../mixins' as *;
    .permission {
        width: 100%;
        height: 10rem;
        display: flex;
        padding: 1rem;
        border-radius: 0.3rem;
        display: flex;
        flex-direction: column;
        justify-content: start;
        background-color: var(--clr-height-1-3);

        span {
            color: var(--clr-text);
        }

        @include widescreen {
            width: 29%;
        }

        @include mobile {
            height: auto;
        }
    }
    .top {
        display: flex;
        justify-content: space-between;
        align-items: center;

        h6 {
            margin: 0;
            margin-bottom: 0.5rem;
            color: var(--clr-text);
        }

        div {
            display: flex;
            align-items: center;
            gap: 1rem;

            pre {
                font-size: 0.8rem;
                background-color: var(--clr-height-3-6);
                padding: 0.1rem 0.2rem;
                border-radius: 0.1rem;
            }
        }
    }
    .description {
        height: 3rem;
        overflow-y: auto;

        @include mobile {
            height: 100%;
            overflow-y: hidden;
            color: var(--clr-text-hint) !important;
        }
    }
</style>
