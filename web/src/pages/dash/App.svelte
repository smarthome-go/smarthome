<script lang="ts">
    import Page from '../../Page.svelte'
    import AutomationsSchedules from './AutomationsSchedules/AutomationsSchedules.svelte'
    import PowerUsage from './PowerUsage.svelte'
    import QuickActions from './QuickActions/QuickActions.svelte'
    import Reminders from './Reminders/Reminders.svelte'
    import Weather from './Weather.svelte'
    import { data, hasPermissionSync } from '../../global'
    import Box from './Box.svelte'
    import Progress from '../../components/Progress.svelte'
    import { runHomescriptById } from '../../homescript'

    let homescripts = undefined
</script>

<Page>
    <div class="dash">
        <PowerUsage />
        <!-- HACK: A note on performance: this can possibly be considered bad practice as this abuses svelte as a spin-lock -->
        <!-- A better alternative would be to use svelte await -->
        {#if $data && hasPermissionSync('homescript')}
            <QuickActions bind:homescripts />
        {/if}
        <Weather />
        {#if ($data && hasPermissionSync('automation')) || hasPermissionSync('scheduler')}
            <AutomationsSchedules />
        {/if}
        {#if $data && hasPermissionSync('reminder')}
            <Reminders />
        {/if}
        {#if $data && hasPermissionSync('homescript') && homescripts !== undefined}
            {#each homescripts.filter(h => h.data.data.isWidget) as hms (hms.data.data.id)}
                <Box loading={false}>
                    <span slot="header" class="title">{hms.data.data.name}</span>
                    <i slot="header-right" class="material-icons text-disabled"
                        >{hms.data.data.mdIcon}</i
                    >
                    <div class="actions" slot="content">
                        {#await runHomescriptById(hms.data.data.id, [], true)}
                            <Progress type="circular" loading={true} />
                        {:then res}
                            {#if res.success}
                                {@html res.output}
                            {:else}
                                <span style:color="var(--clr-error)">Widget Crashed</span>
                                <br />
                                <code style="font-size: .9rem">
                                    {#if res.errors[0].syntaxError !== null}
                                        SyntaxError: {res.errors[0].syntaxError.message}
                                    {:else if res.errors[0].diagnosticError !== null}
                                        SemanticError: {res.errors[0].diagnosticError.message}
                                    {:else}
                                        {res.errors[0].runtimeError.kind}: {res.errors[0]
                                            .runtimeError.message}
                                    {/if}
                                </code>
                            {/if}
                        {/await}
                    </div></Box
                >
            {/each}
        {/if}
        <div class="placeholder" />
        <div class="placeholder" />
    </div>
</Page>

<style lang="scss">
    .placeholder {
        flex-grow: 1;
        width: 30rem;
    }
    .dash {
        padding: 1.5rem 2rem;
        display: flex;
        flex-wrap: wrap;
        gap: 1.5rem;
    }
</style>
