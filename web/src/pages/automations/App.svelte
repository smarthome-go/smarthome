<script lang="ts">
    import { onMount } from 'svelte'
    import Progress from '../../components/Progress.svelte'
    import { createSnackbar } from '../../global'
    import Page from '../../Page.svelte'
    import { automation,automations,loading } from './main'

    $loading = false

    // Fetches the current reminders from the server
    async function loadAutomations() {
        $loading = true
        try {
            const res = (await (
                await fetch('/api/automation/list/personal')
            ).json()) as automation[]
            automations.set(res)
        } catch (err) {
            $createSnackbar('Could not load automations')
        }
        $loading = false
    }

    onMount(() => loadAutomations()) // Load reminders as soon as the component is mounted
</script>

<Page>
    <Progress id="loader" bind:loading={$loading} />
</Page>

<style lang="scss">
    @use '../../mixins' as *;
</style>
