<script lang="ts">
    import Textfield from "@smui/textfield";
    import CharacterCounter from "@smui/textfield/character-counter";
    import Select, { Option } from "@smui/select";
    import type { homescriptArgData } from "../../homescript";
    import { displayOpts } from "../../homescript";
    import { inputTypeOpts } from "../../homescript";

    // Is bound externally in order to allow editing
    export let data: homescriptArgData;
</script>

<div>
    <span class="text-hint"
        >The key and the prompt should be meaningful</span
    >
    <div class="selectors">
        <Select bind:value={data.inputType} label="Input data-type">
            {#each inputTypeOpts as type}
                <Option value={type}>{type}</Option>
            {/each}
        </Select>
        <Select bind:value={data.display} label="Display style">
            {#each displayOpts.filter((o) => o.type === data.inputType) as opt}
                <Option value={opt.identifier}>{opt.label}</Option>
            {/each}
        </Select>
    </div>
    <div class="prompt">
        <Textfield
            bind:value={data.prompt}
            label="Prompt"
            required
            style="width: 100%;"
            helperLine$style="width: 100%;"
        />
    </div>
    <div class="key">
        <Textfield
            bind:value={data.argKey}
            input$maxlength={100}
            label="Key"
            required
            style="width: 100%;"
            helperLine$style="width: 100%;"
        >
            <svelte:fragment slot="helper">
                <CharacterCounter>0 / 100</CharacterCounter>
            </svelte:fragment>
        </Textfield>
    </div>
</div>

<style lang="scss">
    .selectors {
        display: flex;
        flex-wrap: wrap;
        gap: 1rem;
    }
    .prompt,
    .key {
        margin-bottom: 1rem;
    }
    .prompt {
        margin-top: 1rem;
    }
</style>
