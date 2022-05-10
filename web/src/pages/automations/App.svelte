<script lang="ts">
    import { Icon,Label } from '@smui/button'
    import Button from '@smui/button/src/Button.svelte'
    import { onMount } from 'svelte'
    import Progress from '../../components/Progress.svelte'
    import { createSnackbar } from '../../global'
    import Page from '../../Page.svelte'
    import Automation from './Automation.svelte'
    import AddAutomation from './dialogs/AddAutomation.svelte'
    import { automations,homescripts,loading } from './main'

    let addOpen = false

    // Fetches the current automations from the server
    async function loadAutomations() {
        $loading = true
        try {
            const res = (await (
                await fetch('/api/automation/list/personal')
            ).json())
            if (res.success !== undefined && !res.success)
                throw Error(res.error)
            automations.set(res)
        } catch (err) {
            $createSnackbar(`Could not load automations: ${err}`)
        }
        $loading = false
    }

    // Fetches the available homescripts for the selection and naming
    async function loadHomescript() {
        $loading = true
        try {
            const res = await (await (fetch('/api/homescript/list/personal'))).json()
            if (res.success !== undefined && !res.success)
                throw Error(res.error)
            homescripts.set(res)
        }catch(err) {
            $createSnackbar(`Could not load homescript: ${err}`)
        }
        $loading = false
    }

    onMount(() => {
        loadAutomations()
        loadHomescript()
    }) // Load automations as soon as the component is mounted
</script>

<Page>
    <Progress id="loader" bind:loading={$loading} />
    <AddAutomation open={addOpen} on:add={() => console.log('add automation')} />

    <div class="automations">
        {#if $automations.length == 0}
            No automations
            <Button on:click={() => (addOpen = true)}>
                <Label>Create</Label>
                <Icon class="material-icons">add</Icon>
            </Button>
        {:else}
            {#each $automations as automation (automation.id)}
                <Automation bind:data={automation} />
            {/each}
        {/if}
    </div>
</Page>

<style lang="scss">
    @use '../../mixins' as *;
</style>
