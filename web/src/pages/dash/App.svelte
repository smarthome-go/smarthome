<script lang="ts">
    import Page from '../../Page.svelte'
    import AutomationsSchedules from './AutomationsSchedules/AutomationsSchedules.svelte'
    import PowerUsage from './PowerUsage.svelte'
    import QuickActions from './QuickActions/QuickActions.svelte'
    import Reminders from './Reminders/Reminders.svelte'
    import Weather from './Weather.svelte'
    import { data } from '../../global'

    function hasPermission(permission: string): boolean {
        return (
            $data.userData.permissions.includes(permission) ||
            $data.userData.permissions.includes('*')
        )
    }
</script>

<Page>
    <div class="dash">
        <PowerUsage />
        {#if $data && hasPermission('homescript')}
            <QuickActions />
        {/if}
        <Weather />
        {#if ($data && hasPermission('automation')) || hasPermission('scheduler')}
            <AutomationsSchedules />
        {/if}
        {#if $data && hasPermission('reminder')}
            <Reminders />
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
