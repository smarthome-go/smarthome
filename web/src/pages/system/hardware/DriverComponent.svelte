<script lang="ts">
    import type { FetchedDriver } from "../driver";
    import DynamicConfigurator from "./DynamicConfigurator.svelte";

    export let driver: FetchedDriver = null
    $: if (driver) console.log('driver changed')
</script>

<div class="driver">
    {#if driver !== null}
        <div class="driver__header">
            {driver.driver.vendorId}: {driver.driver.modelId}
        </div>
        <div class="driver__config">
            <div class="driver__config__driver">
                <h6>Driver Configuration</h6>
                <DynamicConfigurator bind:spec={driver.info.driver} />
            </div>
            <div class="driver__config__device">
                <h6>Device Configuration: TODO: must be rendered per-device</h6>
                <DynamicConfigurator bind:spec={driver.info.device} />
            </div>
        </div>
    {/if}
</div>

<style lang="scss">

</style>
