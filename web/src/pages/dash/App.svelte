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
        <QuickActions />
        <Weather />
        <AutomationsSchedules />
        <Reminders />
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
