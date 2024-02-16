<script lang="ts">
    import Button from "@smui/button";
    import type { FetchedDriver } from "../driver";
    import DynamicConfigurator from "../../../components/Homescript/DynamicConfigurator.svelte";
    import { createEventDispatcher } from "svelte";

    export let driver: FetchedDriver = null

    const dispatch = createEventDispatcher()

    let textareaContent = ""
    let preventReacttoOutput = false
    let textarea: HTMLTextAreaElement = null

    let lastOutput = null

    function reactToInput() {
        textareaContent = textarea.value
        try {
            driver.configuration = JSON.parse(textareaContent)
            lastOutput = JSON.parse(textareaContent)

            if (!preventReacttoOutput) {
                preventReacttoOutput = true
                setTimeout(() => preventReacttoOutput = false, 10)
            }
        } catch (err) {
            console.error(`JSON parse error: `, err)
        }
    }

    function reactToOutput(data: any) {
        if (textarea.isEqualNode(document.activeElement) || preventReacttoOutput) {
            // TODO: is this even triggered?
            console.warn("Is active element, prevent cycle")
            return
        }
        textareaContent = `\n${JSON.stringify(data, null, 2)}`
        lastOutput = data
    }


    function commitConfig() {
        dispatch('save', lastOutput)
    }
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
                        bind:spec={driver.info.driver.info.config}
                        on:change={ (e) => reactToOutput(e.detail) }
                        bind:inputData={driver.configuration}
                        topLevelLabel={`Driver-wide configuration`}
                    />
                </div>

                <!-- <div class="driver__config__device"> -->
                <!--     <h6>Device Configuration: TODO: must be rendered per-device</h6> -->
                <!--     <DynamicConfigurator bind:spec={driver.info.device} topLevelLabel={"PER-DEVICE"} /> -->
                <!-- </div> -->
            </div>

            <h6>JSON configuration output</h6>
            <!-- <textarea bind:this={textarea} on:input={(_) => {}} rows="10" cols="40" value={textareaContent}></textarea> -->
            <textarea bind:this={textarea} on:input={(_) => reactToInput()} rows="20" cols="40" value={textareaContent}></textarea>

            <Button
                variant="outlined"
                on:click={commitConfig}>commit
            </Button>
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
