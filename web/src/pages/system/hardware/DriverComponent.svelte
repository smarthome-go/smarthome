<script lang="ts">
    import type { FetchedDriver } from "../driver";
    import DynamicConfigurator from "./DynamicConfigurator.svelte";

    export let driver: FetchedDriver = null
    $: if (driver) console.log('driver changed')

    let configuratorData: any = {}
</script>

<div class="driver mdc-elevation--z3">
    {#if driver !== null}
        <div class="driver__header">
            <h6>
                {driver.driver.vendorId}: {driver.driver.modelId}
            </h6>
        </div>
        {#if driver.validationErrors.length === 0}
        <div class="driver__config">
            <div class="driver__config">
                <DynamicConfigurator
                    bind:spec={driver.info.driver}
                    on:change={(e) => {configuratorData = e.detail; console.dir(e.detail)}}
                    topLevelLabel={`Driver-wide configuration`}
                />
            </div>

            <!-- <div class="driver__config__device"> -->
            <!--     <h6>Device Configuration: TODO: must be rendered per-device</h6> -->
            <!--     <DynamicConfigurator bind:spec={driver.info.device} topLevelLabel={"PER-DEVICE"} /> -->
            <!-- </div> -->
        </div>

        <h6>JSON configuration output</h6>
        <textarea rows="10" cols="40">
            { `\n${JSON.stringify(configuratorData, null, 2).replace('\t', '')}` }
        </textarea>
        {:else}
            <h6>Driver is broken: TODO</h6>
        {/if}
    {/if}
</div>

<style lang="scss">
    h6 {
        margin: 0;
    }

    .driver {
        background-color: var(--clr-height-1-3);
        margin-bottom: 1rem;
        padding: 1rem 2rem;
        border-radius: .3rem;
    }
</style>
