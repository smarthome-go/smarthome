<script lang="ts">
    import FormField from "@smui/form-field";
    import { Label } from "@smui/list";
    import SegmentedButton, { Segment } from "@smui/segmented-button";
    import Switch from "@smui/switch/src/Switch.svelte";
    import Textfield from "@smui/textfield";
    import CharacterCounter from "@smui/textfield/character-counter";
    import { onMount } from "svelte";
    import TimePicker from "../../../components/TimePicker.svelte";
    import { sleep } from "../../../global";
    import type { homescriptData } from "../main";
    import HmsSelector from "./HmsSelector.svelte";
    import IconPicker from "../../../components/IconPicker/IconPicker.svelte";
    import Button from "@smui/button";

    let iconPickerOpen = false;

    // Data which is dispatched as soon as the create button is pressed
    export let data: homescriptData;
</script>

<IconPicker bind:open={iconPickerOpen} bind:selected={data.mdIcon} />

<div class="container">
    <!-- Names and Text -->
    <div class="text">
        <span class="text-hint">Name and description of the Homescript</span>
        <Textfield
            bind:value={data.name}
            input$maxlength={30}
            label="Name"
            required
            style="width: 100%;"
            helperLine$style="width: 100%;"
        >
            <svelte:fragment slot="helper">
                <CharacterCounter>0 / 30</CharacterCounter>
            </svelte:fragment>
        </Textfield>
        <Textfield
            bind:value={data.description}
            label="Description"
            style="width: 100%;"
            helperLine$style="width: 100%;"
        />
    </div>
    <div class="toggles">
        <span class="text-hint">Attributes and visibility</span>
        <div class="toggles__item">
            <Switch bind:checked={data.schedulerEnabled} />
            <span class="text-hint">
                Automation selection {data.schedulerEnabled
                    ? "shown"
                    : "hidden"}
            </span>
        </div>
        <div class="right__toggles__item">
            <Switch bind:checked={data.quickActionsEnabled} />
            <span class="text-hint">
                Quick actions selection {data.quickActionsEnabled
                    ? "shown"
                    : "hidden"}
            </span>
        </div>
    </div>
    <Button
        on:click={() => {
            iconPickerOpen = true;
        }}
    >
        Pick Icon
    </Button>
</div>

<style lang="scss">
    @use "../../../mixins" as *;
    .container {
        display: flex;
        flex-wrap: wrap;
        gap: 2rem;

        @include not-widescreen {
            flex-direction: column;
        }
    }

    .toggles {
        background-color: var(--clr-height-1-2);
        width: 100%;
        padding: 1rem;
        border-radius: 0.3rem;
    }

    .text {
        width: 100%;
    }
</style>
