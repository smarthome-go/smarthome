<script lang="ts">
    import Page from "../../Page.svelte";
    import PowerUsage from "./PowerUsage.svelte";
    import Weather from "./Weather.svelte";
</script>

<Page>
    <div class="dash">
        <PowerUsage />
        <Weather />
        <Weather />
    </div>
</Page>

<style lang="scss">
    .dash {
        padding: 1.5rem 2rem;
        display: flex;
        gap: 1.5rem;
        justify-content: center;
        flex-wrap: wrap;
        align-content: flex-start;

        height: 100vh;
        overflow-y: auto;
        box-sizing: border-box;
    }
</style>
