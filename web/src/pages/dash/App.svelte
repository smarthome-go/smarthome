<script lang="ts">
    import Page from "../../Page.svelte";
    import AutomationsSchedules from "./AutomationsSchedules/AutomationsSchedules.svelte";
    import PowerUsage from "./PowerUsage.svelte";
    import QuickActions from "./QuickActions/QuickActions.svelte";
    import Reminders from "./Reminders.svelte";
    import Weather from "./Weather.svelte";
</script>

<Page>
    <div class="dash">
        <PowerUsage />
        <Weather />
        <QuickActions />
        <AutomationsSchedules />
        <Reminders />
    </div>
</Page>

<style lang="scss">
    .dash {
        padding: 1.5rem 2rem;
        display: flex;
        gap: 1.5rem;
        flex-wrap: wrap;
        align-content: flex-start;
    }
</style>
