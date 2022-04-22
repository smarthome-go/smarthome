<script lang="ts">
    import FormField from '@smui/form-field/'
    import Switch from '@smui/switch'
    import Progress from '../../../components/Progress.svelte'
    import { sleep } from '../../../global'

    export let name: string
    export let description: string
    export let permission: string

    export let active // Shows if the user currently has the permission
    let loading = false

    export let grantFunc: (_: string) => {}
    export let removeFunc: (_: string) => {}

    // Handle switch updates
    async function toggle(event: CustomEvent<{ selected: boolean }>) {
        loading = true
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
        <Switch on:SMUISwitch:change={toggle} bind:checked={active} />
        <span slot="label">{active ? 'granted' : 'denied'}</span>
    </FormField>
</div>

<style lang="scss">
    @use '../../../mixins' as *;
    .permission {
        width: 100%;
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
            height: 2rem;
        }
    }
</style>
