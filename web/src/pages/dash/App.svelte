<script lang="ts">
    import Page from "../../Page.svelte";
    import PowerUsage from "./PowerUsage.svelte";
</script>

<Page>
    <div class="dash">
        <PowerUsage />
    </div>
</Page>

<style lang="scss">
    .dash {
        padding: 1.5rem 2rem;
    }
</style>
