<script lang="ts">
    import type { FetchedDriver } from "../driver";
    import DynamicConfigurator from "./DynamicConfigurator.svelte";

    export let driver: FetchedDriver = null
    $: if (driver) console.log('driver changed')

    // <!--     jsonValue = `\n${JSON.stringify(configuratorOutputData, null, 2).replace('\t', '')}` -->

    let textareaContent = ""

    let preventReacttoOutput = false

    function reactToInput() {
        textareaContent = textarea.value
        console.log(`TEXTAREA CONTENT: ${textareaContent}`)

        try {
            let parsedTemp = JSON.parse(textareaContent)
            inputData = parsedTemp

            if (!preventReacttoOutput) {
                preventReacttoOutput = true
                setTimeout(() => preventReacttoOutput = false, 10)
            }
        } catch (err) {
            console.error(`JSON parse error: `, err)
        }
    }

    let textarea: HTMLTextAreaElement = null

    function reactToOutput(data: any) {
        if (textarea.isEqualNode(document.activeElement) || preventReacttoOutput) {
            console.warn("is active element, prevent cycle")
            return
        }
        textareaContent = `\n${JSON.stringify(data, null, 2)}`
    }

    let inputData = null
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
                    on:change={ (e) => reactToOutput(e.detail) }
                    bind:inputData
                    topLevelLabel={`Driver-wide configuration`}
                />
            </div>

            <!-- <div class="driver__config__device"> -->
            <!--     <h6>Device Configuration: TODO: must be rendered per-device</h6> -->
            <!--     <DynamicConfigurator bind:spec={driver.info.device} topLevelLabel={"PER-DEVICE"} /> -->
            <!-- </div> -->
        </div>

        <h6>JSON configuration output</h6>
            <textarea bind:this={textarea} on:input={(_) => reactToInput()} rows="10" cols="40" value={textareaContent}></textarea>
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
